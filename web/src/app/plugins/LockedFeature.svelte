<script lang="ts">
  import { t } from '../../shared/i18n';

  /** A plugin section whose licence feature is off (ADR 0021): what it does, and that it needs a key. */
  interface Props {
    title: string;
    description: string;
  }
  let { title, description }: Props = $props();
</script>

<div class="locked" role="note">
  <p><strong>{title}</strong></p>
  <p>{description}</p>
  <p class="muted">{t('plugins.needs_licence')}</p>
</div>

<style>
  .locked {
    padding: var(--space-3);
    border-radius: var(--radius);
    background: var(--color-surface-2, rgba(127, 127, 127, 0.12));
  }
  .locked p {
    margin: 0 0 var(--space-2);
  }
</style>
