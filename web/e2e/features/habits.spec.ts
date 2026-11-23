import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('habits are checked off with streaks and a heatmap (#106)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Routine');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('gewohnheit');
  await page.getByRole('button', { name: /^Habits/ }).click();
  await page.getByRole('button', { name: 'Save', exact: true }).click(); // widget settings
  await account(page, 'Done');

  await page.getByRole('button', { name: '+ New habit' }).click();
  const dialog = page.getByRole('dialog', { name: 'New habit' });
  await dialog.getByLabel('Name').fill('Read');
  await dialog.getByLabel('Emoji').fill('📖');
  await dialog.getByRole('button', { name: 'Save', exact: true }).click();
  await expect(dialog).toHaveCount(0);

  const item = page.getByRole('listitem').filter({ hasText: 'Read' });
  await item.getByRole('checkbox', { name: 'Read done today' }).check();
  await expect(item.locator('.streak')).toHaveText(/^\s*1\s*$/);
  // Forgot to tick it yesterday: the heatmap cell of yesterday is a button.
  await item.getByRole('button', { name: 'Read: check off yesterday' }).click();
  await expect(item.locator('.streak')).toHaveText(/^\s*2\s*$/);
  // Undoing today keeps the streak from yesterday (today is still open).
  await item.getByRole('checkbox', { name: 'Read done today' }).uncheck();
  await expect(item.locator('.streak')).toHaveText(/^\s*1\s*$/);

  await page.reload();
  await expect(page.getByRole('listitem').filter({ hasText: 'Read' }).locator('.streak')).toHaveText(
    /^\s*1\s*$/,
  );
});
