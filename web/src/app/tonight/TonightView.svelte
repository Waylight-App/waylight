<script lang="ts">
  import { TonightWidget, type TonightUseCases } from '../../features/tonight';
  import WeatherIcon from '../../features/weather/ui/WeatherIcon.svelte';
  import type { WeatherIcon as Kind } from '../../features/weather/domain/weather';

  /** The Tonight card with the weather feature's icons (composition, see weatherSource.ts). */
  let { tonight, config }: { tonight: TonightUseCases; config: Record<string, unknown> } = $props();
</script>

{#snippet icon(kind: string)}<WeatherIcon icon={kind as Kind} />{/snippet}
<TonightWidget {tonight} {config} {icon} />
