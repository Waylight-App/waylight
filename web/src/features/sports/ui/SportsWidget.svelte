<script lang="ts">
  import { usePolling } from '../../../shared/refresh';
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import {
    outcome,
    parseConfig,
    score,
    signed,
    type SportsMatch,
    type SportsOverview,
  } from '../domain/sports';
  import type { SportsUseCases } from '../usecase/sportsUseCases';

  /** Props of the sports widget (#285). */
  interface Props {
    sports: SportsUseCases;
    config: Record<string, unknown>;
  }
  let { sports, config }: Props = $props();
  const poll = usePolling();
  const cfg = $derived(parseConfig(config));
  let o = $state<SportsOverview | null>(null);
  let error = $state('');

  $effect(() => {
    const { league, team } = cfg;
    return poll(
      () =>
        sports
          .overview(league, team)
          .then((r) => ((o = r), (error = '')))
          .catch((err) => {
            error = errorMessage(err);
            throw err;
          }),
      15 * 60_000,
    );
  });

  const when = (m: SportsMatch) =>
    m.kickoff.toLocaleString(intlLocale(), {
      weekday: 'short',
      day: 'numeric',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit',
    });
</script>

{#snippet match(m: SportsMatch | null, teamId: number)}
  {#if m}
    {@const res = outcome(m, teamId)}
    <p class="match">
      <span class="teams">{m.home.shortName} – {m.away.shortName}</span>
      {#if score(m)}<strong class="score {res}">{score(m)}</strong>{/if}
    </p>
    <p class="muted small">
      {when(m)}{#if m.matchday}
        · {m.matchday}{/if}{#if res}
        · {t(`sports.${res}`)}{/if}
    </p>
  {:else}
    <p class="muted">{t('sports.none')}</p>
  {/if}
{/snippet}

{#if o}
  {#if o.team}
    <h4 class="team">{o.team.name}</h4>
    {#if o.position}<p class="muted small">
        {t('sports.position', {
          position: o.position.position,
          points: o.position.points,
          played: o.position.played,
        })}
      </p>{/if}
    <h5>{t('sports.last')}</h5>
    {@render match(o.last, o.team.id)}
    <h5>{t('sports.next')}</h5>
    {@render match(o.next, o.team.id)}
  {:else}
    <h4 class="team">{o.league.name}</h4>
    <table>
      <thead><tr><th></th><th></th><th>{t('sports.goal_diff')}</th><th>{t('sports.points')}</th></tr></thead>
      <tbody>
        {#each o.table as s (s.team.id)}
          <tr
            ><td>{s.position}.</td><td class="name">{s.team.shortName}</td><td>{signed(s.goalDiff)}</td><td
              ><strong>{s.points}</strong></td
            ></tr
          >
        {/each}
      </tbody>
    </table>
  {/if}
{:else if error}
  <p class="error">{error}</p>
{:else}
  <p class="muted">{t('common.loading')}</p>
{/if}

<style>
  .team {
    margin: 0;
    font-size: var(--fs-sm);
  }
  h5 {
    margin: var(--space-2) 0 var(--space-1);
    font-size: var(--fs-xs);
    color: var(--color-text-subtle);
  }
  .match {
    display: flex;
    justify-content: space-between;
    gap: var(--space-2);
    margin: 0;
  }
  .teams {
    min-width: 0;
    overflow-wrap: anywhere;
  }
  .score {
    flex: none;
    font-variant-numeric: tabular-nums;
  }
  .score.win {
    color: var(--color-success);
  }
  .score.loss {
    color: var(--color-danger);
  }
  .small {
    font-size: var(--fs-xs);
    margin: 0;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    font-variant-numeric: tabular-nums;
  }
  th {
    font-weight: 500;
    color: var(--color-text-subtle);
    font-size: var(--fs-xs);
    text-align: right;
  }
  td {
    text-align: right;
    padding: var(--space-1) 0;
  }
  td.name {
    text-align: left;
    padding-left: var(--space-2);
    width: 100%;
  }
</style>
