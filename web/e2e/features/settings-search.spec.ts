import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('the settings search finds sections and fields in both languages (#601)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await account(page, 'Settings');
  const dialog = page.getByRole('dialog', { name: 'Settings' });

  // "/" focuses the field from anywhere in the dialog.
  await dialog
    .getByRole('navigation', { name: 'Settings areas' })
    .getByRole('button', { name: 'My account' })
    .click();
  await page.keyboard.press('/');
  const field = dialog.getByRole('combobox', { name: 'Search the settings' });
  await expect(field).toBeFocused();

  // A German word in the English interface finds the field; Enter opens the first hit.
  await page.keyboard.type('passwort ändern');
  await expect(dialog.getByRole('option').first()).toContainText('Change password');
  await expect(dialog.getByText('1 result')).toBeAttached();
  await page.keyboard.press('Enter');
  await expect(dialog.getByRole('region', { name: 'Account', exact: true })).toBeVisible();
  await expect(dialog.locator('.settings-found')).toHaveText('Change password');

  // A section by a synonym, chosen with the arrow keys.
  await page.keyboard.press('/');
  await page.keyboard.type('backup');
  await expect(dialog.getByRole('option').first()).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(field).toHaveValue('');
  await page.keyboard.type('nothing like this');
  await expect(dialog.getByText('No setting matches.').first()).toBeVisible();
});
