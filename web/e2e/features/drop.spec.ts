import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('the drop zone shares text and files and serves files as downloads (#101)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Ablage');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('teilen');
  await page.getByRole('button', { name: /^Drop zone/ }).click();
  await account(page, 'Done');

  await page.getByRole('textbox', { name: 'Text or link' }).fill('https://example.org/von-handy');
  await page.getByRole('button', { name: 'Drop', exact: true }).click();
  await expect(page.getByRole('link', { name: 'https://example.org/von-handy' })).toBeVisible();
  await expect(page.getByText(/expires in 2[34] h/)).toBeVisible();

  await page.getByLabel('File…').setInputFiles({
    name: 'evil.html',
    mimeType: 'text/html',
    buffer: Buffer.from('<script>alert(1)</script>'),
  });
  const download = page.getByRole('link', { name: 'Download evil.html' });
  await expect(download).toBeVisible();
  const res = await page.request.get((await download.getAttribute('href')) ?? '');
  expect(res.headers()['content-type']).toBe('application/octet-stream');
  expect(res.headers()['content-disposition']).toContain('attachment');

  // Images get a thumbnail (#259), although the file is served as an attachment.
  await page.getByLabel('File…').setInputFiles({
    name: 'punkt.png',
    mimeType: 'image/png',
    buffer: Buffer.from(
      'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==',
      'base64',
    ),
  });
  const thumb = page.locator('li', { hasText: 'punkt.png' }).locator('img.thumb');
  await expect(thumb).toBeVisible();
  await expect.poll(() => thumb.evaluate((img: HTMLImageElement) => img.naturalWidth)).toBe(1);
  await expect(page.locator('li', { hasText: 'evil.html' }).locator('img')).toHaveCount(0);

  for (let i = 0; i < 3; i++) await page.getByRole('button', { name: 'Delete', exact: true }).first().click();
  await expect(page.getByText('Nothing here.', { exact: false })).toBeVisible();
});
