import { describe, expect, it, vi } from 'vitest';
import { capabilitiesOf, LOCAL_CAPABILITIES, parseTodosConfig, type TodoSource } from './todo';
import { TodoSourcesUseCases, type TodoSourcesGateway } from '../usecase/todosUseCases';
import { TodosApi } from '../adapter/api/todosApi';
import type { HttpClient } from '../../../shared/api-client/http';

const nc: TodoSource = {
  id: 'nc',
  name: 'Nextcloud',
  kind: 'caldav',
  url: 'https://cloud.example/tasks/',
  username: 'anna',
  listId: '',
  capabilities: { ordering: false, repeat: false, reminders: false },
};

describe('connected to-do lists (#169)', () => {
  it('reads the list of a widget', () => {
    expect(parseTodosConfig({})).toEqual({ showDone: false, source: '' });
    expect(parseTodosConfig({ source: 'nc', showDone: true })).toEqual({ showDone: true, source: 'nc' });
    expect(parseTodosConfig({ source: 'local' }).source).toBe('');
    expect(parseTodosConfig({ source: '../x' }).source).toBe('');
  });

  it('knows what a list supports', () => {
    expect(capabilitiesOf('', [])).toEqual(LOCAL_CAPABILITIES);
    expect(capabilitiesOf('nc', [nc]).repeat).toBe(false);
    expect(capabilitiesOf('gone', [nc])).toEqual({ ordering: false, repeat: false, reminders: false });
  });

  it('sends the list with every request', async () => {
    const http = {
      get: vi.fn().mockResolvedValue([]),
      post: vi.fn().mockResolvedValue({ removed: 0 }),
      put: vi.fn(),
      delete: vi.fn().mockResolvedValue(undefined),
    };
    const api = new TodosApi(http as unknown as HttpClient, 'nc');
    await api.list(true);
    await api.clearCompleted();
    await api.delete('a b.ics');
    expect(http.get).toHaveBeenCalledWith('/api/todos?done=true&source=nc');
    expect(http.post).toHaveBeenCalledWith('/api/todos/clear-completed?source=nc');
    expect(http.delete).toHaveBeenCalledWith('/api/todos/a%20b.ics?source=nc');
    await new TodosApi(http as unknown as HttpClient).list(false);
    expect(http.get).toHaveBeenLastCalledWith('/api/todos');
  });

  it('checks the form, caches the lists and hides them from guests', async () => {
    const g = {
      list: vi.fn().mockResolvedValue([nc]),
      save: vi.fn().mockResolvedValue(nc),
      delete: vi.fn().mockResolvedValue(undefined),
      discover: vi.fn().mockResolvedValue([]),
      oauthRedirect: vi.fn().mockResolvedValue('https://home.example/api/todos/settings/oauth/callback'),
      startOAuth: vi.fn().mockResolvedValue({ url: 'https://login.example', state: 's1' }),
      oauthStatus: vi.fn().mockResolvedValueOnce('pending').mockResolvedValueOnce('ready'),
    } satisfies TodoSourcesGateway;
    const uc = new TodoSourcesUseCases(g);
    const input = {
      kind: 'caldav' as const,
      name: ' NC ',
      url: ' https://cloud.example/dav ',
      username: 'anna',
      password: 'pw',
    };
    await uc.list();
    await uc.list();
    expect(g.list).toHaveBeenCalledTimes(1);
    await expect(uc.save('', { ...input, url: 'cloud.example' })).rejects.toThrow();
    await expect(uc.save('', { ...input, password: '' })).rejects.toThrow();
    await expect(uc.save('', { ...input, name: ' ' })).rejects.toThrow();
    await uc.save('', input);
    expect(g.save).toHaveBeenCalledWith('', { ...input, name: 'NC', url: 'https://cloud.example/dav' });
    await uc.list();
    expect(g.list).toHaveBeenCalledTimes(2); // saving drops the cache
    // Editing or discovering with a stored list needs no password.
    await uc.discover({ ...input, password: '' }, 'nc');
    expect(g.discover).toHaveBeenCalledWith({
      ...input,
      password: '',
      url: 'https://cloud.example/dav',
      id: 'nc',
    });
    // Todoist: token and project only.
    const td = { kind: 'todoist' as const, name: 'Inbox', url: '', username: '', password: 'tok' };
    await expect(uc.save('', td)).rejects.toThrow(); // no project chosen yet
    await expect(uc.discover({ ...td, password: '' })).rejects.toThrow();
    await uc.discover(td);
    await uc.save('', { ...td, listId: 'p1' });
    expect(g.save).toHaveBeenLastCalledWith('', { ...td, listId: 'p1' });
    // OAuth (#171): a client ID is needed, then the sign-in is awaited.
    const ms = { kind: 'microsoft' as const, name: 'To Do', url: '', username: '', password: '' };
    await expect(uc.startOAuth(ms)).rejects.toThrow();
    expect(await uc.startOAuth({ ...ms, username: ' app ' })).toEqual({
      url: 'https://login.example',
      state: 's1',
    });
    expect(g.startOAuth).toHaveBeenCalledWith({ ...ms, username: 'app' });
    const slept: number[] = [];
    expect(
      await uc.waitForSignIn(
        's1',
        () => false,
        async (ms) => void slept.push(ms),
      ),
    ).toBe('ready');
    expect(slept).toEqual([2000]);
    expect(await uc.waitForSignIn('s1', () => true)).toBe('pending'); // stopped: no request
    await expect(uc.save('', { ...ms, username: 'app', oauthState: 's1' })).rejects.toThrow(); // no list chosen
    await uc.save('', { ...ms, username: 'app', oauthState: 's1', listId: 'L1' });
    g.oauthRedirect.mockRejectedValue(new Error('403'));
    expect(await uc.oauthRedirect()).toBe('');
    g.list.mockRejectedValue(new Error('403'));
    expect(await uc.list(true)).toEqual([]);
  });
});
