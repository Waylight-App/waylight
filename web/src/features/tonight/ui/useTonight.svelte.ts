import { SvelteDate } from 'svelte/reactivity';
import { errorMessage } from '../../../shared/api-client/http';
import { usePolling } from '../../../shared/refresh';
import type { TonightWarning, TonightWeather } from '../domain/tonight';
import type { TonightUseCases } from '../usecase/tonightUseCases';

/** The backend refreshes the weather every 15 minutes; polling more often is pointless. */
const POLL_MS = 10 * 60 * 1000;
/** The card shows minutes only, so a tick every ten seconds keeps it exact enough. */
const TICK_MS = 10_000;

/**
 * Loads the weather and warnings for the card and keeps the time current.
 * Call it while the component initialises; both the full and the compact
 * view use it, so they load the same data the same way.
 */
export function useTonight(source: () => { tonight: TonightUseCases; config: Record<string, unknown> }) {
  const poll = usePolling();
  const now = new SvelteDate();
  let weather = $state<TonightWeather | null>(null);
  let warnings = $state<TonightWarning[]>([]);
  let loaded = $state(false);
  let error = $state('');

  $effect(() => {
    const timer = setInterval(() => now.setTime(Date.now()), TICK_MS);
    return () => clearInterval(timer);
  });

  $effect(() => {
    const { tonight, config } = source();
    return poll(
      () =>
        tonight
          .snapshot(config)
          .then((s) => {
            weather = s.weather;
            warnings = s.warnings;
            error = '';
            loaded = true;
          })
          .catch((err) => {
            error = errorMessage(err);
            throw err;
          }),
      POLL_MS,
    );
  });

  return {
    get now() {
      return now;
    },
    get weather() {
      return weather;
    },
    get warnings() {
      return warnings;
    },
    get loaded() {
      return loaded;
    },
    get error() {
      return error;
    },
  };
}
