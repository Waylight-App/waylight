import { expect, test } from '@playwright/test';
import { openSettings, password } from '../support';

test('the wall tablet assistant creates a display with large text and night mode (#680)', async ({
  page,
  browser,
}) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await openSettings(page, 'People & access', 'Kiosk displays');
  await page.getByRole('button', { name: 'Set up a wall tablet …' }).click();

  const assistant = page.getByRole('dialog', { name: 'Set up a wall tablet' });
  await expect(assistant.getByRole('checkbox', { name: /Large text/ })).toBeChecked();
  await expect(assistant.getByRole('checkbox', { name: 'Dim the screen at night' })).toBeChecked();
  await assistant.getByLabel('Name of the tablet').fill('kitchen');
  await assistant.getByRole('radio', { name: 'Tabs I already have' }).check();
  await assistant.getByRole('checkbox', { name: 'Home' }).check();
  await assistant.getByRole('button', { name: 'Set up the tablet' }).click();

  await expect(assistant.getByRole('img', { name: 'QR code of the kiosk link' })).toBeVisible();
  const link = (await assistant.locator('code').textContent()) ?? '';
  expect(link).toMatch(/\/kiosk#.+/);
  await assistant.getByRole('button', { name: 'Done' }).click();
  const summary = page.getByRole('dialog').getByRole('listitem').filter({ hasText: 'kitchen' });
  await expect(summary).toContainText('dimmed 22:00–06:30');
  await expect(summary).toContainText('Large text');

  const context = await browser.newContext();
  const display = await context.newPage();
  await display.goto(link);
  await expect(display.getByTestId('dashboard')).toBeVisible();
  await expect(display.locator('html')).toHaveClass(/kiosk-large/);
  await context.close();

  const settings = page.getByRole('dialog');
  page.once('dialog', (d) => d.accept());
  await settings.getByRole('button', { name: 'Delete' }).first().click();
  await expect(settings.getByText('No kiosk displays yet.')).toBeVisible();
});
