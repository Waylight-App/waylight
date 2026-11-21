<script lang="ts">
  import { serverMessage } from '../shared/i18n';
  import { usePolling } from '../shared/refresh';
  import { COMPACT_REFRESH_MS, compactText, type CompactState } from '../features/compact/domain/compact';
  import type { CompactUseCases } from '../features/compact/usecase/compactUseCases';
  import SystemSummary from '../features/compact/ui/SystemSummary.svelte';
  import {
    byUrgency,
    listedEntries,
    sourceEntry,
    type StatusSource,
  } from '../features/dashboard/domain/attentionLine';
  import AttentionLine from '../features/dashboard/ui/AttentionLine.svelte';
  import { attention, integrationName } from './attention.svelte';

  /**
   * The lines above a tab's grid (#593): the attention line and the system
   * summary. This is where the compact states of the status widgets and the
   * account's attention list (#592) meet; the features themselves stay
   * apart.
   */
  interface Props {
    compact: CompactUseCases;
    tabId: string;
    /** One status widget per module (see uniqueSources). */
    sources: StatusSource[];
    /** Whether the status widgets are folded into the summary. */
    collapsed: boolean;
    ontoggle: () => void;
    onview: (widgetId: string) => void;
  }
  let { compact, tabId, sources, collapsed, ontoggle, onview }: Props = $props();

  let states = $state<Record<string, CompactState>>({});

  async function load() {
    const wanted = sources;
    const results = await Promise.allSettled(wanted.map((s) => compact.state(s.module)));
    const next: Record<string, CompactState> = {};
    results.forEach((r, i) => {
      const s = wanted[i];
      if (s && r.status === 'fulfilled') next[s.module] = r.value;
    });
    states = next;
  }

  const poll = usePolling();
  $effect(() => poll(load, COMPACT_REFRESH_MS));

  // The shell keeps the attention list fresh; it stays empty where the list
  // is not offered (guests, kiosk displays, local mode).
  const entries = $derived.by(() => {
    const fromWidgets = sources.flatMap((s) => {
      const state = states[s.module];
      if (!state) return [];
      const text = compactText(state.title) + (state.detail ? ` · ${compactText(state.detail)}` : '');
      const entry = sourceEntry(s, { level: state.level, text });
      return entry ? [entry] : [];
    });
    const listed = listedEntries(
      attention.items.map((i) => ({
        id: i.id,
        kind: i.kind,
        title: integrationName(i.integration),
        text: serverMessage(i.reason.messageKey, i.reason.params) ?? i.reason.message,
        ...(i.widgetId ? { widgetId: i.widgetId } : {}),
        ...(i.tabId ? { tabId: i.tabId } : {}),
      })),
      tabId,
    );
    return byUrgency([...fromWidgets, ...listed]);
  });
</script>

<AttentionLine {entries} {onview} />
<!-- One status widget needs no summary; two or more fold into one line. -->
{#if sources.length > 1}
  <SystemSummary
    items={sources.flatMap((s) => {
      const state = states[s.module];
      return state ? [{ module: s.module, title: s.title, state }] : [];
    })}
    {collapsed}
    {ontoggle}
  />
{/if}
