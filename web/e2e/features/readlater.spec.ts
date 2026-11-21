import { expect, test } from '@playwright/test';
import { account, password, openSettings } from '../support';

test('read later keeps links, refuses internal addresses and checks off articles (#99)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Lesen');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('pocket');
  await page.getByRole('button', { name: /^Read later/ }).click();
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await account(page, 'Done');

  // The server itself is on loopback: the guarded client refuses to fetch it,
  // the link is kept without a reader copy.
  const url = new URL('/healthz', page.url()).toString();
  const input = page.getByRole('textbox', { name: 'Address of the page to read later' });
  await input.fill(url);
  await input.press('Enter');
  const host = new URL(url).host;
  const item = page.locator('li', { hasText: host });
  await expect(item).toBeVisible();
  await item.getByRole('button').click();
  const reader = page.getByRole('dialog');
  await expect(reader).toContainText('could not be loaded');
  await expect(reader.getByRole('link', { name: 'Original' })).toHaveAttribute('href', url);
  await reader.getByRole('button', { name: 'Archive' }).click();
  await expect(page.getByText('Nothing to read.')).toBeVisible();

  // Settings: the archived article can be deleted.
  await openSettings(page, 'Dashboard & widgets', 'Read later');
  await page.locator('.tabs label', { hasText: 'Archived' }).click();
  page.once('dialog', (d) => d.accept());
  await page.getByRole('button', { name: new RegExp(`^Delete “${host}”`) }).click();
  await expect(page.getByRole('dialog').getByText('Nothing to read.')).toBeVisible();

  // Part 2: a Pocket export is imported at once, the copies follow later;
  // archived articles can be cleaned up automatically.
  const dialog = page.getByRole('dialog');
  await dialog.getByLabel('Delete archived articles').selectOption('30');
  const csv =
    'title,url,time_added,tags,status\nPocket eins,https://pocket.example/1,1700000000,,unread\nPocket zwei,https://pocket.example/2,1700000100,,archive\n';
  await dialog
    .getByLabel('Import Pocket export…')
    .setInputFiles({ name: 'part_000000.csv', mimeType: 'text/csv', buffer: Buffer.from(csv) });
  await expect(dialog.getByText('2 added, 0 already saved, 0 not importable.')).toBeVisible();
  await dialog.locator('.tabs label', { hasText: 'Unread' }).click();
  await expect(dialog.getByText('Pocket eins')).toBeVisible();
  await expect(dialog.getByText('Reader copy follows in the background')).toBeVisible();
  const settings = await (await page.request.get('/api/readlater/settings')).json();
  expect(settings.archiveDays).toBe(30);
});
