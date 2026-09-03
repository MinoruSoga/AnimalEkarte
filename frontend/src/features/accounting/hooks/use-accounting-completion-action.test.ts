import { startTransition } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import type { QueryClient } from "@tanstack/react-query";
import type { NavigateFunction } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { handleApiError } from "@/lib/handle-api-error";
import { toast } from "sonner";
import { completeAccounting } from "../api/complete-accounting";
import { updateAccounting } from "../api/update-accounting";
import type { Accounting } from "../types";
import {
  focusAccountingCompletionError,
  resolveAccountingCompletionFocusTarget,
  useAccountingCompletionAction,
} from "./use-accounting-completion-action";

vi.mock("../api/complete-accounting", () => ({
  completeAccounting: vi.fn(),
  createAccountingCompletionIdempotencyKey: vi.fn(() => "test-idempotency-key"),
}));
vi.mock("../api/update-accounting", () => ({
  updateAccounting: vi.fn(),
}));
vi.mock("@/lib/handle-api-error", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/handle-api-error")>();
  return {
    ...actual,
    handleApiError: vi.fn(),
  };
});
vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

const completeAccountingMock = vi.mocked(completeAccounting);
const updateAccountingMock = vi.mocked(updateAccounting);
const handleApiErrorMock = vi.mocked(handleApiError);

const POST_CLOSE_REASON_400 = "レジ締め済み期間の会計編集には post_close_reason の入力が必要です";
const GENERIC_400 = "支払金額の合計が請求額と一致しません";

function axiosError(status: number, data: Record<string, unknown>): AxiosError {
  const config = {
    headers: new AxiosHeaders(),
  } as InternalAxiosRequestConfig;
  return new AxiosError("request failed", AxiosError.ERR_BAD_REQUEST, config, undefined, {
    config,
    data,
    headers: new AxiosHeaders(),
    status,
    statusText: "Bad Request",
  });
}

function waitingAccounting(): Accounting {
  return {
    id: "123",
    clinicId: "1",
    ownerId: "10",
    ownerName: "テスト飼い主",
    petId: "20",
    petName: "ポチ",
    status: "waiting",
    scheduledDate: "2026-05-01",
    items: [],
    totalRefundedAmount: 0,
  };
}

function mountFocusFields(options?: { receivedAmount?: boolean; paymentSplit?: boolean }) {
  const root = document.createElement("div");
  const textarea = document.createElement("textarea");
  textarea.id = "postCloseReason";
  root.appendChild(textarea);
  if (options?.receivedAmount !== false) {
    const received = document.createElement("input");
    received.id = "receivedAmount";
    root.appendChild(received);
  }
  if (options?.paymentSplit === true) {
    const splitReceived = document.createElement("input");
    splitReceived.id = "payment-split-0-received";
    root.appendChild(splitReceived);
  }
  document.body.appendChild(root);
  return root;
}

afterEach(() => {
  document.getElementById("postCloseReason")?.parentElement?.remove();
});

function buildHookArgs(
  overrides: {
    accountingId?: string;
  } = {},
) {
  const queryClient = {
    invalidateQueries: vi.fn(),
  } as unknown as QueryClient;
  const navigate = vi.fn() as unknown as NavigateFunction;
  const setCompletedPayment = vi.fn();

  return {
    accountingId: "accountingId" in overrides ? overrides.accountingId : "123",
    accounting: waitingAccounting(),
    calculation: {
      subtotal: 1000,
      taxTotal: 100,
      totalAmount: 1100,
      insuranceAmount: 0,
      billingAmount: 1100,
    },
    displayItems: [],
    hasInsurance: false,
    insuranceRatio: "0",
    paymentSplits: [{ method: "cash" as const, amount: "1100", receivedAmount: "1100" }],
    queryClient,
    navigate,
    setCompletedPayment,
    postCloseReason: "",
    // FE-RC-001: このテスト群は「権限あり」時の通常フローを検証するため既定で全許可する。
    permissions: { canCreate: true, canEdit: true },
  };
}

async function submitCompletionAction(
  formAction: ReturnType<typeof useAccountingCompletionAction>["formAction"],
) {
  await act(async () => {
    startTransition(() => {
      formAction(new FormData());
    });
  });
}

describe("resolveAccountingCompletionFocusTarget", () => {
  it("null / undefined は receivedAmount を返す", () => {
    expect(resolveAccountingCompletionFocusTarget(null)).toBe("receivedAmount");
    expect(resolveAccountingCompletionFocusTarget(undefined)).toBe("receivedAmount");
  });

  it("空メッセージは receivedAmount を返す", () => {
    expect(resolveAccountingCompletionFocusTarget(axiosError(400, { error: "" }))).toBe(
      "receivedAmount",
    );
    expect(resolveAccountingCompletionFocusTarget(new Error(""))).toBe("receivedAmount");
  });

  it("日本語 400 に post_close_reason が含まれると postCloseReason を返す", () => {
    expect(
      resolveAccountingCompletionFocusTarget(axiosError(400, { error: POST_CLOSE_REASON_400 })),
    ).toBe("postCloseReason");
  });

  it("post_close_reason を含まない 400 は receivedAmount を返す", () => {
    expect(resolveAccountingCompletionFocusTarget(axiosError(400, { error: GENERIC_400 }))).toBe(
      "receivedAmount",
    );
  });

  it("Axios でないが body.error に post_close_reason がある場合も postCloseReason を返す", () => {
    const duckTyped = {
      response: {
        status: 400,
        data: { error: POST_CLOSE_REASON_400 },
      },
    };
    expect(resolveAccountingCompletionFocusTarget(duckTyped)).toBe("postCloseReason");
  });
});

describe("focusAccountingCompletionError", () => {
  it("postCloseReason 指定時は textarea#postCloseReason にフォーカスする", () => {
    mountFocusFields({ receivedAmount: true });
    const textarea = document.getElementById("postCloseReason");
    const received = document.getElementById("receivedAmount");
    expect(textarea).toBeInstanceOf(HTMLTextAreaElement);
    expect(received).toBeInstanceOf(HTMLInputElement);

    focusAccountingCompletionError("postCloseReason");

    expect(document.activeElement).toBe(textarea);
    expect(document.activeElement).not.toBe(received);
  });

  it("receivedAmount 指定時は input#receivedAmount にフォーカスする", () => {
    mountFocusFields({ receivedAmount: true, paymentSplit: true });

    focusAccountingCompletionError("receivedAmount");

    expect(document.activeElement).toBe(document.getElementById("receivedAmount"));
    expect(document.activeElement).not.toBe(document.getElementById("postCloseReason"));
  });

  it("receivedAmount id が無いときは payment-split-0-received にフォールバックする", () => {
    mountFocusFields({ receivedAmount: false, paymentSplit: true });

    focusAccountingCompletionError("receivedAmount");

    expect(document.activeElement).toBe(document.getElementById("payment-split-0-received"));
    expect(document.activeElement).not.toBe(document.getElementById("postCloseReason"));
  });

  it("対象要素が無くても throw しない", () => {
    expect(() => focusAccountingCompletionError("postCloseReason")).not.toThrow();
    expect(() => focusAccountingCompletionError("receivedAmount")).not.toThrow();
  });
});

describe("useAccountingCompletionAction post-close 400 focus (BUG-009)", () => {
  beforeEach(() => {
    completeAccountingMock.mockReset();
    updateAccountingMock.mockReset();
    handleApiErrorMock.mockReset();
    vi.mocked(toast.error).mockReset();
    vi.mocked(toast.success).mockReset();
  });

  it("post_close_reason の 400 では textarea#postCloseReason にフォーカスし toast は handleApiError のみ", async () => {
    updateAccountingMock.mockRejectedValue(axiosError(400, { error: POST_CLOSE_REASON_400 }));
    mountFocusFields({ receivedAmount: true, paymentSplit: true });
    const { result } = renderHook(() => useAccountingCompletionAction(buildHookArgs()));

    await submitCompletionAction(result.current.formAction);

    await waitFor(() => {
      expect(result.current.formState.focusTarget).toBe("postCloseReason");
    });
    expect(document.activeElement).toBe(document.getElementById("postCloseReason"));
    expect(document.activeElement).not.toBe(document.getElementById("receivedAmount"));
    expect(handleApiErrorMock).toHaveBeenCalledTimes(1);
    expect(handleApiErrorMock).toHaveBeenCalledWith(expect.anything(), "会計の処理");
    expect(toast.error).not.toHaveBeenCalled();
  });

  it("post_close_reason を含まない 400 では receivedAmount にフォーカスし postCloseReason は触らない", async () => {
    updateAccountingMock.mockRejectedValue(axiosError(400, { error: GENERIC_400 }));
    mountFocusFields({ receivedAmount: true, paymentSplit: true });
    const { result } = renderHook(() => useAccountingCompletionAction(buildHookArgs()));

    await submitCompletionAction(result.current.formAction);

    await waitFor(() => {
      expect(result.current.formState.focusTarget).toBe("receivedAmount");
    });
    expect(document.activeElement).toBe(document.getElementById("receivedAmount"));
    expect(document.activeElement).not.toBe(document.getElementById("postCloseReason"));
    expect(handleApiErrorMock).toHaveBeenCalledTimes(1);
  });

  it("新規 complete の post_close_reason 400 でも postCloseReason にフォーカスする", async () => {
    completeAccountingMock.mockRejectedValue(axiosError(400, { error: POST_CLOSE_REASON_400 }));
    mountFocusFields({ receivedAmount: true });
    const { result } = renderHook(() =>
      useAccountingCompletionAction(buildHookArgs({ accountingId: undefined })),
    );

    await submitCompletionAction(result.current.formAction);

    await waitFor(() => {
      expect(result.current.formState.focusTarget).toBe("postCloseReason");
    });
    expect(document.activeElement).toBe(document.getElementById("postCloseReason"));
    expect(completeAccountingMock).toHaveBeenCalledTimes(1);
    expect(updateAccountingMock).not.toHaveBeenCalled();
  });
});

// FE-RC-001: fieldset disabled 等の render 側ガードをバイパスされても action 側で権限を再検証し、fail-closed で API を叩かないことを保証する。
describe("useAccountingCompletionAction permissions (FE-RC-001 fail-closed)", () => {
  beforeEach(() => {
    completeAccountingMock.mockReset();
    updateAccountingMock.mockReset();
    handleApiErrorMock.mockReset();
    vi.mocked(toast.error).mockReset();
    vi.mocked(toast.success).mockReset();
  });

  it("canEdit=false（既存会計）では updateAccounting を呼ばず権限エラー toast を出す", async () => {
    const { result } = renderHook(() =>
      useAccountingCompletionAction({
        ...buildHookArgs(),
        permissions: { canCreate: true, canEdit: false },
      }),
    );

    await submitCompletionAction(result.current.formAction);

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
    });
    expect(updateAccountingMock).not.toHaveBeenCalled();
  });

  it("canCreate=false（新規会計）では completeAccounting を呼ばない", async () => {
    const { result } = renderHook(() =>
      useAccountingCompletionAction({
        ...buildHookArgs({ accountingId: undefined }),
        permissions: { canCreate: false, canEdit: true },
      }),
    );

    await submitCompletionAction(result.current.formAction);

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
    });
    expect(completeAccountingMock).not.toHaveBeenCalled();
  });

  it("permissions 未指定（既定 deny）では API を呼ばない", async () => {
    const { permissions, ...argsWithoutPermissions } = buildHookArgs();
    void permissions;
    const { result } = renderHook(() => useAccountingCompletionAction(argsWithoutPermissions));

    await submitCompletionAction(result.current.formAction);

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
    });
    expect(updateAccountingMock).not.toHaveBeenCalled();
  });
});
