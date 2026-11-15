// Lints the extension without dependencies: manifest shape, minimal
// permissions, every referenced file present, JavaScript syntax (#65).
import { readFileSync, existsSync } from 'node:fs';
import { execFileSync } from 'node:child_process';
import { appFiles, files, ownFiles } from './files.mjs';

const fail = (msg) => {
  console.error(`✗ ${msg}`);
  process.exitCode = 1;
};

const manifest = JSON.parse(readFileSync('manifest.json', 'utf8'));
if (manifest.manifest_version !== 3) fail('manifest_version must be 3');
if (!/^\d+\.\d+\.\d+$/.test(manifest.version)) fail('version must be x.y.z');
const allowed = new Set(['storage', 'contextMenus', 'activeTab']);
for (const p of manifest.permissions ?? []) if (!allowed.has(p)) fail(`permission not allowed: ${p}`);
for (const key of ['host_permissions', 'optional_host_permissions', 'content_scripts', 'web_accessible_resources']) {
  if (manifest[key]) fail(`${key} is not allowed: the extension only opens the configured address`);
}
// Firefox requires new add-ons to declare what they transmit (#290). Saving a
// page sends its URL and title to the user's own server, so the honest
// declaration is browsingActivity, never "none".
const dcp = manifest.browser_specific_settings?.gecko?.data_collection_permissions;
if (!dcp || !Array.isArray(dcp.required) || !dcp.required.includes('browsingActivity')) {
  fail('browser_specific_settings.gecko.data_collection_permissions must declare browsingActivity');
}
const referenced = [
  manifest.chrome_url_overrides?.newtab,
  manifest.options_ui?.page,
  manifest.background?.service_worker,
  ...Object.values(manifest.icons ?? {}),
].filter(Boolean);
for (const f of [...referenced, ...files]) if (!existsSync(f)) fail(`missing file: ${f}`);
// The app's chunks are ES modules built by Vite and type-checked in web/;
// node --check would parse them as scripts, so it covers the own files only.
for (const f of ownFiles.filter((f) => f.endsWith('.js'))) {
  try {
    execFileSync(process.execPath, ['--check', f], { stdio: 'pipe' });
  } catch (err) {
    fail(`${f}: ${String(err.stderr).trim()}`);
  }
}
// Remote code is not allowed in MV3 and not wanted here.
for (const f of files.filter((f) => f.endsWith('.html'))) {
  if (/<script[^>]+src=["']https?:/i.test(readFileSync(f, 'utf8'))) fail(`${f} loads a remote script`);
}
// The local-mode app must not load code from anywhere but the package, and
// must not use eval, which the extension's content security policy forbids.
if (appFiles.length && !appFiles.includes('app/extension.html')) fail('app/ has no extension.html');
for (const f of appFiles.filter((f) => f.endsWith('.js'))) {
  const code = readFileSync(f, 'utf8');
  if (/\bimport\s*\(\s*["']https?:/.test(code) || /from\s*["']https?:/.test(code)) fail(`${f} imports remote code`);
  if (/\beval\s*\(|new Function\s*\(/.test(code)) fail(`${f} uses eval or new Function`);
}
if (!process.exitCode)
  console.log(`✓ manifest and ${files.length} files OK (${appFiles.length} of them the local-mode app)`);
