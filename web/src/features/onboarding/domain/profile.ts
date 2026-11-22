import type { ThemeMode } from '../../../shared/theme/theme';

/** The personal settings the wizard asks; they live in the appearance settings. */
export interface Personal {
  greetingName: string;
  language: '' | 'en' | 'de';
  theme: ThemeMode;
  accentColor: string;
}

/** A saved place (#591) as the wizard and the template choice need it. */
export interface Place {
  id: string;
  name: string;
  label: string;
  lat: number;
  lon: number;
}
