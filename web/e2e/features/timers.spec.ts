import { expect, test } from '@playwright/test';
import { password } from '../support';

test('a timer from the search runs out with a toast and survives a reload (#474)', async ({ page }) => {
  await page.clock.install({ time: new Date('2026-11-16T12:00:00') });
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  const box = page.getByRole('combobox', { name: 'Quick search (Ctrl+K or /)' });

  await page.keyboard.press('Control+k');
  await box.fill('timer 2 min Tee');
  await expect(page.getByRole('option', { name: /Start a 2:00 timer · Tee/ })).toBeVisible();
  await box.press('Enter');
  await expect(page.getByText('Timer started, runs out at')).toBeVisible();

  // Stored in the browser: still running after a reload, listed under "timer".
  await page.reload();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await page.keyboard.press('Control+k');
  await box.fill('timer');
  await expect(page.getByRole('option', { name: /Tee/ })).toBeVisible();
  await page.keyboard.press('Escape');

  await page.clock.fastForward('02:05');
  const toast = page.getByRole('alert').filter({ hasText: 'Tee' });
  await expect(toast).toBeVisible();
  await toast.getByRole('button', { name: 'Dismiss' }).click();
  await expect(toast).toHaveCount(0);

  // An alarm time that already passed today rings tomorrow.
  await page.keyboard.press('Control+k');
  await box.fill('wecker 6:30');
  await expect(page.getByRole('option', { name: /Set an alarm for/ })).toContainText('Tomorrow');
});
