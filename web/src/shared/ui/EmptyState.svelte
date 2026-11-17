<script lang="ts">
  import { t } from '../i18n';
  import { useOpenSettings } from './settingsLink';

  /**
   * What a widget shows before it has anything to show (#512): an icon, one
   * sentence and, when the widget needs setting up, a button that opens that
   * settings section.
   */
  interface Props {
    text: string;
    /** Settings section that sets the widget up, e.g. "calendar". */
    section?: string;
  }
  let { text, section }: Props = $props();
  const open = useOpenSettings();
</script>

<div class="empty">
  <svg
    width="28"
    height="28"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.75"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
    ><path d="M22 12h-6l-2 3h-4l-2-3H2" /><path
      d="M5.45 5.11 2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z"
    /></svg
  >
  <p>{text}</p>
  {#if section && open}
    <button class="btn" type="button" onclick={() => open(section)}>{t('common.set_up')}</button>
  {/if}
</div>

<style>
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    height: 100%;
    padding: var(--space-3) var(--space-2);
    text-align: center;
    color: var(--color-text-muted);
  }
  svg {
    color: var(--color-accent-text);
    opacity: 0.8;
  }
  p {
    margin: 0;
    max-width: 28ch;
    font-size: 0.9em;
  }
</style>
