import { normalise } from './settingsSearch';

/**
 * Jumping to a field found by the settings search (#601). Sections are
 * written by many features, so instead of asking each for anchors the
 * field is found by the label the search showed: a heading, legend or
 * label whose text starts with it.
 */

/** Whether an element's text stands for the label (it may add a hint or a value after it). */
export function textMatchesLabel(text: string, label: string): boolean {
  const t = normalise(text);
  const l = normalise(label);
  return l !== '' && (t === l || t.startsWith(`${l} `) || t.startsWith(`${l}:`) || t.startsWith(`${l}(`));
}

const CANDIDATES = 'h3, h4, legend, label, summary';
/** How long the field stays highlighted; long enough to find it, short enough not to distract. */
const HIGHLIGHT_MS = 2000;
/** Sections load lazily; give a slow one this long to render the field. */
const WAIT_MS = 3000;

function find(root: HTMLElement, label: string): HTMLElement | null {
  for (const el of root.querySelectorAll<HTMLElement>(CANDIDATES)) {
    if (textMatchesLabel(el.textContent ?? '', label)) return el;
  }
  return null;
}

/** The control to focus for a found label: its input, else the label itself. */
function focusTarget(el: HTMLElement): HTMLElement {
  if (el instanceof HTMLLabelElement && el.control instanceof HTMLElement) return el.control;
  if (!el.hasAttribute('tabindex')) el.setAttribute('tabindex', '-1');
  return el;
}

/**
 * Scrolls to the field, focuses it and highlights it briefly. Resolves
 * false when the field did not appear, e.g. because the viewer's role
 * hides that part of the section.
 */
export function highlightField(root: HTMLElement, label: string): Promise<boolean> {
  return new Promise((resolve) => {
    const show = (el: HTMLElement) => {
      el.scrollIntoView({ block: 'center' });
      focusTarget(el).focus({ preventScroll: true });
      el.classList.add('settings-found');
      setTimeout(() => el.classList.remove('settings-found'), HIGHLIGHT_MS);
      resolve(true);
    };
    const now = find(root, label);
    if (now) return show(now);
    const observer = new MutationObserver(() => {
      const el = find(root, label);
      if (!el) return;
      observer.disconnect();
      clearTimeout(timer);
      show(el);
    });
    const timer = setTimeout(() => {
      observer.disconnect();
      resolve(false);
    }, WAIT_MS);
    observer.observe(root, { childList: true, subtree: true });
  });
}
