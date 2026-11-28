import { expect, test } from '@playwright/test';
import { account, password, openSettings } from '../support';

test('a shopping list on Waylight works without CalDAV and can be exported (#471)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Errands');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('einkauf');
  await page.getByRole('button', { name: /^Shopping list/ }).click();
  await account(page, 'Done');

  // Without any list the widget offers to create one; no CalDAV needed.
  await page.getByRole('button', { name: 'Create list' }).click();

  const item = page.getByRole('textbox', { name: 'New item' });
  await item.fill('Milk');
  await item.press('Enter');
  await item.fill('Bread');
  await item.press('Enter');
  await page.getByRole('checkbox', { name: 'Milk' }).check();
  await page.getByRole('button', { name: 'Remove “Bread”' }).click();
  await expect(page.getByRole('checkbox', { name: 'Bread' })).toHaveCount(0);
  // Ticking off and removing wait for their undo toast (#867) before the server hears of them.
  await expect(page.getByTestId('undo-toast')).toHaveCount(0, { timeout: 12_000 });

  // The ticked item survives a reload: it is stored on the server.
  await page.reload();
  await expect(page.getByRole('checkbox', { name: 'Milk' })).toBeChecked();

  // A second list brings up the choice in the widget.
  await openSettings(page, 'Dashboard & widgets', 'Shopping list');
  await expect(page.getByRole('heading', { name: 'Lists on Waylight' })).toBeVisible();
  await page.getByPlaceholder('Name of the new list').fill('Hardware store');
  await page.getByRole('button', { name: 'Create list' }).click();
  await expect(page.getByRole('textbox', { name: 'Name', exact: true }).last()).toHaveValue('Hardware store');

  // Export as text: one item per line.
  const download = page.waitForEvent('download');
  await page.getByRole('link', { name: 'as text' }).first().click();
  const file = await download;
  const text = await (await file.createReadStream()).toArray();
  expect(Buffer.concat(text).toString()).toContain('Milk');
  await page.keyboard.press('Escape');

  // The widget reads the lists when it loads; with two it offers the choice.
  await page.reload();
  await expect(page.getByRole('combobox', { name: 'List' })).toBeVisible();
});
