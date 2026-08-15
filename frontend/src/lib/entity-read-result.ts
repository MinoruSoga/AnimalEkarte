import axios from "axios";

/**
 * Discriminated read result for clinic-scoped entity edit loaders.
 * Prevents mapping fetch failures into blank editable form models (BUG-016 / BUG-019).
 *
 * - `idle`: create route (no id)
 * - `loading`: edit route still fetching
 * - `found`: entity loaded and editable
 * - `notFound` / `forbiddenOrHidden`: non-disclosure UI (same client message; 404 vs 403 kept distinct for logs)
 * - `error`: network / 5xx / unexpected — keep retry path; never treat as 404
 */
export type EntityReadResult<T> =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "found"; data: T }
  | { status: "notFound" }
  | { status: "forbiddenOrHidden" }
  | { status: "error"; error: unknown; retry: (() => void) | undefined };

export type EntityReadStatus = EntityReadResult<unknown>["status"];

/** 404 and 403 must share the same non-disclosure UI (tenant existence not leaked). */
export function isNonDisclosureReadStatus(
  status: EntityReadStatus,
): status is "notFound" | "forbiddenOrHidden" {
  return status === "notFound" || status === "forbiddenOrHidden";
}

export function resolveEntityReadResult<T>(input: {
  id: string | undefined | null;
  data: T | undefined | null;
  isLoading: boolean;
  isError: boolean;
  error: unknown;
  refetch?: (() => unknown) | undefined;
}): EntityReadResult<T> {
  const id = input.id;
  if (id == null || id === "") {
    return { status: "idle" };
  }

  // First paint / in-flight without cached data
  if (input.isLoading && (input.data === undefined || input.data === null)) {
    return { status: "loading" };
  }

  if (input.isError) {
    if (axios.isAxiosError(input.error)) {
      const status = input.error.response?.status;
      if (status === 404) {
        return { status: "notFound" };
      }
      if (status === 403) {
        return { status: "forbiddenOrHidden" };
      }
    }
    const retry = input.refetch
      ? () => {
          void input.refetch?.();
        }
      : undefined;
    return { status: "error", error: input.error, retry };
  }

  if (input.data !== undefined && input.data !== null) {
    return { status: "found", data: input.data };
  }

  // Settled success with no body (or disabled query edge): treat as missing.
  return { status: "notFound" };
}
