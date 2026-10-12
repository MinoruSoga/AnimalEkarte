import { useCallback, useSyncExternalStore } from "react";

import { todayJSTISO } from "@/lib/jst-date";

const DAY_MS = 24 * 60 * 60 * 1000;
const listeners = new Map<() => void, number | null>();
let timer: ReturnType<typeof setTimeout> | undefined;
let clockTime = Date.now();

function getClockTime(): number {
  return clockTime;
}

function getServerClockTime(): number {
  return 0;
}

function clearTimer(): void {
  if (timer !== undefined) {
    clearTimeout(timer);
    timer = undefined;
  }
}

function refresh(): void {
  clearTimer();
  clockTime = Date.now();
  for (const listener of listeners.keys()) listener();
  if (listeners.size === 0 || document.visibilityState === "hidden") return;
  const now = Date.now();
  let nextBoundary = Date.parse(`${todayJSTISO()}T00:00:00+09:00`) + DAY_MS;
  for (const boundary of listeners.values()) {
    if (boundary !== null && boundary > now) nextBoundary = Math.min(nextBoundary, boundary);
  }
  timer = setTimeout(refresh, Math.max(1, nextBoundary - now));
}

function subscribe(listener: () => void, boundary: number | null = null): () => void {
  listeners.set(listener, boundary);
  if (listeners.size === 1) {
    document.addEventListener("visibilitychange", refresh);
    window.addEventListener("focus", refresh);
    window.addEventListener("pageshow", refresh);
  }
  refresh();
  return () => {
    listeners.delete(listener);
    if (listeners.size === 0) {
      clearTimer();
      document.removeEventListener("visibilitychange", refresh);
      window.removeEventListener("focus", refresh);
      window.removeEventListener("pageshow", refresh);
    } else {
      refresh();
    }
  };
}

/**
 * Share one timer for JST midnight and registered appointment boundaries.
 * Hidden tabs pause the timer; visibility, focus and pageshow catch up after sleep.
 * Date strings provide stable snapshots, including server rendering, without a
 * stale module-level date when the last consumer unmounts and later remounts.
 */
export function useTodayJst(): string {
  return useSyncExternalStore(subscribe, todayJSTISO, todayJSTISO);
}

/** Re-render at an instant without making date-only consumers observe clock ticks. */
export function useTimeBoundary(boundary: number | null): void {
  const subscribeBoundary = useCallback(
    (listener: () => void) => subscribe(listener, boundary),
    [boundary],
  );
  useSyncExternalStore(subscribeBoundary, getClockTime, getServerClockTime);
}
