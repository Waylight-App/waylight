import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test.describe('space widget', () => {
  // The service worker would answer the API itself; page.route only sees requests without it.
  test.use({ serviceWorkers: 'block' });

  test('shows launches and the picture of the day once switched on (#64)', async ({ page }) => {
    const net = new Date(Date.now() + 5 * 3600_000).toISOString();
    await page.route(
      (u) => u.pathname === '/api/space/launches',
      (r) =>
        r.fulfill({
          json: [
            {
              id: '1',
              mission: 'Starlink 12-3',
              missionType: 'Communications',
              description: '',
              rocket: 'Falcon 9',
              provider: 'SpaceX',
              pad: 'SLC-40',
              location: 'Cape Canaveral',
              net,
              status: 'go',
              webcastLive: true,
            },
          ],
        }),
    );
    await page.route(
      (u) => u.pathname === '/api/space/apod',
      (r) =>
        r.fulfill({
          json: {
            date: '2026-11-13',
            title: 'The Pleiades',
            explanation: 'Seven sisters.',
            copyright: 'Jane Doe',
            mediaType: 'image',
            image: '/api/space/apod/image?d=2026-11-13',
            link: '',
          },
        }),
    );
    await page.route(
      (u) => u.pathname === '/api/space/apod/image',
      (r) =>
        r.fulfill({
          contentType: 'image/png',
          body: Buffer.from(
            'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
            'base64',
          ),
        }),
    );

    await page.goto('/');
    await page.getByLabel('Password').fill(password);
    await page.getByRole('button', { name: 'Log in' }).click();
    await expect(page.getByTestId('dashboard')).toBeVisible();
    await account(page, 'Settings');
    await page.getByRole('button', { name: 'Space', exact: true }).click();
    await page.getByLabel('Show rocket launches (Launch Library 2)').check();
    await page.getByLabel('Show the Astronomy Picture of the Day (NASA)').check();
    await page.getByRole('button', { name: 'Save', exact: true }).click();
    await expect(page.getByText('Saved.')).toBeVisible();
    // The stable background address of today's picture is offered for copying.
    await expect(page.getByText(/\/api\/space\/apod\/image$/)).toBeVisible();

    await page.goto('/');
    await account(page, 'Edit');
    await page.getByRole('button', { name: '+ Tab' }).click();
    await page.getByRole('textbox', { name: 'Name' }).fill('Orbit');
    await page.getByRole('button', { name: 'Create' }).click();
    await page.getByRole('button', { name: '+ Add widget' }).click();
    await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('rakete');
    await page.getByRole('button', { name: /^Space/ }).click();
    await page.getByRole('button', { name: 'Save', exact: true }).click();
    await account(page, 'Done');
    await expect(page.getByText('Starlink 12-3')).toBeVisible();
    await expect(page.getByText('in 5 h')).toBeVisible();
    await expect(page.getByText('Live', { exact: true })).toBeVisible();
    await expect(page.getByRole('img', { name: 'The Pleiades' })).toBeVisible();

    // Switch the sources off again so no later test makes real requests.
    const status = await (await page.request.get('/api/auth/status')).json();
    const res = await page.request.put('/api/space/settings', {
      headers: { 'X-CSRF-Token': status.csrfToken as string },
      data: { launches: false, picture: false, keepKey: true },
    });
    expect(res.ok()).toBeTruthy();
  });
});
