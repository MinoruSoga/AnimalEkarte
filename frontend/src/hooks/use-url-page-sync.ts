import { useEffect } from "react";
import type { SetURLSearchParams } from "react-router";

interface UseUrlPageSyncOptions {
  urlPage: number;
  totalPages: number;
  isLoading: boolean;
  setSearchParams: SetURLSearchParams;
  /** When true, page clamps keep the current location state. */
  preserveNavigationState?: boolean;
  navigationState?: unknown;
}

function nextListSearchParamsWithPage(prev: URLSearchParams, page: number): URLSearchParams {
  const next = new URLSearchParams(prev);
  if (page <= 1) {
    next.delete("page");
  } else {
    next.set("page", String(page));
  }
  return next;
}

/**
 * Clamp URL `page` into [1, totalPages] after list data loads.
 * Replaces the duplicated effect in 7 list routes (FE-RC-028).
 */
export function useUrlPageSync({
  urlPage,
  totalPages,
  isLoading,
  setSearchParams,
  preserveNavigationState = false,
  navigationState,
}: UseUrlPageSyncOptions): void {
  useEffect(() => {
    if (isLoading) return;
    const clampedPage = Math.max(1, Math.min(urlPage, Math.max(totalPages, 1)));
    if (clampedPage !== urlPage) {
      const options = preserveNavigationState
        ? { replace: true, state: navigationState }
        : { replace: true };
      setSearchParams((prev) => nextListSearchParamsWithPage(prev, clampedPage), options);
    }
    // setSearchParams is a stable router setter; re-run only on page/total/loading.
    // eslint-disable-next-line react-hooks/exhaustive-deps -- intentional FE-144 / FE-RC-028 contract
  }, [urlPage, totalPages, isLoading, preserveNavigationState, navigationState]);
}
