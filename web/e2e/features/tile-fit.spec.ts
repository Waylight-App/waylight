import { expect, test, type Page } from '@playwright/test';
import { password } from '../support';

// #737: a tile shows its first lines and its primary action without scrolling,
// at the size a template gives it, and never cuts a sentence mid-line.

/** Controls that are the point of a tile: setup buttons, add buttons. */
const primary = '.empty button, button.btn-primary, form button[type="submit"]';

/** The primary controls of every tile that lie (partly) outside the tile's content box. */
async function cutOff(page: Page): Promise<string[]> {
  return page.evaluate((selector) => {
    const found: string[] = [];
    for (const tile of document.querySelectorAll('article.widget')) {
      const box = tile.querySelector('.content')?.getBoundingClientRect();
      if (!box) continue;
      const name = tile.querySelector('h3')?.textContent?.trim() ?? '?';
      for (const el of tile.querySelectorAll<HTMLElement>(selector)) {
        const r = el.getBoundingClientRect();
        if (r.width === 0) continue;
        if (r.top < box.top - 1 || r.bottom > box.bottom + 1)
          found.push(`${name}: ${el.textContent?.trim()}`);
      }
      // A sentence of a not-set-up tile is whole or ends in an ellipsis; it is never clipped by the tile.
      for (const p of tile.querySelectorAll<HTMLElement>('.empty p')) {
        const r = p.getBoundingClientRect();
        if (r.bottom > box.bottom + 1) found.push(`${name}: text`);
      }
    }
    return found;
  }, primary);
}

test('the default templates show every primary action without scrolling (#737)', async ({ page }) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  const status = await (await page.request.get('/api/auth/status')).json();
  const headers = { 'X-CSRF-Token': status.csrfToken as string };

  const templates = (await (await page.request.get('/api/dashboard/templates/builtin')).json()) as {
    id: string;
  }[];
  expect(templates.length).toBeGreaterThan(0);
  // Not named after the template: other specs look for a tab called "Home" by substring.
  const tabName = (id: string) => `Layout ${templates.findIndex((x) => x.id === id) + 1}`;
  for (const tpl of templates) {
    const res = await page.request.post(`/api/dashboard/templates/builtin/${tpl.id}/tabs`, {
      headers,
      data: { name: tabName(tpl.id) },
    });
    expect(res.ok(), tpl.id).toBeTruthy();
  }
  await page.reload();

  for (const tpl of templates) {
    await page.getByRole('button', { name: tabName(tpl.id), exact: true }).click();
    await expect(page.locator('[aria-busy="true"]')).toHaveCount(0, { timeout: 15_000 });
    await expect.poll(() => cutOff(page), { message: tpl.id }).toEqual([]);
  }

  // The family template has a shopping list: however long it gets, the add field stays in place.
  await page.getByRole('button', { name: tabName('family'), exact: true }).click();
  const tile = page.locator('article.widget', { has: page.getByRole('heading', { name: /Shopping list/ }) });
  // The specs share one account: an earlier one may have made a list already.
  const create = tile.getByRole('button', { name: 'Create list' });
  if (await create.isVisible()) await create.click();
  const item = tile.getByRole('textbox', { name: 'New item' });
  for (let i = 1; i <= 15; i++) {
    await item.fill(`Item ${i}`);
    await item.press('Enter');
    await expect(tile.getByRole('checkbox', { name: `Item ${i}` })).toBeAttached();
  }
  await expect(item).toBeInViewport({ ratio: 1 });
  await expect(tile.getByRole('button', { name: 'Add', exact: true })).toBeInViewport({ ratio: 1 });
  await expect.poll(() => cutOff(page)).toEqual([]);
});
