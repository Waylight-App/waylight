import { describe, expect, it, vi } from 'vitest';
import type { HttpClient } from '../../../../shared/api-client/http';
import { AttentionApi } from './attentionApi';

describe('AttentionApi (#593)', () => {
  it('maps the items of GET /api/attention', async () => {
    const get = vi.fn().mockResolvedValue({
      count: 1,
      items: [
        {
          id: 'conn:immich',
          kind: 'connection',
          integration: 'immich',
          title: 'Immich',
          reason: { message: 'The server did not answer.', messageKey: 'srv_unknown' },
          since: '2026-11-21T10:00:00Z',
          action: 'check',
        },
      ],
    });
    const items = await new AttentionApi({ get } as unknown as HttpClient).list();
    expect(get).toHaveBeenCalledWith('/api/attention');
    expect(items).toEqual([
      { id: 'conn:immich', kind: 'connection', title: 'Immich', reason: 'The server did not answer.' },
    ]);
  });
});
