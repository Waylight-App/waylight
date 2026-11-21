import { expect, test } from '@playwright/test';
import { expectAccessible } from '../a11y';
import { openSettings, password } from '../support';

// A failed run must not leave the simple view on for the specs that follow.
test.afterEach(async ({ page }) => {
  const status = await (await page.request.get('/api/auth/status')).json();
  const current = await (await page.request.get('/api/appearance')).json();
  await page.request.put('/api/appearance', {
    headers: { 'X-CSRF-Token': status.csrfToken as string },
    data: { ...current, textSize: 'normal', highContrast: false, simpleView: false },
  });
});

test('text size, high contrast and the simple view apply per account (#681)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  const html = page.locator('html');
  const rootFontSize = () =>
    page.evaluate(() => parseFloat(getComputedStyle(document.documentElement).fontSize));
  const before = await rootFontSize();

  await openSettings(page, 'Appearance & language', 'Appearance');
  const options = page.getByRole('group', { name: 'Accessibility' });
  await options.getByLabel('Text size').selectOption('larger');
  await options.getByLabel('High contrast').check();
  await options.getByLabel('Simple view').check();
  await page.getByRole('button', { name: 'Save appearance' }).click();

  await expect(html).toHaveAttribute('data-text-size', 'larger');
  await expect(html).toHaveAttribute('data-contrast', 'high');
  await expect(html).toHaveAttribute('data-simple', 'true');
  expect(await rootFontSize()).toBeCloseTo(before * 1.3, 0);
  const saved = await (await page.request.get('/api/appearance')).json();
  expect(saved).toMatchObject({ textSize: 'larger', highContrast: true, simpleView: true });

  await page.keyboard.press('Escape');
  await expect(page.getByTestId('simple-view')).toBeVisible();
  // No edit mode in the simple view: neither the menu entry nor the shortcut.
  await page.keyboard.press('e');
  await expect(page.getByText(/^Edit mode:/)).toHaveCount(0);
  await page.getByRole('button', { name: 'Account menu', exact: true }).click();
  await expect(page.getByRole('menuitem', { name: 'Settings', exact: true })).toBeVisible();
  await expect(page.getByRole('menuitem', { name: 'Edit', exact: true })).toHaveCount(0);
  await page.keyboard.press('Escape');
  await expectAccessible(page, 'dashboard in the simple view with high contrast');

  // Back to the defaults, so the following specs see the usual dashboard.
  await openSettings(page, 'Appearance & language', 'Appearance');
  await options.getByLabel('Text size').selectOption('normal');
  await options.getByLabel('High contrast').uncheck();
  await options.getByLabel('Simple view').uncheck();
  await page.getByRole('button', { name: 'Save appearance' }).click();
  await expect(html).not.toHaveAttribute('data-simple');
  await expect(html).not.toHaveAttribute('data-contrast');
  await expect(html).not.toHaveAttribute('data-text-size');
});
