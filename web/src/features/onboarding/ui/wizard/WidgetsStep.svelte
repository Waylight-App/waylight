<script lang="ts">
  import { onMount } from 'svelte';
  import { t, type MessageKey } from '../../../../shared/i18n';
  import { errorMessage } from '../../../../shared/api-client/http';
  import Icon from '../../../../shared/ui/icons/Icon.svelte';
  import WidgetPicker from '../../../dashboard/ui/WidgetPicker.svelte';
  import { PLACE_ID_KEY } from '../../../../shared/places/places';
  import type { BuiltinTemplate, Tab, Widget, WidgetConfig } from '../../../dashboard/domain/dashboard';
  import type { DashboardUseCases } from '../../../dashboard/usecase/dashboardUseCases';
  import { suggestionsOf, suggestionTiles, templateCards, templateKinds } from '../../domain/wizard';
  import type { StepProps } from './wizardContext';

  let { ctx, session, ctl }: StepProps = $props();
  const widgets = $derived(ctx.widgets);

  const EMPTY = 'empty';

  let cards = $state<BuiltinTemplate[]>([]);
  // The earlier choice comes back when the user steps back to this screen.
  // svelte-ignore state_referenced_locally
  let template = $state(session.selection?.template ?? '');
  // svelte-ignore state_referenced_locally
  let kinds = $state<string[]>(session.selection?.kinds ?? []);
  let error = $state('');
  let picking = $state(false);

  const available = $derived(new Set(widgets?.types.map((x) => x.type.kind)));
  const suggestions = $derived(suggestionsOf(template));
  const tiles = $derived(
    suggestionTiles(suggestions, kinds, available).flatMap((kind) => {
      const entry = widgets?.types.find((x) => x.type.kind === kind);
      return entry ? [{ kind, title: entry.type.title, icon: entry.icon }] : [];
    }),
  );
  const current = $derived(cards.find((c) => c.id === template));

  onMount(async () => {
    try {
      cards = templateCards((await widgets?.uc.listBuiltinTemplates()) ?? []);
      // The first visit starts with the family template, like the old picker did.
      const first = cards[0];
      if (!session.selection && first) choose(first);
    } catch (err) {
      error = errorMessage(err);
    }
  });

  function choose(card: BuiltinTemplate) {
    template = card.id;
    kinds = suggestionsOf(card.id).selected.filter((k) => available.has(k));
  }

  function chooseEmpty() {
    template = EMPTY;
    kinds = [];
  }

  function toggle(kind: string) {
    kinds = kinds.includes(kind) ? kinds.filter((k) => k !== kind) : [...kinds, kind];
  }

  const describe = (card: BuiltinTemplate) =>
    templateKinds(card)
      .slice(0, 4)
      .map((k) => ctx.widgetName(k))
      .join(', ');

  const titleOf = (card: BuiltinTemplate) => {
    const key = `templates.${card.id}` as MessageKey;
    const text = t(key);
    return text === key ? card.title : text;
  };

  const sameKinds = (a: readonly string[], b: readonly string[]) =>
    a.length === b.length && a.every((k) => b.includes(k));

  /** The place of a new weather widget, so it is not asked again. */
  function configOf(kind: string): WidgetConfig | undefined {
    const place = session.place;
    if (kind !== 'weather' || !place) return undefined;
    return { lat: place.lat, lon: place.lon, label: place.label, [PLACE_ID_KEY]: place.id };
  }

  /** Adds one widget of a kind to the tab, with the saved place for the weather. */
  async function addOfKind(uc: DashboardUseCases, tab: Tab, kind: string): Promise<Widget> {
    const type = widgets?.types.find((x) => x.type.kind === kind)?.type;
    if (!type) throw new Error(`unknown widget kind ${kind}`);
    const config = { ...structuredClone(type.defaultConfig ?? {}), ...configOf(kind) };
    return uc.addWidget(tab, type, config);
  }

  /** Creates the chosen widgets; the setup of those that need one happens on the dashboard. */
  async function apply() {
    const uc = widgets?.uc;
    if (!uc) return;
    if (session.selection?.template === template && sameKinds(session.selection.kinds, kinds)) return;
    const place = session.place;
    // A template keeps its layout; the widgets that were switched off are
    // removed from the new tab and the others are added to it.
    if (current && kinds.length > 0) {
      let tab = await uc.createTabFromBuiltin(
        current.id,
        titleOf(current),
        place ? { lat: place.lat, lon: place.lon, label: place.label, placeId: place.id } : undefined,
      );
      for (const w of tab.widgets.filter((w) => !kinds.includes(w.kind))) await uc.deleteWidget(w.id);
      tab = { ...tab, widgets: tab.widgets.filter((w) => kinds.includes(w.kind)) };
      for (const kind of kinds.filter((k) => !tab.widgets.some((w) => w.kind === k))) {
        tab = { ...tab, widgets: [...tab.widgets, await addOfKind(uc, tab, kind)] };
      }
    } else if (kinds.length > 0) {
      let tab: Tab = await uc.createTab(t('wizard.widgets.tab_name'));
      for (const kind of kinds) tab = { ...tab, widgets: [...tab.widgets, await addOfKind(uc, tab, kind)] };
    }
    // A choice, also the empty one, replaces the example tabs of the setup.
    if (widgets?.sample && (kinds.length > 0 || template === EMPTY)) await uc.discardSample();
    session.selection = { template, kinds: [...kinds] };
    session.created = [...kinds];
  }

  // The shell hands each step a fresh control, so the first value is the one.
  // svelte-ignore state_referenced_locally
  ctl.save = apply;
  $effect(() => {
    ctl.status = t(kinds.length === 1 ? 'wizard.widgets.count_one' : 'wizard.widgets.count', {
      n: kinds.length,
    });
  });
</script>

<h1 class="wz-title wz-title-md" tabindex="-1">{t('wizard.widgets.title')}</h1>
<p class="wz-lead wz-lead-sm">{t('wizard.widgets.body')}</p>
<h2 class="wz-eyebrow">{t('wizard.widgets.templates')}</h2>
<div class="wz-cards">
  {#each cards as card (card.id)}
    <button class="wz-card" type="button" aria-pressed={template === card.id} onclick={() => choose(card)}>
      <b>{titleOf(card)}</b><span>{describe(card)}</span>
    </button>
  {/each}
  <button class="wz-card" type="button" aria-pressed={template === EMPTY} onclick={chooseEmpty}>
    <b>{t('wizard.widgets.empty')}</b><span>{t('wizard.widgets.empty_hint')}</span>
  </button>
</div>
<h2 class="wz-eyebrow">
  {template === EMPTY || !current
    ? t('wizard.widgets.suggestions')
    : t('wizard.widgets.suggestions_for', { name: titleOf(current) })}
</h2>
<div class="wz-tiles">
  {#each tiles as s (s.kind)}
    {@const on = kinds.includes(s.kind)}
    <button class="wz-tile" type="button" aria-pressed={on} onclick={() => toggle(s.kind)}>
      {#if on}
        <span class="wz-tick" aria-hidden="true">
          <svg
            width="12"
            height="12"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="3.5"
            stroke-linecap="round"
            stroke-linejoin="round"><path d="M5 12l5 5 9-10" /></svg
          >
        </span>
      {/if}
      <Icon name={s.icon} size={28} />
      {s.title}
    </button>
  {/each}
  <button class="wz-tile wz-tile-more" type="button" onclick={() => (picking = true)}>
    <svg
      width="28"
      height="28"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="1.8"
      stroke-linecap="round"
      stroke-linejoin="round"
      aria-hidden="true"><path d="M12 5v14M5 12h14" /></svg
    >
    {t('wizard.widgets.more')}
  </button>
</div>
{#if error}<p class="wz-error" role="alert">{error}</p>{/if}
<p class="wz-flag" role="note">{t('wizard.widgets.flag')}</p>

{#if picking}
  <WidgetPicker
    types={widgets?.types.map((x) => x.type) ?? []}
    icon={(kind) => widgets?.types.find((x) => x.type.kind === kind)?.icon}
    selected={new Set(kinds)}
    sizes={false}
    onpick={(type) => toggle(type.kind)}
    onclose={() => (picking = false)}
  >
    {#snippet footer()}
      <div class="picker-end">
        <button class="btn btn-primary" type="button" onclick={() => (picking = false)}
          >{t('dashboard.done')}</button
        >
      </div>
    {/snippet}
  </WidgetPicker>
{/if}

<style>
  .picker-end {
    display: flex;
    justify-content: flex-end;
  }
</style>
