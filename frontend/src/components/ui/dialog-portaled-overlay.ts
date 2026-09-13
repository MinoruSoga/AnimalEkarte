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
