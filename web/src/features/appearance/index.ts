/** Public API of the appearance feature. */
export { AppearanceApi } from './adapter/api/appearanceApi';
export { AppearanceUseCases } from './usecase/appearanceUseCases';
export {
  type Appearance,
  accessibilityAttributes,
  cssVariables,
  defaultAppearance,
} from './domain/appearance';
export { default as AppearanceSettings } from './ui/AppearanceSettings.svelte';
