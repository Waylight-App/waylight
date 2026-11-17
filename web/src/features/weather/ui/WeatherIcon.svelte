<script lang="ts" module>
  import type { WeatherIcon } from '../domain/weather';

  /*
   * Line icons for the weather (#508), one per kind of weather, drawn with
   * currentColor so they follow the theme. Emoji looked different on every
   * system and could not take the theme's colours. The shapes are from
   * Lucide (ISC licence, see THIRD_PARTY_NOTICES.md), bundled, no request.
   */
  const cloud = 'M4 14.899A7 7 0 1 1 15.71 8h1.79a4.5 4.5 0 0 1 2.5 8.242';
  const paths: Record<WeatherIcon, string[]> = {
    'clear-day': [
      'M12 8a4 4 0 1 0 0 8a4 4 0 1 0 0-8',
      'M12 2v2',
      'M12 20v2',
      'm4.93 4.93 1.41 1.41',
      'm17.66 17.66 1.41 1.41',
      'M2 12h2',
      'M20 12h2',
      'm6.34 17.66-1.41 1.41',
      'm19.07 4.93-1.41 1.41',
    ],
    'clear-night': ['M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z'],
    'partly-day': [
      'M12 2v2',
      'm4.93 4.93 1.41 1.41',
      'M20 12h2',
      'm19.07 4.93-1.41 1.41',
      'M15.947 12.65a4 4 0 0 0-5.925-4.128',
      'M13 22H7a5 5 0 1 1 4.9-6H13a3 3 0 0 1 0 6Z',
    ],
    'partly-night': [
      'M13 16a3 3 0 1 1 0 6H7a5 5 0 1 1 4.9-6Z',
      'M10.1 9A6 6 0 0 1 16 4a4.24 4.24 0 0 0 6 6 6 6 0 0 1-3 5.197',
    ],
    cloudy: ['M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z'],
    fog: [cloud, 'M16 17H7', 'M17 21H9'],
    drizzle: [cloud, 'M8 19v1', 'M8 14v1', 'M16 19v1', 'M16 14v1', 'M12 21v1', 'M12 16v1'],
    rain: [cloud, 'M16 14v6', 'M8 14v6', 'M12 16v6'],
    sleet: [cloud, 'M16 14v2', 'M8 14v2', 'M16 20h.01', 'M8 20h.01', 'M12 16v2', 'M12 22h.01'],
    snow: [cloud, 'M8 15h.01', 'M8 19h.01', 'M12 17h.01', 'M12 21h.01', 'M16 15h.01', 'M16 19h.01'],
    thunder: ['M6 16.326A7 7 0 1 1 15.71 8h1.79a4.5 4.5 0 0 1 .5 8.973', 'm13 12-3 5h4l-3 5'],
    unknown: [
      'M12 2a10 10 0 1 0 0 20a10 10 0 1 0 0-20',
      'M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3',
      'M12 17h.01',
    ],
  };
</script>

<script lang="ts">
  /**
   * One weather icon. It is decorative: the condition is always written next
   * to it or given as `label`, which makes the icon an image for screen
   * readers.
   */
  interface Props {
    icon: WeatherIcon;
    /** The condition in words, when no text next to the icon says it. */
    label?: string;
    /** Width and height; 1em follows the surrounding font size. */
    size?: string;
  }
  let { icon, label, size = '1em' }: Props = $props();
</script>

<svg
  class="weather-icon"
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width="2"
  stroke-linecap="round"
  stroke-linejoin="round"
  role={label ? 'img' : undefined}
  aria-label={label}
  aria-hidden={label ? undefined : 'true'}
>
  {#each paths[icon] as d (d)}<path {d} />{/each}
</svg>

<style>
  .weather-icon {
    display: inline-block;
    flex: none;
    vertical-align: -0.125em;
  }
</style>
