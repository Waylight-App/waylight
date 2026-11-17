import { expect, test, type Page } from '@playwright/test';
import { account, password } from '../support';

/** Calls a state-changing endpoint with the session's CSRF token. */
async function post(page: Page, path: string) {
  await page.evaluate(async (p) => {
    const status = (await (await fetch('/api/auth/status')).json()) as { csrfToken?: string };
    const res = await fetch(p, { method: 'POST', headers: { 'X-CSRF-Token': status.csrfToken ?? '' } });
    if (!res.ok) throw new Error(`${p}: ${res.status}`);
  }, path);
}

async function signIn(page: Page) {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
}

test('the sign-in page shows the theme of the last account on this browser (#526)', async ({ page }) => {
  const look = page.locator('style[data-source="default-theme"]');

  await signIn(page);
  await post(page, '/api/appearance/themes/nord/apply');
  try {
    await page.reload();
    await expect(page.getByTestId('dashboard')).toBeVisible();
    await account(page, 'Log out');
    await expect(page.getByRole('heading', { name: 'Welcome back' })).toBeVisible();
    await expect(look).toHaveCount(1);
    expect(await look.textContent()).toContain('/* waylight theme: nord */');
    // Only the look is kept, nothing about the account.
    const kept = await page.evaluate(() => localStorage.getItem('waylight:sign-in-look') ?? '');
    expect(kept).not.toContain('owner');
  } finally {
    await signIn(page);
    await post(page, '/api/appearance/themes/restore');
  }

  // Without a gallery theme of its own the kept look is gone again, and the
  // sign-in page falls back to the instance's default theme (Amber).
  await page.reload();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await account(page, 'Log out');
  await expect(page.getByRole('heading', { name: 'Welcome back' })).toBeVisible();
  await expect(look).toHaveCount(1);
  const fallback = (await look.textContent()) ?? '';
  expect(fallback).toContain('/* waylight theme: amber */');
  expect(fallback).not.toContain('nord');
});
