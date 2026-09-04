import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import { useGetInventoryItem } from "../api/inventory";
import { useInventoryForm, type InventoryMutationPermissions } from "./use-inventory-form";

// FE-RC-220: 既存テスト群は「権限あり」の通常フローを検証するため、
// 既定で create/edit を全許可する。権限拒否の fail-closed 挙動は専用の describe で検証する。
const ALLOW_ALL_PERMISSIONS: Readonly<InventoryMutationPermissions> = {
  canCreate: true,
  canEdit: true,
};

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

function renderInventoryForm(
  id?: string,
  permissions: Readonly<InventoryMutationPermissions> = ALLOW_ALL_PERMISSIONS,
) {
  return renderHook(() => useInventoryForm(id, { permissions }));
}

// ──────────────────────────────────────────────────────────
// モック定義
// vi.mock はホイストされるため、参照する変数は vi.hoisted で先に定義する
// ──────────────────────────────────────────────────────────

const { mockToast, mockCreateMutateAsync, mockUpdateMutateAsync, foundItem } = vi.hoisted(() => ({
  mockToast: { error: vi.fn(), success: vi.fn() },
  mockCreateMutateAsync: vi.fn().mockResolvedValue({}),
  mockUpdateMutateAsync: vi.fn().mockResolvedValue({}),
  foundItem: {
    id: "42",
    clinicId: "1",
    name: "テスト商品",
    category: "medicine",
    quantity: 100,
    unit: "個",
    minStockLevel: 10,
    status: "in_stock",
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
  },
}));

vi.mock("react-router", () => ({
  useNavigate: vi.fn(() => vi.fn()),
}));

vi.mock("sonner", () => ({ toast: mockToast }));

vi.mock("../api/inventory", () => ({
  useGetInventoryItem: vi.fn((id: string) => ({
    data: id ? foundItem : undefined,
    isLoading: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
  })),
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
    vi.mocked(useGetInventoryItem).mockImplementation(
      (id: string) =>
        ({
          data: id ? foundItem : undefined,
          isLoading: false,
          isError: false,
          error: null,
          refetch: vi.fn(),
        }) as unknown as ReturnType<typeof useGetInventoryItem>,
    );
  });

  // ──────────────────────────
  // 最低在庫数は発注点なので、現在庫より大きくても保存できる
  // ──────────────────────────
  describe("バリデーション: minStockLevel と quantity の関係", () => {
    it("minStockLevel > quantity のとき在庫不足状態として保存できる", async () => {
      const { result } = renderInventoryForm();

      await act(async () => {
        await result.current.formAction(makeFormData({ quantity: "100", minStockLevel: "150" }));
      });

      expect(result.current.formState.fieldErrors?.minStockLevel).toBeUndefined();
      expect(mockCreateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ quantity: 100, min_stock_level: 150 }),
      );
    });

    it("minStockLevel === quantity のとき バリデーション通過（createMutation が呼ばれる）", async () => {
      const { result } = renderInventoryForm();

      await act(async () => {
        await result.current.formAction(makeFormData({ quantity: "100", minStockLevel: "100" }));
      });

      expect(result.current.formState.fieldErrors?.minStockLevel).toBeUndefined();
      expect(mockCreateMutateAsync).toHaveBeenCalled();
    });

    it("minStockLevel < quantity のとき バリデーション通過（createMutation が呼ばれる）", async () => {
      const { result } = renderInventoryForm();

      await act(async () => {
        await result.current.formAction(makeFormData({ quantity: "100", minStockLevel: "50" }));
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
      const { result } = renderInventoryForm();

      await act(async () => {
        await result.current.formAction(makeFormData());
      });

      expect(mockCreateMutateAsync).toHaveBeenCalledTimes(1);
      expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    });

    it("成功時に toast.success が呼ばれ formState.success = true になる", async () => {
      const { result } = renderInventoryForm();

      await act(async () => {
        await result.current.formAction(makeFormData());
      });

      expect(mockToast.success).toHaveBeenCalledWith("在庫情報を登録しました");
      expect(result.current.formState.success).toBe(true);
    });

    it("mutateAsync が reject した場合 formState.success = false になる", async () => {
      mockCreateMutateAsync.mockRejectedValueOnce(new Error("API Error"));

      const { result } = renderInventoryForm();

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
      const { result } = renderInventoryForm("42");

      await act(async () => {
        await result.current.formAction(makeFormData());
      });

      expect(mockUpdateMutateAsync).toHaveBeenCalledTimes(1);
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("updateMutation.mutateAsync には { id, req } の形で渡される", async () => {
      const { result } = renderInventoryForm("42");

      await act(async () => {
        await result.current.formAction(
          makeFormData({ name: "編集商品", quantity: "200", minStockLevel: "20" }),
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
        }),
      );
    });

    it("成功時に toast.success が呼ばれ formState.success = true になる", async () => {
      const { result } = renderInventoryForm("42");

      await act(async () => {
        await result.current.formAction(makeFormData());
      });

      expect(mockToast.success).toHaveBeenCalledWith("在庫情報を更新しました");
      expect(result.current.formState.success).toBe(true);
    });

    it("mutateAsync が reject した場合 formState.success = false になる", async () => {
      mockUpdateMutateAsync.mockRejectedValueOnce(new Error("API Error"));

      const { result } = renderInventoryForm("42");

      await act(async () => {
        await result.current.formAction(makeFormData());
      });

      expect(result.current.formState.success).toBe(false);
    });
  });

  // ──────────────────────────
  // BUG-507: 不在ID / 読取失敗は空フォームに折り畳まない
  // ──────────────────────────
  describe("読取分類 (BUG-507)", () => {
    function axiosError(status: number | undefined) {
      const config = { headers: new AxiosHeaders() } as InternalAxiosRequestConfig;
      if (status === undefined) {
        return new AxiosError(
          "Network Error",
          AxiosError.ERR_NETWORK,
          config,
          undefined,
          undefined,
        );
      }
      return new AxiosError("request failed", AxiosError.ERR_BAD_RESPONSE, config, undefined, {
        config,
        data: { error: "not found" },
        headers: new AxiosHeaders(),
        status,
        statusText: "Error",
      });
    }

    it("404 → isReadNotFound、formAction で update/create 0 回", async () => {
      vi.mocked(useGetInventoryItem).mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
        error: axiosError(404),
        refetch: vi.fn(),
      } as unknown as ReturnType<typeof useGetInventoryItem>);

      const { result } = renderInventoryForm("999999001");
      expect(result.current.isReadNotFound).toBe(true);
      expect(result.current.entityRead.status).toBe("notFound");
      expect(result.current.existingItem).toBeUndefined();

      await act(async () => {
        await result.current.formAction(makeFormData());
      });
      expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("403 → isReadNotFound（非開示）で mutation 0 回", async () => {
      vi.mocked(useGetInventoryItem).mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
        error: axiosError(403),
        refetch: vi.fn(),
      } as unknown as ReturnType<typeof useGetInventoryItem>);

      const { result } = renderInventoryForm("42");
      expect(result.current.isReadNotFound).toBe(true);
      expect(result.current.entityRead.status).toBe("forbiddenOrHidden");

      await act(async () => {
        await result.current.formAction(makeFormData());
      });
      expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    });

    it("network error → isReadError と retry、mutation 0 回", async () => {
      const refetch = vi.fn();
      vi.mocked(useGetInventoryItem).mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
        error: axiosError(undefined),
        refetch,
      } as unknown as ReturnType<typeof useGetInventoryItem>);

      const { result } = renderInventoryForm("999999001");
      expect(result.current.isReadError).toBe(true);
      expect(result.current.isReadNotFound).toBe(false);
      result.current.retryRead?.();
      expect(refetch).toHaveBeenCalledTimes(1);

      await act(async () => {
        await result.current.formAction(makeFormData());
      });
      expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    });

    it("create route: idle で isReadNotFound=false", () => {
      vi.mocked(useGetInventoryItem).mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: false,
        error: null,
        refetch: vi.fn(),
      } as unknown as ReturnType<typeof useGetInventoryItem>);

      const { result } = renderInventoryForm();
      expect(result.current.entityRead.status).toBe("idle");
      expect(result.current.isReadNotFound).toBe(false);
      expect(result.current.isEdit).toBe(false);
    });
  });

  // ──────────────────────────
  // 返り値の初期値
  // ──────────────────────────
  describe("返り値の初期値", () => {
    it("id なし → isEdit = false", () => {
      const { result } = renderInventoryForm();
      expect(result.current.isEdit).toBe(false);
    });

    it("id あり → isEdit = true", () => {
      const { result } = renderInventoryForm("1");
      expect(result.current.isEdit).toBe(true);
    });

    it("formState の初期値は success: false", () => {
      const { result } = renderInventoryForm();
      expect(result.current.formState.success).toBe(false);
    });

    it("isPending の初期値は false", () => {
      const { result } = renderInventoryForm();
      expect(result.current.isPending).toBe(false);
    });

    it("category の初期値は 'medicine'", () => {
      const { result } = renderInventoryForm();
      expect(result.current.category).toBe("medicine");
    });
  });

  // ──────────────────────────
  // setCategory / setExpiryDate / setLastRestocked
  // ──────────────────────────
  describe("状態更新ハンドラ", () => {
    it("setCategory でカテゴリを変更できる", () => {
      const { result } = renderInventoryForm();

      act(() => {
        result.current.setCategory("food" as Parameters<typeof result.current.setCategory>[0]);
      });

      expect(result.current.category).toBe("food");
    });

    it("setExpiryDate で有効期限を変更できる", () => {
      const { result } = renderInventoryForm();

      act(() => {
        result.current.setExpiryDate("2026-12-31");
      });

      expect(result.current.resolvedExpiry).toBe("2026-12-31");
    });

    it("setLastRestocked で最終補充日を変更できる", () => {
      const { result } = renderInventoryForm();

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
      const { result } = renderInventoryForm();

      await act(async () => {
        await result.current.formAction(makeFormData({ quantity: "0", minStockLevel: "0" }));
      });

      expect(result.current.formState.fieldErrors?.minStockLevel).toBeUndefined();
      expect(mockCreateMutateAsync).toHaveBeenCalled();
    });

    it("quantity = 1, minStockLevel = 2 → 残少状態として保存できる", async () => {
      const { result } = renderInventoryForm();

      await act(async () => {
        await result.current.formAction(makeFormData({ quantity: "1", minStockLevel: "2" }));
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
      const { result } = renderInventoryForm();

      await act(async () => {
        await result.current.formAction(makeFormData({ quantity, minStockLevel }));
      });

      expect(result.current.formState.success).toBe(false);
      expect(Object.values(result.current.formState.fieldErrors ?? {}).join(" ")).toContain(
        fieldLabel,
      );
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });
  });

  // ──────────────────────────
  // BUG-009: 必須未入力・負値は JS fieldErrors で拒否する
  // ──────────────────────────
  describe("バリデーション: 必須未入力と負値 (BUG-009)", () => {
    it("品名が空なら fieldErrors.name を返し mutation しない", async () => {
      const { result } = renderInventoryForm();

      await act(async () => {
        await result.current.formAction(makeFormData({ name: "" }));
      });

      expect(result.current.formState.success).toBe(false);
      expect(result.current.formState.fieldErrors?.name).toBe("品名を入力してください");
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("品名が空白のみなら fieldErrors.name を返す", async () => {
      const { result } = renderInventoryForm();

      await act(async () => {
        await result.current.formAction(makeFormData({ name: "   " }));
      });

      expect(result.current.formState.fieldErrors?.name).toBe("品名を入力してください");
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("単位が空なら fieldErrors.unit を返し mutation しない", async () => {
      const { result } = renderInventoryForm();

      await act(async () => {
        await result.current.formAction(makeFormData({ unit: "" }));
      });

      expect(result.current.formState.success).toBe(false);
      expect(result.current.formState.fieldErrors?.unit).toBe("単位を入力してください");
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("単位が空白のみなら fieldErrors.unit を返す", async () => {
      const { result } = renderInventoryForm();

      await act(async () => {
        await result.current.formAction(makeFormData({ unit: "   " }));
      });

      expect(result.current.formState.fieldErrors?.unit).toBe("単位を入力してください");
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("現在庫数が負値なら fieldErrors.quantity を返す", async () => {
      const { result } = renderInventoryForm();

      await act(async () => {
        await result.current.formAction(makeFormData({ quantity: "-1" }));
      });

      expect(result.current.formState.success).toBe(false);
      expect(result.current.formState.fieldErrors?.quantity).toBe(
        "現在庫数は0以上の整数で入力してください",
      );
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("最低在庫数が負値なら fieldErrors.minStockLevel を返す", async () => {
      const { result } = renderInventoryForm();

      await act(async () => {
        await result.current.formAction(makeFormData({ minStockLevel: "-1" }));
      });

      expect(result.current.formState.success).toBe(false);
      expect(result.current.formState.fieldErrors?.minStockLevel).toBe(
        "最低在庫数は0以上の整数で入力してください",
      );
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("現在庫数が空なら fieldErrors.quantity を返す", async () => {
      const { result } = renderInventoryForm();

      await act(async () => {
        await result.current.formAction(makeFormData({ quantity: "" }));
      });

      expect(result.current.formState.fieldErrors?.quantity).toBe(
        "現在庫数は0以上の整数で入力してください",
      );
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("最低在庫数が空なら fieldErrors.minStockLevel を返す", async () => {
      const { result } = renderInventoryForm();

      await act(async () => {
        await result.current.formAction(makeFormData({ minStockLevel: "" }));
      });

      expect(result.current.formState.fieldErrors?.minStockLevel).toBe(
        "最低在庫数は0以上の整数で入力してください",
      );
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });
  });
});

// ──────────────────────────────────────────────────────────
// FE-RC-220: permissionsRef + isMutationAllowed による mutation 直前の権限再検査。
// UI の canSubmit をバイパスされても fail-closed で API を呼ばない。
// ──────────────────────────────────────────────────────────

describe("useInventoryForm permissions (FE-RC-220 fail-closed)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateMutateAsync.mockResolvedValue({});
    mockUpdateMutateAsync.mockResolvedValue({});
    vi.mocked(useGetInventoryItem).mockImplementation(
      (id: string) =>
        ({
          data: id ? foundItem : undefined,
          isLoading: false,
          isError: false,
          error: null,
          refetch: vi.fn(),
        }) as unknown as ReturnType<typeof useGetInventoryItem>,
    );
  });

  it("canCreate=false（新規作成）→ createMutation.mutateAsync は呼ばれない", async () => {
    const { result } = renderInventoryForm(undefined, { canCreate: false, canEdit: true });

    await act(async () => {
      await result.current.formAction(makeFormData());
    });

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    expect(result.current.formState.success).toBe(false);
    expect(result.current.formState.error).toBe(PERMISSION_DENIED_MESSAGE);
  });

  it("canEdit=false（編集）→ updateMutation.mutateAsync は呼ばれない", async () => {
    const { result } = renderInventoryForm("42", { canCreate: true, canEdit: false });

    await act(async () => {
      await result.current.formAction(makeFormData());
    });

    expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    expect(result.current.formState.success).toBe(false);
    expect(result.current.formState.error).toBe(PERMISSION_DENIED_MESSAGE);
  });

  it("permissions 未指定（既定 deny）→ create/update いずれも呼ばれない", async () => {
    const { result } = renderHook(() => useInventoryForm());

    await act(async () => {
      await result.current.formAction(makeFormData());
    });

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    expect(result.current.formState.success).toBe(false);
    expect(result.current.formState.error).toBe(PERMISSION_DENIED_MESSAGE);
  });

  it("permissions が後から canCreate=false になると createMutation しない", async () => {
    const { result, rerender } = renderHook(
      ({ permissions }: { permissions: Readonly<InventoryMutationPermissions> }) =>
        useInventoryForm(undefined, { permissions }),
      { initialProps: { permissions: ALLOW_ALL_PERMISSIONS } },
    );

    rerender({ permissions: { canCreate: false, canEdit: true } });

    await act(async () => {
      await result.current.formAction(makeFormData());
    });

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    expect(result.current.formState.success).toBe(false);
  });
});
