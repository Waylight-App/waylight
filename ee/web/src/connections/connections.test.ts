import { describe, expect, it, vi } from "vitest";
import { ConnectionsApi, outcome } from "./connections";

describe("connections", () => {
  it("reads the outcome of the round trip", () => {
    expect(outcome("?settings=plugin-connections&connected=slack")).toEqual({
      connected: "slack",
      failed: undefined,
    });
    expect(outcome("?connection_error=microsoft")).toEqual({
      connected: undefined,
      failed: "microsoft",
    });
  });

  it("starts a connection and trims the client registration", async () => {
    const http = {
      get: vi.fn(),
      post: vi.fn().mockResolvedValue({ url: "https://slack.example/oauth" }),
      put: vi.fn(),
      delete: vi.fn(),
      putBlob: vi.fn(),
    };
    const api = new ConnectionsApi(http);
    expect(await api.start("slack")).toBe("https://slack.example/oauth");
    await api.saveClient("microsoft", {
      clientId: " id ",
      clientSecret: "s",
      tenant: " t ",
    });
    expect(http.put).toHaveBeenCalledWith(
      "/api/connections/clients/microsoft",
      {
        clientId: "id",
        clientSecret: "s",
        tenant: "t",
      },
    );
  });
});
