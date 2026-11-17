<script lang="ts">
  /**
   * The icon on the single sign-on button (#514): the Microsoft mark or a
   * neutral key. Both are drawn here, no image is fetched. More providers
   * can be added as further branches.
   */
  interface Props {
    icon: 'microsoft' | 'key' | '';
  }
  let { icon }: Props = $props();
</script>

{#if icon === 'microsoft'}
  <svg class="sso-icon" width="18" height="18" viewBox="0 0 21 21" aria-hidden="true">
    <rect x="1" y="1" width="9" height="9" fill="#f25022" />
    <rect x="11" y="1" width="9" height="9" fill="#7fba00" />
    <rect x="1" y="11" width="9" height="9" fill="#00a4ef" />
    <rect x="11" y="11" width="9" height="9" fill="#ffb900" />
  </svg>
{:else if icon === 'key'}
  <svg
    class="sso-icon"
    width="18"
    height="18"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <circle cx="7.5" cy="15.5" r="5.5" />
    <path d="m21 2-9.6 9.6" />
    <path d="m15.5 7.5 3 3L22 7l-3-3" />
  </svg>
{/if}

<style>
  .sso-icon {
    flex: none;
    margin-right: var(--space-2);
    vertical-align: -0.2em;
  }
</style>
