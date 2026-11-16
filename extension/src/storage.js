// Two settings, in storage.sync so they follow the browser profile: the
// dashboard address and the mode ('server' or 'local', ADR 0019). The local
// mode's own data lives in the app's IndexedDB, not here.
const KEY = 'baseUrl';
const MODE = 'mode';
const api = globalThis.browser ?? globalThis.chrome;

export async function loadBase() {
  const got = await api.storage.sync.get(KEY);
  return typeof got[KEY] === 'string' ? got[KEY] : '';
}

export async function saveBase(base) {
  await api.storage.sync.set({ [KEY]: base });
}

export async function loadMode() {
  const got = await api.storage.sync.get(MODE);
  return got[MODE] === 'local' || got[MODE] === 'server' ? got[MODE] : '';
}

export async function saveMode(mode) {
  await api.storage.sync.set({ [MODE]: mode });
}

export { api };
