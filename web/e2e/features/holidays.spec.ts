import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('the holidays widget lists the public holidays of the chosen state (#55)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Days off');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('ferien');
  await page.getByRole('button', { name: /^Holidays/ }).click();
  // The widget has settings; they open right after adding.
  await expect(page.getByRole('combobox', { name: 'State' })).toHaveValue('NI');
  await page.getByRole('combobox', { name: 'State' }).selectOption('BY');
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await account(page, 'Done');

  await expect(page.getByText('Bavaria', { exact: true })).toBeVisible();
  // Public holidays are computed locally: there is always a next one, named and with a countdown.
  const first = page
    .locator('li')
    .filter({ hasText: /in \d+ days|today|tomorrow/ })
    .first();
  await expect(first).toBeVisible();
  await expect(first).toContainText(/Day|Christmas|Friday|Monday|Epiphany|Corpus Christi|Saints|Year/);
});
