import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('the phone navigation ends with "Widgets", which opens the widget picker (#847)', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  const nav = page.getByRole('navigation', { name: 'Dashboard tabs' });
  const widgets = nav.getByRole('button', { name: 'Widgets', exact: true });
  // Earlier tests leave many tabs behind: the entry stays reachable while the tabs scroll.
  await expect(nav.getByRole('button').last()).toHaveText('Widgets');
  await expect(widgets).toBeInViewport();
  const box = await widgets.boundingBox();
  expect(box?.height).toBeGreaterThanOrEqual(44);

  const before = await page.locator('[data-widget-id]').count();
  await widgets.click();
  const picker = page.getByRole('dialog', { name: 'Add widget' });
  await expect(picker).toBeVisible();
  await picker.getByRole('searchbox', { name: 'Search widgets…' }).fill('countdown');
  await picker.getByRole('button', { name: /^Countdowns/ }).click();
  // As from the edit mode, the widget's settings open right after adding.
  await page
    .getByRole('dialog', { name: /settings/ })
    .getByRole('button', { name: 'Save', exact: true })
    .click();
  await account(page, 'Done');
  await expect(page.locator('[data-widget-id]')).toHaveCount(before + 1);

  // Wide screens keep the tab pills without the entry.
  await page.setViewportSize({ width: 1280, height: 900 });
  await expect(widgets).toBeHidden();
});
