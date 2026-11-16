<script lang="ts">
  import {
    captureRest,
    mergePrefixes,
    palettePrefixes,
    prefixesFromConfig,
    QuickActionError,
    type QuickCommand,
  } from '../features/quickactions';
  import type { SearchBarOptions } from '../features/search/usecase/ports';
  import { describeCommand } from './quickCapture';
  import { t } from '../shared/i18n';
  import Markdown from '../shared/ui/Markdown.svelte';
  import { errorMessage } from '../shared/api-client/http';
  import type { Container } from './container';
  import { itemTarget, type QuickItem, search } from './quickSearch';
  import { answerTarget, searchUsage, unansweredQueries } from './searchUsage';
  import { quickItemLoaders } from './quickItems';
  import {
    MIN_QUERY_LENGTH as MIN_WIKI_QUERY,
    pageDetail,
    wikiOnlyQuery,
  } from '../features/wiki/domain/wiki';
  import { createSmartSearch } from './searchSources';
  import { buildAiContext } from './aiContext';
  import { buildRows, keyIntent, move, type Row } from './searchRows';
  import { matchActions, searchActions, type SearchAction } from './searchActions';
  import { type CommandMatch, matchCommands, parseCommandQuery } from '../features/search/domain/commands';
  import { buildCommands, type CommandHandlers } from './appCommands';
  import { parseHelpQuery } from '../features/search/domain/help';
  import { buildHelp } from './searchHelp';
  import SearchHelp from './SearchHelp.svelte';
  import { detectAddress, displayAddress } from '../features/search/domain/address';
  import { BANG_ENGINES, suggestBangs, type Bang } from '../features/search/domain/bangs';
  import { matchKeyword, type OwnKeyword } from '../features/search/domain/keywords';
  import { asksForTimers, formatRemaining, parseTimer, type Timer } from '../features/search/domain/timer';
  import { timers, timerTitle } from './timers';
  import AnswerCard from '../features/search/ui/answers/AnswerCard.svelte';
  import type { ActionItem, Answer } from '../features/search/domain/smart/answer';
  import type { WeatherConfig } from '../features/weather/domain/weather';
  import { requestWidgetFocus } from '../shared/ui/widgetFocus';
  import { copyToClipboard } from '../shared/ui/clipboard';

  /**
   * Results of the hero search field (#222): a panel directly below the
   * field instead of a dialog. The field keeps the text and the focus; this
   * component gets the text, renders the groups of searchRows.ts and handles
   * the keys the field forwards. Loaded on first use (#116).
   */
  interface Props {
    container: Container;
    /** Enabled modules; disabled ones are not searched. */
    modules: Set<string>;
    /** Id of the listbox, referenced by the field's aria-controls. */
    listboxId: string;
    /** The text of the field. */
    query: string;
    /** Id of the selected option, for the field's aria-activedescendant. */
    activeId?: string;
    /** Whether results are shown, for the field's aria-expanded. */
    expanded?: boolean;
    /** Esc: close the panel, keep the text. */
    onclose: () => void;
    /** After opening a result elsewhere: close the panel and clear the field. */
    ondone: () => void;
    /** Location of the first weather widget for weather answers (#156). */
    weatherLocation?: WeatherConfig | null;
    /** First widget config of a kind, e.g. the transit stop (#161). */
    widgetConfig?: (kind: string) => Record<string, unknown> | null;
    /** Handlers of the app commands (#476), supplied by the shell. */
    commandHandlers?: CommandHandlers;
  }
  let {
    container,
    modules,
    listboxId,
    query = $bindable(),
    activeId = $bindable(),
    expanded = $bindable(),
    onclose,
    ondone,
    weatherLocation = null,
    widgetConfig = () => null,
    commandHandlers,
  }: Props = $props();

  // svelte-ignore state_referenced_locally
  const commands = commandHandlers ? buildCommands(commandHandlers) : null;

  // svelte-ignore state_referenced_locally
  const smart = createSmartSearch(container, modules, {
    canAct: !container.auth.guest,
    weatherLocation: () => weatherLocation,
    widgetConfig: (kind) => widgetConfig(kind),
    score: (q, target, score) => searchUsage.score(q, target, score),
  });

  let items = $state<QuickItem[]>([]);
  let notes = new Map<string, { title: string; body: string }>();
  let todos = new Map<string, Parameters<Container['todos']['toggle']>[0]>();
  let active = $state(0);
  let status = $state('');
  let note = $state<{ title: string; body: string } | null>(null);
  let undoLast = $state<(() => Promise<void>) | null>(null);
  let openLast = $state<string | null>(null);

  // Options of the search field (#221) and the name of the search engine.
  let bar = $state<SearchBarOptions | null>(null);
  let engine = $state('');
  let engineId = $state('');
  let ownBangs = $state<Bang[]>([]);
  // Own keywords (#478); the server gives guests and kiosk displays none.
  let ownKeywords = $state<OwnKeyword[]>([]);
  // svelte-ignore state_referenced_locally
  if (modules.has('search'))
    container.search
      .settings()
      .then((s) => {
        bar = s.bar;
        engine = s.providers.find((p) => p.id === s.defaultProvider)?.name ?? '';
        engineId = s.defaultProvider;
        ownBangs = s.shortcuts.map((x) => ({ bang: x.keyword, name: x.name }));
        ownKeywords = s.keywords;
      })
      .catch(() => {});
  const prefixes = $derived(mergePrefixes(palettePrefixes, prefixesFromConfig(bar?.capturePrefixes)));

  // "t Milch" → to-do; the same text as a note is the Ctrl+Enter alternative.
  const command = $derived(container.quickActions.parse(query, prefixes));
  const alternative = $derived.by((): QuickCommand | null => {
    const rest = captureRest(query, prefixes);
    if (!command || rest === null) return null;
    const other = command.kind === 'note' ? 'todo:' : 'note:';
    return container.quickActions.parse(`${other} ${rest}`, palettePrefixes);
  });

  // Body matches from the notes full-text search (#73).
  let noteHits = $state(new Map<string, string>());
  // Pages from connected wikis (#321), asked while typing; "w …" asks only them.
  let wikiItems = $state<QuickItem[]>([]);
  let wikiFailed = $state<string[]>([]);
  const wikiOnly = $derived(wikiOnlyQuery(query));
  const searchable = $derived([
    ...items.map((i) => {
      const snippet = i.kind === 'note' ? noteHits.get(i.id) : undefined;
      return snippet === undefined ? i : { ...i, detail: snippet, keywords: [...(i.keywords ?? []), query] };
    }),
    ...wikiItems,
  ]);
  // ">" limits the rows to app commands (#476).
  const commandOnly = $derived(parseCommandQuery(query).only);
  const commandTarget = (id: string) => `command:${id}`;
  const commandMatches = $derived(
    command || !commands
      ? []
      : matchCommands(
          commands.specs(),
          query,
          {
            guest: container.auth.guest,
            kiosk: container.auth.kiosk,
            admin: container.auth.admin,
            owner: container.auth.owner,
          },
          (id, s) => searchUsage.score(query, commandTarget(id), s),
        ),
  );
  // The text a search keyword put into the field is not matched again, so a
  // keyword whose search is itself a keyword cannot bounce back (#478).
  let keywordQuery = $state<string | null>(null);
  const keyword = $derived(
    command || commandOnly || query === keywordQuery ? null : matchKeyword(ownKeywords, query),
  );
  // "github.com" or "192.168.178.1" opens directly (#475).
  const address = $derived(command || commandOnly ? null : detectAddress(query));
  // "!g" suggests bangs: own shortcuts always, built-in ones when the engine has bangs.
  const bangs = $derived(
    command || !modules.has('search') ? [] : suggestBangs(query, ownBangs, BANG_ENGINES.has(engineId)),
  );
  // Timers and alarms (#474); kiosk displays get none.
  // svelte-ignore state_referenced_locally
  const timersAllowed = !container.auth.kiosk;
  let runningTimers = $state<Timer[]>(timers.list());
  let now = $state(Date.now());
  $effect(() => timers.onChange((list) => (runningTimers = list)));
  const timerRequest = $derived(
    timersAllowed && !command && !commandOnly ? parseTimer(query, new Date()) : null,
  );
  const showRunning = $derived(
    timersAllowed && runningTimers.length > 0 && (query.trim() === '' || asksForTimers(query)),
  );
  // The remaining time ticks while running timers are listed.
  $effect(() => {
    if (!showRunning) return;
    const id = setInterval(() => (now = Date.now()), 1000);
    return () => clearInterval(id);
  });
  const clock = (end: number) => new Date(end).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  const notificationsBlocked = timers.notifier.permission() === 'denied';
  // Results chosen before for this query rank higher (#159).
  const groups = $derived(
    command || commandOnly
      ? []
      : search(searchable, query, 5, (it, s) => searchUsage.score(query, itemTarget(it), s)).filter(
          (g) => wikiOnly === null || g.kind === 'wiki',
        ),
  );
  // Commands of plugins (#453); the loader registers none for guests.
  const pluginActions = $derived(command || commandOnly ? [] : matchActions(searchActions, query));
  const rowGroups = $derived(
    buildRows({
      command,
      alternative,
      groups,
      web: modules.has('search') && query.trim() !== '' && !commandOnly,
      actions: pluginActions,
      commands: commandMatches,
      keyword,
      address,
      bangs,
      timer: timerRequest,
      runningTimers: showRunning ? runningTimers : [],
    }),
  );
  const flat = $derived(rowGroups.flatMap((g) => g.rows));
  const optionId = (i: number) => `${listboxId}-option-${i}`;

  let noteTimer: ReturnType<typeof setTimeout> | undefined;
  $effect(() => {
    const q = query;
    clearTimeout(noteTimer);
    if (!modules.has('notes') || q.trim().length < 2) {
      noteHits = new Map();
      return;
    }
    noteTimer = setTimeout(() => {
      container.notes
        .search(q)
        .then((hits) => {
          if (q === query) noteHits = new Map(hits.map((h) => [h.id, h.snippet.map((x) => x.text).join('')]));
        })
        .catch(() => (noteHits = new Map()));
    }, 200);
    return () => clearTimeout(noteTimer);
  });

  // The wiki is asked from two characters on, after a short pause; the
  // newest query wins and a slow wiki never holds up the other results.
  let wikiTimer: ReturnType<typeof setTimeout> | undefined;
  $effect(() => {
    const q = query;
    const wq = (wikiOnlyQuery(q) ?? q).trim();
    clearTimeout(wikiTimer);
    if (!modules.has('wiki') || container.auth.guest || command || wq.length < MIN_WIKI_QUERY) {
      wikiItems = [];
      wikiFailed = [];
      return;
    }
    wikiTimer = setTimeout(() => {
      container.wiki
        .search(wq)
        .then((r) => {
          if (q !== query) return;
          wikiFailed = r.failed;
          // The wiki already matched the page; the query as keyword keeps
          // the local ranking from dropping it.
          wikiItems = r.pages.map((p, i) => ({
            kind: 'wiki' as const,
            id: `${p.wiki}:${i}:${p.url}`,
            title: p.title,
            detail: pageDetail(p),
            keywords: [q],
            url: p.url,
          }));
        })
        .catch(() => ((wikiItems = []), (wikiFailed = [])));
    }, 250);
    return () => clearTimeout(wikiTimer);
  });

  $effect(() => {
    // Load everything once when the panel opens; one failing source does not
    // block the others. The module calls live in quickItems.ts (#156).
    for (const load of quickItemLoaders(container, modules)) {
      load
        .then((r) => {
          items = [...items, ...r.items];
          if (r.notes) notes = r.notes;
          if (r.todos) todos = r.todos;
        })
        .catch(() => {});
    }
    void smart.prepare();
  });

  // Smart answers (#156): debounced, the newest query wins.
  let answers = $state<Answer[]>([]);
  $effect(() => {
    const q = query;
    if (command || commandOnly || wikiOnlyQuery(q) !== null || q.trim().length < 2) {
      answers = [];
      return;
    }
    const ctrl = new AbortController();
    const timer = setTimeout(() => {
      smart
        .answer(q, ctrl.signal)
        .then((a) => (answers = a))
        .catch(() => {});
    }, 150);
    return () => {
      clearTimeout(timer);
      ctrl.abort();
    };
  });

  // Typing starts over: first row selected, the last confirmation goes away.
  $effect(() => {
    if (!query) return; // keep the confirmation after a command cleared the field
    active = 0;
    status = '';
    undoLast = null;
    openLast = null;
  });

  // "?", "hilfe" or "help" lists what the search can do (#479).
  const helpFilter = $derived(parseHelpQuery(query));
  const helpEntries = $derived(
    helpFilter === null
      ? []
      : buildHelp({
          modules,
          viewer: {
            guest: container.auth.guest,
            kiosk: container.auth.kiosk,
            admin: container.auth.admin,
            owner: container.auth.owner,
          },
          prefixes: prefixesFromConfig(bar?.capturePrefixes),
          commands: commands ?? undefined,
        }),
  );

  const visible = $derived(query.trim() !== '' || status !== '' || note !== null || showRunning);

  // Ask the AI (#162): only for a query starting with "?", only when a local
  // Ollama is set up, never for guests; the component is loaded on demand.
  let aiReady = $state(false);
  $effect(() => {
    // The summary uses the instance's language model; admins only (ADR 0018).
    if (modules.has('summary') && container.auth.admin)
      void container.summary.available().then((ok) => (aiReady = ok));
  });
  const aiQuestion = $derived(
    aiReady && query.trim().startsWith('?') && query.trim().length > 3 ? query.trim() : '',
  );
  const askAi = (onChunk: (text: string) => void, signal: AbortSignal) =>
    buildAiContext(container, modules).then((context) =>
      container.summary.ask(aiQuestion, context, onChunk, signal),
    );

  $effect(() => {
    activeId = visible && flat.length ? optionId(Math.min(active, flat.length - 1)) : undefined;
    expanded = visible && (flat.length > 0 || answers.length > 0 || aiQuestion !== '' || helpFilter !== null);
  });

  /** Opens a typed address, in a new tab when the search is set up that way. */
  function openAddress(url: string) {
    if (bar?.openInNewTab) {
      window.open(url, '_blank', 'noopener');
      ondone();
    } else window.location.assign(url);
  }

  /** Opens what an own keyword points to (#478). */
  function openKeyword(k: OwnKeyword) {
    switch (k.kind) {
      case 'url':
        return openAddress(k.target);
      case 'query':
        keywordQuery = k.target;
        query = k.target;
        return;
      case 'widget':
        requestWidgetFocus([k.target]);
        ondone();
        return;
      case 'service':
      case 'bookmark': {
        const item = searchable.find((i) => i.kind === k.kind && i.id === k.target);
        if (item) void choose(item);
        else status = t('search.keyword_gone', { keyword: k.keyword });
        return;
      }
    }
  }

  // "Add keyword…" for the selected tile or bookmark (#478). It sits below
  // the list, because a button inside an option is not accessible.
  const keywordFor = $derived.by((): QuickItem | null => {
    const row = flat[active];
    if (!modules.has('search') || container.auth.kiosk || row?.kind !== 'result') return null;
    return row.item.kind === 'service' || row.item.kind === 'bookmark' ? row.item : null;
  });
  let keywordItem = $state<QuickItem | null>(null);
  let keywordWord = $state('');
  async function addKeyword(e: SubmitEvent) {
    e.preventDefault();
    const item = keywordItem;
    if (!item || (item.kind !== 'service' && item.kind !== 'bookmark')) return;
    try {
      const s = await container.search.addKeyword({
        keyword: keywordWord,
        kind: item.kind,
        target: item.id,
        label: item.title,
      });
      ownKeywords = s.keywords;
      status = t('search.keyword_added', { keyword: keywordWord.trim().toLowerCase() });
      keywordItem = null;
      keywordWord = '';
    } catch (err) {
      status = errorMessage(err);
    }
  }

  function searchWeb() {
    const target = modules.has('search') ? container.search.targetFor(query) : null;
    if (!target) return;
    // Remembered only when switched on (#481) and when nothing else matched:
    // no result, command or capture row and no answer.
    if (!command && answers.length === 0 && flat.every((r) => r.kind === 'web'))
      unansweredQueries.record(query);
    if (bar?.openInNewTab) {
      window.open(target, '_blank', 'noopener');
      ondone();
    } else window.location.assign(target);
  }

  async function runCommand(c: QuickCommand) {
    if (
      (c.kind === 'restart' || c.kind === 'wake') &&
      !confirm(describeCommand(c, container.quickActions) + '?')
    )
      return;
    try {
      const res = await container.quickActions.execute(c);
      query = '';
      status = describeCommand(res.command, container.quickActions, true);
      undoLast = res.undo ?? null;
      openLast = res.openUrl ?? null;
    } catch (err) {
      status =
        err instanceof QuickActionError
          ? t(err.message === 'restart' ? 'quick.not_allowed_restart' : 'quick.not_allowed_wake')
          : errorMessage(err);
    }
  }

  async function runAction(a: SearchAction) {
    try {
      const done = await a.run();
      query = '';
      status = done || a.title();
      undoLast = null;
      openLast = null;
    } catch (err) {
      status = errorMessage(err);
    }
  }

  function runAppCommand(m: CommandMatch) {
    if (!commands) return;
    // Remembered like a chosen result, so it ranks higher next time (#159).
    searchUsage.record(query, commandTarget(m.spec.id));
    ondone();
    commands.run(m.spec, m.rest);
  }

  async function undo() {
    const u = undoLast;
    undoLast = null;
    openLast = null; // the bookmark is gone
    if (!u) return;
    try {
      await u();
      status = t('quick.undone');
    } catch (err) {
      status = errorMessage(err);
    }
  }

  async function choose(item: QuickItem) {
    // Only a chosen result is remembered, never the query alone (#159).
    searchUsage.record(query, itemTarget(item));
    if (item.url) {
      window.open(item.url, '_blank', 'noopener,noreferrer');
      ondone();
    } else if (item.kind === 'note') {
      note = notes.get(item.id) ?? null;
    } else if (item.kind === 'todo') {
      const todo = todos.get(item.id);
      if (!todo) return;
      try {
        await container.todos.toggle(todo);
        items = items.filter((i) => !(i.kind === 'todo' && i.id === item.id));
        status = t('app.marked_done', { title: item.title });
      } catch (err) {
        status = errorMessage(err);
      }
    }
  }

  function run(row: Row | undefined) {
    if (!row) return searchWeb();
    switch (row.kind) {
      case 'capture':
      case 'action':
        return void runCommand(row.command);
      case 'plugin':
        return void runAction(row.action);
      case 'command':
        return runAppCommand(row.match);
      case 'keyword':
        return openKeyword(row.keyword);
      case 'address':
        return openAddress(row.url);
      case 'bang':
        // Completes the bang; the field keeps the focus for the search terms.
        query = `!${row.bang.bang} `;
        return;
      case 'timer': {
        const started = timers.start(row.request);
        query = '';
        status = t(started.kind === 'alarm' ? 'timers.alarm_set' : 'timers.timer_started', {
          time: clock(started.end),
        });
        return;
      }
      case 'running':
        timers.cancel(row.timer.id);
        status = t('timers.cancelled', { name: timerTitle(row.timer) });
        return;
      case 'result':
        return void choose(row.item);
      case 'web':
        return searchWeb();
    }
  }

  async function execute(item: ActionItem, intent: string) {
    searchUsage.record(query, answerTarget(intent));
    try {
      await smart.execute(item);
      status = t('smart.action_done', { name: item.target.name });
      answers = await smart.answer(query, new AbortController().signal);
    } catch (err) {
      status = errorMessage(err);
    }
  }

  function openWidget(kind: string, intent: string) {
    searchUsage.record(query, answerTarget(intent));
    requestWidgetFocus([kind]);
    ondone();
  }

  async function copyResult(text: string) {
    status = (await copyToClipboard(text)) ? `${t('smart.copied')}: ${text}` : '';
  }

  /** Handles a key the field forwards; returns whether it was used. */
  export function keydown(e: KeyboardEvent): boolean {
    const intent = keyIntent(e);
    switch (intent) {
      case 'down':
      case 'up':
        active = move(active, flat.length, intent);
        document.getElementById(optionId(active))?.scrollIntoView({ block: 'nearest' });
        return true;
      case 'close':
        if (note) note = null;
        else onclose();
        return true;
      case 'web':
        searchWeb();
        return true;
      case 'alternative':
        if (alternative) void runCommand(alternative);
        return true;
      case 'run': {
        // A calculator result is copied (#160) while its row is the default.
        const calc =
          command || active > 0 ? undefined : smart.direct(query).find((a) => a.copy !== undefined);
        if (calc?.copy !== undefined) void copyResult(calc.copy);
        else run(flat[active]);
        return true;
      }
      default:
        return false;
    }
  }
</script>

{#if visible}
  <div class="panel">
    {#if note}
      <div class="note">
        <button class="btn btn-icon" type="button" onclick={() => (note = null)}>{t('app.back')}</button>
        <h4>{note.title}</h4>
        <Markdown source={note.body} />
      </div>
    {:else}
      {#if helpFilter !== null}
        <SearchHelp entries={helpEntries} filter={helpFilter} onpick={(example) => (query = example)} />
      {/if}
      {#if answers.length > 0}
        <!-- Instant answers above the rows (#156). -->
        <section class="answers" aria-label={t('smart.answers')} aria-live="polite">
          {#each answers as a (a.intent + a.title)}
            <AnswerCard
              answer={a}
              onopen={(kind) => openWidget(kind, a.intent)}
              onexecute={(item) => execute(item, a.intent)}
              onpick={(name) => (query = `${query.trim()} ${name}`)}
            />
          {/each}
        </section>
      {/if}
      {#if flat.length}
        <div id={listboxId} role="listbox" aria-label={t('app.results')} class="rows">
          {#each rowGroups as g (g.id)}
            <div role="group" aria-labelledby="{listboxId}-group-{g.id}">
              <div class="group" id="{listboxId}-group-{g.id}">{t(g.label)}</div>
              {#each g.rows as row (row.id)}
                {@const i = flat.indexOf(row)}
                <!-- The field keeps the focus; rows are reached with the arrow keys. -->
                <div
                  id={optionId(i)}
                  role="option"
                  tabindex="-1"
                  aria-selected={i === active}
                  class="row"
                  class:active={i === active}
                  onmousedown={(e) => e.preventDefault()}
                  onclick={() => run(row)}
                  onkeydown={(e) => e.key === 'Enter' && run(row)}
                  onmousemove={() => (active = i)}
                >
                  {#if row.kind === 'capture'}
                    <span class="title">{describeCommand(row.command, container.quickActions)}</span>
                    <kbd>{row.shortcut}</kbd>
                  {:else if row.kind === 'result'}
                    <span class="title">{row.item.title}</span>
                    {#if row.item.detail}<span class="muted detail">{row.item.detail}</span>{/if}
                  {:else if row.kind === 'keyword'}
                    <span class="title"
                      >{row.keyword.keyword} → {row.keyword.label || row.keyword.target}</span
                    >
                    <kbd>Enter</kbd>
                  {:else if row.kind === 'address'}
                    <span class="title">{t('app.open_address', { address: displayAddress(row.url) })}</span>
                    <kbd>Enter</kbd>
                  {:else if row.kind === 'bang'}
                    <span class="title">!{row.bang.bang}</span>
                    <span class="muted detail"
                      >{row.bang.name}{row.bang.own ? ` · ${t('app.bang_own')}` : ''}</span
                    >
                  {:else if row.kind === 'timer'}
                    <span class="title"
                      >{row.request.kind === 'alarm'
                        ? t('timers.set_alarm', { time: clock(row.request.end) })
                        : t('timers.start_timer', {
                            duration: formatRemaining(row.request.end - Date.now()),
                          })}{row.request.label ? ` · ${row.request.label}` : ''}</span
                    >
                    <span class="muted detail"
                      >{row.request.tomorrow
                        ? t('timers.tomorrow')
                        : t('timers.ends_at', { time: clock(row.request.end) })}{notificationsBlocked
                        ? ` · ${t('timers.notifications_blocked')}`
                        : ''}</span
                    >
                  {:else if row.kind === 'running'}
                    <span class="title"
                      >⏱ {timerTitle(row.timer)} · {formatRemaining(row.timer.end - now)}</span
                    >
                    <span class="muted detail">{t('timers.enter_cancels')}</span>
                  {:else if row.kind === 'command'}
                    <span class="title">{commands?.title(row.match.spec)}</span>
                    {#if row.match.spec.shortcut}<kbd>{row.match.spec.shortcut}</kbd>{/if}
                  {:else if row.kind === 'plugin'}
                    <span class="title">{row.action.title()}</span>
                    {#if row.action.detail()}<span class="muted detail">{row.action.detail()}</span>{/if}
                  {:else if row.kind === 'action'}
                    <span class="title">{describeCommand(row.command, container.quickActions)}</span>
                    <span class="muted detail"
                      >{row.command.kind === 'restart' ? 'Docker' : 'Wake-on-LAN'}</span
                    >
                  {:else}
                    <span class="title"
                      >{engine
                        ? t('app.search_web_with', { query: query.trim(), engine })
                        : t('app.search_web', { query: query.trim() })}</span
                    >
                    <kbd>Alt Enter</kbd>
                  {/if}
                </div>
              {/each}
            </div>
          {/each}
        </div>
      {/if}
      {#if keywordItem}
        <form class="keyword-form" onsubmit={addKeyword}>
          <label
            >{t('search.add_keyword_prompt', { label: keywordItem.title })}
            <input class="input" bind:value={keywordWord} maxlength="24" required /></label
          >
          <button class="btn" type="submit">{t('search.add')}</button>
          <button class="btn" type="button" onclick={() => (keywordItem = null)}>{t('common.cancel')}</button>
        </form>
      {:else if keywordFor}
        <p class="hint">
          <button class="btn link" type="button" onclick={() => (keywordItem = keywordFor)}
            >{t('search.add_keyword')}</button
          >
        </p>
      {/if}
      {#if aiQuestion}
        {#await import('./AskAi.svelte') then m}
          <m.default question={aiQuestion} ask={askAi} />
        {/await}
      {/if}
      {#if wikiFailed.length}
        <p class="muted hint">{t('wiki.failed', { names: wikiFailed.join(', ') })}</p>
      {/if}
      <p class="muted hint" role="status">
        {status}
        {#if openLast}<a href={openLast} target="_blank" rel="noopener noreferrer">{t('quick.open')}</a>{/if}
        {#if undoLast}<button class="btn" type="button" onclick={undo}>{t('dashboard.undo')}</button>{/if}
      </p>
      <p class="sr-only" aria-live="polite">
        {query.trim() ? t('app.results_count', { n: flat.length }) : ''}
      </p>
    {/if}
  </div>
{/if}

<style>
  .keyword-form {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    align-items: end;
    padding: var(--space-2) var(--space-3);
  }
  .panel {
    position: absolute;
    top: calc(100% + var(--space-2));
    left: 0;
    right: 0;
    z-index: 40;
    max-height: min(70vh, 34rem);
    overflow: auto;
    padding: var(--space-2);
    border: 1px solid var(--color-border);
    border-radius: 14px;
    background: var(--color-surface);
    box-shadow: 0 12px 36px color-mix(in srgb, #000 16%, transparent);
    text-align: left;
  }
  .answers {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(14rem, 1fr));
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }
  .group {
    padding: var(--space-2) var(--space-2) 2px;
    font-size: max(0.8rem, 0.75em); /* at least 12 px (#317) */
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--color-text-muted);
  }
  .row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    padding: var(--space-2);
    border-radius: 8px;
    cursor: pointer;
  }
  .row.active {
    background: color-mix(in srgb, var(--color-accent) 14%, transparent);
  }
  .title {
    flex: 1 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .detail {
    flex: 0 1 auto;
    min-width: 0;
    max-width: 40%;
    font-size: 0.8em;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  kbd {
    flex-shrink: 0;
    font-size: max(0.8rem, 0.75rem); /* at least 12 px (#317) */
    border: 1px solid var(--color-border);
    border-radius: 4px;
    padding: 0 6px;
    color: var(--color-text-muted);
  }
  .hint {
    font-size: 0.8em;
    margin: var(--space-2) var(--space-2) 0;
  }
  .hint:empty {
    display: none;
  }
  .note h4 {
    margin: var(--space-2) 0;
  }
  @media (max-width: 600px) {
    /* No keyboard on a phone: the shortcut hints only take room. */
    kbd {
      display: none;
    }
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
    white-space: nowrap;
  }
</style>
