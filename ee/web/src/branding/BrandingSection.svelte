<script lang="ts">
  import type { PluginContext } from "../../../../web/src/app/plugins/api";
  import { text } from "../lib/texts";
  import { assetKinds, BrandingApi, imageProblem, type AssetKind, type BrandingSettings } from "./branding";
  import { texts } from "./texts";

  /** Settings → Branding (#305): name, colours and images of the instance. */
  interface Props {
    context: PluginContext;
  }
  let { context }: Props = $props();

  const api = $derived(new BrandingApi(context.http));
  const tr = (key: string, params: Record<string, string | number> = {}) => text(texts, key, context.language, params);
  const readOnly = $derived(context.level("branding") === "read");

  let s = $state<BrandingSettings | null>(null);
  let useAccent = $state(false);
  let useSecondary = $state(false);
  let error = $state("");
  let message = $state("");

  function show(v: BrandingSettings) {
    s = v;
    useAccent = !!v.accent;
    useSecondary = !!v.secondary;
  }

  $effect(() => {
    api
      .settings()
      .then(show)
      .catch((err) => (error = context.errorMessage(err)));
  });

  async function save(e: SubmitEvent) {
    e.preventDefault();
    if (!s) return;
    error = message = "";
    try {
      show(
        await api.save({
          name: s.name,
          accent: useAccent ? s.accent || "#2563eb" : "",
          secondary: useSecondary ? s.secondary || "#64748b" : "",
          hidePoweredBy: s.hidePoweredBy,
        }),
      );
      message = tr("saved");
    } catch (err) {
      error = context.errorMessage(err);
    }
  }

  async function upload(kind: AssetKind, input: HTMLInputElement) {
    const file = input.files?.[0];
    input.value = "";
    if (!file) return;
    error = message = "";
    const problem = imageProblem(file);
    if (problem) {
      error = tr(problem);
      return;
    }
    try {
      show(await api.upload(kind, file));
      message = tr("saved");
    } catch (err) {
      error = context.errorMessage(err);
    }
  }

  async function remove(kind: AssetKind) {
    error = message = "";
    try {
      const v = await api.remove(kind);
      show(v ?? (await api.settings()));
    } catch (err) {
      error = context.errorMessage(err);
    }
  }
</script>

<p class="muted">{tr("description")}</p>
{#if s}
  {#if !s.applied}<p class="notice" role="status">{tr("not_applied")}</p>{/if}
  <form onsubmit={save}>
    <label class="field"><span>{tr("name")}</span><input class="input" bind:value={s.name} maxlength="60" disabled={readOnly} /></label>
    <div class="colours">
      <label class="check"
        ><input type="checkbox" bind:checked={useAccent} disabled={readOnly} />
        {tr("accent")}
        {#if useAccent}<input type="color" bind:value={s.accent} disabled={readOnly} aria-label={tr("accent")} />{/if}</label
      >
      <label class="check"
        ><input type="checkbox" bind:checked={useSecondary} disabled={readOnly} />
        {tr("secondary")}
        {#if useSecondary}<input type="color" bind:value={s.secondary} disabled={readOnly} aria-label={tr("secondary")} />{/if}</label
      >
    </div>
    {#each s.warnings as w (w)}<p class="warning" role="note">⚠ {w}</p>{/each}
    <label class="check"><input type="checkbox" bind:checked={s.hidePoweredBy} disabled={readOnly} /> {tr("hide_powered_by")}</label>
    {#if !readOnly}<button class="btn btn-primary" type="submit">{tr("save")}</button>{/if}
  </form>

  <h3>{tr("images")}</h3>
  <p class="muted small">{tr("image_hint")}</p>
  <ul>
    {#each assetKinds as kind (kind)}
      {@const at = s.assets[kind]}
      <li>
        <div>
          <strong>{tr(`kind.${kind}`)}</strong>
          <span class="muted small"
            >{at ? tr("uploaded", { date: new Date(at).toLocaleDateString(context.language) }) : tr("none_uploaded")}</span
          >
        </div>
        {#if !readOnly}
          <label class="btn file"
            >{tr("upload")}<input
              type="file"
              accept="image/png,image/jpeg,image/gif"
              onchange={(e) => upload(kind, e.currentTarget)}
            /></label
          >
          {#if at}<button class="btn" type="button" onclick={() => remove(kind)}>{tr("remove")}</button>{/if}
        {/if}
      </li>
    {/each}
  </ul>
{/if}
{#if error}<p class="error" role="alert">{error}</p>{/if}
{#if message}<p class="muted" role="status">{message}</p>{/if}

<style>
  form {
    display: grid;
    gap: var(--space-2);
  }
  .colours {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-3);
  }
  .check {
    display: flex;
    gap: var(--space-1);
    align-items: center;
  }
  .notice,
  .warning {
    padding: var(--space-2);
    border-radius: var(--radius);
    background: var(--color-warning-bg, rgba(234, 179, 8, 0.18));
    margin: 0;
  }
  h3 {
    font-size: 1rem;
    margin: var(--space-3) 0 var(--space-1);
  }
  ul {
    list-style: none;
    padding: 0;
  }
  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-1) 0;
  }
  li div {
    flex: 1;
    display: flex;
    flex-direction: column;
  }
  .small {
    font-size: 0.85em;
  }
  .file {
    position: relative;
    overflow: hidden;
  }
  .file input {
    position: absolute;
    inset: 0;
    opacity: 0;
    cursor: pointer;
  }
</style>
