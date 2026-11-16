import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('own keywords are made in the settings or from a result and open first (#478)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  // A keyword for a search, made in the settings.
  await account(page, 'Settings');
  await page.getByRole('button', { name: 'Search', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Own keywords' })).toBeVisible();
  await page.getByRole('textbox', { name: 'Own keyword' }).fill('Merk');
  await page.getByRole('combobox', { name: 'Opens' }).selectOption('query');
  await page.getByRole('textbox', { name: 'Target' }).fill('merkliste');
  await page.getByRole('button', { name: 'Add' }).last().click();
  await expect(page.getByText('merk', { exact: true })).toBeVisible();

  // A quick capture prefix cannot become a keyword; the message says why.
  await page.getByRole('textbox', { name: 'Own keyword' }).fill('n');
  await page.getByRole('textbox', { name: 'Target' }).fill('notes');
  await page.getByRole('button', { name: 'Add' }).last().click();
  await expect(page.getByRole('alert')).toContainText('quick capture prefix for notes');
  await page.keyboard.press('Escape');

  // Typed alone, the keyword is the first row; Enter runs its search.
  await page.keyboard.press('Control+k');
  const field = page.getByRole('combobox', { name: 'Quick search (Ctrl+K or /)' });
  await field.fill('MERK');
  const first = page.getByRole('option').first();
  await expect(first).toContainText('merk → merkliste');
  await expect(first).toHaveAttribute('aria-selected', 'true');
  await field.press('Enter');
  await expect(field).toHaveValue('merkliste');
  // More than the keyword is a normal search again.
  await field.fill('merk mal');
  await expect(page.getByRole('option', { name: /merk →/ })).toHaveCount(0);

  // A keyword straight from a result: save a bookmark, find it, add "ex".
  await field.fill('b example.org Beispielseite');
  await field.press('Enter');
  await expect(page.getByText('Bookmark “Beispielseite” saved')).toBeVisible();
  // The panel reads bookmarks when it loads; a reload brings in the new one.
  await page.reload();
  await page.keyboard.press('Control+k');
  await field.fill('Beispielseite');
  await expect(page.getByRole('option', { name: /Beispielseite/ }).first()).toHaveAttribute(
    'aria-selected',
    'true',
  );
  await page.getByRole('button', { name: 'Add keyword…' }).click();
  await page.getByRole('textbox', { name: 'Keyword for “Beispielseite”' }).fill('ex');
  await page.getByRole('button', { name: 'Add', exact: true }).click();
  await expect(page.getByText('Keyword “ex” added')).toBeVisible();
  await field.fill('ex');
  await expect(page.getByRole('option').first()).toContainText('ex → Beispielseite');
});
