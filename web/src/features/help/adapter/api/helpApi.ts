import type { HttpClient } from '../../../../shared/api-client/http';
import type { HelpResponse, CompactText } from '../../../../shared/api-client/generated/modules';
import type { Index } from '../../../../shared/api-client/generated/docsindex';
import { serverMessage } from '../../../../shared/i18n';
import type { DocsPage, WidgetHelp } from '../../domain/help';
import type { HelpGateway } from '../../usecase/helpUseCases';

/** The text in the current language, falling back to the server's English. */
const text = (t: CompactText) => serverMessage(t.key, t.params) ?? t.text;

/** Reads the widget help and the documentation index from the server (#678). */
export class HelpApi implements HelpGateway {
  constructor(private readonly http: HttpClient) {}

  async widgets(): Promise<WidgetHelp[]> {
    const res = await this.http.get<HelpResponse>('/api/help/widgets');
    return res.widgets.map((w) => ({
      kind: w.kind,
      shows: text(w.shows),
      setUp: text(w.setUp),
      docsUrl: w.docsUrl,
    }));
  }

  async docs(): Promise<{ baseUrl: string; pages: DocsPage[] }> {
    return this.http.get<Index>('/api/help/docs');
  }
}
