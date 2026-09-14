/** Portaled Radix overlays nested under Dialog (Popover/Select). Clicks must not dismiss Dialog. */
export function isDialogPortaledOverlayTarget(target: EventTarget | null): boolean {
  if (!(target instanceof Element)) {
    return false;
  }
  return Boolean(
    target.closest("[data-radix-popper-content-wrapper]") ||
    target.closest('[data-slot="popover-content"]') ||
    target.closest('[data-slot="select-content"]'),
  );
}

/** Resolve the originating EventTarget from a Radix outside-interaction event. */
export function getDialogOutsideInteractionTarget(event: {
  target: EventTarget | null;
  detail?: unknown;
}): EventTarget | null {
  const detail = event.detail && typeof event.detail === "object" ? event.detail : null;
  const originalEvent =
    detail && "originalEvent" in detail
      ? (detail as { originalEvent?: Event }).originalEvent
      : undefined;
  return originalEvent?.target ?? event.target;
}

/**
 * Prevent Dialog dismiss when the interaction landed on a portaled Popover/Select.
 * Used for both onPointerDownOutside (touch) and onInteractOutside.
 */
export function shouldPreventDialogOutsideInteraction(event: {
  target: EventTarget | null;
  detail?: unknown;
}): boolean {
  return isDialogPortaledOverlayTarget(getDialogOutsideInteractionTarget(event));
}
