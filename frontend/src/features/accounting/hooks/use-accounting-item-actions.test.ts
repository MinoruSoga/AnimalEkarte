import { describe, expect, it, vi, beforeEach } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { QueryClient } from "@tanstack/react-query";

import { toast } from "sonner";

import { createBillingItem } from "../api/create-billing-item";
import { deleteBillingItem } from "../api/delete-billing-item";
import { updateBillingItem } from "../api/update-billing-item";
import type { AccountingItem } from "../types";
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
    accountingId: "accountingId" in overrides ? overrides.accountingId : "42",
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

// EMR-65 / BUG-BILLING-TAX-TYPE-DROPPED: 物販マスタ由来の tax_type/tax_rate が
// 作成リクエストとローカルプレビューへ伝播する。一律 "excluded"/既定率への潰しを防ぐ。
describe("useAccountingItemActions tax propagation (EMR-65)", () => {
  beforeEach(() => {
    createBillingItemMock.mockReset();
    deleteBillingItemMock.mockReset();
    updateBillingItemMock.mockReset();
    createBillingItemMock.mockResolvedValue({} as never);
  });

  it("taxType=included の追加で createBillingItem に included と税率を送る", async () => {
    const params = buildParams();
    const { result } = renderHook(() => useAccountingItemActions(params));

    act(() => {
      result.current.handleAddItem({
        name: "内税商品",
        price: "1100",
        category: "goods",
        taxType: "included",
        taxRate: 0.08,
        merchandiseItemId: "9",
      });
    });

    await waitFor(() => {
      expect(createBillingItemMock).toHaveBeenCalledTimes(1);
    });
    expect(createBillingItemMock).toHaveBeenCalledWith(
      expect.objectContaining({
        tax_type: "included",
        tax_rate: 0.08,
        merchandise_item_id: 9,
      }),
    );

    // ローカルプレビューも同一の課税区分・税率で税額を出す（内税は税抜部分を抽出）
    const applyUpdate = params.setLocalItems.mock.calls[0]?.[0] as (
      prev: AccountingItem[] | null,
    ) => AccountingItem[];
    const items = applyUpdate(null);
    expect(items.at(-1)).toMatchObject({
      taxType: "included",
      taxRate: 0.08,
      taxAmount: Math.round((1100 * 0.08) / 1.08),
    });
  });

  it("taxType=exempt の追加では税額0で tax_type=exempt を送る", async () => {
    const params = buildParams();
    const { result } = renderHook(() => useAccountingItemActions(params));

    act(() => {
      result.current.handleAddItem({
        name: "非課税商品",
        price: "500",
        category: "goods",
        taxType: "exempt",
        taxRate: 0,
      });
    });

    await waitFor(() => {
      expect(createBillingItemMock).toHaveBeenCalledTimes(1);
    });
    expect(createBillingItemMock).toHaveBeenCalledWith(
      expect.objectContaining({ tax_type: "exempt", tax_rate: 0 }),
    );

    const applyUpdate = params.setLocalItems.mock.calls[0]?.[0] as (
      prev: AccountingItem[] | null,
    ) => AccountingItem[];
    expect(applyUpdate(null).at(-1)).toMatchObject({ taxType: "exempt", taxAmount: 0 });
  });

  it("taxType 未指定の追加は従来どおり excluded + 既定税率で送る", async () => {
    const params = buildParams();
    const { result } = renderHook(() => useAccountingItemActions(params));

    act(() => {
      result.current.handleAddItem({ name: "手入力明細", price: "1000", category: "test" });
    });

    await waitFor(() => {
      expect(createBillingItemMock).toHaveBeenCalledTimes(1);
    });
    expect(createBillingItemMock).toHaveBeenCalledWith(
      expect.objectContaining({ tax_type: "excluded", tax_rate: 0.1 }),
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

  it("canEdit=false では handleUpdateItemName / Quantity / Amount が API を呼ばない (EMR-229/230)", () => {
    const params = buildParams();
    const item = { id: "9", quantity: 1, discountAmount: 0 } as AccountingItem;
    const { result } = renderHook(() =>
      useAccountingItemActions({
        ...params,
        permissions: { canCreate: true, canEdit: false, canDelete: true },
      }),
    );

    act(() => {
      result.current.handleUpdateItemName("7", "社販 2024-05");
      result.current.handleUpdateItemQuantity("8", 2);
      result.current.handleUpdateItemAmount(item, 1500);
    });

    expect(updateBillingItemMock).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
  });
});

// EMR-229/230: 項目名・数量・金額（税抜小計）の明細編集。
// 既存会計は PATCH、新規会計ドラフトは localItems への immutable 反映。
describe("useAccountingItemActions item name/quantity/amount (EMR-229/230)", () => {
  const baseManualItem: AccountingItem = {
    id: "manual_draft",
    category: "goods",
    name: "社販フード",
    unitPrice: 800,
    quantity: 1,
    discountRate: 0,
    discountAmount: 0,
    taxType: "excluded",
    taxRate: 0.1,
    taxAmount: 80,
    subtotal: 800,
    isInsuranceApplicable: false,
    source: "manual",
  };

  beforeEach(() => {
    createBillingItemMock.mockReset();
    deleteBillingItemMock.mockReset();
    updateBillingItemMock.mockReset();
    vi.mocked(toast.error).mockClear();
    updateBillingItemMock.mockResolvedValue({} as never);
  });

  it("項目名の更新は trim して name を PATCH する (EMR-229)", async () => {
    const params = buildParams();
    const { result } = renderHook(() => useAccountingItemActions(params));

    act(() => {
      result.current.handleUpdateItemName("7", "  社販 2024-05  ");
    });

    await waitFor(() => {
      expect(updateBillingItemMock).toHaveBeenCalledTimes(1);
    });
    expect(updateBillingItemMock).toHaveBeenCalledWith(
      "7",
      expect.objectContaining({ name: "社販 2024-05" }),
    );
  });

  it("空白のみの項目名は API を呼ばずエラーを表示する (EMR-229)", () => {
    const params = buildParams();
    const { result } = renderHook(() => useAccountingItemActions(params));

    act(() => {
      result.current.handleUpdateItemName("7", "   ");
    });

    expect(updateBillingItemMock).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("項目名は必須です");
  });

  it("数量の更新は quantity を PATCH する (EMR-230)", async () => {
    const params = buildParams();
    const { result } = renderHook(() => useAccountingItemActions(params));

    act(() => {
      result.current.handleUpdateItemQuantity("7", 3);
    });

    await waitFor(() => {
      expect(updateBillingItemMock).toHaveBeenCalledTimes(1);
    });
    expect(updateBillingItemMock).toHaveBeenCalledWith(
      "7",
      expect.objectContaining({ quantity: 3 }),
    );
  });

  it("数量0以下は API を呼ばない (EMR-230)", () => {
    const params = buildParams();
    const { result } = renderHook(() => useAccountingItemActions(params));

    act(() => {
      result.current.handleUpdateItemQuantity("7", 0);
    });

    expect(updateBillingItemMock).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("数量は正の値で入力してください");
  });

  it("金額の更新は (amount + discountAmount) / quantity を unit_price へ換算して PATCH する (EMR-230)", async () => {
    const params = buildParams();
    const item: AccountingItem = { ...baseManualItem, id: "7", quantity: 2, discountAmount: 100 };
    const { result } = renderHook(() => useAccountingItemActions(params));

    act(() => {
      result.current.handleUpdateItemAmount(item, 1500);
    });

    await waitFor(() => {
      expect(updateBillingItemMock).toHaveBeenCalledTimes(1);
    });
    // (1500 + 100) / 2 = 800
    expect(updateBillingItemMock).toHaveBeenCalledWith(
      "7",
      expect.objectContaining({ unit_price: 800 }),
    );
  });

  it("追加直後の仮ID行（manual_<uuid>）は PATCH せず案内する (EMR-229/230)", () => {
    const params = buildParams();
    const item: AccountingItem = { ...baseManualItem, id: "manual_abc" };
    const { result } = renderHook(() => useAccountingItemActions(params));

    act(() => {
      result.current.handleUpdateItemName("manual_abc", "社販 R7.5");
      result.current.handleUpdateItemQuantity("manual_abc", 2);
      result.current.handleUpdateItemAmount(item, 1000);
    });

    expect(updateBillingItemMock).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("明細の登録が反映されてから編集してください");
  });

  it("新規会計ドラフトの項目名変更は localItems へ immutable に反映し API は呼ばない (EMR-229)", () => {
    const params = buildParams({ accountingId: undefined });
    const { result } = renderHook(() =>
      useAccountingItemActions({ ...params, baseItems: [baseManualItem] }),
    );

    act(() => {
      result.current.handleUpdateItemName("manual_draft", "社販フード R7.5");
    });

    const applyUpdate = params.setLocalItems.mock.calls[0]?.[0] as (
      prev: AccountingItem[] | null,
    ) => AccountingItem[];
    const items = applyUpdate(null);
    expect(items[0]?.name).toBe("社販フード R7.5");
    // 非対象行・元オブジェクトは変更しない（immutable）
    expect(baseManualItem.name).toBe("社販フード");
    expect(updateBillingItemMock).not.toHaveBeenCalled();
  });

  it("ドラフトの数量変更は subtotal/taxAmount を BE 同式で再計算する (EMR-230)", () => {
    const params = buildParams({ accountingId: undefined });
    const { result } = renderHook(() =>
      useAccountingItemActions({ ...params, baseItems: [baseManualItem] }),
    );

    act(() => {
      result.current.handleUpdateItemQuantity("manual_draft", 3);
    });

    const applyUpdate = params.setLocalItems.mock.calls[0]?.[0] as (
      prev: AccountingItem[] | null,
    ) => AccountingItem[];
    const items = applyUpdate(null);
    // 800 * 3 - 0 = 2400、外税 10% → 240
    expect(items[0]).toMatchObject({ quantity: 3, subtotal: 2400, taxAmount: 240 });
    expect(updateBillingItemMock).not.toHaveBeenCalled();
  });

  it("ドラフトの金額変更は unit_price 換算して localItems を更新する (EMR-230)", () => {
    const params = buildParams({ accountingId: undefined });
    const item: AccountingItem = {
      ...baseManualItem,
      quantity: 2,
      discountAmount: 100,
      subtotal: 1500,
      taxAmount: 150,
    };
    const { result } = renderHook(() => useAccountingItemActions({ ...params, baseItems: [item] }));

    act(() => {
      result.current.handleUpdateItemAmount(item, 2000);
    });

    const applyUpdate = params.setLocalItems.mock.calls[0]?.[0] as (
      prev: AccountingItem[] | null,
    ) => AccountingItem[];
    const items = applyUpdate(null);
    // unit_price = (2000 + 100) / 2 = 1050 → subtotal = 1050*2 - 100 = 2000, tax = 200
    expect(items[0]).toMatchObject({ unitPrice: 1050, subtotal: 2000, taxAmount: 200 });
    expect(updateBillingItemMock).not.toHaveBeenCalled();
  });
});

// EMR-229: 手入力追加行への飼主マスタ割引率の optimistic 事前適用。
describe("useAccountingItemActions owner discount prefill (EMR-229)", () => {
  beforeEach(() => {
    createBillingItemMock.mockReset();
    updateBillingItemMock.mockReset();
    createBillingItemMock.mockResolvedValue({} as never);
  });

  it("ownerDiscountRate がある追加は割引を optimistic 適用し、POST には割引を送らない", async () => {
    const params = buildParams();
    const { result } = renderHook(() =>
      useAccountingItemActions({ ...params, ownerDiscountRate: 10 }),
    );

    act(() => {
      result.current.handleAddItem({ name: "社販フード", price: "1000", category: "food" });
    });

    // ローカルプレビュー: 1000 * 10% = 100 引、小計 900、外税 90
    const applyUpdate = params.setLocalItems.mock.calls[0]?.[0] as (
      prev: AccountingItem[] | null,
    ) => AccountingItem[];
    expect(applyUpdate(null).at(-1)).toMatchObject({
      discountRate: 10,
      discountAmount: 100,
      subtotal: 900,
      taxAmount: 90,
    });

    await waitFor(() => {
      expect(createBillingItemMock).toHaveBeenCalledTimes(1);
    });
    // POST body には割引を載せない（BE が飼主率 vs キャンペーンの大きい方で再解決する）
    const req = createBillingItemMock.mock.calls[0]?.[0];
    expect(req).not.toHaveProperty("discount_amount");
    expect(req).not.toHaveProperty("discount_rate");
  });

  it("ownerDiscountRate 未指定の追加は従来どおり割引0で表示する（回帰）", async () => {
    const params = buildParams();
    const { result } = renderHook(() => useAccountingItemActions(params));

    act(() => {
      result.current.handleAddItem({ name: "通常明細", price: "500", category: "goods" });
    });

    const applyUpdate = params.setLocalItems.mock.calls[0]?.[0] as (
      prev: AccountingItem[] | null,
    ) => AccountingItem[];
    expect(applyUpdate(null).at(-1)).toMatchObject({
      discountRate: 0,
      discountAmount: 0,
      subtotal: 500,
      taxAmount: 50,
    });
  });
});
