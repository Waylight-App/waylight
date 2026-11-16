import { expect, test } from '@playwright/test';
import { password } from '../support';

// The panel is loaded on first use; the service worker would serve the chunk
// itself, so page.route could not slow it down.
test.use({ serviceWorkers: 'block', permissions: ['clipboard-read', 'clipboard-write'] });

test('Enter before the search panel has loaded still gets the answer, not the web (#501)', async ({
  page,
}) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  // Hold the panel back long enough to press Enter before it arrives.
  await page.route('**/assets/SearchPanel-*.js', async (route) => {
    await new Promise((resolve) => setTimeout(resolve, 1000));
    await route.continue();
  });
  const start = page.url();
  await page.keyboard.press('Control+k');
  const box = page.getByRole('combobox', { name: 'Quick search (Ctrl+K or /)' });
  await box.fill('uhrzeit');
  await box.press('Enter');

  // Enter waited for the panel and copied the time, as it does once loaded.
  await expect.poll(() => page.evaluate(() => navigator.clipboard.readText())).toMatch(/^\d{1,2}:\d{2}/);
  expect(page.url()).toBe(start);
});
