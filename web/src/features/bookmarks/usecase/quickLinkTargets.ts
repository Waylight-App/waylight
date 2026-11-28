import { SHORTCUT_LINKS } from '../domain/quickLinks';

/**
 * The links the digits 1 to 9 open (#859). Each quick links widget on the
 * visible tab publishes its links here; the first one that registered wins,
 * so two widgets on a tab do not fight over the digits.
 */
export class QuickLinkTargets {
  private readonly owners = new Map<symbol, string[]>();

  /** @param open opens a URL (window.open in the app; injected for tests). */
  constructor(private readonly open: (url: string) => void) {}

  /** Publishes (or replaces) the links of a widget; returns the removal. */
  set(owner: symbol, urls: readonly string[]): () => void {
    this.owners.set(owner, urls.slice(0, SHORTCUT_LINKS));
    return () => this.owners.delete(owner);
  }

  /** Opens the link at the 0-based index; false when there is none. */
  openAt(index: number): boolean {
    const first = this.owners.values().next();
    const url = first.done ? undefined : first.value[index];
    if (!url) return false;
    this.open(url);
    return true;
  }
}
