import { useSyncExternalStore } from "react";

const QUERY = "(prefers-reduced-motion: reduce)";
const subscribers = new Set<() => void>();
let mediaQueryList: MediaQueryList | null = null;
let currentMatches = false;
let listening = false;

function getMediaQueryList(): MediaQueryList | null {
  if (typeof window === "undefined") return null;
  if (!mediaQueryList) {
    mediaQueryList = window.matchMedia(QUERY);
    currentMatches = mediaQueryList.matches;
  }
  return mediaQueryList;
}

function notifySubscribers(event: MediaQueryListEvent): void {
  currentMatches = event.matches;
  subscribers.forEach((subscriber) => subscriber());
}

function subscribe(subscriber: () => void): () => void {
  const mql = getMediaQueryList();
  subscribers.add(subscriber);
  if (mql && !listening) {
    mql.addEventListener("change", notifySubscribers);
    listening = true;
  }

  return () => {
    subscribers.delete(subscriber);
    if (subscribers.size === 0 && mediaQueryList && listening) {
      mediaQueryList.removeEventListener("change", notifySubscribers);
      mediaQueryList = null;
      currentMatches = false;
      listening = false;
    }
  };
}

function getSnapshot(): boolean {
  getMediaQueryList();
  return currentMatches;
}

/**
 * Returns true when the user has requested reduced motion via
 * their OS / browser preference (prefers-reduced-motion: reduce).
 */
export function useReducedMotion(): boolean {
  return useSyncExternalStore(subscribe, getSnapshot, () => false);
}
