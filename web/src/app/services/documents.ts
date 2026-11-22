import { DocumentsApi } from '../../features/documents/adapter/api/documentsApi';
import { DocumentsUseCases } from '../../features/documents/usecase/documentsUseCases';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    documents: DocumentsUseCases;
  }
}

export const service = defineService('documents', (http) => new DocumentsUseCases(new DocumentsApi(http)));
