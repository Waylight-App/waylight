// Design guide: font sizes and radii come from the tokens (#755).
const FONT_SIZE = [
  /^var\(--fs-(xs|sm|md|lg)\)$/,
  'inherit',
  '0',
  /^clamp\(.*cq/, // fluid sizes of widget content, scaled by the widget's container
  'calc(100% * var(--text-scale))', // the root size, set once in base.css
];
const RADIUS = [/^(var\(--radius-(sm|md)\)|0|50%|999px|inherit)( |$)/, /^0 0 var\(--radius-(sm|md)\)/];

export default {
  rules: {
    'declaration-property-value-allowed-list': {
      'font-size': FONT_SIZE,
      'border-radius': RADIUS,
    },
  },
  overrides: [
    { files: ['**/*.svelte'], customSyntax: 'postcss-html' },
    // The tokens themselves hold the raw values.
    { files: ['src/shared/theme/tokens.css'], rules: { 'declaration-property-value-allowed-list': null } },
  ],
};
