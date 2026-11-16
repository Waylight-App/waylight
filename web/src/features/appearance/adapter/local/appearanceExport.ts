import { EXPORT_OWNER, type ExportTables } from '../../../../shared/local/exportRows';
import type { Appearance } from '../../domain/appearance';

/**
 * The appearance as a row of appearance_settings (migrations of
 * internal/modules/appearance, per account since 0008). Uploaded pictures
 * and themes only exist on a server, so their tables stay empty.
 */
export function appearanceExport(a: Appearance): ExportTables {
  return {
    appearance_settings: [
      {
        account_id: EXPORT_OWNER,
        theme: a.theme,
        accent_color: a.accentColor,
        background_image: a.backgroundImage,
        custom_css: a.customCss,
        language: a.language,
        greeting_name: a.greetingName,
        launcher: a.launcher,
        profile: a.profile,
      },
    ],
  };
}
