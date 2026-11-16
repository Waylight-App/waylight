import { describe, expect, it, vi } from "vitest";
import {
  nextStatus,
  parseConfig,
  poll,
  PresenceApi,
  type Person,
} from "./presence";

const me = (
  status: Person["status"],
  source: Person["source"] = "self",
): Person => ({
  accountId: "a",
  name: "Anna",
  status,
  source,
});

describe("presence", () => {
  it("reads the config", () => {
    expect(parseConfig(undefined)).toEqual({ group: "" });
    expect(parseConfig({ group: 3 })).toEqual({ group: "" });
    expect(parseConfig({ group: "g1" })).toEqual({ group: "g1" });
  });

  it("toggles the own status", () => {
    expect(nextStatus(me(""), "office")).toBe("office");
    expect(nextStatus(me("office"), "office")).toBe("");
    expect(nextStatus(me("office"), "home")).toBe("home");
    expect(nextStatus(me("away", "absence"), "office")).toBeNull();
    expect(nextStatus(null, "office")).toBeNull();
  });

  it("polls only while visible", () => {
    vi.useFakeTimers();
    const doc = {
      visibilityState: "visible" as DocumentVisibilityState,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    };
    const fn = vi.fn();
    const stop = poll(fn, 1000, doc);
    expect(fn).toHaveBeenCalledTimes(1);
    vi.advanceTimersByTime(1000);
    expect(fn).toHaveBeenCalledTimes(2);
    doc.visibilityState = "hidden";
    vi.advanceTimersByTime(3000);
    expect(fn).toHaveBeenCalledTimes(2);
    const onVisible = doc.addEventListener.mock.calls[0]?.[1] as () => void;
    doc.visibilityState = "visible";
    onVisible();
    expect(fn).toHaveBeenCalledTimes(3);
    stop();
    expect(doc.removeEventListener).toHaveBeenCalledWith(
      "visibilitychange",
      onVisible,
    );
    vi.advanceTimersByTime(5000);
    expect(fn).toHaveBeenCalledTimes(3);
    vi.useRealTimers();
  });

  it("calls the API", async () => {
    const http = {
      get: vi.fn().mockResolvedValue({ enabled: true }),
      post: vi.fn(),
      put: vi.fn().mockResolvedValue({ enabled: false }),
      delete: vi.fn(),
    };
    const api = new PresenceApi(http);
    await api.board("");
    expect(http.get).toHaveBeenLastCalledWith("/api/presence");
    await api.board("a b");
    expect(http.get).toHaveBeenLastCalledWith("/api/presence?group=a%20b");
    await api.set("away", "2026-11-18");
    expect(http.put).toHaveBeenLastCalledWith("/api/presence/me", {
      status: "away",
      returnOn: "2026-11-18",
    });
    expect(await api.enabled()).toBe(true);
    expect(await api.setEnabled(false)).toBe(false);
    expect(http.put).toHaveBeenLastCalledWith("/api/presence/settings", {
      enabled: false,
    });
  });
});
