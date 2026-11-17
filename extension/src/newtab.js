import { dashboardUrl, healthUrl } from './url.js';
import { loadBase, loadMode, saveMode } from './storage.js';
import { LOCAL_APP, startTarget } from './start.js';

// The new tab: the dashboard on the user's server, the bundled app in local
// mode (ADR 0019), or a choice between the two. The health check is a
// no-cors request, so the extension needs no host permission; an opaque
// answer is enough to know the server is there.
const message = document.getElementById('message');
const retry = document.getElementById('retry');
const choose = document.getElementById('choose');

function show(text) {
  message.textContent = text;
  document.body.classList.add('problem');
}

async function go() {
  const [base, mode] = await Promise.all([loadBase(), loadMode()]);
  const target = startTarget({ base, mode });
  if (target === 'local') {
    location.replace(LOCAL_APP);
    return;
  }
  if (target === 'choose') {
    show('Use waylight right here in your browser, or connect it to your own server.');
    choose.hidden = false;
    return;
  }
  try {
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), 5000);
    await fetch(healthUrl(base), { mode: 'no-cors', cache: 'no-store', signal: controller.signal });
    clearTimeout(timer);
    location.replace(dashboardUrl(base));
  } catch {
    show(`${base} is not reachable right now.`);
    retry.hidden = false;
  }
}

retry.addEventListener('click', () => location.reload());
document.getElementById('local').addEventListener('click', async () => {
  await saveMode('local');
  location.replace(LOCAL_APP);
});
for (const id of ['options', 'server']) {
  document.getElementById(id).addEventListener('click', (e) => {
    e.preventDefault();
    (globalThis.browser ?? globalThis.chrome).runtime.openOptionsPage();
  });
}
go();
