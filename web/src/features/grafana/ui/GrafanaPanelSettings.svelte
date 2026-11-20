<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import { type GrafanaDashboard, type GrafanaPanel, parsePanelConfig } from '../domain/grafana';
  import type { GrafanaUseCases } from '../usecase/grafanaUseCases';

  /** Props of the Grafana panel widget settings. */
  interface Props {
    config: Record<string, unknown>;
    grafana: GrafanaUseCases;
    onsave: (config: Record<string, unknown>) => Promise<void>;
  }
  let { config, grafana, onsave }: Props = $props();
  // svelte-ignore state_referenced_locally
  let form = $state(parsePanelConfig(config));
  let query = $state('');
  let dashboards = $state<GrafanaDashboard[]>([]);
  let panels = $state<GrafanaPanel[]>([]);
  let error = $state('');

  $effect(() => {
    const q = query;
    const timer = setTimeout(() => {
      grafana
        .dashboards(q)
        .then((list) => ((dashboards = list), (error = '')))
        .catch((err) => (error = errorMessage(err)));
    }, 250);
    return () => clearTimeout(timer);
  });

  $effect(() => {
    const uid = form.dashboardUid;
    if (!uid) return;
    grafana
      .panels(uid)
      .then((list) => (panels = list))
      .catch((err) => (error = errorMessage(err)));
  });

  function chooseDashboard(uid: string) {
    const d = dashboards.find((x) => x.uid === uid);
    form.dashboardUid = uid;
    form.dashboardTitle = d?.title ?? '';
    form.dashboardUrl = d?.url ?? '';
    form.panelId = 0;
    form.panelTitle = '';
  }

  function choosePanel(id: number) {
    form.panelId = id;
    form.panelTitle = panels.find((p) => p.id === id)?.title ?? '';
  }
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onsave({ ...form });
  }}
>
  <label class="field"
    ><span>{t('grafana.search_dashboards')}</span><input
      class="input"
      type="search"
      bind:value={query}
    /></label
  >
  <label class="field">
    <span>{t('grafana.dashboard')}</span>
    <select class="input" value={form.dashboardUid} onchange={(e) => chooseDashboard(e.currentTarget.value)}>
      <option value="" disabled>{dashboards.length ? '' : t('grafana.no_dashboards')}</option>
      {#if form.dashboardUid && !dashboards.some((d) => d.uid === form.dashboardUid)}
        <option value={form.dashboardUid}>{form.dashboardTitle}</option>
      {/if}
      {#each dashboards as d (d.uid)}
        <option value={d.uid}>{d.folder ? `${d.folder} / ${d.title}` : d.title}</option>
      {/each}
    </select>
  </label>
  {#if form.dashboardUid}
    <label class="field">
      <span>{t('grafana.panel')}</span>
      <select class="input" value={form.panelId} onchange={(e) => choosePanel(Number(e.currentTarget.value))}>
        <option value={0} disabled></option>
        {#each panels as p (p.id)}
          <option value={p.id}>{p.title || `#${p.id}`}</option>
        {/each}
      </select>
    </label>
  {/if}
  <label class="field"
    ><span>{t('grafana.range')}</span><input
      class="input"
      bind:value={form.range}
      pattern="[1-9][0-9]{'{0,2}'}[mhdw]"
      required
    /></label
  >
  <label class="field"
    ><span>{t('grafana.refresh_minutes')}</span><input
      class="input"
      type="number"
      min="1"
      max="60"
      bind:value={form.refreshMinutes}
    /></label
  >
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  <p><button class="btn btn-primary" type="submit" disabled={!form.panelId}>{t('common.save')}</button></p>
</form>
