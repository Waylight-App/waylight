import { describe, expect, it, vi } from "vitest";
import { TeamTabsApi } from "./teamtabs";

describe("TeamTabsApi", () => {
  it("shares a tab by id", async () => {
    const http = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn().mockResolvedValue({}),
      delete: vi.fn(),
      putBlob: vi.fn(),
    };
    await new TeamTabsApi(http).share("t/1", "group:office");
    expect(http.put).toHaveBeenCalledWith("/api/team-tabs/t%2F1", {
      visibility: "group:office",
    });
  });
});
