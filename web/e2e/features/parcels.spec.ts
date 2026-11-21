import { expect, test } from '@playwright/test';
import { account, password, openSettings } from '../support';

test('parcels of other carriers link to their tracking page and the mailbox is tested before saving (#166)', async ({
  page,
}) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();
  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Pakete');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('parcels');
  await page
    .getByRole('button', { name: /^Parcels/ })
    .first()
    .click();
  await account(page, 'Done');

  await page.getByRole('button', { name: '+ Parcel' }).click();
  await page.getByLabel('Tracking number').fill('1234 5678 901');
  await page.getByLabel('Carrier').selectOption('gls');
  await page.getByRole('textbox', { name: 'Label' }).fill('Kaufland');
  await page.getByRole('button', { name: 'Track', exact: true }).click();
  const link = page.getByRole('link', { name: 'Track Kaufland at GLS' });
  await expect(link).toHaveAttribute(
    'href',
    'https://www.gls-pakete.de/sendungsverfolgung?match=12345678901',
  );
  await expect(page.getByText('No live status for GLS')).toBeVisible();

  await openSettings(page, 'Dashboard & widgets', 'Parcels');
  await expect(page.getByRole('heading', { name: 'Parcels from e-mails' })).toBeVisible();
  await page.getByLabel('Look for parcels in e-mails').check();
  await page.getByLabel('IMAP server').fill('localhost');
  await page.getByLabel('Port').fill('1');
  await page.getByLabel('User name').fill('me@example.com');
  await page.getByLabel(/^App password/).fill('app-pw');
  await page.getByLabel('Folder').fill('Pakete');
  await page.getByRole('button', { name: 'Save mailbox' }).click();
  await expect(page.getByRole('alert')).toContainText('could not be reached');
  await page.getByRole('button', { name: 'Close' }).click();
});
