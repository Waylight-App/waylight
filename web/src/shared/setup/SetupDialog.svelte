<script lang="ts">
  import type { Snippet } from 'svelte';
  import { t } from '../i18n';
  import Icon from '../ui/icons/Icon.svelte';
  import type { IconName } from '../ui/icons/paths';
  import Modal from '../ui/Modal.svelte';

  /**
   * The one frame every guided setup uses (#738): a symbol, a plain title,
   * one sentence on what the widget does, everything the setup needs in the
   * body, a "where do I find this" note and a single primary action. Later
   * leaves without saving. Enter in any field takes the primary action.
   */
  interface Props {
    title: string;
    icon?: IconName;
    /** One sentence for what the widget does and what the setup needs. */
    lead: string;
    /** The one primary action, e.g. Check and save, then Done. */
    primary: string;
    /** Disables the primary action while a check runs. */
    busy?: boolean;
    onprimary: () => void;
    onclose: () => void;
    children: Snippet;
    /** The note on where to find the values, shown below the body. */
    help?: Snippet;
  }
  let { title, icon, lead, primary, busy = false, onprimary, onclose, children, help }: Props = $props();

  function submit(e: SubmitEvent) {
    e.preventDefault();
    if (!busy) onprimary();
  }
</script>

<Modal {title} {icon} size="medium" {onclose}>
  <form class="setup" onsubmit={submit}>
    <p class="lead">{lead}</p>
    {@render children()}
    {#if help}
      <aside class="help">
        <Icon name="info" size={22} />
        <div>{@render help()}</div>
      </aside>
    {/if}
    <div class="foot">
      <button class="btn link" type="button" onclick={onclose}>{t('setup.later')}</button>
      <button class="btn btn-primary" type="submit" disabled={busy}>{primary}</button>
    </div>
  </form>
</Modal>

<style>
  .setup {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }
  .lead {
    margin: 0;
    color: var(--color-text-muted);
    font-size: var(--fs-sm);
    line-height: 1.5;
  }
  .help {
    display: flex;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-md);
    background: color-mix(in srgb, var(--color-accent) 10%, transparent);
    line-height: 1.5;
  }
  .help :global(svg) {
    flex: none;
    margin-top: 2px;
    color: var(--color-accent-text);
  }
  .foot {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    gap: var(--space-4);
  }
  .foot .btn-primary {
    padding: var(--space-3) var(--space-6);
    font-weight: 600;
  }
  .link {
    border-color: transparent;
    background: none;
    color: var(--color-text-muted);
  }
  @media (prefers-reduced-motion: no-preference) {
    .foot .btn-primary {
      transition: background-color 0.15s;
    }
  }
</style>
