import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useInventoryForm } from "./use-inventory-form";

// ──────────────────────────────────────────────────────────
// モック定義
// vi.mock はホイストされるため、参照する変数は vi.hoisted で先に定義する
// ──────────────────────────────────────────────────────────

const { mockToast, mockCreateMutateAsync, mockUpdateMutateAsync } = vi.hoisted(() => ({
  mockToast: { error: vi.fn(), success: vi.fn() },
  mockCreateMutateAsync: vi.fn().mockResolvedValue({}),
  mockUpdateMutateAsync: vi.fn().mockResolvedValue({}),
}));

vi.mock("react-router", () => ({
  useNavigate: vi.fn(() => vi.fn()),
}));

vi.mock("sonner", () => ({ toast: mockToast }));

vi.mock("../api/inventory", () => ({
  useGetInventoryItem: vi.fn(() => ({ data: null, isLoading: false })),
  useCreateInventoryItem: vi.fn(() => ({ mutateAsync: mockCreateMutateAsync })),
  useUpdateInventoryItem: vi.fn(() => ({ mutateAsync: mockUpdateMutateAsync })),
}));

// ──────────────────────────────────────────────────────────
// ヘルパー: テスト用 FormData を生成
// ──────────────────────────────────────────────────────────

function makeFormData(overrides: Record<string, string> = {}): FormData {
  const defaults: Record<string, string> = {
    name: "テスト商品",
    quantity: "100",
    minStockLevel: "10",
    unit: "個",
    location: "",
    supplier: "",
    expiryDate: "",
    lastRestocked: "",
  };
  const fd = new FormData();
  const merged = { ...defaults, ...overrides };
  for (const [key, value] of Object.entries(merged)) {
    fd.append(key, value);
  }
  return fd;
}

// ──────────────────────────────────────────────────────────
// テスト
// ──────────────────────────────────────────────────────────

describe("useInventoryForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateMutateAsync.mockResolvedValue({});
    mockUpdateMutateAsync.mockResolvedValue({});
  });

  // ──────────────────────────
  // 最低在庫数は発注点なので、現在庫より大きくても保存できる
  // ──────────────────────────
  describe("バリデーション: minStockLevel と quantity の関係", () => {
    it("minStockLevel > quantity のとき在庫不足状態として保存できる", async () => {
      const { result } = renderHook(() => useInventoryForm());

      await act(async () => {
        await result.current.formAction(
          makeFormData({ quantity: "100", minStockLevel: "150" })
        );
      });

      expect(result.current.formState.fieldErrors?.minStockLevel).toBeUndefined();
      expect(mockCreateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ quantity: 100, min_stock_level: 150 }),
      );
    });

    it("minStockLevel === quantity のとき バリデーション通過（createMutation が呼ばれる）", async () => {
      const { result } = renderHook(() => useInventoryForm());

      await act(async () => {
        await result.current.formAction(
          makeFormData({ quantity: "100", minStockLevel: "100" })
        );
      });

      expect(result.current.formState.fieldErrors?.minStockLevel).toBeUndefined();
      expect(mockCreateMutateAsync).toHaveBeenCalled();
    });

    it("minStockLevel < quantity のとき バリデーション通過（createMutation が呼ばれる）", async () => {
      const { result } = renderHook(() => useInventoryForm());

      await act(async () => {
        await result.current.formAction(
          makeFormData({ quantity: "100", minStockLevel: "50" })
        );
      });

      expect(result.current.formState.fieldErrors?.minStockLevel).toBeUndefined();
      expect(mockCreateMutateAsync).toHaveBeenCalled();
    });
  });

  // ──────────────────────────
  // 新規作成
  // ──────────────────────────
  describe("新規作成モード（id なし）", () => {
    it("バリデーション通過時に createMutation.mutateAsync が呼ばれる", async () => {
      const { result } = renderHook(() => useInventoryForm());

      await act(async () => {
        await result.current.formAction(makeFormData());
      });

      expect(mockCreateMutateAsync).toHaveBeenCalledTimes(1);
      expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    });

    it("成功時に toast.success が呼ばれ formState.success = true になる", async () => {
      const { result } = renderHook(() => useInventoryForm());

      await act(async () => {
        await result.current.formAction(makeFormData());
      });

      expect(mockToast.success).toHaveBeenCalledWith("在庫情報を登録しました");
      expect(result.current.formState.success).toBe(true);
    });

    it("mutateAsync が reject した場合 formState.success = false になる", async () => {
      mockCreateMutateAsync.mockRejectedValueOnce(new Error("API Error"));

      const { result } = renderHook(() => useInventoryForm());

      await act(async () => {
        await result.current.formAction(makeFormData());
      });

      expect(result.current.formState.success).toBe(false);
    });
  });

  // ──────────────────────────
  // 編集モード
  // ──────────────────────────
  describe("編集モード（id あり）", () => {
    it("バリデーション通過時に updateMutation.mutateAsync が呼ばれる", async () => {
      const { result } = renderHook(() => useInventoryForm("42"));

      await act(async () => {
        await result.current.formAction(makeFormData());
      });

      expect(mockUpdateMutateAsync).toHaveBeenCalledTimes(1);
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("updateMutation.mutateAsync には { id, req } の形で渡される", async () => {
      const { result } = renderHook(() => useInventoryForm("42"));

      await act(async () => {
        await result.current.formAction(
          makeFormData({ name: "編集商品", quantity: "200", minStockLevel: "20" })
        );
      });

      expect(mockUpdateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          id: "42",
          req: expect.objectContaining({
            name: "編集商品",
            quantity: 200,
            min_stock_level: 20,
          }),
        })
      );
    });

    it("成功時に toast.success が呼ばれ formState.success = true になる", async () => {
      const { result } = renderHook(() => useInventoryForm("42"));

      await act(async () => {
        await result.current.formAction(makeFormData());
      });

      expect(mockToast.success).toHaveBeenCalledWith("在庫情報を更新しました");
      expect(result.current.formState.success).toBe(true);
    });

    it("mutateAsync が reject した場合 formState.success = false になる", async () => {
      mockUpdateMutateAsync.mockRejectedValueOnce(new Error("API Error"));

      const { result } = renderHook(() => useInventoryForm("42"));

      await act(async () => {
        await result.current.formAction(makeFormData());
      });

      expect(result.current.formState.success).toBe(false);
    });
  });

  // ──────────────────────────
  // 返り値の初期値
  // ──────────────────────────
  describe("返り値の初期値", () => {
    it("id なし → isEdit = false", () => {
      const { result } = renderHook(() => useInventoryForm());
      expect(result.current.isEdit).toBe(false);
    });

    it("id あり → isEdit = true", () => {
      const { result } = renderHook(() => useInventoryForm("1"));
      expect(result.current.isEdit).toBe(true);
    });

    it("formState の初期値は success: false", () => {
      const { result } = renderHook(() => useInventoryForm());
      expect(result.current.formState.success).toBe(false);
    });

    it("isPending の初期値は false", () => {
      const { result } = renderHook(() => useInventoryForm());
      expect(result.current.isPending).toBe(false);
    });

    it("category の初期値は 'medicine'", () => {
      const { result } = renderHook(() => useInventoryForm());
      expect(result.current.category).toBe("medicine");
    });
  });

  // ──────────────────────────
  // setCategory / setExpiryDate / setLastRestocked
  // ──────────────────────────
  describe("状態更新ハンドラ", () => {
    it("setCategory でカテゴリを変更できる", () => {
      const { result } = renderHook(() => useInventoryForm());

      act(() => {
        result.current.setCategory("food" as Parameters<typeof result.current.setCategory>[0]);
      });

      expect(result.current.category).toBe("food");
    });

    it("setExpiryDate で有効期限を変更できる", () => {
      const { result } = renderHook(() => useInventoryForm());

      act(() => {
        result.current.setExpiryDate("2026-12-31");
      });

      expect(result.current.resolvedExpiry).toBe("2026-12-31");
    });

    it("setLastRestocked で最終補充日を変更できる", () => {
      const { result } = renderHook(() => useInventoryForm());

      act(() => {
        result.current.setLastRestocked("2026-04-01");
      });

      expect(result.current.resolvedLastRestocked).toBe("2026-04-01");
    });
  });

  // ──────────────────────────
  // バリデーション: 境界値
  // ──────────────────────────
  describe("バリデーション境界値", () => {
    it("quantity = 0, minStockLevel = 0 → バリデーション通過（0 === 0）", async () => {
      const { result } = renderHook(() => useInventoryForm());

      await act(async () => {
        await result.current.formAction(
          makeFormData({ quantity: "0", minStockLevel: "0" })
        );
      });

      expect(result.current.formState.fieldErrors?.minStockLevel).toBeUndefined();
      expect(mockCreateMutateAsync).toHaveBeenCalled();
    });

    it("quantity = 1, minStockLevel = 2 → 残少状態として保存できる", async () => {
      const { result } = renderHook(() => useInventoryForm());

      await act(async () => {
        await result.current.formAction(
          makeFormData({ quantity: "1", minStockLevel: "2" })
        );
      });

      expect(result.current.formState.success).toBe(true);
      expect(mockCreateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ quantity: 1, min_stock_level: 2 }),
      );
    });

    it.each([
      ["-1", "1", "現在庫数"],
      ["1", "-1", "最低在庫数"],
      ["not-a-number", "1", "現在庫数"],
    ])("quantity=%s, min=%s は拒否する", async (quantity, minStockLevel, fieldLabel) => {
      const { result } = renderHook(() => useInventoryForm());

      await act(async () => {
        await result.current.formAction(makeFormData({ quantity, minStockLevel }));
      });

      expect(result.current.formState.success).toBe(false);
      expect(Object.values(result.current.formState.fieldErrors ?? {}).join(" ")).toContain(fieldLabel);
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });
  });
});
