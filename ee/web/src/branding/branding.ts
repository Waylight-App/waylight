import type { PluginHttp } from "../../../../web/src/app/plugins/api";

/** The image kinds the server takes (ee/branding, #305). */
export const assetKinds = ["logo", "logo-dark", "favicon"] as const;
export type AssetKind = (typeof assetKinds)[number];

/** GET/PUT /api/branding/settings. */
export interface BrandingSettings {
  name: string;
  accent: string;
  secondary: string;
  hidePoweredBy: boolean;
  /** When each image was uploaded, by kind; missing = none. */
  assets: Partial<Record<AssetKind, string>>;
  /** False while the licence does not cover branding: stored, not shown. */
  applied: boolean;
  /** Contrast warnings (WCAG AA) for the chosen colours. */
  warnings: string[];
}

/** Largest image the server accepts. */
export const MAX_IMAGE_BYTES = 2 * 1024 * 1024;

/** A file the server would refuse, as a message key, or null. */
export function imageProblem(file: {
  size: number;
  type: string;
}): "too_big" | "svg" | null {
  if (file.size > MAX_IMAGE_BYTES) return "too_big";
  // SVG can carry scripts; the server refuses it, so say so before uploading.
  if (file.type === "image/svg+xml") return "svg";
  return null;
}

export class BrandingApi {
  constructor(private readonly http: PluginHttp) {}

  settings(): Promise<BrandingSettings> {
    return this.http.get<BrandingSettings>("/api/branding/settings");
  }

  save(
    s: Pick<
      BrandingSettings,
      "name" | "accent" | "secondary" | "hidePoweredBy"
    >,
  ): Promise<BrandingSettings> {
    return this.http.put<BrandingSettings>("/api/branding/settings", {
      name: s.name.trim(),
      accent: s.accent,
      secondary: s.secondary,
      hidePoweredBy: s.hidePoweredBy,
    });
  }

  upload(kind: AssetKind, file: Blob): Promise<BrandingSettings> {
    return this.http.putBlob<BrandingSettings>(
      `/api/branding/assets/${kind}`,
      file,
    );
  }

  remove(kind: AssetKind): Promise<BrandingSettings | undefined> {
    return this.http.delete<BrandingSettings | undefined>(
      `/api/branding/assets/${kind}`,
    );
  }
}
