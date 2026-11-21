import { createServer } from 'node:http';
import type { AddressInfo } from 'node:net';
import { expect, test } from '@playwright/test';
import { account, password, openSettings } from '../support';

test('notes are summarised by a local Ollama once it is set up (#75)', async ({ page }) => {
  page.on('pageerror', (e) => console.log('PAGEERROR', e.message.slice(0, 300)));
  page.on('framenavigated', () => console.log('NAV', page.url()));
  // A tiny stand-in for Ollama on the loopback interface.
  const prompts: string[] = [];
  const ollama = createServer((req, res) => {
    let body = '';
    req.on('data', (c) => (body += c));
    req.on('end', () => {
      res.setHeader('Content-Type', 'application/json');
      if (req.url === '/api/tags') return res.end(JSON.stringify({ models: [{ name: 'tiny:1b' }] }));
      const request = JSON.parse(body);
      prompts.push(request.prompt);
      if (request.stream) {
        // Streamed answer of the search question (#162), one JSON object per line.
        res.write(JSON.stringify({ response: 'Jellyfin war ', done: false }) + '\n');
        return res.end(JSON.stringify({ response: 'nicht erreichbar.', done: true }) + '\n');
      }
      res.end(JSON.stringify({ response: '- Radtour an den See am Samstag', done: true }));
    });
  });
  await new Promise<void>((resolve) => ollama.listen(0, '127.0.0.1', resolve));
  const port = (ollama.address() as AddressInfo).port;

  await page.goto('/');
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();
  await expect(page.getByTestId('dashboard')).toBeVisible();

  await openSettings(page, 'Connections & services', 'AI summaries');
  const url = page.getByLabel('Ollama address (home network)');
  await url.fill('https://api.example.com');
  await page.getByRole('button', { name: 'Load models' }).click();
  await expect(page.getByRole('alert')).toContainText('home network');
  await url.fill(`http://127.0.0.1:${port}`);
  await page.getByRole('button', { name: 'Load models' }).click();
  await expect(page.getByRole('combobox', { name: 'Model' })).toHaveValue('tiny:1b');
  await page.getByRole('button', { name: 'Save', exact: true }).click();
  await expect(page.getByRole('status').filter({ hasText: 'Connected and saved' })).toBeVisible();
  await page.getByRole('button', { name: 'Close' }).click();

  await page.reload(); // the notes widget asks once whether summaries are available
  const field = page.getByRole('combobox', { name: 'Quick search (Ctrl+K or /)' });
  await field.fill('n Ausflug mit dem Fahrrad an den See');
  await field.press('Enter');
  await expect(page.getByText('Note “Ausflug mit dem Fahrrad an den See” created')).toBeVisible();
  await page.keyboard.press('Escape');
  await account(page, 'Edit');
  await page.getByRole('button', { name: '+ Tab' }).click();
  await page.getByRole('textbox', { name: 'Name' }).fill('Notizen');
  await page.getByRole('button', { name: 'Create' }).click();
  await page.getByRole('button', { name: '+ Add widget' }).click();
  await page.getByRole('searchbox', { name: 'Search widgets…' }).fill('notes');
  await page
    .getByRole('button', { name: /^Notes/ })
    .first()
    .click();
  const notesSettings = page.getByRole('dialog', { name: 'Notes settings' });
  await notesSettings.getByRole('radio', { name: 'All notes' }).check();
  await notesSettings.getByRole('button', { name: 'Save' }).click();
  await account(page, 'Done');
  await page
    .getByRole('button', { name: /Ausflug mit dem Fahrrad/ })
    .first()
    .click();
  await page.getByRole('dialog').getByRole('button', { name: 'Summarise' }).click();
  await expect(page.getByRole('note', { name: 'Summary' })).toContainText('Radtour an den See');
  expect(prompts[0]).toContain('Ausflug mit dem Fahrrad an den See');
  await page.keyboard.press('Escape');

  // Ask the AI in the search (#162): only after "?", only on request, marked as AI-generated.
  await field.fill('? Warum war Jellyfin offline');
  await page.getByRole('button', { name: 'Ask the AI' }).click();
  await expect(page.getByText('Jellyfin war nicht erreichbar.')).toBeVisible();
  await expect(page.getByText('AI-generated answer')).toBeVisible();
  expect(prompts[1]).toContain('Question: Warum war Jellyfin offline');
  expect(prompts[1]).toMatch(/now: \d{4}-\d\d-\d\d/);
  await field.fill('');
  ollama.close();

  // Switched off again so later tests see the default.
  await page.keyboard.press('Escape');
  const status = await (await page.request.get('/api/auth/status')).json();
  await page.request.delete('/api/summary/settings', {
    headers: { 'X-CSRF-Token': status.csrfToken as string },
  });
});
