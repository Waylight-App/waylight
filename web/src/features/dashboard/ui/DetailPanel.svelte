<script lang="ts">
  import { t } from '../../../shared/i18n';
  import type { Snippet } from 'svelte';

  /**
   * Props of the detail panel (#594, #595): the full view of a widget whose
   * tile shows its compact view. A side panel on wide screens, so the
   * dashboard stays visible beside it, and a sheet from the bottom on
   * phones. It is a modal dialog, so focus stays inside until it closes.
   */
  interface Props {
    title: string;
    /** Called on Escape, the close button or a click beside the panel. */
    onclose: () => void;
    /** Element that gets the focus back when the panel closes (the tile). */
    opener?: HTMLElement | null;
    children: Snippet;
    /** Buttons below the content, e.g. "Edit widget". */
    actions?: Snippet;
  }
  let { title, onclose, opener = null, children, actions }: Props = $props();
  let dialog: HTMLDialogElement;

  $effect(() => {
    dialog.showModal();
    const returnTo = opener;
    return () => {
      dialog.close();
      // Browsers differ in where focus goes after a modal dialog; the tile
      // that opened it is where keyboard users expect to continue.
      if (returnTo?.isConnected) returnTo.focus();
    };
  });
</script>

<dialog
  bind:this={dialog}
  class="panel card"
  aria-labelledby="detail-panel-title"
  data-testid="detail-panel"
  oncancel={(e) => {
    e.preventDefault();
    onclose();
  }}
  onclick={(e) => {
    if (e.target === dialog) onclose();
  }}
>
  <header>
    <h2 id="detail-panel-title">{title}</h2>
    <button class="btn btn-icon" type="button" aria-label={t('ui.close')} onclick={onclose}>✕</button>
  </header>
  <div class="content"><div class="body">{@render children()}</div></div>
  {#if actions}<footer>{@render actions()}</footer>{/if}
</dialog>

<style>
  .panel {
    position: fixed;
    inset: 0 0 0 auto;
    width: min(460px, 100vw);
    height: 100dvh;
    max-height: 100dvh;
    margin: 0;
    padding: 0;
    border-radius: 0;
    color: var(--color-text);
    display: none;
    flex-direction: column;
  }
  .panel[open] {
    display: flex;
  }
  /* A light backdrop: the dashboard stays readable beside the panel. */
  .panel::backdrop {
    background: rgb(0 0 0 / 15%);
  }
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-3) var(--space-4);
    border-bottom: 1px solid var(--color-border);
  }
  h2 {
    margin: 0;
    font-size: var(--fs-sm);
  }
  /* The same size container as a widget card, so the full view lays itself
     out as in a large tile. */
  .content {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding: var(--space-3) var(--space-4);
    container: widget / size;
  }
  .body {
    height: 100%;
  }
  footer {
    display: flex;
    gap: var(--space-2);
    justify-content: flex-end;
    padding: var(--space-3) var(--space-4);
    border-top: 1px solid var(--color-border);
  }
  @media (max-width: 720px) {
    .panel {
      inset: auto 0 0 0;
      width: 100vw;
      max-width: 100vw;
      height: 85dvh;
      border-radius: var(--radius-md) var(--radius-md) 0 0;
    }
  }
</style>
