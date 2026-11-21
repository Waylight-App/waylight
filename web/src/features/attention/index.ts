/** Public API of the attention line feature (#593). */
export { AttentionApi } from './adapter/api/attentionApi';
export { AttentionUseCases } from './usecase/attentionUseCases';
export type { AttentionSource, SourceState } from './domain/attention';
export { default as AttentionLine } from './ui/AttentionLine.svelte';
