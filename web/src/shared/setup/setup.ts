/**
 * The setup assistant of a widget (#588): each integration describes its
 * setup as data (fields, credential steps, a deep link and a check) and the
 * shared component renders it, so a new integration adds data, not UI.
 */
import { getContext, setContext } from 'svelte';
import type { FullAutoFill } from 'svelte/elements';
import type { MessageKey } from '../i18n';
import type { SetupSignIn } from './signin';

/** Which step of the assistant a field belongs to. */
export type SetupStage = 'address' | 'credential';

/** One option of a choice field; the label is a product name, not translated. */
export interface SetupOption {
  value: string;
  label: string;
}

/** One input of the assistant. */
export interface SetupField {
  key: string;
  label: MessageKey;
  stage: SetupStage;
  /** A choice picks one of several products, e.g. Sonarr or Radarr. */
  type: 'url' | 'text' | 'password' | 'choice';
  /** The options of a choice; the first one is preselected. */
  options?: SetupOption[];
  /** An optional field may stay empty, e.g. a Pi-hole without a password. */
  optional?: boolean;
  placeholder?: string;
  autocomplete?: FullAutoFill;
}

/** What the check shows after a working connection: real data. */
export interface SetupPreview {
  /** Short lines such as "12,340 photos" or "Calendars: Personal, Work". */
  lines: string[];
  /** Up to three image URLs, e.g. thumbnails. */
  images?: string[];
}

/** An address found in the home network (#590). */
export interface FoundService {
  url: string;
  /** Where it was found, e.g. "Docker container immich_server". */
  hint: string;
}

/** Everything the assistant needs to know about one integration. */
export interface SetupDescriptor {
  /** Product name in its own spelling, e.g. "Immich". */
  service: string;
  fields: SetupField[];
  /** Numbered steps to get the credential, in the service's own wording. */
  steps: MessageKey[];
  /** The service page where the credential is created, from the address. */
  credentialPage?: (address: string, values: Record<string, string>) => string | null;
  /**
   * Saves the connection after the server checked it and returns the
   * preview; a failure throws an error whose message names the cause.
   */
  check: (values: Record<string, string>) => Promise<SetupPreview>;
  /** Services of this kind found in the home network; optional (#590). */
  discover?: () => Promise<FoundService[]>;
  /**
   * Signing in with the service instead of pasting a key (#654); the steps
   * and fields above stay as the fallback.
   */
  signIn?: SetupSignIn;
}

/** The address a sign-in starts from, or null when it cannot be offered. */
export function signInAddress(d: SetupDescriptor, values: Record<string, string>): string | null {
  if (!d.signIn) return null;
  const address = d.signIn.address ? d.signIn.address(values) : addressOf(d, values);
  return address ? address.trim().replace(/\/+$/, '') : null;
}

/** The value of the first address field, normalised for links. */
export function addressOf(d: SetupDescriptor, values: Record<string, string>): string {
  const field = d.fields.find((f) => f.stage === 'address');
  return field ? (values[field.key] ?? '').trim().replace(/\/+$/, '') : '';
}

/** Whether every field of a stage is filled in. */
export function stageComplete(
  d: SetupDescriptor,
  stage: SetupStage,
  values: Record<string, string>,
): boolean {
  return d.fields
    .filter((f) => f.stage === stage && !f.optional)
    .every((f) => (values[f.key] ?? '').trim() !== '');
}

/** The starting values: the first option of every choice. */
export function initialValues(d: SetupDescriptor): Record<string, string> {
  const values: Record<string, string> = {};
  for (const f of d.fields) {
    const first = f.options?.[0];
    if (f.type === 'choice' && first) values[f.key] = first.value;
  }
  return values;
}

/**
 * Who may set up what in place: instance integrations need an admin,
 * personal ones any account that is not a guest. The shell provides it;
 * without it (tests, the extension's local mode) widgets offer no setup.
 */
export interface SetupRights {
  admin: boolean;
  guest: boolean;
}

const key = Symbol('setup-rights');

/** Called once by the shell. */
export function provideSetupRights(rights: SetupRights): void {
  setContext(key, rights);
}

/** Whether the viewer may set up an integration of the given kind here. */
export function useMaySetUp(instance: boolean): boolean {
  const r = getContext<SetupRights | undefined>(key);
  if (!r || r.guest) return false;
  return instance ? r.admin : true;
}
