import { beforeEach, describe, expect, it, vi } from "vitest";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import { toast } from "sonner";

import { handleApiError } from "@/lib/handle-api-error";

vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

const CONFLICT_MESSAGE = "既に予約が存在します";

function axiosError(status: number, data: Record<string, unknown>): AxiosError {
  const config = {
    headers: new AxiosHeaders(),
  } as InternalAxiosRequestConfig;
  return new AxiosError("request failed", AxiosError.ERR_BAD_RESPONSE, config, undefined, {
    config,
    data,
    headers: new AxiosHeaders(),
    status,
    statusText: "Conflict",
  });
}

/**
 * EMR-76 / BUG-TRIM-EXCL-TIMERANGE-500: the trimming form's save path routes every
 * mutation error through handleApiError(error, "保存") (use-trimming-form.ts). A
 * PostgreSQL exclusion violation now returns 409 + 既に予約が存在します, which must
 * reach the toast — previously a 500 surfaced the generic server-error message.
 */
describe("trimming form 409 conflict display (EMR-76)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("toasts 既に予約が存在します for a 409 reservation_time_conflict body", () => {
    handleApiError(
      axiosError(409, { error: CONFLICT_MESSAGE, code: "reservation_time_conflict" }),
      "保存",
    );
    expect(toast.error).toHaveBeenCalledWith(CONFLICT_MESSAGE);
  });

  it("still toasts the same message on a retried overlapping save", () => {
    const err = axiosError(409, { error: CONFLICT_MESSAGE });
    handleApiError(err, "保存");
    handleApiError(err, "保存");
    expect(toast.error).toHaveBeenCalledTimes(2);
    expect(toast.error).toHaveBeenNthCalledWith(1, CONFLICT_MESSAGE);
    expect(toast.error).toHaveBeenNthCalledWith(2, CONFLICT_MESSAGE);
  });

  it("does not leak the reload fallback for the Japanese conflict body", () => {
    handleApiError(
      axiosError(409, { error: CONFLICT_MESSAGE, code: "reservation_time_conflict" }),
      "保存",
    );
    expect(toast.error).not.toHaveBeenCalledWith(
      "他のユーザーによって更新されています。一度リロードしてください。",
    );
    expect(toast.error).not.toHaveBeenCalledWith(
      "サーバーエラーが発生しました。しばらく経ってから再度お試しください。",
    );
  });
});
