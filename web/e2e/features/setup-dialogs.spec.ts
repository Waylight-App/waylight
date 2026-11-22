import { expect, test } from '@playwright/test';
import { expectAccessible } from '../a11y';
import { password } from '../support';

// #738: every guided setup is the same dialog: symbol, plain title, one
// sentence, all values visible with a hint, the result of the test and one
// primary action.

test('the waste calendar and an integration use the same guided setup (#738)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  const status = await (await page.request.get('/api/auth/status')).json();
  const headers = { 'X-CSRF-Token': status.csrfToken as string };

  const tab = await (
    await page.request.post('/api/dashboard/tabs', { headers, data: { name: 'Guided' } })
  ).json();
  for (const [i, kind] of ['waste', 'immich'].entries()) {
    const res = await page.request.post(`/api/dashboard/tabs/${tab.id}/widgets`, {
      headers,
      data: { kind, x: i * 5, y: 0, w: 5, h: 5 },
    });
    expect(res.ok(), kind).toBeTruthy();
  }
  await page.reload();
  await page.getByRole('button', { name: 'Guided', exact: true }).click();

  // Waste calendar: link and file side by side, and where to find the calendar.
  await page.getByRole('button', { name: 'Set up', exact: true }).click();
  const waste = page.getByRole('dialog', { name: 'Set up the waste calendar' });
  await expect(waste.getByRole('textbox', { name: 'Link to the calendar' })).toBeVisible();
  await expect(waste.getByText('Drop the file here or choose it')).toBeVisible();
  await expect(waste.getByText('There is no central source.')).toBeVisible();
  await expectAccessible(page, 'waste calendar setup', 'dialog');

  await waste.getByRole('button', { name: 'Done' }).click();
  await expect(waste.getByRole('alert')).toHaveText('Paste a link or choose a file first.');

  const day = new Date();
  day.setDate(day.getDate() + 1);
  const ymd = `${day.getFullYear()}${String(day.getMonth() + 1).padStart(2, '0')}${String(day.getDate()).padStart(2, '0')}`;
  const ics = `BEGIN:VCALENDAR\r\nVERSION:2.0\r\nBEGIN:VEVENT\r\nUID:1\r\nSUMMARY:Biotonne\r\nDTSTART;VALUE=DATE:${ymd}\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n`;
  await waste
    .locator('input[type="file"]')
    .setInputFiles({ name: 'abfall.ics', mimeType: 'text/calendar', buffer: Buffer.from(ics) });
  await expect(waste.getByRole('status')).toContainText('File checked and saved.');
  await waste.getByRole('button', { name: 'Done' }).click();
  await expect(page.getByRole('dialog')).toHaveCount(0);
  await expect(page.getByText('Biotonne')).toBeVisible();
  // The other specs expect an unconfigured waste calendar.
  await page.request.delete('/api/waste/calendar', { headers });

  // An integration with credentials: numbered values with a hint each, the
  // connection test and Later.
  await page.getByRole('button', { name: 'Connect', exact: true }).click();
  const immich = page.getByRole('dialog', { name: 'Connect Immich' });
  await expect(immich.getByText('Waylight shows pictures from your Immich server.')).toBeVisible();
  await expect(immich.getByText('Type it the way you open Immich in the browser.')).toBeVisible();
  await expect(immich.getByText('Choose Account Settings')).toBeVisible();
  await expect(immich.getByRole('button', { name: 'Enter on phone' })).toBeVisible();
  await expectAccessible(page, 'Immich setup', 'dialog');

  await immich.getByLabel('Address').fill('http://127.0.0.1:9');
  await immich.getByLabel('Access key').fill('not-a-key');
  await immich.getByRole('button', { name: 'Connect' }).click();
  await expect(immich.getByRole('alert')).toBeVisible();
  // The dialog stays open to fix the value; Later leaves without saving.
  await immich.getByRole('button', { name: 'Later' }).click();
  await expect(page.getByRole('dialog')).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Connect', exact: true })).toBeVisible();
});
