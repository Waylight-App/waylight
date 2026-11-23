<script lang="ts">
  import type { Container } from './container';
  import type { OnboardingStep } from './plugins/loader';
  import PluginHost from './plugins/PluginHost.svelte';
  import { definitionFor, widgetDefinitions } from './widgets';
  import { setupWalk } from '../shared/setup/walk.svelte';
  import { DOCS_URL } from '../shared/product';
  import { wizardSearch } from './wizardSearch';
  import type { Appearance } from '../features/appearance';
  import type { Personal } from '../features/onboarding/domain/profile';
  import { startIndex, wizardSteps } from '../features/onboarding/domain/wizard';
  import Wizard from '../features/onboarding/ui/wizard/Wizard.svelte';
  import type { WizardContext } from '../features/onboarding/ui/wizard/wizardContext';

  /**
   * Wires the onboarding wizard (#739) to the app: the account's settings,
   * places, dashboard and the plugin steps. It is loaded on first use, so
   * none of this weighs on the first paint.
   */
  interface Props {
    container: Container;
    modules: Set<string>;
    widgetKinds: Set<string>;
    /** The appearance settings as they are now. */
    appearance: Personal;
    onboardingSteps: OnboardingStep[];
    /** The step the account stopped at; empty before it started. */
    stored: string;
    savePersonal: (patch: Partial<Personal>) => Promise<void>;
    applyDesign: (id: string) => Promise<Appearance>;
    preview: (patch: Partial<Personal> | null) => void;
    ondone: () => void;
  }
  let {
    container,
    modules,
    widgetKinds,
    appearance,
    onboardingSteps,
    stored,
    savePersonal,
    applyDesign,
    preview,
    ondone,
  }: Props = $props();

  /**
   * Whether the dashboard is still as setup left it: one tab with at most
   * the default widget ('empty'), or only the untouched example tabs of the
   * sample data ('sample', #670). An applied waylight.yaml or a dashboard
   * someone arranged counts as set up ('no').
   */
  async function freshDashboard(): Promise<'empty' | 'sample' | 'no'> {
    try {
      const tabs = await container.dashboard.load();
      if (tabs.length > 0 && tabs.every((tab) => tab.sample)) return 'sample';
      return tabs.length <= 1 && tabs.every((tab) => tab.widgets.length <= 1) ? 'empty' : 'no';
    } catch {
      return 'no';
    }
  }

  // The steps are fixed when the wizard opens, so the progress does not jump.
  async function plan() {
    const fresh = await freshDashboard();
    const steps = wizardSteps({
      personal: modules.has('appearance'),
      owner: container.auth.owner,
      widgets: fresh !== 'no',
      extra: onboardingSteps.map((s) => s.id),
    });
    return { steps, start: startIndex(steps, stored), sample: fresh === 'sample' };
  }
  const opening = plan();

  // Created once, so the answers of the search step are cached.
  // svelte-ignore state_referenced_locally
  const search = wizardSearch(container, modules);

  const context = (sample: boolean, withWidgets: boolean): WizardContext => ({
    owner: container.auth.owner,
    personal: modules.has('appearance') ? appearance : undefined,
    savePersonal,
    design: modules.has('appearance')
      ? { gallery: () => container.appearance.gallery(), apply: applyDesign }
      : undefined,
    preview,
    places: modules.has('places')
      ? {
          search: modules.has('weather')
            ? async (q) =>
                (await container.weather.searchPlaces(q)).map((p) => ({
                  name: '',
                  label: [p.name, p.region, p.country].filter(Boolean).join(', '),
                  lat: p.lat,
                  lon: p.lon,
                }))
            : undefined,
          save: (p) => container.places.add(p),
        }
      : undefined,
    widgets: withWidgets
      ? {
          uc: container.dashboard,
          types: widgetDefinitions
            .filter((d) => widgetKinds.has(d.type.kind))
            .map((d) => ({ type: d.type, icon: d.icon })),
          sample,
        }
      : undefined,
    extraSteps: onboardingSteps.map((s) => ({ id: s.id, title: s.title() })),
    extra: pluginStep,
    search,
    widgetName: (kind) => definitionFor(kind)?.type.title ?? kind,
    widgetIcon: (kind) => definitionFor(kind)?.icon,
    setupHelp: async (kind) => {
      const help = await container.help.widget(kind).catch(() => undefined);
      return {
        needs: help?.setUp ?? definitionFor(kind)?.type.description ?? '',
        docsUrl: help?.docsUrl ?? DOCS_URL,
      };
    },
    startSetup: (kinds) => setupWalk.start(kinds),
  });
</script>

{#snippet pluginStep(id: string, setSubmit: (save: () => Promise<void>) => void)}
  {@const step = onboardingSteps.find((s) => s.id === id)}
  {#if step}
    <PluginHost mountable={step.mountable} context={step.context} props={{ setSubmit }} />
  {/if}
{/snippet}

{#await opening then { steps, start, sample }}
  <Wizard
    {steps}
    {start}
    ctx={context(sample, steps.includes('widgets'))}
    onstep={(step) => void container.onboarding.saveStep(step).catch(() => undefined)}
    onfinish={ondone}
  />
{/await}
