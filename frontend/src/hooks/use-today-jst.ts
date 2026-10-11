import { useSyncExternalStore } from "react";

import { todayJSTISO } from "@/lib/jst-date";

const DAY_MS = 24 * 60 * 60 * 1000;
const listeners = new Set<() => void>();
let timer: ReturnType<typeof setTimeout> | undefined;

function clearTimer(): void {
  if (timer !== undefined) {
    clearTimeout(timer);
    timer = undefined;
  }
}

function refresh(): void {
  clearTimer();
  for (const listener of listeners) listener();
  if (listeners.size === 0 || document.visibilityState === "hidden") return;
  const nextMidnight = Date.parse(`${todayJSTISO()}T00:00:00+09:00`) + DAY_MS;
  timer = setTimeout(refresh, Math.max(1, nextMidnight - Date.now()));
}

function subscribe(listener: () => void): () => void {
  listeners.add(listener);
  if (listeners.size === 1) {
    document.addEventListener("visibilitychange", refresh);
    window.addEventListener("focus", refresh);
    window.addEventListener("pageshow", refresh);
    refresh();
  }
  return () => {
    listeners.delete(listener);
    if (listeners.size === 0) {
      clearTimer();
      document.removeEventListener("visibilitychange", refresh);
      window.removeEventListener("focus", refresh);
      window.removeEventListener("pageshow", refresh);
    }
  };
}

/**
 * Share one JST day-boundary timer across mounted date/birthday consumers.
 * Hidden tabs pause the timer; visibility, focus and pageshow catch up after sleep.
 * Date strings provide stable snapshots, including server rendering, without a
 * stale module-level date when the last consumer unmounts and later remounts.
 */
export function useTodayJst(): string {
  return useSyncExternalStore(subscribe, todayJSTISO, todayJSTISO);
}
