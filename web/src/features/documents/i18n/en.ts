/** English UI texts of the documents feature (#688). */
export const en = {
  'widget.documents.title': 'Warranty and documents',
  'widget.documents.description':
    'Receipts and warranty papers with the end of each warranty; photos and PDFs stay on your server.',
  'compact.example.documents': 'Headphones: warranty ends 12 Dec',
  'documents.new': 'New document',
  'documents.edit': 'Change document',
  'documents.title': 'Product',
  'documents.purchased_on': 'Purchased on',
  'documents.warranty_end': 'Warranty ends on',
  'documents.note': 'Note',
  'documents.file': 'Photo or PDF (up to 10 MB)',
  'documents.file_replace': 'Choose a new file to replace {name}.',
  'documents.status.running': 'Warranty until {date}',
  'documents.status.ending': 'Warranty ends on {date}',
  'documents.status.ended': 'Warranty ended on {date}',
  'documents.status.none': 'No warranty entered',
  'documents.empty': 'No documents yet.',
  'documents.download': 'Download {name}',
  'documents.edit_named': 'Change {name}',
  'documents.remove': 'Remove {name}',
  'documents.error.title': 'Enter a product name of up to 80 characters.',
  'documents.error.dates': 'The warranty cannot end before the purchase.',
  'documents.error.file_type': 'Choose a photo (JPEG, PNG or WebP) or a PDF.',
  'documents.error.file_size': 'The file must not be empty or larger than 10 MB.',
} as const;

declare module '../../../shared/i18n/registry' {
  interface FeatureMessages extends Record<keyof typeof en, string> {}
}
