<script lang="ts">
  import { onMount } from 'svelte';
  import { t, type MessageKey } from '../../../../shared/i18n';
  import { errorMessage } from '../../../../shared/api-client/http';
  import Icon from '../../../../shared/ui/icons/Icon.svelte';
  import { PLACE_ID_KEY } from '../../../../shared/places/places';
  import type { BuiltinTemplate, Tab, WidgetConfig } from '../../../dashboard/domain/dashboard';
  import { SINGLE_WIDGETS, templateCards, templateKinds } from '../../domain/wizard';
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

  const singles = $derived(
    SINGLE_WIDGETS.flatMap((kind) => {
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
    kinds = templateKinds(card);
  }

  function chooseEmpty() {
    template = EMPTY;
    kinds = [];
  }

  function toggle(kind: string) {
    template = '';
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

  /** The place of a new weather widget, so it is not asked again. */
  function configOf(kind: string): WidgetConfig | undefined {
    const place = session.place;
    if (kind !== 'weather' || !place) return undefined;
    return { lat: place.lat, lon: place.lon, label: place.label, [PLACE_ID_KEY]: place.id };
  }

  /** Creates the chosen widgets; the setup of those that need one happens on the dashboard. */
  async function apply() {
    const uc = widgets?.uc;
    if (!uc) return;
    const sameAsBefore =
      session.selection?.template === template &&
      session.selection.kinds.length === kinds.length &&
      session.selection.kinds.every((k) => kinds.includes(k));
    if (sameAsBefore) return;
    const place = session.place;
    const unchanged = current !== undefined && kinds.join() === templateKinds(current).join();
    if (current && unchanged) {
      await uc.createTabFromBuiltin(
        current.id,
        titleOf(current),
        place ? { lat: place.lat, lon: place.lon, label: place.label, placeId: place.id } : undefined,
      );
    } else if (kinds.length > 0) {
      let tab: Tab = await uc.createTab(t('wizard.widgets.tab_name'));
      for (const kind of kinds) {
        const type = widgets?.types.find((x) => x.type.kind === kind)?.type;
        if (!type) continue;
        const config = { ...structuredClone(type.defaultConfig ?? {}), ...configOf(kind) };
        tab = { ...tab, widgets: [...tab.widgets, await uc.addWidget(tab, type, config)] };
      }
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
<h2 class="wz-eyebrow">{t('wizard.widgets.single')}</h2>
<div class="wz-tiles">
  {#each singles as s (s.kind)}
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
</div>
{#if error}<p class="wz-error" role="alert">{error}</p>{/if}
