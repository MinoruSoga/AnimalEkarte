import { useCallback, useEffect, useState } from "react";

/**
 * Hook to warn users about unsaved changes when navigating away.
 * - Centralizes the beforeunload subscription for shared dirty-state hooks.
 * - Uses the browser's beforeunload event for tab close / browser navigation.
 * - Returns `isDirty` so the caller can pass it to <NavigationBlocker when={isDirty} />
 *   for SPA in-app navigation protection via React Router's useBlocker.
 *
 * @returns { markDirty, markClean, isDirty }
 */
export function useUnsavedChanges() {
  const [isDirty, setIsDirty] = useState(false);

  useEffect(() => {
    if (!isDirty) return;

    const handler = (e: BeforeUnloadEvent) => {
      e.preventDefault();
      // Modern browsers ignore custom messages but still show a dialog
      e.returnValue = "";
    };

    window.addEventListener("beforeunload", handler);
    return () => window.removeEventListener("beforeunload", handler);
  }, [isDirty]);

  const markDirty = useCallback(() => setIsDirty(true), []);
  const markClean = useCallback(() => setIsDirty(false), []);

  return { isDirty, markDirty, markClean };
}
