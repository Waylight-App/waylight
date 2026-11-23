import { expect, test } from '@playwright/test';
import { password, setupToken } from './support';

// The first start with the default SAMPLE_DATA=true, on a server of its own
// (see playwright.config.ts): the onboarding wizard (#739) asks the profile
// and the widgets, then the tour follows. A template replaces the example
// tabs with a zoned tab.
test.use({ geolocation: { latitude: 53.14, longitude: 8.21 }, permissions: ['geolocation'] });

test('first start asks the essentials, offers the templates and Family replaces the example dashboard', async ({
  page,
}) => {
  await page.goto('/');
  await page.getByLabel('Setup token').fill(setupToken);
  await page.getByLabel('Password (at least 12 characters)').fill(password);
  await page.getByLabel('Repeat password').fill(password);
  await page.getByRole('button', { name: 'Create account' }).click();

  // The dashboard stays hidden until the wizard is over.
  const wizard = page.getByRole('region', { name: 'Set up Waylight' });
  await expect(wizard.getByRole('heading', { name: 'Glad you are here.' })).toBeVisible();
  await expect(page.getByTestId('dashboard')).toHaveCount(0);
  await expect(wizard.getByRole('progressbar')).toHaveAttribute('aria-valuetext', 'Step 1 of 6');
  await wizard.getByRole('button', { name: "Let's go" }).click();

  await expect(wizard.getByRole('heading', { name: 'What should we call you?' })).toBeVisible();
  await wizard.getByLabel('Name').fill('Sam');
  // Picking a language shows it at once.
  await wizard.getByRole('radio', { name: 'Deutsch' }).check();
  await expect(page.getByRole('heading', { name: 'Wie sollen wir dich nennen?' })).toBeVisible();
  await page.getByRole('radio', { name: 'English' }).check();
  await wizard.getByRole('button', { name: 'Use my location' }).click();
  await expect(wizard.getByRole('status')).toHaveText('Chosen: My location');
  await wizard.getByRole('button', { name: 'Next' }).click();

  // The look step applies a design with one click on its card.
  await expect(wizard.getByRole('heading', { name: 'Pick a look' })).toBeVisible();
  const nord = wizard.getByRole('button', { name: 'Apply theme Nord' });
  await nord.click();
  await expect(nord).toHaveAttribute('aria-pressed', 'true');
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
  await wizard.getByRole('button', { name: 'Next' }).click();

  // Back and forth keeps the step; a reload continues where it stopped.
  await expect(wizard.getByRole('heading', { name: 'What goes on your start page?' })).toBeVisible();
  await page.reload();
  await expect(wizard.getByRole('heading', { name: 'What goes on your start page?' })).toBeVisible();

  await expect(wizard.getByRole('button', { name: 'Family' })).toHaveAttribute('aria-pressed', 'true');
  // The suggestions follow the template: its core is chosen, a few more are offered.
  await expect(wizard.getByRole('heading', { name: 'Suggestions for Family' })).toBeVisible();
  await expect(wizard.getByRole('button', { name: 'Calendar', exact: true })).toHaveAttribute(
    'aria-pressed',
    'true',
  );
  // A tile can be switched off and on again.
  const calendar = wizard.getByRole('button', { name: 'Calendar', exact: true });
  await calendar.click();
  await expect(calendar).toHaveAttribute('aria-pressed', 'false');
  await calendar.click();
  await expect(calendar).toHaveAttribute('aria-pressed', 'true');
  await expect(wizard.getByText('You can change everything anytime')).toBeVisible();
  // More opens the add-widget dialog, which marks what is chosen already.
  await wizard.getByRole('button', { name: 'More' }).click();
  const picker = page.getByRole('dialog', { name: 'Add widget' });
  await expect(picker.getByRole('button', { name: /^Calendar/ })).toContainText('Already added');
  await expect(picker.getByRole('button', { name: /^Services/ })).not.toContainText('Already added');
  await picker.getByRole('button', { name: /^Services/ }).click();
  await expect(picker.getByRole('button', { name: /^Services/ })).toContainText('Already added');
  await picker.getByRole('button', { name: 'Done' }).click();
  await expect(wizard.getByRole('button', { name: 'Services', exact: true })).toHaveAttribute(
    'aria-pressed',
    'true',
  );
  await wizard.getByRole('button', { name: 'Next' }).click();

  // The search step answers with the real search; an example types itself in.
  await expect(wizard.getByRole('heading', { name: 'Just ask.' })).toBeVisible();
  const field = wizard.getByRole('textbox', { name: 'Try a question' });
  // Without a set-up widget the example shows a note instead of an answer card.
  const answerOrNote = wizard.getByRole('article').or(wizard.getByText('Once the widget is set up'));
  // The examples follow the template chosen before: Family starts with the agenda.
  await expect(field).toHaveValue('What is on today?');
  await expect(answerOrNote.first()).toBeVisible();
  await wizard.getByRole('button', { name: 'Who has a birthday soon?' }).click();
  await expect(field).toHaveValue('Who has a birthday soon?');
  await expect(answerOrNote.first()).toBeVisible();
  await field.fill('3 * 17 + 4');
  await expect(wizard.getByRole('article', { name: '3 * 17 + 4 =' })).toContainText('55');
  await wizard.getByRole('button', { name: 'Next' }).click();

  await expect(wizard.getByRole('heading', { name: "All done, let's go." })).toBeVisible();
  await expect(page.getByTestId('dashboard')).toHaveCount(0);
  await wizard.getByRole('button', { name: 'Go to the start page' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  const tour = page.getByRole('dialog', { name: 'Welcome to Waylight 👋' });
  await expect(tour).toBeVisible();
  await tour.getByRole('button', { name: 'Skip tour' }).click();
  await expect(page.getByRole('dialog')).toHaveCount(0);
  await expect(page.locator('.hero h1')).toHaveText(/, Sam$/);

  // Only the template's tab is left (a single tab shows no tab bar), laid
  // out in zones; the owner sees System.
  await expect(page.getByRole('button', { name: 'Server', exact: true })).toHaveCount(0);
  for (const [zone, heading] of [
    ['today', 'Today'],
    ['mine', 'My things'],
    ['system', 'System'],
  ] as const) {
    await expect(
      page.locator(`section[data-zone="${zone}"]`).getByRole('heading', { name: heading }),
    ).toBeVisible();
  }

  const dashboard = (await (await page.request.get('/api/dashboard')).json()) as {
    tabs: {
      name: string;
      zoned: boolean;
      sample?: boolean;
      widgets: { kind: string; size?: string; zone?: string }[];
    }[];
  };
  expect(dashboard.tabs.map((t) => t.name)).toEqual(['Family']);
  expect(dashboard.tabs[0]?.zoned).toBe(true);
  expect(dashboard.tabs[0]?.widgets.map((w) => w.kind).sort()).toEqual(
    ['calendar', 'meals', 'services', 'shopping', 'todos'].sort(),
  );

  // The answers are stored where the settings pages store them.
  const stored = (await (await page.request.get('/api/appearance')).json()) as Record<string, unknown>;
  expect(stored).toMatchObject({ greetingName: 'Sam', language: 'en' });
  const places = (await (await page.request.get('/api/places')).json()) as { name: string; lat: number }[];
  expect(places).toEqual([expect.objectContaining({ name: 'Home', lat: 53.14 })]);
  // The example data of the modules stays.
  expect(((await (await page.request.get('/api/todos')).json()) as unknown[]).length).toBeGreaterThan(0);
});
