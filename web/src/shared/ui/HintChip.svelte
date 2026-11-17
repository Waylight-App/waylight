<script lang="ts">
  /**
   * A short hint in a tinted pill (#509): a bold lead word and, optionally,
   * the rest, e.g. "**Frost** tomorrow" or "**Today**". The text keeps the
   * normal text colour and the pill is only lightly tinted, so the contrast
   * stays that of text on a card in every theme.
   */
  interface Props {
    lead: string;
    text?: string;
    tone?: 'accent' | 'warning' | 'danger';
  }
  let { lead, text = '', tone = 'accent' }: Props = $props();
</script>

<span class="chip" data-tone={tone}
  ><strong>{lead}</strong>{#if text}<span>{text}</span>{/if}</span
>

<style>
  .chip {
    display: inline-flex;
    align-items: baseline;
    gap: 0.3em;
    padding: 0.1em 0.65em;
    border-radius: 999px;
    font-size: 0.8em;
    line-height: 1.4;
    white-space: nowrap;
    color: var(--color-text);
    background: color-mix(in srgb, var(--tone) 16%, var(--color-surface));
  }
  .chip[data-tone='accent'] {
    --tone: var(--color-accent);
  }
  .chip[data-tone='warning'] {
    --tone: var(--color-warning);
  }
  .chip[data-tone='danger'] {
    --tone: var(--color-danger);
  }
</style>
