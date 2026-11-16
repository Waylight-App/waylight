import { describe, expect, it, vi } from "vitest";
import { TeamTabsApi, tabsFor } from "./teamtabs";

describe("tabsFor", () => {
  it("lists the tabs a group sees", () => {
    const tabs = [
      { id: "a", name: "A", visibility: "group:office" },
      { id: "b", name: "B", visibility: "everyone" },
      { id: "c", name: "C", visibility: "private" },
    ];
    expect(tabsFor(tabs, "office").map((t) => t.id)).toEqual(["a", "b"]);
    expect(tabsFor(tabs, "family").map((t) => t.id)).toEqual(["b"]);
  });
});

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
