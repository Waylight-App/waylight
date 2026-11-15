/**
 * Helpers shared by the end-to-end tests (#235). The tests run in projects
 * (see playwright.config.ts): setup creates the account, the general and
 * feature specs log in with it, and two-factor runs last because it turns
 * on 2FA for that account.
 */
import { createHmac } from 'node:crypto';
import { expect, type Locator, type Page } from '@playwright/test';

export { setupToken } from '../playwright.config';

/** Password of the account created by setup.spec.ts. */
export const password = 'smoke-test-password-123';

/** Edit, Done, Settings and Log out live in the account menu of the top bar (#189). */
export async function account(page: Page, item: string) {
  await page.getByRole('button', { name: 'Account menu', exact: true }).click();
  await page.getByRole('menuitem', { name: item, exact: true }).click();
}

/** RFC 6238 TOTP for the test (SHA1, 6 digits, 30 s). */
export function totp(base32Secret: string, now = Date.now()): string {
  const alphabet = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567';
  let bits = '';
  for (const ch of base32Secret.replace(/\s/g, '').toUpperCase())
    bits += alphabet.indexOf(ch).toString(2).padStart(5, '0');
  const key = Buffer.from((bits.match(/.{8}/g) ?? []).map((b) => parseInt(b, 2)));
  const counter = Buffer.alloc(8);
  counter.writeBigUInt64BE(BigInt(Math.floor(now / 1000 / 30)));
  const mac = createHmac('sha1', key).update(counter).digest();
  const offset = (mac[mac.length - 1] ?? 0) & 0x0f;
  return String((mac.readUInt32BE(offset) & 0x7fffffff) % 1_000_000).padStart(6, '0');
}

export async function boxOf(locator: Locator) {
  const box = await locator.boundingBox();
  if (!box) throw new Error('not visible');
  return box;
}
export async function expectBeside(left: Locator, right: Locator) {
  const [a, b] = [await boxOf(left), await boxOf(right)];
  expect(b.x).toBeGreaterThanOrEqual(a.x + a.width - 1);
}
export async function expectBelow(top: Locator, bottom: Locator) {
  const [a, b] = [await boxOf(top), await boxOf(bottom)];
  expect(b.y).toBeGreaterThanOrEqual(a.y + a.height - 1);
}
