import { describe, expect, it } from 'vitest';
import { docsTopics, type HelpTopic, isQuestion, matchTopics, questionWords, widgetTopics } from './help';

const topic = (
  id: string,
  title: string,
  detail = '',
  keywords: string[] = [],
  group?: string,
): HelpTopic => ({
  id,
  title,
  detail,
  keywords,
  url: `https://example.org/${id}`,
  group,
});

describe('isQuestion (#678)', () => {
  it('takes question marks and question words in both languages', () => {
    expect(isQuestion('How do I add a calendar')).toBe(true);
    expect(isQuestion('Wie füge ich einen Kalender hinzu?')).toBe(true);
    expect(isQuestion('calendar?')).toBe(true);
    expect(isQuestion('? kalender')).toBe(true);
  });

  it('leaves ordinary searches alone', () => {
    expect(isQuestion('calendar')).toBe(false);
    expect(isQuestion('howto')).toBe(false);
    expect(isQuestion('?')).toBe(false);
  });
});

describe('questionWords', () => {
  it('drops filler words and cuts long words to a stem', () => {
    expect(questionWords('How do I add my calendars?')).toEqual(['add', 'calenda']);
    expect(questionWords('Wie füge ich einen Kalender hinzu')).toEqual(['fuge', 'kalend']);
  });
});

describe('matchTopics', () => {
  const topics = [
    topic('notes', 'Notes', 'Add a note from the search.'),
    topic('calendar', 'Calendar', 'Your appointments.', ['agenda', 'termine']),
    topic('todos', 'To-dos', 'Tasks with due dates; add them from the calendar view.'),
  ];

  it('needs a word in the title or keywords; the detail only ranks', () => {
    expect(matchTopics('How do I add a calendar?', topics).map((t) => t.id)).toEqual(['calendar']);
    expect(matchTopics('how do I add something?', topics)).toEqual([]);
  });

  it('finds keywords, also in another language', () => {
    expect(matchTopics('Wo sehe ich meine Termine?', topics).map((t) => t.id)).toEqual(['calendar']);
  });

  it('lists a group once, with its best topic', () => {
    const docs = [
      topic('dash', 'Tabs and widgets', '', [], 'dash'),
      topic('dash#templates', 'Templates', '', ['Tabs and widgets'], 'dash'),
      topic('other', 'Templates of mails', '', [], 'other'),
    ];
    expect(matchTopics('how do templates work?', docs).map((t) => t.id)).toEqual(['dash#templates', 'other']);
  });
});

describe('topics', () => {
  it('names widget topics like the picker and skips unknown kinds', () => {
    const help = [
      { kind: 'immich', shows: 'Your photos.', setUp: 'Enter a key.', docsUrl: 'https://d/immich/' },
      { kind: 'gone', shows: '', setUp: '', docsUrl: '' },
    ];
    const t = widgetTopics(help, (k) =>
      k === 'immich' ? { title: 'Immich', keywords: ['fotos'] } : undefined,
    );
    expect(t).toEqual([
      {
        id: 'widget:immich',
        title: 'Immich',
        detail: 'Your photos.',
        keywords: ['fotos', 'Enter a key.'],
        kind: 'immich',
        url: 'https://d/immich/',
      },
    ]);
  });

  it('turns a page into a topic per page and section with links', () => {
    const t = docsTopics('https://d/', [
      {
        path: 'dashboard/',
        title: 'Tabs',
        summary: 'Tabs.',
        sections: [{ title: 'Zones', anchor: 'zones' }],
      },
    ]);
    expect(t.map((x) => x.url)).toEqual(['https://d/dashboard/', 'https://d/dashboard/#zones']);
    expect(t.every((x) => x.group === 'dashboard/')).toBe(true);
  });
});
