import { describe, expect, it } from 'vitest';
import { AuthUseCases } from '../../usecase/authUseCases';
import type { AuthGateway } from '../../usecase/ports';
import { LocalAuthGateway, LocalModeUnavailableError, noCsrf } from './localAuthGateway';

describe('LocalAuthGateway', () => {
  it('signs the browser profile in as the owner without a server', async () => {
    const auth = new AuthUseCases(new LocalAuthGateway(), noCsrf);
    expect(await auth.currentView()).toBe('app');
    expect(auth.admin).toBe(true);
    expect(auth.guest).toBe(false);
    expect(auth.member).toBe(false);
    expect(auth.demo).toBe(false);
  });

  it('rejects account features that need a server', async () => {
    // Through the port, as the use cases call it.
    const gateway: AuthGateway = new LocalAuthGateway();
    await expect(gateway.login('secret')).rejects.toBeInstanceOf(LocalModeUnavailableError);
    await expect(gateway.members()).rejects.toBeInstanceOf(LocalModeUnavailableError);
    await expect(gateway.createApiToken('t', [], 1)).rejects.toBeInstanceOf(LocalModeUnavailableError);
  });

  it('signs out without an error', async () => {
    await expect(new LocalAuthGateway().logout()).resolves.toBeUndefined();
  });
});
