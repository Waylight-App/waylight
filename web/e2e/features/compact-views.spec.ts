import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('a small tile shows the compact view and opens the details (#595)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Compact');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('to-do');
  await page.getByRole('button', { name: /^To-dos/ }).click();
  // The settings form opens right after adding; wait for it before going on.
  const settings = page.getByRole('dialog', { name: 'To-dos settings' });
  await settings.getByRole('button', { name: 'Save', exact: true }).click();
  await expect(settings).toHaveCount(0);
  await page.getByRole('combobox', { name: 'Size of To-dos' }).selectOption('small');
  await account(page, 'Done');

  // Other specs share the account, so the number of open to-dos varies.
  const compact = page.getByTestId('compact-view');
  await expect(compact).toContainText(/open|Nothing to do/);
  const open = page.getByRole('button', { name: 'To-dos: show details' });
  await open.click();
  const panel = page.getByRole('dialog', { name: 'To-dos' });
  await expect(panel.getByRole('textbox', { name: 'New to-do' })).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(panel).toHaveCount(0);
  await expect(open).toBeFocused();

  // "Edit widget" in the panel opens the widget's settings (#594).
  await open.click();
  await panel.getByRole('button', { name: 'Edit widget' }).click();
  await expect(panel).toHaveCount(0);
  await expect(page.getByRole('dialog', { name: /To-dos/ })).toBeVisible();
  await page.keyboard.press('Escape');

  // The size is stored with the widget.
  await page.reload();
  await page.getByRole('button', { name: 'Compact', exact: true }).click();
  await expect(page.getByTestId('compact-view')).toBeVisible();
});
