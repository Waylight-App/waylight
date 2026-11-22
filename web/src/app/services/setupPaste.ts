import { PasteApi, type PasteGateway } from '../../shared/setup/paste';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    setupPaste: PasteGateway;
  }
}

// Paste from phone in the setup assistant (#618).
export const service = defineService('setupPaste', (http) => new PasteApi(http));
