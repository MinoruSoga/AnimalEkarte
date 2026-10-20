import { useCallback, useEffect, useRef } from "react";
import { discardTrimmingCourseImage } from "../api/upload-trimming-course-image";

/** Only local uploads which have never been submitted may be discarded. */
export function useCourseImageDrafts() {
  const drafts = useRef(new Set<string>());
  const mounted = useRef(true);
  const discard = useCallback((key: string) => {
    if (!drafts.current.delete(key)) return;
    void discardTrimmingCourseImage(key).catch(() => {
      // Orphan cleanup is best effort; never widen it to submitted/original keys.
    });
  }, []);
  useEffect(() => {
    mounted.current = true;
    const owned = drafts.current;
    return () => {
      mounted.current = false;
      for (const key of owned) discard(key);
    };
  }, [discard]);
  const register = useCallback(
    (key: string) => {
      drafts.current.add(key);
      if (!mounted.current) {
        discard(key);
        return false;
      }
      return true;
    },
    [discard],
  );
  const protectSubmitted = useCallback((key: string) => {
    // Irreversible ownership transfer happens BEFORE the request. A failure can
    // mean commit + response loss, so the key never becomes a draft again.
    drafts.current.delete(key);
  }, []);
  return { register, discard, protectSubmitted };
}
