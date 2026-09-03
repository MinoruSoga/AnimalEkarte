import { describe, expect, it, vi, beforeEach } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { QueryClient } from "@tanstack/react-query";
import type { NavigateFunction } from "react-router";

import { toast } from "sonner";

import { cancelAccounting } from "../api/cancel-accounting";
import { createRefund } from "../api/create-refund";
import { useAccountingSettlementActions } from "./use-accounting-settlement-actions";

vi.mock("../api/cancel-accounting", () => ({
  cancelAccounting: vi.fn(),
}));
vi.mock("../api/create-refund", () => ({
  createRefund: vi.fn(),
}));
vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));
vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

const cancelAccountingMock = vi.mocked(cancelAccounting);
const createRefundMock = vi.mocked(createRefund);

function runNow(cb: () => void) {
  void cb();
}

function buildParams(overrides: { canCancel?: boolean; canEdit?: boolean } = {}) {
  const queryClient = {
    invalidateQueries: vi.fn().mockResolvedValue(undefined),
  } as unknown as QueryClient;

  return {
    accountingId: "42",
    navigate: vi.fn() as unknown as NavigateFunction,
    queryClient,
    setCancelConfirmOpen: vi.fn(),
    setPreviewOpen: vi.fn(),
    startCancelTransition: runNow,
    startRefundTransition: runNow,
    canCancel: overrides.canCancel ?? true,
    canEdit: overrides.canEdit ?? true,
  };
}

describe("useAccountingSettlementActions permissions (FE-RC-109)", () => {
  beforeEach(() => {
    cancelAccountingMock.mockReset();
    createRefundMock.mockReset();
    cancelAccountingMock.mockResolvedValue(undefined);
    createRefundMock.mockResolvedValue({} as never);
    vi.mocked(toast.error).mockClear();
    vi.mocked(toast.success).mockClear();
  });

  it("canEdit=false では handleRefund が createRefund を呼ばず toast する", () => {
    const params = buildParams({ canEdit: false });
    const { result } = renderHook(() => useAccountingSettlementActions(params));

    act(() => {
      result.current.handleRefund(1000, "返金理由");
    });

    expect(createRefundMock).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
  });

  it("canCancel=false では handleCancelConfirm が cancelAccounting を呼ばない", () => {
    const params = buildParams({ canCancel: false });
    const { result } = renderHook(() => useAccountingSettlementActions(params));

    act(() => {
      result.current.handleCancelConfirm();
    });

    expect(cancelAccountingMock).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
  });

  it("権限ありなら handleRefund / handleCancelConfirm が API を呼ぶ", async () => {
    const params = buildParams();
    const { result } = renderHook(() => useAccountingSettlementActions(params));

    act(() => {
      result.current.handleRefund(500, "過請求");
    });
    await waitFor(() => {
      expect(createRefundMock).toHaveBeenCalledWith("42", {
        amount: 500,
        reason: "過請求",
        paymentMethod: undefined,
      });
    });

    act(() => {
      result.current.handleCancelConfirm();
    });
    await waitFor(() => {
      expect(cancelAccountingMock).toHaveBeenCalledWith("42");
    });
  });

  it("rerender で canEdit が false になると最新の ref で返金を拒否する", () => {
    const initial = buildParams({ canEdit: true });
    const { result, rerender } = renderHook(
      (props: ReturnType<typeof buildParams>) => useAccountingSettlementActions(props),
      { initialProps: initial },
    );

    rerender({ ...initial, canEdit: false });

    act(() => {
      result.current.handleRefund(1000, "返金理由");
    });

    expect(createRefundMock).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
  });

  it("rerender で canCancel が false になると最新の ref でキャンセルを拒否する", () => {
    const initial = buildParams({ canCancel: true });
    const { result, rerender } = renderHook(
      (props: ReturnType<typeof buildParams>) => useAccountingSettlementActions(props),
      { initialProps: initial },
    );

    rerender({ ...initial, canCancel: false });

    act(() => {
      result.current.handleCancelConfirm();
    });

    expect(cancelAccountingMock).not.toHaveBeenCalled();
  });
});
