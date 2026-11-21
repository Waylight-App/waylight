import { SignInApi, type SignInGateway } from '../../shared/setup/signin';
import { defineService } from '../serviceKit';

declare module '../serviceKit' {
  interface Services {
    setupSignin: SignInGateway;
  }
}

// Signing in with Nextcloud and Home Assistant from the setup assistant (#654).
export const service = defineService('setupSignin', (http) => new SignInApi(http));
