import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('quick capture saves a shared page only after confirming (#66)', async ({ page }) => {
  const shared = new URLSearchParams({ url: 'https://example.org/article', title: 'An article', text: '' });
  // Not signed in yet: the capture page waits behind the login.
  await page.goto(`/capture?${shared}`);
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByRole('heading', { name: 'Save to trisse-home' })).toBeVisible();
  await expect(page.getByRole('radio', { name: 'Bookmark' })).toBeChecked();
  await expect(page.getByLabel('Title')).toHaveValue('An article');
  await expect(page.getByLabel('Link')).toHaveValue('https://example.org/article');
  // Opening the page saved nothing.
  const before = await (await page.request.get('/api/bookmarks')).text();
  expect(before).not.toContain('example.org/article');

  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await expect(page.getByText('Bookmark saved.')).toBeVisible();
  expect(await (await page.request.get('/api/bookmarks')).text()).toContain('example.org/article');
  await page.getByRole('button', { name: 'Undo' }).click();
  await expect(page.getByText('Removed again.')).toBeVisible();
  expect(await (await page.request.get('/api/bookmarks')).text()).not.toContain('example.org/article');

  // Shared text becomes a note; the bookmarklet is offered in the settings.
  await page.goto(`/capture?${new URLSearchParams({ text: 'Idee: Fahrradtour' })}`);
  await expect(page.getByRole('radio', { name: 'Note' })).toBeChecked();
  await expect(page.getByLabel('Text')).toHaveValue('Idee: Fahrradtour');
  await page.goto('/');
  await account(page, 'Settings');
  await page.getByRole('button', { name: 'Search', exact: true }).click();
  await expect(page.getByRole('link', { name: '＋ trisse-home' })).toHaveAttribute(
    'href',
    /^javascript:.*\/capture\?url=/,
  );
});
