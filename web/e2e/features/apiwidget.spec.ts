import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('the JSON widget shows values from an endpoint picked with JSONPath (#48)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  // The server's own health check is a JSON endpoint on loopback, which counts as private.
  const url = new URL('/healthz', page.url()).toString();
  // JSON APIs are an advanced setting (#602): hidden in the areas, but a
  // direct link opens the section anyway.
  await account(page, 'Settings');
  const nav = page.getByRole('navigation', { name: 'Settings areas' });
  await nav.getByRole('button', { name: 'Connections & services', exact: true }).click();
  await expect(nav.getByRole('button', { name: 'Proxmox', exact: true })).toBeVisible();
  await expect(nav.getByRole('button', { name: 'JSON APIs', exact: true })).toHaveCount(0);
  await page.getByRole('button', { name: 'Close' }).click();
  await page.goto('/?settings=apiwidget');
  await expect(page.getByRole('region', { name: 'JSON APIs', exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Add endpoint' }).click();
  await page.getByLabel('Name').fill('Self');
  await page.getByLabel('URL').fill(url);
  await page.getByLabel('Label').fill('Health');
  await page.getByLabel('JSONPath', { exact: true }).fill('$.status');
  await page.getByRole('button', { name: 'Add value' }).click();
  await page.getByLabel('Label').nth(1).fill('Missing');
  await page.getByLabel('JSONPath', { exact: true }).nth(1).fill('$.nope');
  await page.getByRole('button', { name: 'Test', exact: true }).click();
  const preview = page.getByRole('region', { name: 'Test' });
  await expect(preview).toContainText('Health');
  await expect(preview).toContainText('ok');
  await expect(preview).toContainText('"nope" not found');
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await expect(page.getByText('Saved.')).toBeVisible();
  await expect(page.getByText('Self', { exact: true })).toBeVisible();

  await page.goto('/');
  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('JSON');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('jsonpath');
  await page.getByRole('button', { name: /^JSON API/ }).click();
  await expect(page.getByRole('combobox', { name: 'Endpoint' })).toHaveValue(/.+/);
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await account(page, 'Done');
  const widget = page.locator('dl').filter({ hasText: 'Health' });
  await expect(widget).toContainText('ok');
  await expect(page.getByText('Self', { exact: true })).toBeVisible();
});
