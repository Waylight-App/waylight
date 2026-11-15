// Copies the app built for the local mode (web: npm run build:extension,
// ADR 0019) into app/, where files.mjs picks it up for checking and packing.
// Run from the extension directory after building the web app.
import { cpSync, existsSync, rmSync } from 'node:fs';

const from = '../web/dist-extension';
if (!existsSync(`${from}/extension.html`)) {
  console.error(`✗ ${from}/extension.html is missing: run "npm run build:extension" in web/ first`);
  process.exit(1);
}
rmSync('app', { recursive: true, force: true });
cpSync(from, 'app', { recursive: true });
console.log('✓ copied the local-mode app into app/');
