// Entry point of the browser extension's local mode (ADR 0019): the same app
// as src/main.ts, with the local container and without the service worker,
// which an extension page cannot register and does not need.
import { mount } from 'svelte';
import './app/featureMessages';
import App from './app/App.svelte';
import { downloadExport } from './app/local/downloadExport';
import { createLocalContainer } from './app/local/localContainer';
import { IndexedDbDocumentStore } from './shared/local/documentStore';
import './shared/theme/tokens.css';
import './shared/theme/base.css';

const target = document.getElementById('app');
if (!target) throw new Error('missing #app element');

const store = new IndexedDbDocumentStore();

// "Download my data" on the options page opens the app with #export.
if (location.hash === '#export') {
  history.replaceState(null, '', location.pathname);
  void downloadExport(store);
}

export default mount(App, { target, props: { localContainer: createLocalContainer(store) } });
