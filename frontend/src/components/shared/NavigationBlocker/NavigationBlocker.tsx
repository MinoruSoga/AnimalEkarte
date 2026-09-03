import { useState, useEffect, useRef } from "react";
import { useBlocker } from "react-router";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";

interface NavigationBlockerProps {
  /** Whether the blocker should be active */
  when: boolean;
  /** Dialog title */
  title?: string;
  /** Dialog description */
  description?: string;
}

/* ------------------------------------------------------------------ */
/*  Inner component — only mounted when blocking is actually needed.  */
/*  Because it is conditionally rendered, useBlocker is never called  */
/*  during initial page load (when `when` is false), which avoids the */
/*  React Router warning about blocking a POP navigation to a         */
/*  location not created by the router.                               */
/* ------------------------------------------------------------------ */
function NavigationBlockerDialog({
  title,
  description,
  onCleanup,
}: {
  title: string;
  description: string;
  onCleanup: () => void;
}) {
  const blocker = useBlocker(
    ({ currentLocation, nextLocation }) => currentLocation.pathname !== nextLocation.pathname,
  );

  // FE-RC-031: cleanup は unmount 時にしか実行されないため、mount 時点の
  // `blocker` を deps=[] で固定 closure すると常に初期状態（unblocked）しか
  // 見えず no-op になっていた。ref で常に最新の blocker を参照する。
  const blockerRef = useRef(blocker);
  useEffect(() => {
    blockerRef.current = blocker;
  }, [blocker]);

  // If blocking was deactivated externally (parent sets when=false
  // → this component unmounts), reset any pending block first.
  useEffect(() => {
    return () => {
      if (blockerRef.current.state === "blocked") {
        blockerRef.current.reset();
      }
    };
  }, []);

  // When the user confirms leaving, proceed and notify parent.
  const handleConfirm = () => {
    blocker.proceed?.();
    onCleanup();
  };

  return (
    <ConfirmDialog
      open={blocker.state === "blocked"}
      onClose={() => blocker.reset?.()}
      onConfirm={handleConfirm}
      title={title}
      description={description}
      confirmLabel="ページを離れる"
      cancelLabel="このページに留まる"
      variant="destructive"
    />
  );
}

/* ------------------------------------------------------------------ */
/*  Public component — safe to render unconditionally.                */
/*  The inner <NavigationBlockerDialog> is only mounted when `when`   */
/*  is true, so useBlocker is never registered on initial page load.  */
/* ------------------------------------------------------------------ */

/**
 * Renders a confirmation dialog when the user attempts SPA navigation
 * while there are unsaved changes.
 *
 * Place this component inside any route component that needs protection.
 */
export function NavigationBlocker({
  when,
  title = "変更が保存されていません",
  description = "このページから離れると、入力中の内容は失われます。ページを離れますか？",
}: NavigationBlockerProps) {
  // Keep the dialog mounted while a block interaction may still be
  // in progress, even if `when` flips to false mid-dialog.
  const [active, setActive] = useState(when);
  const [prevWhen, setPrevWhen] = useState(when);
  if (prevWhen !== when) {
    setPrevWhen(when);
    setActive(when);
  }

  if (!active) {
    return null;
  }

  return (
    <NavigationBlockerDialog
      title={title}
      description={description}
      onCleanup={() => setActive(false)}
    />
  );
}
