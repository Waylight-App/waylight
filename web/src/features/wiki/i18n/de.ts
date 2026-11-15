import type { en } from './en';

/** German texts of the wiki search; every key is mandatory. */
export const de: Record<keyof typeof en, string> = {
  'settings.section.wiki': 'Wiki-Suche',
  'app.group.wiki': 'Wiki',
  'wiki.intro':
    'Finde Wiki-Seiten über das Suchfeld. trisse-home fragt das Wiki beim Tippen und speichert keine Seiteninhalte. Mit „w “ am Anfang suchst du nur in den Wikis.',
  'wiki.shared_token':
    'Alle, die diese Startseite nutzen, sehen, was das API-Token lesen darf. Nimm ein Token eines Kontos, das nur liest, was ihr alle sehen dürft.',
  'wiki.none': 'Noch kein Wiki verbunden.',
  'wiki.add': 'Wiki verbinden',
  'wiki.name': 'Name',
  'wiki.kind': 'Wiki-Software',
  'wiki.url': 'Adresse des Wikis',
  'wiki.token': 'API-Token',
  'wiki.token_stored': 'gespeichert – leer lassen, um es zu behalten',
  'wiki.token_hint_bookstack':
    'In BookStack: dein Profil → API-Tokens. Trage es als „Token-ID:Token-Secret“ ein.',
  'wiki.token_hint_outline': 'In Outline: Einstellungen → API & Apps → Neuer API-Schlüssel.',
  'wiki.token_hint_wikijs':
    'In Wiki.js: Administration → API-Zugriff → Neuer API-Schlüssel (Lesezugriff reicht).',
  'wiki.test': 'Testen',
  'wiki.test_ok': 'Das Wiki hat geantwortet.',
  'wiki.saved': 'Gespeichert.',
  'wiki.delete': 'Entfernen',
  'wiki.failed': 'Keine Antwort von {names}.',
};
