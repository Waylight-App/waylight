// Entry point of the browser extension's local mode (ADR 0019): the same app
// as src/main.ts, with the local container and without the service worker,
// which an extension page cannot register and does not need.
import { mount } from 'svelte';
import './app/featureMessages';
import App from './app/App.svelte';
import { createLocalContainer } from './app/local/localContainer';
import './shared/theme/tokens.css';
import './shared/theme/base.css';

const target = document.getElementById('app');
if (!target) throw new Error('missing #app element');

export default mount(App, { target, props: { localContainer: createLocalContainer() } });
