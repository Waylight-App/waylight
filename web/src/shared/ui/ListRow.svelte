<script lang="ts">
  import type { Snippet } from 'svelte';

  /**
   * The one list row every widget uses (#767): main text, an optional second
   * line, a value on the right and optional leading and trailing controls.
   * Rows keep a 40px minimum and a hairline between them; text wraps instead
   * of being cut off, so the row also works when the widget is resized.
   */
  interface Props {
    /** Quiet second line at the small size. */
    sub?: Snippet;
    /** Right-aligned value; tabular figures keep columns aligned. */
    value?: Snippet;
    /** Colours the value with the accent text colour (due or urgent). */
    urgent?: boolean;
    /** Marker before the text, such as a status dot or a checkbox. */
    lead?: Snippet;
    /** Controls after the value, such as small icon buttons. */
    trail?: Snippet;
    /** Truncates the main text to one line with an ellipsis instead of wrapping. */
    nowrap?: boolean;
    /** False when the row has no value, so no empty value box is drawn. */
    showValue?: boolean;
    title?: string;
    style?: string;
    class?: string;
    children: Snippet;
  }
  let {
    sub,
    value,
    urgent = false,
    lead,
    trail,
    nowrap = false,
    showValue = true,
    style,
    title,
    class: cls = '',
    children,
  }: Props = $props();
</script>

<li class="row {cls}" {title} {style}>
  {#if lead}{@render lead()}{/if}
  <div class="main">
    <div class="text" class:nowrap>{@render children()}</div>
    {#if sub}<div class="sub">{@render sub()}</div>{/if}
  </div>
  {#if value && showValue}<span class="value" class:urgent>{@render value()}</span>{/if}
  {#if trail}{@render trail()}{/if}
</li>

<style>
  .row {
    position: relative;
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-height: 40px;
    padding: var(--space-1) 0;
    border-bottom: 1px solid var(--color-border);
    font-variant-numeric: tabular-nums;
  }
  .row:last-child {
    border-bottom: none;
  }
  .main {
    flex: 1;
    min-width: 0;
  }
  .text {
    overflow-wrap: anywhere;
  }
  .text.nowrap {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .sub {
    font-size: var(--fs-xs);
    color: var(--color-text-muted);
    overflow-wrap: anywhere;
  }
  .value {
    flex: none;
    max-width: 50%;
    text-align: right;
    color: var(--color-text-muted);
    overflow-wrap: anywhere;
  }
  .value.urgent {
    color: var(--color-accent-text);
  }
</style>
