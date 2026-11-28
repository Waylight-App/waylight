<script lang="ts">
  import Icon from './icons/Icon.svelte';
  import { t } from '../i18n';
  import { sharedState, type ShareGroup, type Visibility } from '../sharing/sharing';

  /** Props of the mark that tells who shares a record in a widget (#872). */
  interface Props {
    visibility: Visibility;
    groups: ShareGroup[];
  }
  let { visibility, groups }: Props = $props();
  const mark = $derived(sharedState(visibility, groups));
</script>

{#if mark}
  <span class="shared muted small"
    ><Icon name="users" size={16} />{'group' in mark
      ? t('sharing.shared_with_group', { group: mark.group })
      : 'everyone' in mark
        ? t('sharing.shared_with_everyone')
        : t('sharing.shared_with_a_group')}</span
  >
{/if}

<style>
  .shared {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    min-width: 0;
  }
</style>
