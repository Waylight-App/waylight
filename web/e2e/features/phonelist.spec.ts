import { expect, test } from '@playwright/test';
import { account, password } from '../support';

test('the phone list keeps people, finds them and answers in the search field (#300)', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Phones');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('telefon');
  await page.getByRole('button', { name: /^Phone list/ }).click();
  await account(page, 'Done');

  // The owner is an admin and maintains the list.
  await expect(page.getByText('The phone list is empty. Add people with ＋.')).toBeVisible();
  await page.getByRole('button', { name: 'Add a person' }).click();
  await page.getByRole('textbox', { name: 'Name', exact: true }).fill('Anna Weber');
  await page.getByRole('textbox', { name: 'Role' }).fill('Reception');
  await page.getByRole('textbox', { name: 'Team' }).fill('Front desk');
  await page.getByRole('textbox', { name: 'Phone' }).fill('+49 30 1234-56');
  await page.getByRole('textbox', { name: 'E-mail' }).fill('anna@example.org');
  await page.getByRole('button', { name: 'Save', exact: true }).click();

  // The number is a tel: link that dials, the address a mailto: link.
  const anna = page.getByRole('listitem').filter({ hasText: 'Anna Weber' });
  await expect(anna.getByRole('link', { name: '+49 30 1234-56' })).toHaveAttribute('href', /^tel:\+49/);
  await expect(anna.getByRole('link', { name: 'anna@example.org' })).toHaveAttribute(
    'href',
    'mailto:anna@example.org',
  );

  // The widget's own search covers name, role and team.
  const filter = page.getByRole('searchbox', { name: 'Search name, role or team' });
  await filter.fill('recep');
  await expect(anna).toBeVisible();
  await filter.fill('nobody here');
  await expect(page.getByText('Nobody matches.')).toBeVisible();
  await filter.fill('');

  // "phone anna" in the search field answers with the number and the address.
  await page.keyboard.press('Control+k');
  const box = page.getByRole('combobox', { name: 'Quick search (Ctrl+K or /)' });
  const answers = page.getByRole('region', { name: 'Answers' });
  await box.fill('phone anna');
  await expect(answers.getByRole('article', { name: 'Anna Weber · Reception' })).toContainText(
    '+49 30 1234-56 · anna@example.org',
  );
  await page.keyboard.press('Escape');

  // Leave the shared account as it was.
  page.once('dialog', (d) => d.accept());
  await page.getByRole('button', { name: 'Edit Anna Weber' }).click();
  await page.getByRole('button', { name: 'Delete', exact: true }).click();
  await expect(page.getByText('The phone list is empty. Add people with ＋.')).toBeVisible();
});
