import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('countdowns count the days and can be edited and deleted (#54)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Dates');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('prüfung');
  await page.getByRole('button', { name: /^Countdowns/ }).click();
  // The widget has settings; they open right after adding.
  await page.getByLabel('Highlight when fewer days are left than').fill('5');
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await account(page, 'Done');

  const inDays = (n: number) => {
    const d = new Date();
    d.setDate(d.getDate() + n);
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
  };
  await page.getByRole('button', { name: '+ New countdown' }).click();
  const dialog = page.getByRole('dialog', { name: 'New countdown' });
  await dialog.getByLabel('Title').fill('Exam');
  await dialog.getByLabel('Date').fill(inDays(3));
  await dialog.getByLabel('Emoji').fill('📚');
  await dialog.getByRole('button', { name: 'Save', exact: true }).click();
  await expect(dialog).toHaveCount(0);
  const item = page.getByRole('listitem').filter({ hasText: 'Exam' });
  await expect(item).toContainText('3 days');
  await expect(item).toHaveClass(/soon/);

  await item.getByRole('button', { name: 'Edit Exam' }).click();
  const edit = page.getByRole('dialog', { name: 'Edit countdown' });
  await edit.getByLabel('Title').fill('Holiday');
  await edit.getByLabel('Date').fill(inDays(30));
  await edit.getByRole('button', { name: 'Save', exact: true }).click();
  const holiday = page.getByRole('listitem').filter({ hasText: 'Holiday' });
  await expect(holiday).toContainText('30 days');
  await expect(holiday).not.toHaveClass(/soon/);

  page.once('dialog', (d) => d.accept());
  await holiday.getByRole('button', { name: 'Delete Holiday' }).click();
  await expect(page.getByText('No countdowns yet.')).toBeVisible();
});
