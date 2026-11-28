import AxeBuilder from '@axe-core/playwright';
import { expect, type Page } from '@playwright/test';

/** WCAG 2.0–2.2 A and AA rules (#113). */
const tags = ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'];

/**
 * Runs axe on the page (or a part of it) and fails with a readable list of
 * violations: rule, impact and the offending elements.
 */
export async function expectAccessible(page: Page, view: string, include?: string): Promise<void> {
  // A hover would fade elements in or out while axe measures their contrast.
  await page.mouse.move(0, 0);
  let builder = new AxeBuilder({ page }).withTags(tags);
  if (include) builder = builder.include(include);
  const { violations } = await builder.analyze();
  const report = violations.map(
    // At most three elements per rule keep the report readable.
    (v) =>
      `${v.id} (${v.impact}): ${v.help}\n${v.nodes
        .slice(0, 3)
        .map((n) => `    ${n.target.join(' ')} ${n.failureSummary?.split('\n')[1] ?? ''}`)
        .join('\n')}`,
  );
  // Soft: one run lists the violations of every view.
  expect.soft(report, `accessibility violations in ${view}`).toEqual([]);
  expect.soft(await unnamedButtons(page, include), `buttons without an accessible name in ${view}`).toEqual([]);
}

/**
 * Lists the buttons in the page (or a part of it) that have no accessible
 * name (#857). Unlike axe this includes buttons that only fade in on hover
 * or focus, such as a widget's help button: they are faded out (opacity 0)
 * while axe measures, but a screen reader and the keyboard still reach them.
 * Buttons that are not rendered at all (display: none) are skipped.
 */
export function unnamedButtons(page: Page, include = 'body'): Promise<string[]> {
  return page.evaluate((scope) => {
    // The visible text of an element without its aria-hidden parts.
    const text = (el: Element): string =>
      [...el.childNodes]
        .map((n) => {
          if (n.nodeType === Node.TEXT_NODE) return n.textContent ?? '';
          if (!(n instanceof Element) || n.getAttribute('aria-hidden') === 'true') return '';
          if (n.tagName === 'IMG') return n.getAttribute('alt') ?? '';
          if (n.tagName.toLowerCase() === 'svg') return n.querySelector('title')?.textContent ?? '';
          return text(n);
        })
        .join('');
    const name = (el: Element): string => {
      const ids = el.getAttribute('aria-labelledby')?.split(/\s+/) ?? [];
      const labelled = ids.map((id) => document.getElementById(id)?.textContent ?? '').join(' ');
      return [labelled, el.getAttribute('aria-label') ?? '', text(el), el.getAttribute('title') ?? '']
        .map((s) => s.trim())
        .find((s) => s !== '') ?? '';
    };
    return [...document.querySelectorAll(scope)]
      .flatMap((root) => [...root.querySelectorAll('button, [role="button"]')])
      .filter((el) => el instanceof HTMLElement && el.getClientRects().length > 0 && name(el) === '')
      .map((el) => el.outerHTML.slice(0, 200));
  }, include);
}
