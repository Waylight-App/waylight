import type { ServerMessageKey } from '../../../api-client/generated/server-messages/birthdays';

/** German translations of the backend messages of `birthdays` (#287). */
export const serverDe: Record<ServerMessageKey, string> = {
  // At most 2000 birthdays can be stored.
  srv_5c7e3fc688c672ff: 'Es können höchstens 2000 Geburtstage gespeichert werden.',
  // Enter a valid date (day and month; the year is optional).
  srv_edef2011ce880fc5: 'Gib ein gültiges Datum ein (Tag und Monat; das Jahr ist optional).',
  // Enter the time as HH:MM.
  srv_e2cd3de79cef3345: 'Gib die Uhrzeit als HH:MM ein.',
  // The birthday does not exist (any more).
  srv_83cfa49329c60bec: 'Den Geburtstag gibt es nicht (mehr).',
  // The file contains no contacts with a birthday (vCard, .vcf).
  srv_703e13ba2a8ad5b7: 'Die Datei enthält keine Kontakte mit Geburtstag (vCard, .vcf).',
  // The name must have between 1 and 100 characters.
  srv_0ac00605de9a88c1: 'Der Name muss 1 bis 100 Zeichen lang sein.',
  // The vCard file must be at most 5 MB.
  srv_b6cfe14b52beec78: 'Die vCard-Datei darf höchstens 5 MB groß sein.',
  // This birthday comes from the connected address book; change or remove it there.
  srv_3200c6dd1c3ff975:
    'Dieser Geburtstag stammt aus dem verbundenen Adressbuch; ändere oder lösche ihn dort.',
  // Add birthdays in Settings under Birthdays, import a contacts file or connect an address book.
  srv_df3d1d090d91e1fa:
    'Füge Geburtstage in den Einstellungen unter Geburtstage hinzu, importiere eine Kontaktdatei oder verbinde ein Adressbuch.',
  // The next birthdays and how old everyone turns.
  srv_5b1ecee92864e5b0: 'Die nächsten Geburtstage und wie alt alle werden.',
};
