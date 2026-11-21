import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('the focus timer is kept on the server and controlled from the widget (#58, #631)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Deep work');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('pomodoro');
  await page.getByRole('button', { name: /^Focus timer/ }).click();
  await page.getByLabel('Focus (min)').fill('2');
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await account(page, 'Done');

  const timer = page.getByRole('timer');
  await expect(timer).toHaveText('2:00');
  await expect(page.getByText(/sessions in the last 7 days/)).toBeVisible();
  await page.getByRole('button', { name: 'Start', exact: true }).click();
  await expect(timer).not.toHaveText('2:00');
  await page.getByRole('button', { name: 'Pause', exact: true }).click();

  // The server keeps the paused timer: a reload shows it, not a fresh one.
  await page.reload();
  await expect(page.getByRole('timer')).toHaveText(/^1:5\d$/);
  await expect(page.getByRole('button', { name: 'Start', exact: true })).toBeVisible();

  // Skip moves to the break; reset leaves the timer untouched for other tests.
  await page.getByRole('button', { name: 'Skip', exact: true }).click();
  await expect(page.getByText('Short break', { exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Skip', exact: true }).click();
  await page.getByRole('button', { name: 'Reset', exact: true }).click();
  await expect(page.getByRole('timer')).toHaveText('2:00');
});
