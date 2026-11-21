import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('a small tile shows the compact view and opens the details (#595)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Compact');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('to-do');
  await page.getByRole('button', { name: /^To-dos/ }).click();
  // The settings form opens right after adding; wait for it before going on.
  const settings = page.getByRole('dialog', { name: 'To-dos settings' });
  await settings.getByRole('button', { name: 'Save', exact: true }).click();
  await expect(settings).toHaveCount(0);
  await page.getByRole('combobox', { name: 'Size of To-dos' }).selectOption('small');
  await account(page, 'Done');

  // Other specs share the account, so the number of open to-dos varies.
  const compact = page.getByTestId('compact-view');
  await expect(compact).toContainText(/open|Nothing to do/);
  const open = page.getByRole('button', { name: 'To-dos: show details' });
  await open.press('Enter'); // the quick action may sit where a click would land
  const panel = page.getByRole('dialog', { name: 'To-dos', exact: true });
  await expect(panel.getByRole('textbox', { name: 'New to-do' })).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(panel).toHaveCount(0);
  await expect(open).toBeFocused();

  // "Edit widget" in the panel opens the widget's settings (#594).
  await open.press('Enter'); // the quick action may sit where a click would land
  await panel.getByRole('button', { name: 'Edit widget' }).click();
  await expect(panel).toHaveCount(0);
  await expect(page.getByRole('dialog', { name: /To-dos/ })).toBeVisible();
  await page.keyboard.press('Escape');

  // The size is stored with the widget.
  await page.reload();
  await page.getByRole('button', { name: 'Compact', exact: true }).click();
  await expect(page.getByTestId('compact-view')).toBeVisible();
});

test('the picker previews the sizes and adds the widget in the chosen one (#596)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Sizes');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('focus');
  await page.getByRole('button', { name: 'Choose a size for Focus timer' }).click();
  // Previews use example data and say what each size shows.
  await expect(page.getByRole('button', { name: 'Medium: Full view' })).toBeVisible();
  await page.getByRole('button', { name: 'Strip: 50 min focus today' }).click();
  const settings = page.getByRole('dialog', { name: 'Focus timer settings' });
  await settings.getByRole('button', { name: 'Save', exact: true }).click();
  await expect(page.getByRole('combobox', { name: 'Size of Focus timer' })).toHaveValue('strip');
  await account(page, 'Done');
  await expect(page.getByTestId('compact-view')).toContainText('min focus today');
});

test('a compact view ticks off the next habit with its quick action (#598)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  const status = await (await page.request.get('/api/auth/status')).json();
  const created = await page.request.post('/api/habits', {
    headers: { 'X-CSRF-Token': status.csrfToken as string },
    data: { name: 'Stretch', emoji: '🧘', days: [] },
  });
  expect(created.ok()).toBe(true);

  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Actions');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('habits');
  await page.getByRole('button', { name: 'Choose a size for Habits' }).click();
  await page.getByRole('button', { name: /^Small:/ }).click();
  const settings = page.getByRole('dialog', { name: 'Habits settings' });
  await settings.getByRole('button', { name: 'Save', exact: true }).click();
  await account(page, 'Done');

  // The action sits on the tile, above the button that opens the details.
  const done = page.getByRole('button', { name: 'Done: Stretch' });
  await done.click();
  await expect(done).toHaveCount(0);
  await expect(page.getByRole('dialog', { name: 'Habits' })).toHaveCount(0);
});

test('status widgets fold into the system summary and come back (#593)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('System');
  await page.getByRole('button', { name: 'Create' }).click();
  for (const [query, name] of [
    ['services', /^Services/],
    ['backup', /^Backups/],
  ] as const) {
    await page.getByRole('button', { name: '+ Add widget' }).click();
    await page.getByRole('searchbox', { name: 'Search widgets…' }).fill(query);
    // Inside the picker: the top bar has a "Services and bookmarks" button as well.
    await page.getByRole('dialog', { name: 'Add widget' }).getByRole('button', { name }).first().click();
  }
  await account(page, 'Done');

  const summary = page.getByTestId('system-summary');
  await expect(summary).toContainText('Services:');
  const grid = page.locator('.grid [data-widget-id]');
  await expect(grid).toHaveCount(2);
  await summary.getByRole('button', { name: 'Collapse' }).click();
  await expect(grid).toHaveCount(0);
  // No "empty tab" hint while the widgets are folded away.
  await expect(page.getByText('This tab is empty')).toHaveCount(0);
  await summary.getByRole('button', { name: 'Show all' }).click();
  await expect(grid).toHaveCount(2);
});
