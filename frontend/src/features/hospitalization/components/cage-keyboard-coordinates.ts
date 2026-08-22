export interface CageKeyboardRect {
  top: number;
  left: number;
  bottom: number;
  right: number;
}

function center(rect: CageKeyboardRect): { x: number; y: number } {
  return {
    x: (rect.left + rect.right) / 2,
    y: (rect.top + rect.bottom) / 2,
  };
}

function distance(a: { x: number; y: number }, b: { x: number; y: number }): number {
  const dx = a.x - b.x;
  const dy = a.y - b.y;
  return dx * dx + dy * dy;
}

const DIRECTION_THRESHOLD = 8;

/** Pick the nearest cage rect in the arrow-key direction (BUG-005). */
export function selectNextCageRect(
  code: string,
  collision: CageKeyboardRect | null | undefined,
  candidates: CageKeyboardRect[],
): CageKeyboardRect | undefined {
  if (candidates.length === 0) return undefined;
  if (!collision) return candidates[0];

  const origin = center(collision);
  const filtered = candidates.filter((rect) => {
    const next = center(rect);
    switch (code) {
      case "ArrowDown":
        return next.y > origin.y + DIRECTION_THRESHOLD;
      case "ArrowUp":
        return next.y < origin.y - DIRECTION_THRESHOLD;
      case "ArrowRight":
        return next.x > origin.x + DIRECTION_THRESHOLD;
      case "ArrowLeft":
        return next.x < origin.x - DIRECTION_THRESHOLD;
      default:
        return false;
    }
  });
  if (filtered.length === 0) return undefined;
  return filtered.reduce((best, rect) =>
    distance(origin, center(rect)) < distance(origin, center(best)) ? rect : best,
  );
}

export function cageKeyboardCoordinateGetter(
  event: KeyboardEvent,
  args: {
    currentCoordinates: { x: number; y: number };
    context: {
      collisionRect?: CageKeyboardRect | null;
      droppableRects: { get: (id: PropertyKey) => CageKeyboardRect | undefined };
      droppableContainers: { getEnabled: () => ReadonlyArray<{ id: PropertyKey }> };
    };
  },
): { x: number; y: number } | undefined {
  if (!["ArrowDown", "ArrowUp", "ArrowLeft", "ArrowRight"].includes(event.code)) {
    return undefined;
  }
  event.preventDefault();

  const collision =
    args.context.collisionRect ?? {
      top: args.currentCoordinates.y,
      left: args.currentCoordinates.x,
      bottom: args.currentCoordinates.y + 1,
      right: args.currentCoordinates.x + 1,
    };

  const candidates = args.context.droppableContainers
    .getEnabled()
    .map((entry) => args.context.droppableRects.get(entry.id))
    .filter((rect): rect is CageKeyboardRect => rect != null);

  const next = selectNextCageRect(event.code, collision, candidates);
  if (!next) return undefined;
  return { x: next.left, y: next.top };
}
