<script lang="ts">
  import { intlLocale, t } from '../../../shared/i18n';
  import { errorMessage } from '../../../shared/api-client/http';
  import SetupDialog from '../../../shared/setup/SetupDialog.svelte';
  import SetupResult from '../../../shared/setup/SetupResult.svelte';
  import Icon from '../../../shared/ui/icons/Icon.svelte';
  import { useWidgetIcon } from '../../../shared/ui/widgetIcon';
  import type { CalendarFile } from '../domain/waste';
  import type { WasteUseCases } from '../usecase/wasteUseCases';

  /**
   * The guided setup of the waste calendar (#738): the link and the file side
   * by side, both visible, with the note on where to find a calendar. A file
   * is checked and saved when it is chosen; a link when Done is pressed.
   */
  interface Props {
    waste: WasteUseCases;
    ondone: () => void;
    onclose: () => void;
  }
  let { waste, ondone, onclose }: Props = $props();
  const uid = $props.id();
  const icon = useWidgetIcon();

  let url = $state('');
  let file = $state<CalendarFile | null>(null);
  let error = $state('');
  let busy = $state(false);

  const day = (d: Date) =>
    d.toLocaleDateString(intlLocale(), { day: 'numeric', month: 'short', year: 'numeric' });

  async function upload(chosen: File | undefined) {
    if (!chosen) return;
    error = '';
    busy = true;
    try {
      file = (await waste.uploadCalendar(chosen)).file;
    } catch (err) {
      file = null;
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  async function done() {
    error = '';
    if (!url.trim()) {
      // A file was saved when it was chosen.
      if (file) ondone();
      else error = t('waste.setup_empty');
      return;
    }
    busy = true;
    try {
      await waste.saveSettings(url, []);
      ondone();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      busy = false;
    }
  }

  // A file that was saved already is no reason to lose the dates when
  // the dialog closes.
  const leave = () => (file ? ondone() : onclose());
</script>

<SetupDialog
  title={t('waste.setup_title')}
  {icon}
  lead={t('waste.setup_lead')}
  primary={t('setup.done')}
  {busy}
  onprimary={() => void done()}
  onclose={leave}
>
  {#snippet help()}
    <strong>{t('waste.setup_help_title')}</strong><br />{t('waste.setup_help')}
  {/snippet}

  <div class="two">
    <section class="option" aria-labelledby="{uid}-link">
      <h3 id="{uid}-link">{t('waste.setup_link_title')}</h3>
      <p class="muted">{t('waste.setup_link_hint')}</p>
      <input
        class="input"
        type="url"
        bind:value={url}
        aria-label={t('waste.setup_link_label')}
        placeholder="https://…/abfuhrkalender.ics"
        autocomplete="off"
      />
    </section>
    <section class="option" aria-labelledby="{uid}-file">
      <h3 id="{uid}-file">{t('waste.setup_file_title')}</h3>
      <p class="muted">{t('waste.setup_file_hint')}</p>
      <label class="drop">
        <Icon name="upload" size={26} />
        <span>{file ? file.name : t('waste.setup_drop')}</span>
        <input
          class="file"
          type="file"
          accept=".ics,text/calendar"
          disabled={busy}
          onchange={(e) => void upload(e.currentTarget.files?.[0])}
        />
      </label>
    </section>
  </div>

  {#if file}
    <SetupResult
      ok
      lines={[
        t('waste.file_uploaded'),
        t('waste.file_range', { n: file.events, from: day(file.from), to: day(file.to) }),
      ]}
    />
  {:else if error}
    <SetupResult ok={false} lines={[error]} />
  {/if}
</SetupDialog>

<svelte:window
  ondragover={(e) => {
    // Without this the browser opens a dropped file instead of handing it over.
    if (e.dataTransfer?.types.includes('Files')) e.preventDefault();
  }}
  ondrop={(e) => {
    if (!e.dataTransfer?.types.includes('Files')) return;
    e.preventDefault();
    void upload(e.dataTransfer.files[0]);
  }}
/>

<style>
  .two {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-4);
  }
  @media (max-width: 640px) {
    .two {
      grid-template-columns: 1fr;
    }
  }
  .option {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding: var(--space-4);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-surface-2);
  }
  h3 {
    margin: 0;
    font-size: var(--fs-sm);
  }
  p {
    margin: 0;
    font-size: var(--fs-sm);
    line-height: 1.4;
  }
  .drop {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-4);
    border: 2px dashed var(--color-text-subtle);
    border-radius: var(--radius-md);
    text-align: center;
    cursor: pointer;
    overflow-wrap: anywhere;
  }
  .drop:hover,
  .drop:focus-within {
    border-color: var(--color-accent);
  }
  .drop:focus-within {
    outline: 2px solid var(--color-accent-text);
    outline-offset: 2px;
  }
  /* The input covers the area, so a click anywhere opens the file chooser. */
  .file {
    position: absolute;
    inset: 0;
    opacity: 0;
    cursor: pointer;
  }
</style>
