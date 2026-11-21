import { expect, test } from '@playwright/test';
import { openSettings, password } from '../support';

// The guided restore up to the last click (#677): submitting would restart
// the shared test server, so the restore itself is covered by the Go tests.
test('guided restore: pick, check, confirm', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await openSettings(page, 'Data & backup', 'Backups');
  const section = page
    .getByRole('dialog', { name: 'Settings' })
    .getByRole('region', { name: 'Backups', exact: true });

  // The traffic light sits on top; a fresh backup keeps it green.
  await section.getByRole('button', { name: 'Back up now' }).click();
  await expect(section.getByRole('status').first()).toContainText('all good');

  const pick = section.getByRole('group', { name: '1. Choose a backup' });
  await pick
    .getByRole('radio', { name: /^Local:/ })
    .first()
    .check();
  const check = section.getByRole('group', { name: '2. Check it' });
  await expect(check).toContainText('The backup is intact and fits this version of Waylight.');
  await expect(check).toContainText(/Contents: tabs \d+, widgets \d+/);

  const confirm = section.getByRole('group', { name: '3. Confirm' });
  await expect(confirm).toContainText('Before that, Waylight saves the current state as a new backup.');
  const restore = confirm.getByRole('button', { name: 'Restore and restart' });
  await expect(restore).toBeDisabled();
  await confirm.getByLabel('Your password').fill(password);
  await expect(restore).toBeDisabled();
  await confirm
    .getByRole('checkbox', { name: 'I understand that the current data will be replaced' })
    .check();
  await expect(restore).toBeEnabled();

  // An upload is checked only on confirmation and needs a file first.
  await pick.getByRole('radio', { name: 'Upload a file' }).check();
  await expect(section.getByRole('group', { name: '2. Check it' })).toContainText(
    'Waylight checks this backup when you confirm.',
  );
  await expect(section.getByRole('group', { name: '3. Confirm' })).toHaveCount(0);
});
