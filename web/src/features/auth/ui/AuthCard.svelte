<script lang="ts">
  import type { Snippet } from 'svelte';
  import WaylightMark from '../../../shared/ui/WaylightMark.svelte';
  import { PRODUCT_NAME } from '../../../shared/product';

  /** Name and logos shown above the card; the logos are versioned URLs. */
  export interface Brand {
    name: string;
    logo?: string;
    logoDark?: string;
  }

  /** Props of the centered card used by the setup and login screens. */
  interface Props {
    title: string;
    subtitle?: string;
    /** Logo and name above the card (#513); without it the card stands alone. */
    brand?: Brand;
    /** Lamp logo centred above the title (#653); decorative, the title names the card. */
    logo?: boolean;
    children: Snippet;
  }
  let { title, subtitle, brand, logo = false, children }: Props = $props();
</script>

<main>
  {#if brand}
    <p class="brand">
      {#if brand.logo}
        <img class="brand-logo light" src={brand.logo} alt="" height="32" />
        <img class="brand-logo dark" src={brand.logoDark || brand.logo} alt="" height="32" />
      {:else}
        <span class="mark"><WaylightMark size={28} /></span>
      {/if}
      <span class:wordmark={brand.name === PRODUCT_NAME}>{brand.name}</span>
    </p>
  {/if}
  <!-- The frame carries the theme's glow (#505), apart from the card's own
       shadow: a comma list with "none" would void both. -->
  <div class="frame">
    <section class="card">
      {#if logo}<div class="logo"><WaylightMark size={48} /></div>{/if}
      <h1>{title}</h1>
      {#if subtitle}<p class="muted">{subtitle}</p>{/if}
      {@render children()}
    </section>
  </div>
</main>

<style>
  main {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-4);
    padding: var(--space-4);
    box-sizing: border-box;
  }
  .frame {
    width: min(420px, 100%);
    border-radius: var(--radius);
    box-shadow: var(--glow);
  }
  section {
    box-sizing: border-box;
    padding: var(--space-6);
    box-shadow: var(--shadow-pop);
  }
  h1 {
    margin: 0 0 var(--space-2);
    font-size: 1.6rem;
  }
  .logo {
    display: flex;
    justify-content: center;
    margin-bottom: var(--space-3);
    color: var(--color-accent-text);
  }
  .brand {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin: 0;
    font-weight: 600;
    font-size: 1.1rem;
  }
  .wordmark {
    font-family: var(--font-wordmark);
    font-weight: 700;
  }
  .mark {
    display: inline-grid;
    place-items: center;
    color: var(--color-accent-text);
  }
  .brand-logo {
    width: auto;
    max-width: 160px;
  }
  .brand-logo.dark,
  :global(html[data-theme='dark']) .brand-logo.light {
    display: none;
  }
  :global(html[data-theme='dark']) .brand-logo.dark {
    display: block;
  }
</style>
