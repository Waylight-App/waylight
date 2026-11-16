import { normaliseBase } from './url.js';
import { loadBase, loadMode, saveBase, saveMode } from './storage.js';

const form = document.getElementById('form');
const input = document.getElementById('base');
const status = document.getElementById('status');
const serverFields = document.getElementById('server-fields');
const move = document.getElementById('move');

const chosen = () => form.elements.namedItem('mode').value;

// The address only matters, and is only required, for the server mode.
function update() {
  const server = chosen() === 'server';
  serverFields.hidden = !server;
  input.required = server;
}

async function init() {
  const [base, mode] = await Promise.all([loadBase(), loadMode()]);
  input.value = base;
  form.elements.namedItem('mode').value = mode || (base ? 'server' : 'local');
  // Someone who used the local mode may want to take their data along.
  move.hidden = mode !== 'local';
  update();
}

form.addEventListener('change', update);
form.addEventListener('submit', async (e) => {
  e.preventDefault();
  if (chosen() === 'local') {
    await saveMode('local');
    move.hidden = false;
    status.textContent = 'Saved. New tabs now open trisse-home without a server.';
    return;
  }
  const base = normaliseBase(input.value);
  if (!base) {
    status.textContent = 'Enter an http(s) address such as http://homeserver:8090 (no user name, query or #).';
    return;
  }
  await Promise.all([saveBase(base), saveMode('server')]);
  input.value = base;
  status.textContent = 'Saved. New tabs now open your dashboard.';
});

init();
