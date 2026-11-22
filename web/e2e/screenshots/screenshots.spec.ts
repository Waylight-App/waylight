import { expect, test, type Page } from '@playwright/test';
import { account } from '../support';

/** Where the images go; referenced from the README. */
const outDir = '../docs/screenshots';

const variants = [
  { name: 'desktop-light', scheme: 'light', viewport: { width: 1440, height: 900 } },
  { name: 'desktop-dark', scheme: 'dark', viewport: { width: 1440, height: 900 } },
  { name: 'mobile-light', scheme: 'light', viewport: { width: 390, height: 844 } },
  { name: 'mobile-dark', scheme: 'dark', viewport: { width: 390, height: 844 } },
] as const;

async function openDashboard(page: Page) {
  await page.goto('/');
  // The demo opens on the sign-in page with its password shown (#539).
  // A fresh demo seeds its showcase before it answers, so wait for either
  // screen instead of guessing a delay.
  const fill = page.getByRole('button', { name: 'Fill in password' });
  const search = page.getByRole('combobox', { name: /Quick search/ });
  await expect(fill.or(search)).toBeVisible({ timeout: 60_000 });
  if (await fill.isVisible()) {
    await fill.click();
    await page.getByRole('button', { name: 'Log in' }).click();
  }
  // The guided tour starts for the demo account: skip it once (it is stored).
  const skip = page.getByRole('button', { name: 'Skip tour' });
  await skip.click({ timeout: 10_000 }).catch(() => undefined);
  await expect(page.locator('.tour')).toHaveCount(0);
  await expect(page.getByRole('combobox', { name: /Quick search/ })).toBeVisible({ timeout: 15_000 });
  // The demo banner is not part of the product screenshots.
  await page.addStyleTag({ content: '.demo-banner { display: none !important; }' });
  await settle(page);
}

/** Waits until widgets have loaded their data. */
async function settle(page: Page) {
  await expect(page.locator('[aria-busy="true"]')).toHaveCount(0, { timeout: 15_000 });
  await page.waitForTimeout(1500);
}

for (const v of variants) {
  test(`dashboard ${v.name}`, async ({ page }) => {
    await page.emulateMedia({ colorScheme: v.scheme, reducedMotion: 'reduce' });
    await page.setViewportSize(v.viewport);
    await openDashboard(page);
    await page.screenshot({
      path: `${outDir}/home-${v.name}.png`,
      fullPage: v.name.startsWith('mobile') ? false : true,
    });
    if (v.name.startsWith('desktop')) {
      await page.getByRole('button', { name: 'Server', exact: true }).click();
      await expect(page.getByText(/CPU \(\d+ cores\)/).first()).toBeVisible({ timeout: 15_000 });
      await settle(page);
      await page.screenshot({ path: `${outDir}/server-${v.name}.png`, fullPage: true });
      // Widget picker (#104).
      await account(page, 'Edit');
      await page.getByRole('button', { name: '+ Add widget' }).click();
      await expect(page.getByRole('searchbox', { name: 'Search widgets…' })).toBeFocused();
      await page.screenshot({ path: `${outDir}/picker-${v.name}.png` });
      await page.keyboard.press('Escape');
      await account(page, 'Done');
    }
  });
}

/** Calls a state-changing appearance endpoint with the session's CSRF token. */
async function post(page: Page, path: string) {
  await page.evaluate(async (p) => {
    const status = (await (await fetch('/api/auth/status')).json()) as { csrfToken?: string };
    const res = await fetch(p, { method: 'POST', headers: { 'X-CSRF-Token': status.csrfToken ?? '' } });
    if (!res.ok) throw new Error(`${p}: ${res.status}`);
  }, path);
}

// Amber is the default theme, so the pictures above already show it. Dark
// Amber is the second Waylight theme; it is always dark.
test('dashboard dark amber', async ({ page }) => {
  await page.emulateMedia({ colorScheme: 'dark', reducedMotion: 'reduce' });
  await page.setViewportSize({ width: 1440, height: 900 });
  await openDashboard(page);
  await post(page, '/api/appearance/themes/amber-dark/apply');
  try {
    await openDashboard(page);
    await page.screenshot({ path: `${outDir}/home-amber-dark-desktop.png`, fullPage: true });
  } finally {
    // Back to the default look, so the order of the tests does not matter.
    await post(page, '/api/appearance/themes/restore');
  }
});

// The guided setup (#588, #618). The demo connects every integration, so the
// Immich widget is shown the way a fresh install shows it: the server answers
// 404 "not configured", the same status it sends when nothing is stored.
// Only placeholder values appear; nothing leaves the demo instance.
test('guided setup', async ({ page }) => {
  await page.emulateMedia({ colorScheme: 'light', reducedMotion: 'reduce' });
  await page.setViewportSize({ width: 1440, height: 900 });
  await page.route('**/api/immich/overview*', (route) =>
    route.fulfill({ status: 404, json: { error: 'Immich is not configured yet.' } }),
  );
  await page.route('**/api/integrations/choices*', (route) =>
    route.fulfill({ json: [{ id: 'photos', name: 'Photos at home' }] }),
  );
  await openDashboard(page);
  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('Immich');
  await page.getByRole('dialog', { name: 'Add widget' }).getByText('Immich', { exact: true }).click();
  await page.getByRole('dialog', { name: 'Immich settings' }).getByRole('button', { name: 'Close' }).click();
  await account(page, 'Done');

  const widget = page.locator('article.widget').filter({ hasText: 'Immich' }).last();
  const shot = async (name: string, height?: number) => {
    // The widget scrolls inside its frame; give it the room the content needs.
    if (height) await widget.evaluate((el, h) => (el.style.height = `${h}px`), height);
    await widget.scrollIntoViewIfNeeded();
    await widget.screenshot({ path: `${outDir}/setup-${name}.png` });
  };
  try {
    // A saved connection is offered before anything is asked.
    await expect(widget.getByRole('button', { name: 'Use Photos at home' })).toBeVisible();
    await shot('connection');

    await widget.getByRole('button', { name: 'Set up', exact: true }).click();
    await widget.getByRole('textbox').first().fill('http://192.168.1.20:2283');
    await widget.getByRole('button', { name: 'Next' }).click();

    // The credential step: the steps in Immich's words, and the QR code that
    // lets the mobile app send the key.
    await widget.getByRole('button', { name: 'Paste from phone' }).click();
    await expect(widget.getByRole('img', { name: 'QR code for Paste from phone' })).toBeVisible();
    await shot('credential', 720);
  } finally {
    // Take the demo widget out again, so the order of the tests does not matter.
    await page.unrouteAll({ behavior: 'ignoreErrors' });
    await account(page, 'Edit');
    await widget.getByRole('button', { name: 'Remove widget' }).click();
    await account(page, 'Done');
  }
});
