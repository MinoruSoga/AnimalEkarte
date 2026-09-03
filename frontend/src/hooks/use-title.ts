import { useEffect } from "react";

/**
 * Standard suffix for all document titles.
 */
const TITLE_SUFFIX = "Animal Ekarte";

/**
 * Hook to dynamically update the browser tab title.
 *
 * @param title - The title for the current page (e.g., "Owners")
 */
export function useTitle(title?: string) {
  useEffect(() => {
    const fullTitle = title ? `${title} | ${TITLE_SUFFIX}` : TITLE_SUFFIX;
    document.title = fullTitle;
  }, [title]);
}
