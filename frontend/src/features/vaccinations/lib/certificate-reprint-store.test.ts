import { describe, expect, it, vi } from "vitest";
import { createReprintStore } from "./certificate-reprint-store";

describe("再印刷 store", () => {
  it("未登録の edition は idle・pending なしを返し、同じ参照を返し続ける", () => {
    const store = createReprintStore();
    expect(store.get(1)).toEqual({ state: { phase: "idle" }, pending: null, layout: null });
    expect(store.get(1)).toBe(store.get(2));
  });

  it("set は購読者へ通知し、edition ごとに独立して保持する", () => {
    const store = createReprintStore();
    const listener = vi.fn();
    const unsubscribe = store.subscribe(listener);
    store.set(1, { state: { phase: "running" }, pending: null, layout: null });
    expect(listener).toHaveBeenCalledTimes(1);
    expect(store.get(1).state.phase).toBe("running");
    expect(store.get(2).state.phase).toBe("idle");
    unsubscribe();
    store.set(1, { state: { phase: "idle" }, pending: null, layout: null });
    expect(listener).toHaveBeenCalledTimes(1);
  });
});
