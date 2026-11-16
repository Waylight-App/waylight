import { LocalUnansweredStore } from '../features/search/adapter/localUnansweredStore';
import { LocalUsageStore } from '../features/search/adapter/localUsageStore';
import { UnansweredQueries } from '../features/search/usecase/unansweredQueries';
import { SearchUsage } from '../features/search/usecase/searchUsage';

/**
 * The search history of this browser (#159), shared by the quick search
 * (opened from the search field) and the settings. Ids of chosen results:
 * "<kind>:<id>" for bookmarks, services, notes and to-dos,
 * "answer:<intent>" for smart answers.
 */
export const searchUsage = new SearchUsage(new LocalUsageStore());

/**
 * Queries that ended in the web search without any match (#481); off
 * unless switched on in the settings, kept only in this browser.
 */
export const unansweredQueries = new UnansweredQueries(new LocalUnansweredStore());

export const answerTarget = (intent: string) => `answer:${intent}`;
