import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('the FRITZ!Box has separate internet and calls widgets (#168)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await page.getByRole('button', { name: 'Home', exact: true }).click();
  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('fritz');
  await expect(page.getByRole('button', { name: /^FRITZ!Box internet/ })).toBeVisible();
  await page.getByRole('button', { name: /^FRITZ!Box calls/ }).click();
  const settings = page.getByRole('dialog', { name: 'FRITZ!Box calls settings' });
  await settings.getByRole('combobox', { name: 'Calls' }).selectOption('missed');
  await settings.getByRole('button', { name: 'Save', exact: true }).click();
  await account(page, 'Done');
  // No box configured in the smoke test: the widget offers to connect it.
  await expect(page.getByText('How your FRITZ!Box is doing.')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Connect', exact: true }).first()).toBeVisible();
});
