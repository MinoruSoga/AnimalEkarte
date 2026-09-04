import { describe, expect, it, vi, beforeEach } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { QueryClient } from "@tanstack/react-query";

import { toast } from "sonner";

import { createBillingItem } from "../api/create-billing-item";
import { deleteBillingItem } from "../api/delete-billing-item";
import { updateBillingItem } from "../api/update-billing-item";
import { buildPostCloseReasonField, useAccountingItemActions } from "./use-accounting-item-actions";

vi.mock("../api/create-billing-item", () => ({
  createBillingItem: vi.fn(),
}));
vi.mock("../api/delete-billing-item", () => ({
  deleteBillingItem: vi.fn(),
}));
vi.mock("../api/update-billing-item", () => ({
  updateBillingItem: vi.fn(),
}));
vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));
vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

const createBillingItemMock = vi.mocked(createBillingItem);
const deleteBillingItemMock = vi.mocked(deleteBillingItem);
const updateBillingItemMock = vi.mocked(updateBillingItem);

function runNow(cb: () => void) {
  void cb();
}

function buildParams(
  overrides: {
    postCloseReason?: string;
    accountingId?: string;
  } = {},
) {
  const setLocalItems = vi.fn();
  const setNewItemOpen = vi.fn();
  const queryClient = {
    refetchQueries: vi.fn().mockResolvedValue(undefined),
    invalidateQueries: vi.fn().mockResolvedValue(undefined),
  } as unknown as QueryClient;

  return {
    accountingId: overrides.accountingId ?? "42",
    baseItems: [],
    queryClient,
    setLocalItems,
    setNewItemOpen,
    startAddItemTransition: runNow,
    startDeleteItemTransition: runNow,
    startItemUpdateTransition: runNow,
    postCloseReason: overrides.postCloseReason,
    // FE-RC-001: このテスト群は「権限あり」時の通常フローを検証するため既定で全許可する。
    // fail-closed の検証は別 describe（権限なしブロック）で行う。
    permissions: { canCreate: true, canEdit: true, canDelete: true },
  };
}

describe("buildPostCloseReasonField", () => {
  it("trim 後の非空理由だけ post_close_reason を返す", () => {
    expect(buildPostCloseReasonField("  入力誤り  ")).toEqual({
      post_close_reason: "入力誤り",
    });
    expect(buildPostCloseReasonField("")).toEqual({});
    expect(buildPostCloseReasonField("   ")).toEqual({});
    expect(buildPostCloseReasonField(undefined)).toEqual({});
  });
});

describe("useAccountingItemActions post_close_reason (BUG-021)", () => {
  beforeEach(() => {
    createBillingItemMock.mockReset();
    deleteBillingItemMock.mockReset();
    updateBillingItemMock.mockReset();
    createBillingItemMock.mockResolvedValue({} as never);
    deleteBillingItemMock.mockResolvedValue(undefined);
    updateBillingItemMock.mockResolvedValue({} as never);
  });

  it("理由ありの明細追加で createBillingItem に post_close_reason を送る", async () => {
    const params = buildParams({ postCloseReason: "締め後の誤入力修正" });
    const { result } = renderHook(() => useAccountingItemActions(params));

    act(() => {
      result.current.handleAddItem({
        name: "追加明細",
        price: "1000",
        category: "other",
        otherReason: "手入力",
      });
    });

    await waitFor(() => {
      expect(createBillingItemMock).toHaveBeenCalledTimes(1);
    });
    expect(createBillingItemMock).toHaveBeenCalledWith(
      expect.objectContaining({
        billing_id: 42,
        name: "追加明細",
        post_close_reason: "締め後の誤入力修正",
      }),
    );
  });

  it("理由なしの明細追加では post_close_reason を送らない（通常フロー回帰）", async () => {
    const params = buildParams({ postCloseReason: "" });
    const { result } = renderHook(() => useAccountingItemActions(params));

    act(() => {
      result.current.handleAddItem({
        name: "通常明細",
        price: "500",
        category: "goods",
      });
    });

    await waitFor(() => {
      expect(createBillingItemMock).toHaveBeenCalledTimes(1);
    });
    expect(createBillingItemMock.mock.calls[0]?.[0]).not.toHaveProperty("post_close_reason");
  });

  it("理由ありの削除で deleteBillingItem に body を渡す", async () => {
    const params = buildParams({ postCloseReason: "誤追加のため削除" });
    const { result } = renderHook(() => useAccountingItemActions(params));

    act(() => {
      result.current.handleDeleteItem("99");
    });

    await waitFor(() => {
      expect(deleteBillingItemMock).toHaveBeenCalledTimes(1);
    });
    expect(deleteBillingItemMock).toHaveBeenCalledWith("99", {
      post_close_reason: "誤追加のため削除",
    });
  });

  it("理由ありの税区分更新で updateBillingItem に post_close_reason を送る", async () => {
    const params = buildParams({ postCloseReason: "税率修正" });
    const { result } = renderHook(() => useAccountingItemActions(params));

    act(() => {
      result.current.handleUpdateItemTax("7", "included", 0.08);
    });

    await waitFor(() => {
      expect(updateBillingItemMock).toHaveBeenCalledTimes(1);
    });
    expect(updateBillingItemMock).toHaveBeenCalledWith(
      "7",
      expect.objectContaining({
        tax_type: "included",
        tax_rate: 0.08,
        post_close_reason: "税率修正",
      }),
    );
  });

  it("理由ありの割引更新で updateBillingItem に post_close_reason を送る", async () => {
    const params = buildParams({ postCloseReason: "割引訂正" });
    const { result } = renderHook(() => useAccountingItemActions(params));

    act(() => {
      result.current.handleUpdateItemDiscount("8", 100);
    });

    await waitFor(() => {
      expect(updateBillingItemMock).toHaveBeenCalledTimes(1);
    });
    expect(updateBillingItemMock).toHaveBeenCalledWith(
      "8",
      expect.objectContaining({
        discount_amount: 100,
        post_close_reason: "割引訂正",
      }),
    );
  });
});

// FE-RC-001: fieldset disabled 等の render 側ガードをバイパスされても各 handler が fail-closed で API を叩かないことを保証する。
describe("useAccountingItemActions permissions (FE-RC-001 fail-closed)", () => {
  beforeEach(() => {
    createBillingItemMock.mockReset();
    deleteBillingItemMock.mockReset();
    updateBillingItemMock.mockReset();
    vi.mocked(toast.error).mockClear();
  });

  it("permissions 未指定（既定 deny）では handleAddItem が createBillingItem を呼ばない", () => {
    const params = buildParams();
    const paramsWithoutPermissions = { ...params, permissions: undefined };
    const { result } = renderHook(() => useAccountingItemActions(paramsWithoutPermissions));

    act(() => {
      result.current.handleAddItem({ name: "追加明細", price: "1000", category: "goods" });
    });

    expect(createBillingItemMock).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
  });

  it("canDelete=false では handleDeleteItem が deleteBillingItem を呼ばない", () => {
    const params = buildParams();
    const { result } = renderHook(() =>
      useAccountingItemActions({
        ...params,
        permissions: { canCreate: true, canEdit: true, canDelete: false },
      }),
    );

    act(() => {
      result.current.handleDeleteItem("99");
    });

    expect(deleteBillingItemMock).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
  });

  it("canEdit=false では handleUpdateItemTax / handleUpdateItemDiscount が API を呼ばない", () => {
    const params = buildParams();
    const { result } = renderHook(() =>
      useAccountingItemActions({
        ...params,
        permissions: { canCreate: true, canEdit: false, canDelete: true },
      }),
    );

    act(() => {
      result.current.handleUpdateItemTax("7", "included", 0.08);
      result.current.handleUpdateItemDiscount("8", 100);
    });

    expect(updateBillingItemMock).not.toHaveBeenCalled();
  });
});
