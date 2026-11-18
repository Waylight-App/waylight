/** A single-use code that pairs the mobile app with this account (#559). */
export interface PairingCode {
  code: string;
  expiresAt: Date;
}

/**
 * The link the app reads from the QR code. It carries the address the
 * browser uses, which is the one the phone needs too, and the code; the
 * code is useless after five minutes or once redeemed.
 */
export function pairingLink(serverUrl: string, code: string): string {
  return `waylight://pair?${new URLSearchParams({ server: serverUrl, code })}`;
}
