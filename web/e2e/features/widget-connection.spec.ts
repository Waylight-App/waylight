import { expect, test } from '@playwright/test';
import { account, openSettings, password } from '../support';

test('a widget reads through the connection chosen in its settings (#709)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await openSettings(page, 'Connections & services', 'Connections');
  const dialog = page.getByRole('dialog', { name: 'Settings' });
  const section = dialog.getByRole('region', { name: 'Connections' });
  for (const name of ['Garage', 'Attic']) {
    await section.getByLabel('Service').selectOption({ label: 'Immich' });
    await section.getByRole('textbox', { name: 'Name', exact: true }).fill(name);
    await section.getByRole('textbox', { name: 'Address', exact: true }).fill('http://immich.invalid');
    for (const input of await section.locator('form input[autocomplete="off"]').all()) {
      await input.fill('secret-value');
    }
    await section.getByRole('button', { name: 'Save and check' }).click();
    await expect(section.getByRole('listitem').filter({ hasText: name })).toBeVisible();
  }
  await page.keyboard.press('Escape');

  await page.getByRole('button', { name: 'Home', exact: true }).click();
  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('immich');
  await page.getByRole('button', { name: /^Immich/ }).click();
  const settings = page.getByRole('dialog', { name: 'Immich settings' });
  await settings.getByRole('combobox', { name: 'Which connection?' }).selectOption({ label: 'Attic' });
  const request = page.waitForRequest(/\/api\/immich\/overview\?connection=/);
  await settings.getByRole('button', { name: 'Save', exact: true }).click();
  await request;
  await account(page, 'Done');

  // Remove the test connections again; the widget falls back to the module settings.
  await openSettings(page, 'Connections & services', 'Connections');
  await expect(section.getByRole('listitem')).toHaveCount(2);
  for (const name of ['Garage', 'Attic']) {
    page.once('dialog', (d) => d.accept());
    await section.getByRole('button', { name: `Remove ${name}` }).click();
    await expect(section.getByRole('listitem').filter({ hasText: name })).toHaveCount(0);
  }
  await expect(section.getByText('No connections saved yet.')).toBeVisible();
});
