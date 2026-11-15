import { expect, test } from '@playwright/test';
import { password } from '../support';

test('the smart search answers questions about the server (#156)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await page.keyboard.press('Control+k');
  const box = page.getByRole('combobox', { name: 'Quick search (Ctrl+K or /)' });
  const answers = page.getByRole('region', { name: 'Answers' });

  await box.fill('cpu');
  await expect(answers.getByRole('article', { name: 'CPU' })).toContainText('%');

  await box.fill('wie voll ist die platte');
  await expect(answers.getByRole('article', { name: 'Disks' })).toContainText('free');

  // Service names with a typo; the sample data has DuckDuckGo.
  await box.fill('duckduckgoo');
  await expect(answers.getByRole('article', { name: 'DuckDuckGo' })).toBeVisible();

  // The weather module is disabled in this test: no weather answer.
  await box.fill('wetter');
  await expect(answers).toHaveCount(0);

  // "Open widget" jumps to the system widget on the Server tab.
  await box.fill('/ram');
  await answers
    .getByRole('article', { name: 'Memory' })
    .getByRole('button', { name: /Open widget/ })
    .click();
  await expect(page.getByRole('button', { name: 'Server', exact: true })).toHaveAttribute(
    'aria-current',
    'page',
  );
});
