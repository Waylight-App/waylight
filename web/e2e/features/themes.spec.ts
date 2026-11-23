import { expect, test } from '@playwright/test';
import { password, openSettings } from '../support';

test('the theme gallery applies, restores, imports and exports themes (#111)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  const bg = () =>
    page.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue('--color-bg').trim());
  const before = await bg();
  await openSettings(page, 'Appearance & language', 'Appearance');

  await page.getByRole('button', { name: 'Apply theme Nord' }).click();
  await expect(page.getByRole('button', { name: 'Apply theme Nord' })).toHaveAttribute(
    'aria-pressed',
    'true',
  );
  await expect.poll(bg).toBe('#2e3440');
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');

  const theme = {
    id: 'e2e-mint',
    name: 'E2E Mint',
    scheme: 'light',
    tokens: {
      '--color-bg': '#e6fff4',
      '--color-surface': '#ffffff',
      '--color-text': '#123',
      '--color-accent': '#0a7',
    },
  };
  await page.getByLabel('Import theme…').setInputFiles({
    name: 'mint.json',
    mimeType: 'application/json',
    buffer: Buffer.from(JSON.stringify(theme)),
  });
  const exported = page.getByRole('link', { name: 'Download theme E2E Mint' });
  await expect(exported).toBeVisible();
  const res = await page.request.get((await exported.getAttribute('href')) ?? '');
  expect((await res.json()).tokens['--color-accent']).toBe('#0a7');
  await page.getByRole('button', { name: 'Delete theme E2E Mint' }).click();
  await expect(exported).toHaveCount(0);

  await page.getByRole('button', { name: 'Restore my own style' }).click();
  await expect.poll(bg).toBe(before);
  await expect(page.getByRole('button', { name: 'Restore my own style' })).toHaveCount(0);
});
