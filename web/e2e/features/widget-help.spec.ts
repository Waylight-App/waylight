import { expect, type Page, test } from '@playwright/test';
import { password } from '../support';

async function logIn(page: Page) {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
}

test('the "?" of a widget opens its help with a link to the documentation (#678)', async ({ page }) => {
  await logIn(page);
  const help = page.getByRole('button', { name: /^Help for / }).first();
  const name = ((await help.getAttribute('aria-label')) ?? '').replace('Help for ', '');
  await help.click();

  const dialog = page.getByRole('dialog', { name: `About ${name}` });
  await expect(dialog.getByRole('heading', { name: 'What it shows' })).toBeVisible();
  await expect(dialog.getByRole('heading', { name: 'How to set it up' })).toBeVisible();
  await expect(dialog.getByRole('link', { name: /Read more in the documentation/ })).toHaveAttribute(
    'href',
    /^https:\/\/getwaylight\.app\/docs\//,
  );
  await page.keyboard.press('Escape');
  await expect(dialog).toHaveCount(0);
});

test('a plain question in the search finds the help of a widget (#678)', async ({ page }) => {
  await logIn(page);
  await page.keyboard.press('Control+k');
  const box = page.getByRole('combobox', { name: 'Quick search (Ctrl+K or /)' });
  await box.fill('How do I add a note?');

  const group = page.getByRole('group', { name: 'Help' });
  await group.getByRole('option', { name: /^Notes/ }).click();
  await expect(
    page.getByRole('dialog', { name: 'About Notes' }).getByRole('heading', { name: 'How to set it up' }),
  ).toBeVisible();
});
