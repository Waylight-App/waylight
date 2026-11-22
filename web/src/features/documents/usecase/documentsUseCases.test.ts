import { describe, expect, it, vi } from 'vitest';
import type { Doc } from '../domain/documents';
import { DocumentsUseCases, type DocumentsGateway } from './documentsUseCases';

const doc = { id: 'd1' } as Doc;
const input = { title: ' Kettle ', purchasedOn: '', warrantyEnd: '2027-03-31', note: '' };

const gateway = (): DocumentsGateway => ({
  list: vi.fn(),
  create: vi.fn().mockResolvedValue(doc),
  update: vi.fn().mockResolvedValue(doc),
  remove: vi.fn(),
  attachFile: vi.fn().mockResolvedValue(doc),
  fileUrl: vi.fn(),
  shareGroups: vi.fn(),
});

describe('DocumentsUseCases', () => {
  it('creates, then attaches the chosen file', async () => {
    const g = gateway();
    const file = new File(['%PDF'], 'receipt.pdf', { type: 'application/pdf' });
    await new DocumentsUseCases(g).save(null, input, file);
    expect(g.create).toHaveBeenCalledWith(expect.objectContaining({ title: 'Kettle' }), undefined);
    expect(g.attachFile).toHaveBeenCalledWith('d1', file);
  });

  it('changes an entry without touching its file', async () => {
    const g = gateway();
    await new DocumentsUseCases(g).save('d1', input, null, 'everyone');
    expect(g.update).toHaveBeenCalledWith('d1', expect.objectContaining({ title: 'Kettle' }), 'everyone');
    expect(g.attachFile).not.toHaveBeenCalled();
  });
});
