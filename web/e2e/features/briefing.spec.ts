import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('the daily briefing sums up the day and leads to the source widget (#52)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await page.getByRole('button', { name: 'Home', exact: true }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('briefing');
  await page.getByRole('button', { name: /^Daily briefing/ }).click();
  await account(page, 'Done');

  // "Steuer" was created for today by the quick capture test.
  const toTodos = page.getByRole('button', { name: 'To-dos', exact: true });
  await expect(toTodos).toBeVisible({ timeout: 15_000 });
  await expect(page.locator('.briefing')).toContainText('Steuer');
  await expect(page.locator('.briefing')).not.toContainText('Milch kaufen'); // due tomorrow
  // The waste calendar is not set up yet: that is no outage.
  await expect(page.locator('.briefing')).not.toContainText('Not available right now');

  // The heading scrolls to and highlights the to-do widget.
  await toTodos.click();
  await expect(page.locator('[data-widget-id].flash')).toContainText('Milch kaufen');

  // Hiding the section in the settings removes it from the widget.
  await account(page, 'Settings');
  await page.getByRole('button', { name: 'Daily briefing', exact: true }).click();
  await page.getByRole('checkbox', { name: 'To-dos' }).uncheck();
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await expect(page.getByText('Saved.', { exact: true })).toBeVisible();
  await page.reload();
  // Sign-ins of this test run still show up under "Since yesterday evening".
  await expect(page.getByRole('button', { name: 'Since yesterday evening', exact: true })).toBeVisible({
    timeout: 15_000,
  });
  await expect(page.getByRole('button', { name: 'To-dos', exact: true })).toHaveCount(0);
});
