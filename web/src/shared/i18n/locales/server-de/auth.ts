import type { ServerMessageKey } from '../../../api-client/generated/server-messages/auth';

/**
 * German translations of the backend error messages of `auth` (#102, #193). The keys come
 * from the generated catalogue of the same name; the type makes a missing translation a
 * compile error. The English original is shown as a comment above each entry.
 */
export const serverDe: Record<ServerMessageKey, string> = {
  // The code is no longer valid. Start again.
  srv_11fa6fa4ace24e2a:
    'Konten dieser Organisation dürfen sich hier nicht anmelden. Bitte die Administration, ihr Verzeichnis zuzulassen.',
  srv_48b65f2a0b803859: 'Der Code ist nicht mehr gültig. Bitte neu starten.',
  // This Microsoft account is not linked to an account here. Sign in with your password and connect it under Settings → Account.
  srv_d0088f948a0bcde7:
    'Dieses Microsoft-Konto ist hier mit keinem Konto verknüpft. Melde dich mit deinem Passwort an und verbinde es unter Einstellungen → Konto.',
  // This sign-in already belongs to an account here. Sign in with it instead.
  srv_cfe702755e13d3ac: 'Diese Anmeldung gehört hier schon zu einem Konto. Melde dich damit an.',
  // This sign-in is the only way into the account. Set a password first, for example with a reset link.
  srv_f70a5614e4b6564d:
    'Diese Anmeldung ist der einzige Zugang zum Konto. Setze zuerst ein Passwort, zum Beispiel mit einem Reset-Link.',
  // Single sign-on for this account needs a licence key with sso.
  srv_b4096594b09b5ea3:
    'Die Anmeldung per Single Sign-on braucht für dieses Konto einen Lizenzschlüssel mit sso.',
  // This sign-in is already linked to another account.
  srv_6e04252984abe6fe: 'Diese Anmeldung ist schon mit einem anderen Konto verknüpft.',
  // This sign-in is not linked.
  srv_1abbaf60db5b22ae: 'Diese Anmeldung ist nicht verknüpft.',
  // A guest with this name already exists.
  srv_083c4fe6f2b25f45: 'Es gibt bereits einen Gast mit diesem Namen.',
  // Choose at least one scope.
  srv_c2c503807749b722: 'Bitte mindestens einen Bereich wählen.',
  // Choose between 0 (never) and 3650 days.
  srv_0054eff4d1acb233: 'Bitte zwischen 0 (nie) und 3650 Tagen wählen.',
  // Enter your authentication code to finish signing in.
  srv_1eb555800ce2927d: 'Gib deinen Authentifizierungscode ein, um die Anmeldung abzuschließen.',
  // Guests can only view this dashboard.
  srv_79bf20d17f6aa6eb: 'Gäste können dieses Dashboard nur ansehen.',
  // Invalid scope {scope}: use module:read or module:write.
  srv_8648b1ab32906623: 'Ungültiger Bereich {scope}: bitte modul:read oder modul:write verwenden.',
  // Missing or invalid CSRF token.
  srv_eccd19aa12e79824: 'CSRF-Token fehlt oder ist ungültig.',
  // No account has been set up yet.
  srv_a53afcf337b06aea: 'Es wurde noch kein Konto eingerichtet.',
  // Only an admin can do this.
  srv_3a7ce116148b4992: 'Das kann nur ein Admin.',
  // Passkeys need HTTPS and PUBLIC_URL (see docs/accounts.md).
  srv_a9f3222a9472a3de: 'Passkeys brauchen HTTPS und PUBLIC_URL (siehe docs/accounts.md).',
  // Password login is disabled. Use single sign-on.
  srv_fd5d4b3fb31aa85f: 'Die Anmeldung mit Passwort ist deaktiviert. Bitte Single Sign-on verwenden.',
  // Please enter the code from your authenticator app or a recovery code.
  srv_2c825edcb1338a1c:
    'Bitte gib den Code aus deiner Authenticator-App oder einen Wiederherstellungscode ein.',
  // Session not found.
  srv_785eadcc6dca712c: 'Sitzung nicht gefunden.',
  // Single sign-on is not configured.
  srv_c1cbece9dc9bd483: 'Single Sign-on ist nicht eingerichtet.',
  // Start the two-factor setup first.
  srv_32f91240e6490e41: 'Bitte zuerst die Einrichtung der Zwei-Faktor-Authentifizierung starten.',
  // The account has already been set up.
  srv_ef560ac46442ad62: 'Das Konto wurde bereits eingerichtet.',
  // The account must be set up first.
  srv_86b531f1833dda66: 'Das Konto muss zuerst eingerichtet werden.',
  // The authentication code is incorrect.
  srv_1f91f70c0f8f35da: 'Der Authentifizierungscode ist falsch.',
  // The expiry must be in the future.
  srv_d08d3ed2e05c6dbc: 'Das Ablaufdatum muss in der Zukunft liegen.',
  // The identity provider could not be reached: {detail}
  srv_62951ebf93a06287: 'Der Identity Provider ist nicht erreichbar: {detail}',
  // The identity provider did not confirm the sign-in: {detail}
  srv_af05df00dfc4dc1a: 'Der Identity Provider hat die Anmeldung nicht bestätigt: {detail}',
  // The name must be 2–32 lowercase letters, digits, dots, dashes or underscores.
  srv_bd0d70cd47d150cb:
    'Der Name muss aus 2–32 Kleinbuchstaben, Ziffern, Punkten, Binde- oder Unterstrichen bestehen.',
  // The name must be at most 60 characters long.
  srv_7a44bdf0298cc83f: 'Der Name darf höchstens 60 Zeichen lang sein.',
  // The owner account is not a guest.
  srv_a8aed831e812c028: 'Das Besitzerkonto ist kein Gast.',
  // The passkey does not exist.
  srv_30c99e5e8d1ae003: 'Der Passkey existiert nicht.',
  // The passkey request expired. Please try again.
  srv_754fe73fc055a801: 'Die Passkey-Anfrage ist abgelaufen. Bitte erneut versuchen.',
  // The password is incorrect.
  srv_bed2be84c7fc7252: 'Das Passwort ist falsch.',
  // The password must be at least 12 characters long.
  srv_65fb602b5cafaba0: 'Das Passwort muss mindestens 12 Zeichen lang sein.',
  // The password must be at most 256 characters long.
  srv_91216e078745b556: 'Das Passwort darf höchstens 256 Zeichen lang sein.',
  // The password must not consist of whitespace only.
  srv_be9e30d725fd25b6: 'Das Passwort darf nicht nur aus Leerzeichen bestehen.',
  // The session does not exist (anymore).
  srv_869ef1a450e9ad99: 'Die Sitzung existiert nicht (mehr).',
  // The setup token is invalid.
  srv_d08501b212d94717: 'Das Setup-Token ist ungültig.',
  // The sign-in took too long or was started elsewhere. Please try again.
  srv_a9878acf59444198:
    'Die Anmeldung hat zu lange gedauert oder wurde woanders begonnen. Bitte erneut versuchen.',
  // The token does not exist.
  srv_10c62f2d6f8f250e: 'Das Token existiert nicht.',
  // The token is not allowed to do this.
  srv_1399b4677ad575ae: 'Das Token darf das nicht.',
  // There are already 10 passkeys.
  srv_7c2cacdb15125f01: 'Es gibt bereits 10 Passkeys.',
  // There are already 20 guests.
  srv_07a824f1534514ba: 'Es gibt bereits 20 Gäste.',
  // There are already 50 tokens.
  srv_f6aacc9caebd94cc: 'Es gibt bereits 50 Tokens.',
  // This account is disabled.
  srv_5b702bf7f69da54c: 'Dieses Konto ist deaktiviert.',
  // This account is not allowed to sign in here.
  srv_bd7029b4a21fcba0: 'Dieses Konto darf sich hier nicht anmelden.',
  // This passkey could not be verified.
  srv_ad0a8083bfc3f124: 'Dieser Passkey konnte nicht überprüft werden.',
  // This passkey is already registered.
  srv_52d3682de07001b1: 'Dieser Passkey ist bereits registriert.',
  // This passkey reported an old signature counter and may have been cloned. Remove it and register it again.
  srv_f30e4d83e431610d:
    'Dieser Passkey hat einen alten Signaturzähler gemeldet und wurde möglicherweise geklont. Entferne ihn und registriere ihn neu.',
  // Too many failed attempts. Please wait a few minutes.
  srv_bc7172983106e8f0: 'Zu viele Fehlversuche. Bitte warte ein paar Minuten.',
  // Two-factor authentication is already enabled.
  srv_a9954572b8880ec0: 'Die Zwei-Faktor-Authentifizierung ist bereits aktiviert.',
  // Two-factor authentication is not enabled.
  srv_326d279028bf3d2f: 'Die Zwei-Faktor-Authentifizierung ist nicht aktiviert.',
  // Your session has expired. Please log in again.
  srv_a1f313de1861076a: 'Deine Sitzung ist abgelaufen. Bitte melde dich erneut an.',
  // id and password are required
  srv_d036d6830bf49051: 'ID und Passwort sind erforderlich',
  // id and type are required
  srv_b672e981cecb99fe: 'ID und Typ sind erforderlich',
  // incomplete credential
  srv_b27eab1b501a12b6: 'unvollständige Zugangsdaten',
  // must be base64url
  srv_0d6f4bdef976a82d: 'muss base64url-kodiert sein',
  // must be positive
  srv_2199d11679859f04: 'muss positiv sein',
  // must not be empty
  srv_705d19e8395a99a3: 'darf nicht leer sein',
  // token, public id and csrf token are required
  srv_1b3daa6bc9a9f8db: 'Token, öffentliche ID und CSRF-Token sind erforderlich',
  // At most 10 kiosk displays are supported.
  srv_d73d8c1a54199408: 'Es werden höchstens 10 Kiosk-Anzeigen unterstützt.',
  // Choose at least one tab for the display.
  srv_5ea683077504d129: 'Wähle mindestens einen Tab für die Anzeige.',
  // Night dimming needs a start and an end such as 22:00 and 06:30.
  srv_7b34c09128db40de: 'Das Nachtdimmen braucht einen Beginn und ein Ende, z. B. 22:00 und 06:30.',
  // Rotate every 10 seconds to 1 hour, or never.
  srv_11e7a84aac8d61b2: 'Wechsle alle 10 Sekunden bis 1 Stunde oder nie.',
  // The kiosk display does not exist.
  srv_57c1f826d1469252: 'Die Kiosk-Anzeige gibt es nicht.',
  // This kiosk link is not valid (any more). Ask for a new one.
  srv_d5825321d90da76c: 'Dieser Kiosk-Link ist (nicht mehr) gültig. Lass dir einen neuen geben.',
  // id is required
  srv_272f687c666a4c51: 'Die ID ist erforderlich',
  // The name this-device is taken by a guest.
  srv_9737b26eae1b7ac5: 'Der Name this-device ist von einem Gast belegt.',
  // A group with this name already exists.
  srv_ed9c5c61fef0f9b0: 'Es gibt schon eine Gruppe mit diesem Namen.',
  // Members are not available yet: personal content is not separated per account in this version.
  srv_381a488cb15b35ce:
    'Mitglieder sind noch nicht verfügbar: Persönliche Inhalte werden in dieser Version noch nicht pro Konto getrennt.',
  // Members cannot change this; ask an admin.
  srv_d8f679283dba6727: 'Mitglieder können das nicht ändern; bitte einen Admin fragen.',
  // The group does not exist.
  srv_aa197fc26a5b3a24: 'Die Gruppe gibt es nicht.',
  // The group name must be 1–40 characters.
  srv_69b3e544a59f8d98: 'Der Gruppenname muss 1–40 Zeichen lang sein.',
  // The member does not exist.
  srv_d7f0b2abe0b557e5: 'Das Mitglied gibt es nicht.',
  // There are already 50 groups.
  srv_ff45a7ead80407b2: 'Es gibt bereits 50 Gruppen.',
  // There are already 50 members.
  srv_7396c7b16437339f: 'Es gibt bereits 50 Mitglieder.',
  // An invitation is valid for 1 to 30 days.
  srv_7446a4ecf0cb75e6: 'Eine Einladung gilt 1 bis 30 Tage.',
  // At most 20 invitations can be open at the same time.
  srv_fb200f2e8942d8e7: 'Es können höchstens 20 Einladungen gleichzeitig offen sein.',
  // Invite a member or a guest.
  srv_c9d91e1ff5289f27: 'Lade ein Mitglied oder einen Gast ein.',
  // A member can be a member or an admin.
  srv_d6a9ea1bfc72e12b: 'Ein Mitglied kann Mitglied oder Admin sein.',
  // An admin's password is set by the owner or another admin.
  srv_44205add2ef753e0: 'Das Passwort eines Admins setzt der Besitzer oder ein anderer Admin.',
  // Only the owner can do this.
  srv_67f0ba90e3a4e5b4: 'Das kann nur der Besitzer.',
  // You cannot do this to your own account; ask the owner or another admin.
  srv_1553d503b1d708f8:
    'Das geht nicht mit deinem eigenen Konto; frag den Besitzer oder einen anderen Admin.',
  // The admin password is reset on the server only (see docs/accounts.md).
  srv_2d855ec17092911d: 'Das Admin-Passwort wird nur auf dem Server zurückgesetzt (siehe docs/accounts.md).',
  // This account has no password.
  srv_920db87f11531520: 'Dieses Konto hat kein Passwort.',
  // This link is not valid (any more). Ask for a new one.
  srv_54c513952f5013f1: 'Dieser Link ist (nicht mehr) gültig. Lass dir einen neuen geben.',
  // This name is already used by an account or an open invitation.
  srv_6330f6a3d5db3d75: 'Dieser Name gehört schon zu einem Konto oder einer offenen Einladung.',
  // id and token are required
  srv_a7f0df8ff2599767: 'ID und Token sind erforderlich',
  // Password reset by e-mail is not set up on this instance. Ask the admin for a reset link.
  srv_3a24b1058006ec4b:
    'Das Zurücksetzen per E-Mail ist auf dieser Instanz nicht eingerichtet. Bitte den Admin um einen Reset-Link.',
  // Set up a second factor … your administrator requires one.
  srv_648748a3f866984b:
    'Richte einen zweiten Faktor ein (eine Authenticator-App oder einen Passkey), um weiterzumachen. Dein Admin verlangt das.',
  // The password must be at least {min} characters long.
  srv_15140a70901d43cd: 'Das Passwort muss mindestens {min} Zeichen lang sein.',
  // This account cannot pair an app.
  srv_ff3dc20386a5f761: 'Dieses Konto kann keine App koppeln.',
  // This pairing code is not valid (any more). Show a new one in the settings.
  srv_872a2ab087913d2c:
    'Dieser Kopplungscode ist nicht (mehr) gültig. Lass dir in den Einstellungen einen neuen anzeigen.',
  // hash and account are required
  srv_1e96e5b609412e0e: 'Hash und Konto sind erforderlich',
  // must be 1–80 characters
  srv_3a9987390ce3d83f: 'muss 1–80 Zeichen lang sein',
  // must be an S256 PKCE code challenge
  srv_fbb05d449c473821: 'muss eine S256-PKCE-Code-Challenge sein',
};
