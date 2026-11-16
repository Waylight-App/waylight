import { expect, test } from '@playwright/test';
import { password } from '../support';

test.use({ permissions: ['clipboard-read', 'clipboard-write'] });

test('small tools answer in the search field and Enter copies their value (#477)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await page.keyboard.press('Control+k');
  const box = page.getByRole('combobox', { name: 'Quick search (Ctrl+K or /)' });
  const answers = page.getByRole('region', { name: 'Answers' });

  // A password: New makes another one, Enter copies the one on screen.
  await box.fill('pw 16 ohne sonderzeichen');
  const card = answers.getByRole('article', { name: 'Password · 16 characters' });
  const value = card.locator('.value span');
  await expect(value).toHaveText(/^[A-Za-z2-9]{16}$/);
  const first = await value.textContent();
  await card.getByRole('button', { name: 'New' }).click();
  await expect(value).not.toHaveText(first ?? '');
  const shown = (await value.textContent())?.trim() ?? '';
  await box.press('Enter');
  await expect.poll(() => page.evaluate(() => navigator.clipboard.readText())).toBe(shown);

  // A colour in all notations.
  await box.fill('#1e90ff');
  await expect(answers.getByRole('article', { name: 'Colour' })).toContainText('rgb(30 144 255)');

  // A subnet.
  await box.fill('192.168.178.0/24');
  const net = answers.getByRole('article', { name: 'Network 192.168.178.0/24' });
  await expect(net).toContainText('192.168.178.255');
  await expect(net).toContainText('254');

  // Something else falls through to the normal search without a tool answer.
  await box.fill('rgb(300, 0, 0)');
  await expect(answers.getByRole('article', { name: 'Colour' })).toHaveCount(0);
});
