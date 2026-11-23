<script lang="ts">
  import { t } from '../../../shared/i18n';
  import { copyToClipboard } from '../../../shared/ui/clipboard';
  import { qrDataUrl } from '../../../shared/ui/qr';

  /** A freshly created invitation or reset link, shown once (#296). */
  interface Props {
    title: string;
    url: string;
    /** When the link stops working, already formatted. */
    expires: string;
    ondone: () => void;
  }
  let { title, url, expires, ondone }: Props = $props();
  let copied = $state(false);
</script>

<div class="link" role="status">
  <p><strong>{title}</strong> {t('auth.link_shown_once', { expires })}</p>
  <img src={qrDataUrl(url, 4)} alt={t('auth.link_qr')} width="164" height="164" />
  <code>{url}</code>
  <p>
    <button class="btn" type="button" onclick={async () => (copied = await copyToClipboard(url))}
      >{copied ? t('auth.kiosk_copied') : t('auth.kiosk_copy')}</button
    >
    <button class="btn" type="button" onclick={ondone}>{t('auth.kiosk_done')}</button>
  </p>
</div>

<style>
  .link {
    padding: var(--space-3);
    border: 2px solid var(--color-accent);
    border-radius: var(--radius-md);
    margin-bottom: var(--space-3);
  }
  .link code {
    display: block;
    word-break: break-all;
    margin: var(--space-2) 0;
  }
</style>
