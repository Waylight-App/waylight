import { expect, test } from '@playwright/test';
import { openSettings, password } from '../support';

test('admins add, rename and remove a named connection (#708)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await openSettings(page, 'Connections & services', 'Connections');
  const section = page.getByRole('dialog', { name: 'Settings' }).getByRole('region', { name: 'Connections' });
  await expect(section.getByText('No connections saved yet.')).toBeVisible();

  // The first listed service is enough: the form asks for whatever it needs.
  await section.getByLabel('Service').selectOption({ index: 1 });
  await section.getByRole('textbox', { name: 'Name', exact: true }).fill('Home');
  const address = section.getByRole('textbox', { name: 'Address', exact: true });
  if (await address.isVisible()) await address.fill('http://service.invalid');
  for (const input of await section.locator('form input[autocomplete="off"]').all()) {
    await input.fill('secret-value');
  }
  await section.getByRole('button', { name: 'Save and check' }).click();

  const item = section.getByRole('listitem').filter({ hasText: 'Home' });
  await expect(item).toBeVisible();
  await expect(item).toContainText('Not used by any widget');

  await item.getByRole('button', { name: 'Rename or move Home' }).click();
  await section.getByRole('textbox', { name: 'Name', exact: true }).fill('Office');
  await section.getByRole('button', { name: 'Save', exact: true }).click();
  const renamed = section.getByRole('listitem').filter({ hasText: 'Office' });
  await expect(renamed).toBeVisible();

  page.once('dialog', (d) => d.accept());
  await renamed.getByRole('button', { name: 'Remove Office' }).click();
  await expect(section.getByText('No connections saved yet.')).toBeVisible();
});
