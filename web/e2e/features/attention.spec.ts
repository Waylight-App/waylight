import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('a widget without its setting shows up under Needs your attention (#592)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Commute');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('departures');
  await page.getByRole('button', { name: /^Departures/ }).click();
  // Leave the settings without choosing a stop.
  await page
    .getByRole('dialog', { name: 'Departures settings' })
    .getByRole('button', { name: 'Close' })
    .click();
  await account(page, 'Done');

  // The dot in the top bar counts the entries and opens the page.
  const badge = page.getByRole('button', { name: /need(s)? your attention$/ });
  await expect(badge).toBeVisible({ timeout: 10_000 });
  await badge.click();
  const settings = page.getByRole('dialog', { name: 'Settings' });
  await expect(settings.getByRole('heading', { name: /Not set up yet/ })).toBeVisible();
  await expect(settings.getByText('Choose a stop for the departures.')).toBeVisible();

  // Its action opens the widget's own settings on its tab.
  await settings.getByRole('button', { name: 'Set up: Departures' }).click();
  await expect(page.getByRole('dialog', { name: 'Departures settings' })).toBeVisible();
});
