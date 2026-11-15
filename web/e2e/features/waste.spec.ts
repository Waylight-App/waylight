import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('the waste calendar can be uploaded as an ICS file and removed (#153)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  const ymd = (offset: number) => {
    const d = new Date();
    d.setDate(d.getDate() + offset);
    return `${d.getFullYear()}${String(d.getMonth() + 1).padStart(2, '0')}${String(d.getDate()).padStart(2, '0')}`;
  };
  const event = (uid: string, summary: string, day: string) =>
    `BEGIN:VEVENT\r\nUID:${uid}\r\nSUMMARY:${summary}\r\nDTSTART;VALUE=DATE:${day}\r\nEND:VEVENT\r\n`;
  const ics = `BEGIN:VCALENDAR\r\nVERSION:2.0\r\n${event('1', 'Biotonne', ymd(1))}${event('2', 'Papier', ymd(9))}END:VCALENDAR\r\n`;

  await account(page, 'Settings');
  await page.getByRole('button', { name: 'Waste', exact: true }).click();
  const errors: string[] = [];
  page.on('pageerror', (e) => errors.push(e.message));
  // The radios are styled as a segmented control; the label is what gets clicked.
  await page.locator('.modes label', { hasText: 'Upload ICS file' }).click();
  await expect(page.getByRole('radio', { name: 'Upload ICS file' })).toBeChecked();
  const section = page.getByRole('region', { name: 'Upload ICS file' });
  // A wrong file is refused before anything is sent.
  await section
    .getByLabel('Choose .ics file…')
    .setInputFiles({ name: 'plan.pdf', mimeType: 'application/pdf', buffer: Buffer.from('%PDF') });
  await expect(page.getByText('Choose an .ics calendar file.')).toBeVisible();

  await section
    .getByLabel('Choose .ics file…')
    .setInputFiles({ name: 'abfall.ics', mimeType: 'text/calendar', buffer: Buffer.from(ics) });
  await expect(page.getByText('File checked and saved.')).toBeVisible();
  await expect(section).toContainText('abfall.ics');
  await expect(section).toContainText('2 collections');
  // The last date is 9 days away: the settings ask for next year's file.
  await expect(section.getByRole('status')).toContainText('Upload next year’s file');

  await section.getByRole('button', { name: 'Remove file' }).click();
  await expect(page.getByText('File removed.')).toBeVisible();
  await expect(section).not.toContainText('abfall.ics');
  // The other smoke tests expect an unconfigured waste calendar.
  const status = await (await page.request.get('/api/waste/settings')).json();
  expect(status.configured).toBe(false);
  expect(errors).toEqual([]);
});
