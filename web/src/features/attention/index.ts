/** Public API of the attention feature (#592). */
export { AttentionApi } from './adapter/api/attentionApi';
export { AttentionUseCases } from './usecase/attentionUseCases';
export { AttentionState } from './ui/attentionState.svelte';
export { actionTarget, type AttentionItem, type AttentionTarget } from './domain/attention';
export { default as AttentionPage } from './ui/AttentionPage.svelte';
export { default as AttentionBadge } from './ui/AttentionBadge.svelte';
