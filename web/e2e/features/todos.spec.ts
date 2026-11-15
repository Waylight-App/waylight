import { expect, test } from '@playwright/test';
import { password } from '../support';

test('to-dos can be added and edited offline and are synced afterwards (#80)', async ({ page, context }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await page.getByRole('button', { name: 'Home', exact: true }).click();
  const input = page.getByRole('textbox', { name: 'New to-do' }).first();
  await expect(input).toBeVisible();

  await context.setOffline(true);
  await input.fill('Offline gekauft');
  await input.press('Enter');
  // The new to-do shows up at once and the change waits.
  await expect(page.getByText('Offline gekauft')).toBeVisible();
  await expect(page.getByRole('status').filter({ hasText: 'not synced yet' })).toContainText('1 change');
  let onServer = await (await page.request.get('/api/todos')).text();
  expect(onServer).not.toContain('Offline gekauft');

  await context.setOffline(false);
  await expect(page.getByRole('status').filter({ hasText: 'not synced yet' })).toHaveCount(0, {
    timeout: 10_000,
  });
  onServer = await (await page.request.get('/api/todos')).text();
  expect(onServer).toContain('Offline gekauft');

  // A change made elsewhere while offline becomes a conflict the user decides.
  const todo = (JSON.parse(onServer) as { id: string; title: string; version: number }[]).find(
    (t) => t.title === 'Offline gekauft',
  );
  expect(todo?.version).toBe(1);
  await context.setOffline(true);
  await page.locator('li', { hasText: 'Offline gekauft' }).getByRole('checkbox').check();
  const status = await (await page.request.get('/api/auth/status')).json();
  const res = await page.request.put(`/api/todos/${todo?.id}`, {
    headers: { 'X-CSRF-Token': status.csrfToken as string },
    data: { title: 'Offline gekauft (geändert)', due: '', done: false },
  });
  expect(res.ok()).toBeTruthy();
  await context.setOffline(false);
  await page.getByRole('button', { name: 'Resolve' }).click();
  const dialog = page.getByRole('dialog', { name: 'Changes that could not be synced' });
  await expect(dialog).toContainText('Changed elsewhere while you were offline.');
  await dialog.getByRole('button', { name: 'Keep mine' }).click();
  await expect(page.getByRole('button', { name: 'Resolve' })).toHaveCount(0, { timeout: 10_000 });
  const after = (await (await page.request.get('/api/todos?done=true')).json()) as {
    id: string;
    done: boolean;
    version: number;
  }[];
  expect(after.find((t) => t.id === todo?.id)).toMatchObject({ done: true, version: 3 });
  await page.request.delete(`/api/todos/${todo?.id}`, {
    headers: { 'X-CSRF-Token': status.csrfToken as string },
  });
});
