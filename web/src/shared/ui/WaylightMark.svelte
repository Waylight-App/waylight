<!--
  The Waylight lamp. It is drawn with currentColor so the surrounding
  element decides the colour; the shell paints it in the theme's accent.
-->
<script lang="ts">
  interface Props {
    size?: number;
  }

  let { size = 28 }: Props = $props();
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 28 28"
  fill="none"
  stroke="currentColor"
  stroke-width="2.2"
  stroke-linecap="round"
  stroke-linejoin="round"
  aria-hidden="true"
  focusable="false"
>
  <path d="M14 3v4" />
  <path d="M8 7h12l2 10H6L8 7z" />
  <path d="M11 17v2a3 3 0 0 0 6 0v-2" />
  <path d="M3 24h22" />
</svg>
