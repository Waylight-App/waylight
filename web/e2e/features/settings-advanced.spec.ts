import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('advanced settings are hidden until the switch is on, and it is kept (#602)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await account(page, 'Settings');
  const dialog = page.getByRole('dialog', { name: 'Settings' });
  const nav = dialog.getByRole('navigation', { name: 'Settings areas' });
  const toggle = dialog.getByRole('checkbox', { name: 'Show advanced settings' });
  await expect(toggle).not.toBeChecked();

  await toggle.check();
  await expect(toggle).toBeChecked();
  await nav.getByRole('button', { name: 'Connections & services', exact: true }).click();
  await expect(nav.getByRole('button', { name: 'API tokens', exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Close' }).click();

  // Stored for the account: still on after a reload.
  await page.reload();
  await account(page, 'Settings');
  await expect(toggle).toBeChecked();
  // Leave the shared account as it was.
  await toggle.uncheck();
  await expect(toggle).not.toBeChecked();
  await nav.getByRole('button', { name: 'Connections & services', exact: true }).click();
  await expect(nav.getByRole('button', { name: 'API tokens', exact: true })).toHaveCount(0);
});
