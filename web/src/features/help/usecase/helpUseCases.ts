import {
  type DocsPage,
  docsTopics,
  type HelpTopic,
  matchTopics,
  widgetTopics,
  type WidgetHelp,
} from '../domain/help';

/** Output port: where the help comes from. */
export interface HelpGateway {
  widgets(): Promise<WidgetHelp[]>;
  docs(): Promise<{ baseUrl: string; pages: DocsPage[] }>;
}

/** How a widget kind is named in the interface. */
export type DescribeKind = (kind: string) => { title: string; keywords: string[] } | undefined;

/**
 * The in-app help (#678). Both lists are loaded once per page on first use;
 * a failed load is retried on the next call.
 */
export class HelpUseCases {
  private widgetHelp?: Promise<WidgetHelp[]>;
  private docsIndex?: Promise<HelpTopic[]>;

  constructor(private readonly gateway: HelpGateway) {}

  /** The help of every widget kind the account can see. */
  widgets(): Promise<WidgetHelp[]> {
    this.widgetHelp ??= this.gateway.widgets().catch((err: unknown) => {
      this.widgetHelp = undefined;
      throw err;
    });
    return this.widgetHelp;
  }

  /** The help of one widget kind, or undefined when the server has none. */
  async widget(kind: string): Promise<WidgetHelp | undefined> {
    return (await this.widgets()).find((w) => w.kind === kind);
  }

  /**
   * The topics that answer a question: widget help first, then the
   * documentation. A missing documentation index leaves only the widgets.
   */
  async answer(question: string, describe: DescribeKind, limit = 4): Promise<HelpTopic[]> {
    const [widgets, docs] = await Promise.all([this.widgets().catch(() => []), this.docs().catch(() => [])]);
    const fromWidgets = matchTopics(question, widgetTopics(widgets, describe), limit);
    const fromDocs = matchTopics(question, docs, limit);
    return [...fromWidgets, ...fromDocs].slice(0, limit);
  }

  private docs(): Promise<HelpTopic[]> {
    this.docsIndex ??= this.gateway
      .docs()
      .then((d) => docsTopics(d.baseUrl, d.pages))
      .catch((err: unknown) => {
        this.docsIndex = undefined;
        throw err;
      });
    return this.docsIndex;
  }
}
