<script lang="ts">
  import Icon from '../ui/icons/Icon.svelte';

  /**
   * The result of the connection test (#738): a clear success message in
   * green, or what went wrong in red. Status and alert roles make screen
   * readers announce it without moving the focus.
   */
  interface Props {
    ok: boolean;
    lines: string[];
    /** Up to three pictures that show real data, e.g. thumbnails. */
    images?: string[];
  }
  let { ok, lines, images = [] }: Props = $props();
</script>

<div class="result" class:ok role={ok ? 'status' : 'alert'}>
  {#if ok}<Icon name="done" size={20} />{:else}<Icon name="info" size={20} />{/if}
  <div>
    {#each lines as line (line)}<div>{line}</div>{/each}
    {#if images.length}
      <div class="images">
        {#each images.slice(0, 3) as src (src)}<img {src} alt="" loading="lazy" />{/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .result {
    display: flex;
    gap: var(--space-3);
    align-items: flex-start;
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-md);
    font-weight: 600;
    background: color-mix(in srgb, var(--color-danger) 14%, transparent);
    color: var(--color-danger);
  }
  .result.ok {
    background: color-mix(in srgb, var(--color-success) 16%, transparent);
    color: var(--color-success);
  }
  .result :global(svg) {
    flex: none;
    margin-top: 2px;
  }
  .images {
    display: grid;
    grid-template-columns: repeat(3, 6rem);
    gap: var(--space-1);
    margin-top: var(--space-2);
  }
  .images img {
    width: 100%;
    aspect-ratio: 1;
    object-fit: cover;
    border-radius: var(--radius-sm);
  }
</style>
