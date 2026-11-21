/**
 * Composition root of the browser extension's local mode (ADR 0019): the
 * same container as the web app, with local implementations where the
 * extension works without a server. Features that have no local
 * implementation keep their API adapters but are hidden through the module
 * list, so their code is never asked to reach a server.
 */
import { AppearanceUseCases } from '../../features/appearance/usecase/appearanceUseCases';
import { LocalAppearanceGateway } from '../../features/appearance/adapter/local/localAppearanceGateway';
import { AuthUseCases } from '../../features/auth/usecase/authUseCases';
import { LocalAuthGateway, noCsrf } from '../../features/auth/adapter/local/localAuthGateway';
import { CountdownUseCases } from '../../features/countdown/usecase/countdownUseCases';
import { LocalCountdownGateway } from '../../features/countdown/adapter/local/localCountdownGateway';
import { browserNotifier } from '../../features/focus/adapter/api/focusApi';
import { FocusUseCases } from '../../features/focus/usecase/focusUseCases';
import { LocalFocusGateway } from '../../features/focus/adapter/local/localFocusGateway';
import { LocalTimer } from '../../features/focus/adapter/local/localTimer';
import { HabitsUseCases } from '../../features/habits/usecase/habitsUseCases';
import { LocalHabitsGateway } from '../../features/habits/adapter/local/localHabitsGateway';
import { DashboardUseCases } from '../../features/dashboard/usecase/dashboardUseCases';
import { LocalDashboardGateway } from '../../features/dashboard/adapter/local/localDashboardGateway';
import { BookmarksUseCases } from '../../features/bookmarks/usecase/bookmarksUseCases';
import { LocalBookmarksGateway } from '../../features/bookmarks/adapter/local/localBookmarksGateway';
import { NotesUseCases } from '../../features/notes/usecase/notesUseCases';
import { LocalNotesGateway } from '../../features/notes/adapter/local/localNotesGateway';
import { TodosUseCases } from '../../features/todos/usecase/todosUseCases';
import { LocalTodosGateway } from '../../features/todos/adapter/local/localTodosGateway';
import { LocalWeatherGateway } from '../../features/weather/adapter/local/localWeatherGateway';
import { WeatherUseCases } from '../../features/weather/usecase/weatherUseCases';
import { locale } from '../../shared/i18n';
import { IndexedDbDocumentStore, type DocumentStore } from '../../shared/local/documentStore';
import { createContainer, type Container } from '../container';
import { localDataKinds, localModules, localWidgetKinds } from './modules';

/**
 * Builds the container for the extension's local mode. The store is a
 * parameter so tests can pass one in memory.
 */
export function createLocalContainer(store: DocumentStore = new IndexedDbDocumentStore()): Container {
  // Nothing can expire without a server; an unauthorised answer would mean
  // a hidden feature reached for the network, which the module list prevents.
  const base = createContainer(() => undefined);
  const todos = new TodosUseCases(new LocalTodosGateway(store));
  const focusSessions = new LocalFocusGateway(store);
  return {
    ...base,
    localMode: true,
    auth: new AuthUseCases(new LocalAuthGateway(), noCsrf),
    dashboard: new DashboardUseCases(new LocalDashboardGateway(store)),
    appearance: new AppearanceUseCases(new LocalAppearanceGateway(store)),
    notes: new NotesUseCases(new LocalNotesGateway(store)),
    bookmarks: new BookmarksUseCases(new LocalBookmarksGateway(store)),
    todos,
    // Only the built-in list exists locally; connected lists need a server.
    todosFor: () => todos,
    countdown: new CountdownUseCases(new LocalCountdownGateway(store)),
    // The browser notifier lives next to the API adapter but needs no server.
    focus: new FocusUseCases(focusSessions, browserNotifier, new LocalTimer(focusSessions)),
    habits: new HabitsUseCases(new LocalHabitsGateway(store)),
    weather: new WeatherUseCases(new LocalWeatherGateway(locale)),
    enabledModules() {
      return Promise.resolve({
        modules: new Set(localModules),
        widgetKinds: new Set(localWidgetKinds),
        dataKinds: { ...localDataKinds },
      });
    },
  };
}
