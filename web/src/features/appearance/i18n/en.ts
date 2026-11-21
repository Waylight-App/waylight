/** English UI texts of the background picker (#277). */
export const en = {
  'appearance.background_image': 'Background',
  'appearance.background_choose': 'Choose…',
  'appearance.background_title': 'Background',
  'appearance.background_scope': 'Where the background applies',
  'appearance.background_scope_global': 'All tabs',
  'appearance.background_scope_tab': 'Only “{name}”',
  'appearance.background_none_tab': 'Uses the background of all tabs',
  'appearance.background_default': 'Default background',
  'appearance.background_overridden': '“{name}” has its own background, which is shown instead of this one.',
  'appearance.background_clear_tab': 'Remove the background of “{name}” so this one shows there too',
  'appearance.background_builtin': 'Built in',
  'appearance.background_builtin_dusk': 'Dusk',
  'appearance.background_builtin_dawn': 'Dawn',
  'appearance.background_builtin_forest': 'Forest',
  'appearance.background_builtin_ocean': 'Ocean',
  'appearance.background_builtin_aurora': 'Aurora',
  'appearance.background_builtin_sand': 'Sand',
  'appearance.background_builtin_slate': 'Slate',
  'appearance.background_builtin_paper': 'Paper',
  'appearance.background_uploads': 'Your pictures',
  'appearance.background_upload': 'Upload a picture',
  'appearance.background_upload_pick': 'Your picture, {w} × {h}',
  'appearance.background_upload_hint':
    'JPEG, PNG or GIF up to 25 MB. Stored on this server, scaled to 2560 px and without location data; backups contain them.',
  'appearance.background_delete': 'Delete picture',
  'appearance.background_delete_confirm':
    'Delete this picture? Where it is used, the background falls back to the default.',
  'appearance.background_url': 'Image from the web',
  'appearance.background_use_url': 'Use',
  'appearance.background_reset': 'Default background',
  'appearance.background_reset_tab': 'Use the background of all tabs',
  'appearance.background_invalid': 'Enter an http(s) address of an image.',
  'appearance.background_too_large': 'The picture is larger than 25 MB.',
  'dashboard.background_button': 'Background',
  'appearance.accessibility': 'Accessibility',
  'appearance.text_size': 'Text size',
  'appearance.text_size_normal': 'Normal',
  'appearance.text_size_large': 'Large',
  'appearance.text_size_larger': 'Larger',
  'appearance.high_contrast': 'High contrast',
  'appearance.high_contrast_hint':
    'Black and white with strong colours, for light and dark. Replaces the theme, accent colour and background while it is on.',
  'appearance.simple_view': 'Simple view',
  'appearance.simple_view_hint':
    'One column with the short state of each widget, larger buttons and no edit mode. Switch it off here to edit the dashboard.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
