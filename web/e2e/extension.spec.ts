import { chromium, expect, test } from '@playwright/test';
import { existsSync, mkdtempSync, readFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';

// The browser extension in local mode (ADR 0019, #291), loaded into a real
// Chromium. It needs no server: everything happens inside the extension. That
// the exported file imports into a server is checked by the Go test
// TestLocalModeExportImports, so this spec leaves the shared test account alone.
const extension = resolve(import.meta.dirname, '../../extension');

test.skip(
  !existsSync(join(extension, 'app/extension.html')),
  'the local-mode app is not built: npm run build:extension, then node scripts/app.mjs in extension/',
);

test('the extension works without a server and exports what was added (#291)', async () => {
  const profile = mkdtempSync(join(tmpdir(), 'trisse-home-extension-'));
  // Extensions only load in Chromium's own build, not in branded Chrome.
  const context = await chromium.launchPersistentContext(profile, {
    channel: 'chromium',
    args: [`--disable-extensions-except=${extension}`, `--load-extension=${extension}`],
  });
  try {
    const worker = context.serviceWorkers()[0] ?? (await context.waitForEvent('serviceworker'));
    const id = new URL(worker.url()).host;

    // A fresh install offers the choice instead of asking for an address.
    const page = await context.newPage();
    await page.goto(`chrome-extension://${id}/newtab.html`);
    await page.getByRole('button', { name: 'Use without a server' }).click();

    // The bundled app starts with the Home tab, without sign-in.
    await expect(page.getByTestId('dashboard')).toBeVisible();
    const input = page.getByRole('textbox', { name: 'New to-do' }).first();
    await input.fill('Buy stamps');
    await input.press('Enter');
    await expect(page.getByText('Buy stamps')).toBeVisible();

    // Kept in the browser: a new tab opens the local app and shows it again.
    const again = await context.newPage();
    await again.goto(`chrome-extension://${id}/newtab.html`);
    await expect(again.getByTestId('dashboard')).toBeVisible();
    await expect(again.getByText('Buy stamps')).toBeVisible();
    // Server-only modules are hidden rather than broken.
    await expect(again.getByText('Docker', { exact: true })).toHaveCount(0);

    // Moving to a server: the options page downloads the export file.
    const options = await context.newPage();
    await options.goto(`chrome-extension://${id}/options.html`);
    const download = options.waitForEvent('download');
    await options.getByRole('link', { name: 'Download my data' }).click();
    const file = JSON.parse(readFileSync((await (await download).path()) ?? '', 'utf8')) as {
      format: string;
      sections: Record<string, Record<string, { title?: string }[]>>;
    };
    expect(file.format).toBe('trisse-home-export');
    expect(file.sections['todos']?.['todos']?.map((t) => t.title)).toContain('Buy stamps');
    expect(file.sections['dashboard']?.['dashboard_tabs']?.length).toBeGreaterThan(0);
  } finally {
    await context.close();
    rmSync(profile, { recursive: true, force: true });
  }
});
