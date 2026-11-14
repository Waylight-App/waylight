<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { parseConfig, type League, type SportsTeam } from '../domain/sports';
  import type { SportsUseCases } from '../usecase/sportsUseCases';

  /** Props of the sports widget settings: league and team. */
  interface Props {
    config: Record<string, unknown>;
    sports: SportsUseCases;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, sports, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  const initial = parseConfig(config);
  let league = $state(initial.league);
  let team = $state(initial.team);
  let leagues = $state<League[]>([]);
  let teams = $state<SportsTeam[]>([]);
  $effect(() => {
    void sports.leagues().then((l) => (leagues = l));
  });
  $effect(() => {
    const l = league;
    void sports
      .teams(l)
      .then((ts) => (teams = [...ts].sort((a, b) => a.name.localeCompare(b.name))))
      .catch(() => (teams = []));
  });
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ league, ...(team > 0 ? { team } : {}) });
  }}
>
  <label class="field"
    ><span>{t('sports.league')}</span><select class="input" bind:value={league} onchange={() => (team = 0)}>
      {#each leagues as l (l.id)}<option value={l.id}>{l.name}</option>{/each}
    </select></label
  >
  <label class="field"
    ><span>{t('sports.team')}</span><select class="input" bind:value={team}>
      <option value={0}>{t('sports.no_team')}</option>
      {#each teams as tm (tm.id)}<option value={tm.id}>{tm.name}</option>{/each}
    </select></label
  >
  <p><button class="btn btn-primary" type="submit">{t('common.save')}</button></p>
</form>
