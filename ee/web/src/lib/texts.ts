import type { Texts } from "../../../../web/src/app/plugins/api";

/**
 * Texts of the plugin in English and German. They stay in the plugin, so the
 * core's catalogues never list ee/ features (ADR 0021).
 */
export type Catalogue = Record<string, Texts>;

/** Looks up a text for the active language with {name} placeholders. */
export function text(
  catalogue: Catalogue,
  key: string,
  language: string,
  params: Record<string, string | number> = {},
): string {
  const entry = catalogue[key];
  const raw = entry ? (entry[language] ?? entry.en) : key;
  return raw.replace(/\{(\w+)\}/g, (m, name: string) =>
    name in params ? String(params[name]) : m,
  );
}
