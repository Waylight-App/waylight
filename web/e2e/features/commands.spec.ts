import { expect, test } from '@playwright/test';
import { password } from '../support';

test('the search runs app commands: theme, settings deep link, widget picker (#476)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  const box = page.getByRole('combobox', { name: 'Quick search (Ctrl+K or /)' });
  const html = page.locator('html');

  // Theme: harmless and undoable, so it runs directly.
  await page.keyboard.press('Control+k');
  await box.fill('dark mode');
  await expect(page.getByRole('option', { name: 'Switch to dark mode' })).toBeVisible();
  await box.press('Enter');
  await expect(html).toHaveAttribute('data-theme', 'dark');
  await page.keyboard.press('Control+k');
  await box.fill('hell');
  await page.getByRole('option', { name: 'Switch to light mode' }).click();
  await expect(html).toHaveAttribute('data-theme', 'light');

  // Settings deep link: opens the dialog on that section.
  await page.keyboard.press('Control+k');
  await box.fill('settings account');
  await page.getByRole('option', { name: 'Settings: Account' }).click();
  const settings = page.getByRole('dialog', { name: 'Settings' });
  await expect(settings.getByRole('region', { name: 'Account' })).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(settings).toHaveCount(0);

  // ">" limits the rows to commands; the words after "add widget" filter the picker.
  await page.keyboard.press('Control+k');
  await box.fill('>');
  await expect(page.getByRole('group', { name: 'Commands' })).toBeVisible();
  await expect(page.getByRole('group', { name: 'Web' })).toHaveCount(0);
  await box.fill('add widget weather');
  await page.getByRole('option', { name: 'Add a widget' }).click();
  await expect(page.getByRole('searchbox', { name: 'Search widgets…' })).toHaveValue('weather');
});
