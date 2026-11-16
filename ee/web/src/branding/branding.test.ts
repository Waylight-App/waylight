import { describe, expect, it, vi } from "vitest";
import { BrandingApi, imageProblem, MAX_IMAGE_BYTES } from "./branding";

describe("branding", () => {
  it("refuses SVG and files over 2 MB before uploading", () => {
    expect(imageProblem({ size: 10, type: "image/png" })).toBeNull();
    expect(imageProblem({ size: MAX_IMAGE_BYTES + 1, type: "image/png" })).toBe(
      "too_big",
    );
    expect(imageProblem({ size: 10, type: "image/svg+xml" })).toBe("svg");
  });

  it("uploads the raw file and trims the name", async () => {
    const http = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
      putBlob: vi.fn(),
    };
    const api = new BrandingApi(http);
    const blob = new Blob(["x"], { type: "image/png" });
    await api.upload("logo-dark", blob);
    expect(http.putBlob).toHaveBeenCalledWith(
      "/api/branding/assets/logo-dark",
      blob,
    );
    await api.save({
      name: " Acme ",
      accent: "#112233",
      secondary: "",
      hidePoweredBy: true,
      defaultTheme: "",
      lockTheme: true,
    });
    expect(http.put).toHaveBeenCalledWith("/api/branding/settings", {
      name: "Acme",
      accent: "#112233",
      secondary: "",
      hidePoweredBy: true,
      defaultTheme: "",
      // No theme to lock to, so the lock is not sent.
      lockTheme: false,
    });
  });
});
