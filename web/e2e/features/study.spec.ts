import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('study progress sums up credits and the weighted average (#57)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Settings');
  await page.getByRole('button', { name: 'Study progress', exact: true }).click();
  await page.getByLabel('ECTS of the degree').fill('210');
  await page.getByRole('button', { name: 'Save', exact: true }).first().click();
  await expect(page.getByText('Saved.')).toBeVisible();
  for (const [name, ects, status, grade] of [
    ['Analysis I', '10', 'passed', '1,0'],
    ['Lineare Algebra', '5', 'passed', '2.5'],
    ['Programmieren', '8', 'in_progress', ''],
  ] as const) {
    await page.getByRole('button', { name: 'Add module' }).click();
    await page.getByLabel('Module', { exact: true }).fill(name);
    await page.getByLabel('ECTS', { exact: true }).fill(ects);
    await page.getByLabel('Status').selectOption(status);
    if (grade) await page.getByLabel('Grade').fill(grade);
    await page.getByRole('button', { name: 'Save', exact: true }).last().click();
    await expect(page.getByRole('cell', { name, exact: true })).toBeVisible();
  }

  await page.goto('/');
  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Uni');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('ects');
  await page.getByRole('button', { name: /^Study progress/ }).click();
  await account(page, 'Done');
  await expect(page.getByText('15 of 210 ECTS')).toBeVisible();
  // (1.0 · 10 + 2.5 · 5) / 15 = 1.5
  await expect(page.getByText('Average 1.5')).toBeVisible();
  await expect(page.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '15');
  await expect(page.getByText('Programmieren')).toBeVisible();
});
