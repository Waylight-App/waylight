// The files that make up the extension package.
import { existsSync, readdirSync, statSync } from "node:fs";
import { join } from "node:path";

/** The forwarding extension itself: small, hand-written, no build step. */
export const ownFiles = [
  "manifest.json",
  "newtab.html",
  "options.html",
  "style.css",
  "icons/icon-192.png",
  "src/background.js",
  "src/newtab.js",
  "src/options.js",
  "src/storage.js",
  "src/url.js",
];

/** Every file below dir, with forward slashes, sorted for a stable zip. */
function walk(dir) {
  return readdirSync(dir)
    .flatMap((entry) => {
      const path = join(dir, entry);
      return statSync(path).isDirectory()
        ? walk(path)
        : [path.split("\\").join("/")];
    })
    .sort();
}

/**
 * The web app for the local mode (ADR 0019), copied into app/ by app.mjs.
 * Empty when it has not been built, so the forwarding extension can still be
 * checked and packed on its own.
 */
export const appFiles = existsSync("app") ? walk("app") : [];

export const files = [...ownFiles, ...appFiles];
