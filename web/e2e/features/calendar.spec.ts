import { expect, test } from '@playwright/test';
import { account, boxOf, expectBelow, expectBeside, password } from '../support';

test.describe('calendar split view', () => {
  // The service worker would answer the API itself; page.route only sees requests without it.
  test.use({ serviceWorkers: 'block' });

  test('the calendar split view selects days and changes months, wide and narrow (#163)', async ({
    page,
  }) => {
    const day = (offsetMonths: number, date: number, hour?: number) => {
      const now = new Date();
      return new Date(now.getFullYear(), now.getMonth() + offsetMonths, date, hour ?? 0);
    };
    const ymd = (d: Date) =>
      `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
    const today = new Date();
    const standupStart = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 23, 0);
    const events = [
      {
        uid: 'standup',
        summary: 'Standup',
        location: 'Room 1',
        description: 'Line one\nLine two',
        url: 'https://example.org/meet',
        calendar: 'work',
        color: '#e5484d',
        start: standupStart.toISOString(),
        end: new Date(standupStart.getTime() + 59 * 60_000).toISOString(),
        allDay: false,
      },
      {
        uid: 'party',
        summary: 'Party next month',
        location: '',
        description: '',
        url: '',
        calendar: 'home',
        color: '',
        start: ymd(day(1, 15)),
        end: ymd(day(1, 16)),
        allDay: true,
      },
    ];
    const ranges: string[] = [];
    await page.route(
      (url) => url.pathname === '/api/calendar/events' && url.searchParams.has('from'),
      async (route) => {
        const url = new URL(route.request().url());
        ranges.push(url.searchParams.get('from') ?? '');
        const from = new Date(url.searchParams.get('from') ?? '');
        const to = new Date(url.searchParams.get('to') ?? '');
        const inRange = events.filter((e) => new Date(e.start) < to && new Date(e.end) > from);
        await route.fulfill({ json: { events: inRange, updatedAt: new Date().toISOString(), error: '' } });
      },
    );

    await page.goto('/');
    await page.getByLabel('Password').fill(password);
    await page.getByRole('button', { name: 'Log in' }).click();
    await expect(page.getByTestId('dashboard')).toBeVisible();
    await account(page, 'Edit');
    await page.getByRole('button', { name: '+ Tab' }).click();
    await page.getByRole('textbox', { name: 'Name' }).fill('Agenda');
    await page.getByRole('button', { name: 'Create' }).click();
    await page.getByRole('button', { name: '+ Add widget' }).click();
    await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('termine');
    await expect(page.getByRole('status')).toHaveText('1 widgets');
    await page.getByRole('button', { name: /^Calendar/ }).click();
    // New calendars start in the split view.
    await expect(page.getByRole('combobox', { name: 'View' })).toHaveValue('split');
    await page.getByRole('button', { name: 'Save', exact: true }).click();
    await account(page, 'Done');

    const monthName = (d: Date) => d.toLocaleDateString('en', { month: 'long', year: 'numeric' });
    const grid = page.getByRole('grid', { name: monthName(today) });
    await expect(grid).toBeVisible();
    const upcomingList = page.getByRole('region', { name: 'Upcoming' });
    await expect(upcomingList.getByRole('button', { name: /Standup/ })).toBeVisible();
    // Wide: grid and events side by side.
    await expectBeside(grid, upcomingList);

    // Clicking a day shows its events; the details show the description as text and the link.
    await grid.locator('[aria-current="date"]').click();
    const dayList = page.getByRole('region', {
      name: today.toLocaleDateString('en', { weekday: 'long', day: 'numeric', month: 'long' }),
    });
    await dayList.getByRole('button', { name: /Standup/ }).click();
    const details = page.getByRole('dialog', { name: 'Event details' });
    await expect(details).toContainText('Line one');
    await expect(details.getByRole('link', { name: 'Open link' })).toHaveAttribute(
      'href',
      'https://example.org/meet',
    );
    await page.keyboard.press('Escape');
    await expect(details).toHaveCount(0);

    // Keyboard: PageDown moves to the next month, the arrows to the 15th with the party.
    await grid.locator('[aria-current="date"]').focus();
    const target = day(1, 15);
    const sameDayNextMonth = new Date(
      today.getFullYear(),
      today.getMonth() + 1,
      Math.min(today.getDate(), new Date(today.getFullYear(), today.getMonth() + 2, 0).getDate()),
    );
    await page.keyboard.press('PageDown');
    const next = page.getByRole('grid', { name: monthName(target) });
    await expect(next).toBeVisible();
    const steps = target.getDate() - sameDayNextMonth.getDate();
    for (let i = 0; i < Math.abs(steps); i++)
      await page.keyboard.press(steps > 0 ? 'ArrowRight' : 'ArrowLeft');
    await expect(next.locator(`[data-key="${ymd(target)}"]`)).toBeFocused();
    await expect(next.locator(`[data-key="${ymd(target)}"]`)).toHaveAttribute('aria-label', /1 event$/);
    await expect(
      page.getByRole('region', { name: /15/ }).getByRole('button', { name: /Party next month/ }),
    ).toBeVisible();

    // Back and forth reads the per-month cache: no new request for a month already shown.
    await page.getByRole('button', { name: 'Today', exact: true }).click();
    await expect(page.getByRole('grid', { name: monthName(today) })).toBeVisible();
    const loaded = ranges.length;
    await page.getByRole('button', { name: 'Next month' }).click();
    await expect(next).toBeVisible();
    await page.getByRole('button', { name: 'Previous month' }).click();
    await expect(page.getByRole('grid', { name: monthName(today) })).toBeVisible();
    expect(ranges.length).toBe(loaded + 1); // only the prefetch of the month before

    // Narrow (phone): the list sits under the grid, without horizontal scrolling.
    await page.setViewportSize({ width: 390, height: 900 });
    await expect(page.getByRole('grid', { name: monthName(today) })).toBeVisible();
    await page
      .getByRole('grid', { name: monthName(today) })
      .locator('[aria-current="date"]')
      .click();
    await expect(dayList.getByRole('button', { name: /Standup/ })).toBeVisible();
    await expectBelow(page.getByRole('grid', { name: monthName(today) }), dayList);
    // No horizontal scrolling inside the widget (the page header has its own issue, #226).
    const widget = page
      .getByRole('grid', { name: monthName(today) })
      .locator('xpath=ancestor::*[contains(@class,"content")][1]');
    const overflow = await widget.evaluate((e) => e.scrollWidth - e.clientWidth);
    expect(overflow).toBeLessThanOrEqual(0);
    const right = (await boxOf(dayList)).x + (await boxOf(dayList)).width;
    expect(right).toBeLessThanOrEqual(390);
  });
});
