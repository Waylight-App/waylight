<script lang="ts">
  import { t, type MessageKey } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import Modal from '../../../shared/ui/Modal.svelte';
  import type { BuiltinTemplate, Tab, TabTemplate } from '../domain/dashboard';
  import type { DashboardUseCases } from '../usecase/dashboardUseCases';

  /** Props of the dialog that creates a tab, optionally from a template. */
  interface Props {
    uc: DashboardUseCases;
    oncreated: (tab: Tab) => void;
    onclose: () => void;
  }
  let { uc, oncreated, onclose }: Props = $props();

  let name = $state('');
  let templateId = $state('');
  let templates = $state<TabTemplate[]>([]);
  let builtins = $state<BuiltinTemplate[]>([]);
  let error = $state('');

  // Built-in templates (#589) are offered next to the account's own; their
  // ids carry a prefix so the two never collide.
  const BUILTIN = 'builtin:';
  const builtinTitle = (tpl: BuiltinTemplate) => {
    const key = `templates.${tpl.id}` as MessageKey;
    const text = t(key);
    return text === key ? tpl.title : text;
  };

  $effect(() => {
    uc.listTemplates()
      .then((t) => (templates = t))
      .catch(() => {}); // the dialog works without templates
    uc.listBuiltinTemplates()
      .then((b) => (builtins = b))
      .catch(() => {});
  });

  function create(): Promise<Tab> {
    if (templateId.startsWith(BUILTIN)) {
      const tpl = builtins.find((b) => BUILTIN + b.id === templateId);
      return uc.createTabFromBuiltin(
        templateId.slice(BUILTIN.length),
        name || (tpl ? builtinTitle(tpl) : ''),
      );
    }
    return templateId ? uc.createTabFromTemplate(templateId, name) : uc.createTab(name);
  }

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    try {
      const tab = await create();
      oncreated(tab);
      onclose();
    } catch (err) {
      error = errorMessage(err);
    }
  }

  async function remove(tpl: TabTemplate) {
    if (!confirm(t('dashboard.confirm_delete_template', { name: tpl.name }))) return;
    try {
      await uc.deleteTemplate(tpl.id);
      templates = templates.filter((x) => x.id !== tpl.id);
      if (templateId === tpl.id) templateId = '';
    } catch (err) {
      error = errorMessage(err);
    }
  }
</script>

<Modal title={t('dashboard.new_tab')} {onclose}>
  <form onsubmit={submit}>
    <label class="field">
      <span>{t('dashboard.name_2', { value: templateId ? t('dashboard.empty_template_name') : '' })}</span>
      <!-- svelte-ignore a11y_autofocus -->
      <input class="input" bind:value={name} autofocus required={!templateId} maxlength="40" />
    </label>
    {#if templates.length || builtins.length}
      <fieldset>
        <legend>{t('dashboard.start_from')}</legend>
        <label><input type="radio" bind:group={templateId} value="" /> {t('dashboard.empty_tab')}</label>
        {#each builtins as tpl (tpl.id)}
          <label
            ><input type="radio" bind:group={templateId} value={BUILTIN + tpl.id} />
            {builtinTitle(tpl)}
            <span class="muted">{t('dashboard.widgets', { widgets: tpl.widgets.length })}</span></label
          >
        {/each}
        {#each templates as tpl (tpl.id)}
          <div class="tpl">
            <label
              ><input type="radio" bind:group={templateId} value={tpl.id} />
              {tpl.name} <span class="muted">{t('dashboard.widgets', { widgets: tpl.widgets })}</span></label
            >
            <button
              class="btn btn-icon"
              type="button"
              aria-label={t('dashboard.delete_template', { name: tpl.name })}
              onclick={() => remove(tpl)}>✕</button
            >
          </div>
        {/each}
      </fieldset>
    {/if}
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <button class="btn btn-primary" type="submit">{t('dashboard.create')}</button>
  </form>
</Modal>

<style>
  fieldset {
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
    display: grid;
    gap: var(--space-1);
    margin: var(--space-2) 0 var(--space-3);
  }
  .tpl {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
</style>
