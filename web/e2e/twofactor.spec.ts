import { expect, test } from '@playwright/test';
import { account, password, totp } from './support';

test('two-factor authentication can be enabled and is required at login', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Settings');
  await page.getByRole('button', { name: 'Account', exact: true }).click();
  await page.getByRole('button', { name: 'Set up two-factor authentication' }).click();
  await expect(page.getByAltText('QR code for the authenticator app')).toBeVisible();
  const secret = (await page.getByRole('dialog').locator('code').textContent()) ?? '';
  await page.getByLabel('Code from the app').fill(totp(secret));
  await page.getByRole('button', { name: 'Turn on' }).click();
  await expect(page.getByText('Two-factor authentication is on.')).toBeVisible();
  await page.getByRole('button', { name: 'I saved them' }).click();
  await page.getByRole('button', { name: 'Close' }).click();

  await account(page, 'Log out');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByLabel('Authentication code (or recovery code)')).toBeVisible();
  // Wait for the next time step: the code used for enabling cannot be replayed.
  const step = Math.floor(Date.now() / 30000);
  await expect.poll(() => Math.floor(Date.now() / 30000), { timeout: 35_000 }).toBeGreaterThan(step);
  await page.getByLabel('Authentication code (or recovery code)').fill(totp(secret));
  await page.getByRole('button', { name: 'Verify' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
});
