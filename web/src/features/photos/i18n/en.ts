/** English UI texts of the photo frame (#286). */
export const en = {
  'widget.photo_frame.title': 'Photo frame',
  'widget.photo_frame.description':
    'Your own pictures one after another – uploaded here, no photo server needed.',
  'photos.empty': 'No pictures in “{album}” yet. Add some in the widget settings.',
  'photos.picture_of': 'Picture {n} of {total}',
  'photos.album': 'Album',
  'photos.interval': 'Seconds per picture',
  'photos.fit': 'Fit',
  'photos.fit_contain': 'Show whole picture',
  'photos.fit_cover': 'Fill widget',
  'photos.pictures': 'Pictures in “{album}”',
  'photos.upload': 'Upload pictures …',
  'photos.uploading': 'Uploading {done} of {total} …',
  'photos.upload_failed': '{n} picture(s) not added, e.g. “{name}”: {reason}',
  'photos.move_earlier': 'Move picture {n} earlier',
  'photos.move_later': 'Move picture {n} later',
  'photos.delete': 'Delete picture {n}',
  'photos.confirm_delete': 'Delete this picture?',
  'photos.usage': '{used} of {quota} used · {count} of {max} pictures (all albums)',
  'photos.usage_label': 'Photo storage used',
  'photos.backup_note':
    'Pictures are scaled to 1920 px and stored without location data. Backups contain them, data exports do not.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
