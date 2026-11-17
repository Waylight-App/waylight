/** A sign-in at the OpenID Connect provider linked to this account (#527). */
export interface LinkedSignIn {
  /** The provider's id of the person; used to disconnect. */
  subject: string;
  /** Shown so people recognise the account; it decides nothing. */
  email: string;
  linkedAt: Date;
}

/** Result of returning from the provider after linking, from ?link=… */
export type LinkResult = 'ok' | 'taken' | 'licence';

/** Reads the link result the server put into the address, if any. */
export function linkResult(search: URLSearchParams): LinkResult | null {
  const v = search.get('link');
  return v === 'ok' || v === 'taken' || v === 'licence' ? v : null;
}
