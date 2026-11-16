import { expect, type Page, test } from '@playwright/test';
import { password } from '../support';

async function openSearch(page: Page) {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await page.keyboard.press('Control+k');
  return page.getByRole('combobox', { name: 'Quick search (Ctrl+K or /)' });
}

test('"?" lists what the search can do, without disabled modules (#479)', async ({ page }) => {
  const box = await openSearch(page);
  const help = page.getByRole('region', { name: 'What the search can do' });

  await box.fill('?');
  await expect(help.getByRole('heading', { name: 'Time and dates' })).toBeVisible();
  await expect(help.getByText('How hard the processor of your server is working')).toBeVisible();
  // The weather module is disabled in the test server.
  await expect(help.getByText('The weather today or on another day')).toHaveCount(0);

  // "? disk" filters; choosing an example puts it into the field.
  await box.fill('? disk');
  await expect(help.getByText('How full the disks are')).toBeVisible();
  await expect(help.getByText('How hard the processor of your server is working')).toHaveCount(0);
  await help.getByRole('button', { name: 'disk space' }).click();
  await expect(box).toHaveValue('disk space');
  await expect(
    page.getByRole('region', { name: 'Answers' }).getByRole('article', { name: 'Disks' }),
  ).toBeVisible();
});

test.describe('with the weather module', () => {
  // The service worker fetches /api/modules itself, so page.route would not
  // see the request; blocking it lets the test enable weather in the list.
  test.use({ serviceWorkers: 'block' });

  test('"?" lists the weather once the weather module is enabled (#479)', async ({ page }) => {
    // The test server runs without weather; the help only reads the module list.
    await page.route('**/api/modules', async (route) => {
      const response = await route.fetch();
      const info = (await response.json()) as { modules: string[] };
      await route.fulfill({ response, json: { ...info, modules: [...info.modules, 'weather'] } });
    });
    const box = await openSearch(page);
    await box.fill('? weather');
    await expect(
      page
        .getByRole('region', { name: 'What the search can do' })
        .getByText('The weather today or on another day'),
    ).toBeVisible();
  });
});
