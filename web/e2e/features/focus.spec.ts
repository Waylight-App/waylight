import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('the focus timer runs on timestamps and counts finished sessions (#58)', async ({ page }) => {
  await page.clock.install();
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
  await page.getByLabel('Focus (min)').fill('1');
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await account(page, 'Done');

  const timer = page.getByRole('timer');
  await expect(timer).toHaveText('1:00');
  await expect(page.getByText('0 sessions in the last 7 days')).toBeVisible();
  await page.getByRole('button', { name: 'Start', exact: true }).click();
  await page.clock.runFor(30_000);
  await expect(timer).toHaveText('0:30');

  // A sleeping tab: the phase ends once, the break waits to be started.
  await page.clock.runFor(10 * 60_000);
  await expect(page.getByText('Short break', { exact: true })).toBeVisible();
  await expect(timer).toHaveText('5:00');
  await expect(page.getByText('1 sessions in the last 7 days')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Start', exact: true })).toBeVisible();
});
