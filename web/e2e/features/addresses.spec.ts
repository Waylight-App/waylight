import { expect, test } from '@playwright/test';
import { password } from '../support';

test('addresses open directly and "!" suggests bangs (#475)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await page.keyboard.press('Control+k');
  const box = page.getByRole('combobox', { name: 'Quick search (Ctrl+K or /)' });
  const options = page.getByRole('option');

  // An address is the first row, so Enter opens it.
  await box.fill('github.com');
  await expect(options.first()).toHaveText(/Open github\.com/);
  await box.fill('192.168.178.1:8080');
  await expect(options.first()).toHaveText(/Open 192\.168\.178\.1:8080/);
  // Version numbers and file names stay a search.
  await box.fill('v1.2.3');
  await expect(page.getByRole('option', { name: /^Open / })).toHaveCount(0);

  // "!" lists bangs; choosing one completes it and keeps the field focused.
  await box.fill('!g');
  const bangs = page.getByRole('group', { name: 'Bangs' });
  await expect(bangs.getByRole('option', { name: /^!gh/ })).toBeVisible();
  await bangs.getByRole('option', { name: /^!gh/ }).click();
  await expect(box).toHaveValue('!gh ');
  await expect(box).toBeFocused();
  await expect(bangs).toHaveCount(0);
});
