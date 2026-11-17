import { expect, test } from '@playwright/test';

// The sign-in page (#513): a welcome heading, the mark above the card, the
// guest sign-in as a real second button, and no sideways scrolling on a
// phone. The screenshots are attached to the report for review.
for (const v of [
  { name: 'light', scheme: 'light', viewport: { width: 1280, height: 800 } },
  { name: 'dark', scheme: 'dark', viewport: { width: 1280, height: 800 } },
  { name: 'mobile', scheme: 'light', viewport: { width: 390, height: 844 } },
] as const) {
  test(`sign-in page ${v.name} (#513)`, async ({ page }, info) => {
    await page.emulateMedia({ colorScheme: v.scheme });
    await page.setViewportSize(v.viewport);
    await page.goto('/');

    await expect(page.getByRole('heading', { name: 'Welcome back' })).toBeVisible();
    await expect(page.getByText('waylight', { exact: true })).toBeVisible();
    const login = page.getByRole('button', { name: 'Log in' });
    const guest = page.getByRole('button', { name: 'Sign in as guest' });
    await expect(login).toBeVisible();
    await expect(guest).toBeVisible();
    // Same size as the main button, not a text link.
    const [a, b] = [await login.boundingBox(), await guest.boundingBox()];
    expect(Math.round(b?.width ?? 0)).toBe(Math.round(a?.width ?? -1));

    const sideways = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
    expect(sideways).toBeLessThanOrEqual(0);

    // A wrong password shows the message and moves the focus to it.
    await page.getByLabel('Password').fill('not the password');
    await login.click();
    await expect(page.getByRole('alert')).toBeFocused();

    const path = info.outputPath(`sign-in-${v.name}.png`);
    await page.screenshot({ path });
    await info.attach(`sign-in-${v.name}`, { path, contentType: 'image/png' });
  });
}
