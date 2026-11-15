<script lang="ts">
  import { t } from '../i18n';
  import { groupVisibility, type ShareGroup, type Visibility } from '../sharing/sharing';

  /** Props of the "who can see this" choice of a record (ADR 0018, #301). */
  interface Props {
    value: Visibility;
    groups: ShareGroup[];
    onchange: (v: Visibility) => void;
    disabled?: boolean;
  }
  let { value, groups, onchange, disabled = false }: Props = $props();
</script>

<label class="field visibility"
  ><span>{t('sharing.label')}</span>
  <select class="input" {value} {disabled} onchange={(e) => onchange(e.currentTarget.value)}>
    <option value="private">{t('sharing.private')}</option>
    {#each groups as g (g.id)}
      <option value={groupVisibility(g.id)}>{t('sharing.group', { name: g.name })}</option>
    {/each}
    <option value="everyone">{t('sharing.everyone')}</option>
  </select>
</label>
