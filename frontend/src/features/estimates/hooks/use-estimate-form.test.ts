import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useEstimateForm, type EstimateMutationPermissions } from "./use-estimate-form";
import type { Estimate } from "../types";

// FE-RC-001: このファイルの既存テスト群は「権限あり」の通常フローを検証するため、
// 既定で create/edit を全許可する。権限拒否のfail-closed挙動は専用のdescribeで検証する。
const ALLOW_ALL_PERMISSIONS: Readonly<EstimateMutationPermissions> = {
  canCreate: true,
  canEdit: true,
};

// ──────────────────────────────────────────────────────────
// モック定義
// vi.mock はホイストされるため、参照する変数は vi.hoisted で先に定義する
// ──────────────────────────────────────────────────────────

const {
  mockNavigate,
  mockToast,
  mockCreateMutateAsync,
  mockUpdateMutateAsync,
} = vi.hoisted(() => ({
  mockNavigate: vi.fn(),
  mockToast: { error: vi.fn(), success: vi.fn(), info: vi.fn() },
  mockCreateMutateAsync: vi.fn().mockResolvedValue({}),
  mockUpdateMutateAsync: vi.fn().mockResolvedValue({}),
}));

vi.mock("react-router", () => ({
  useNavigate: () => mockNavigate,
}));

vi.mock("sonner", () => ({ toast: mockToast }));

vi.mock("../api/create-estimate", () => ({
  useCreateEstimate: vi.fn(() => ({ mutateAsync: mockCreateMutateAsync })),
}));

vi.mock("../api/update-estimate", () => ({
  useUpdateEstimate: vi.fn(() => ({ mutateAsync: mockUpdateMutateAsync })),
}));

vi.mock("@/config/paths", () => ({
  paths: {
    estimates: {
      getHref: () => "/estimates",
      detail: { getHref: (id: string) => `/estimates/${id}` },
    },
  },
}));

// ──────────────────────────────────────────────────────────
// テスト用フィクスチャ
// ──────────────────────────────────────────────────────────

const mockEstimate: Estimate = {
  id: "1",
  clinicId: "1",
  estimateNo: "EST-001",
  title: "テスト見積書",
  status: "draft",
  subtotal: 1000,
  taxTotal: 100,
  totalAmount: 1100,
  insuranceAmount: 0,
  discountAmount: 0,
  validUntil: null,
  comment: "",
  notes: "",
  items: [],
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
};

// ──────────────────────────────────────────────────────────
// テスト
// ──────────────────────────────────────────────────────────

describe("useEstimateForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateMutateAsync.mockResolvedValue({});
    mockUpdateMutateAsync.mockResolvedValue({});
  });

  // ──────────────────────────
  // 初期状態
  // ──────────────────────────
  describe("初期状態", () => {
    it("estimate 未指定時 → form.title は空文字", () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));
      expect(result.current.form.title).toBe("");
    });

    it("estimate 指定時 → form.title は estimate.title の値", () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "edit", estimate: mockEstimate, permissions: ALLOW_ALL_PERMISSIONS })
      );
      expect(result.current.form.title).toBe("テスト見積書");
    });

    it("estimate 未指定時 → form.status は 'draft'", () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));
      expect(result.current.form.status).toBe("draft");
    });

    it("formState の初期値は success: false", () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));
      expect(result.current.formState.success).toBe(false);
    });

    it("isPending の初期値は false", () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));
      expect(result.current.isPending).toBe(false);
    });
  });

  // ──────────────────────────
  // バリデーション: title が空の場合
  // ──────────────────────────
  describe("バリデーション: title が空の場合", () => {
    it("title が空文字 → success: false かつ fieldErrors.title が定義される", async () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));
      // form.title は初期値空文字

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(result.current.formState.success).toBe(false);
      expect(result.current.formState.fieldErrors?.title).toBe(
        "タイトルを入力してください"
      );
    });

    it("title が空文字 → createEstimate.mutateAsync は呼ばれない", async () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("title がスペースのみ → success: false かつ fieldErrors.title が定義される（trim チェック）", async () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));

      // title をスペースのみに変更
      act(() => {
        result.current.handleChange("title", "   ");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(result.current.formState.success).toBe(false);
      expect(result.current.formState.fieldErrors?.title).toBe(
        "タイトルを入力してください"
      );
    });

    it("title がスペースのみ → createEstimate.mutateAsync は呼ばれない", async () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));

      act(() => {
        result.current.handleChange("title", "   ");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });
  });

  // ──────────────────────────
  // 新規作成モード
  // ──────────────────────────
  describe("新規作成モード（estimate なし）", () => {
    it("title あり → createEstimate.mutateAsync が呼ばれる", async () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));

      act(() => {
        result.current.handleChange("title", "新規見積書");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockCreateMutateAsync).toHaveBeenCalledTimes(1);
      expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    });

    it("title あり → createEstimate.mutateAsync に title が渡される", async () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));

      act(() => {
        result.current.handleChange("title", "新規見積書");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockCreateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ title: "新規見積書" })
      );
    });

    it("initialPetId あり → create payload に pet_id を載せる（BUG-009）", async () => {
      const { result } = renderHook(() =>
        useEstimateForm({
          mode: "create",
          initialPetId: "8",
          initialOwnerId: "5",
          permissions: ALLOW_ALL_PERMISSIONS,
        })
      );

      act(() => {
        result.current.handleChange("title", "ペット紐付き見積");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockCreateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ pet_id: 8, owner_id: 5 })
      );
    });

    it("成功時 → toast.success はhook側から呼ばれない（FE-RC-005: useCreateEstimateのonSuccessに一元化済み）", async () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));

      act(() => {
        result.current.handleChange("title", "新規見積書");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockToast.success).not.toHaveBeenCalled();
    });

    it("成功時 → formState.success = true", async () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));

      act(() => {
        result.current.handleChange("title", "新規見積書");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(result.current.formState.success).toBe(true);
    });

    it("createEstimate.mutateAsync が失敗した場合 → formState.success = false", async () => {
      mockCreateMutateAsync.mockRejectedValueOnce(new Error("API Error"));

      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));

      act(() => {
        result.current.handleChange("title", "新規見積書");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(result.current.formState.success).toBe(false);
    });

    it("createEstimate.mutateAsync が失敗した場合 → toast.error はhook側から呼ばれない（FE-RC-005: onErrorに一元化済み）", async () => {
      mockCreateMutateAsync.mockRejectedValueOnce(new Error("API Error"));

      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));

      act(() => {
        result.current.handleChange("title", "新規見積書");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockToast.error).not.toHaveBeenCalled();
    });

    it("status が approved → create は呼ばれず fieldErrors.status が返る", async () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));

      act(() => {
        result.current.handleChange("title", "新規見積書");
        result.current.handleChange("status", "approved");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(result.current.formState.success).toBe(false);
      expect(result.current.formState.fieldErrors?.status).toBe(
        "作成時は下書きまたは送付済みのみ選択できます"
      );
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("status が rejected → create は呼ばれず fieldErrors.status が返る", async () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));

      act(() => {
        result.current.handleChange("title", "新規見積書");
        result.current.handleChange("status", "rejected");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(result.current.formState.success).toBe(false);
      expect(result.current.formState.fieldErrors?.status).toBe(
        "作成時は下書きまたは送付済みのみ選択できます"
      );
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("status が sent → create に status: 'sent' が渡る", async () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));

      act(() => {
        result.current.handleChange("title", "新規見積書");
        result.current.handleChange("status", "sent");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockCreateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ title: "新規見積書", status: "sent" })
      );
    });
  });

  // ──────────────────────────
  // 編集モード
  // ──────────────────────────
  describe("編集モード（estimate あり）", () => {
    it("title あり & estimate あり → updateEstimate.mutateAsync が呼ばれる", async () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "edit", estimate: mockEstimate, permissions: ALLOW_ALL_PERMISSIONS }));

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockUpdateMutateAsync).toHaveBeenCalledTimes(1);
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("updateEstimate.mutateAsync に id と title が渡される", async () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "edit", estimate: mockEstimate, permissions: ALLOW_ALL_PERMISSIONS }));

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockUpdateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          id: "1",
          data: expect.objectContaining({ title: "テスト見積書" }),
        })
      );
    });

    it("成功時 → toast.success はhook側から呼ばれない（FE-RC-005: useUpdateEstimateのonSuccessに一元化済み）", async () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "edit", estimate: mockEstimate, permissions: ALLOW_ALL_PERMISSIONS }));

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockToast.success).not.toHaveBeenCalled();
    });

    it("成功時 → formState.success = true", async () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "edit", estimate: mockEstimate, permissions: ALLOW_ALL_PERMISSIONS }));

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(result.current.formState.success).toBe(true);
    });

    it("updateEstimate.mutateAsync が失敗した場合 → formState.success = false", async () => {
      mockUpdateMutateAsync.mockRejectedValueOnce(new Error("API Error"));

      const { result } = renderHook(() => useEstimateForm({ mode: "edit", estimate: mockEstimate, permissions: ALLOW_ALL_PERMISSIONS }));

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(result.current.formState.success).toBe(false);
    });

    it("updateEstimate.mutateAsync が失敗した場合 → toast.error はhook側から呼ばれない（FE-RC-005: onErrorに一元化済み）", async () => {
      mockUpdateMutateAsync.mockRejectedValueOnce(new Error("API Error"));

      const { result } = renderHook(() => useEstimateForm({ mode: "edit", estimate: mockEstimate, permissions: ALLOW_ALL_PERMISSIONS }));

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockToast.error).not.toHaveBeenCalled();
    });

    it("status が approved → update に status: 'approved' が渡る", async () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "edit", estimate: mockEstimate, permissions: ALLOW_ALL_PERMISSIONS }));

      act(() => {
        result.current.handleChange("status", "approved");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockUpdateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          id: "1",
          data: expect.objectContaining({ status: "approved" }),
        })
      );
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("既存 status が approved → update は呼ばれず toast.info を出す", async () => {
      const locked: Estimate = { ...mockEstimate, status: "approved" };
      const { result } = renderHook(() => useEstimateForm({ mode: "edit", estimate: locked, permissions: ALLOW_ALL_PERMISSIONS }));

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
      expect(mockToast.info).toHaveBeenCalledWith(
        "承認済みまたは却下済みの見積書は編集できません",
      );
      expect(result.current.formState.success).toBe(false);
    });

    it("既存 status が rejected → update は呼ばれず toast.info を出す", async () => {
      const locked: Estimate = { ...mockEstimate, status: "rejected" };
      const { result } = renderHook(() => useEstimateForm({ mode: "edit", estimate: locked, permissions: ALLOW_ALL_PERMISSIONS }));

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
      expect(mockToast.info).toHaveBeenCalledWith(
        "承認済みまたは却下済みの見積書は編集できません",
      );
      expect(result.current.formState.success).toBe(false);
    });
  });

  // ──────────────────────────
  // handleChange
  // ──────────────────────────
  describe("handleChange", () => {
    it("handleChange('title', ...) でタイトルを更新できる", () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));

      act(() => {
        result.current.handleChange("title", "更新タイトル");
      });

      expect(result.current.form.title).toBe("更新タイトル");
    });

    it("handleChange('status', ...) でステータスを更新できる", () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));

      act(() => {
        result.current.handleChange("status", "sent");
      });

      expect(result.current.form.status).toBe("sent");
    });

    it("handleChange('subtotal', ...) で小計を更新できる", () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));

      act(() => {
        result.current.handleChange("subtotal", 5000);
      });

      expect(result.current.form.subtotal).toBe(5000);
    });
  });

  // ──────────────────────────
  // handleCancel
  // ──────────────────────────
  describe("handleCancel", () => {
    it("estimate なし → 見積一覧ページへナビゲートする", () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));

      act(() => {
        result.current.handleCancel();
      });

      expect(mockNavigate).toHaveBeenCalledWith("/estimates");
    });

    it("estimate あり → 見積詳細ページへナビゲートする", () => {
      const { result } = renderHook(() => useEstimateForm({ mode: "edit", estimate: mockEstimate, permissions: ALLOW_ALL_PERMISSIONS }));

      act(() => {
        result.current.handleCancel();
      });

      expect(mockNavigate).toHaveBeenCalledWith("/estimates/1");
    });
  });
});

// ──────────────────────────────────────────────────────────
// FE-RC-001: permissionsRef + isMutationAllowed による mutation 直前の権限再検査。
// UI の disabled/非表示をバイパスされても fail-closed で API を呼ばない。
// ──────────────────────────────────────────────────────────

describe("useEstimateForm permissions (FE-RC-001 fail-closed)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateMutateAsync.mockResolvedValue({});
    mockUpdateMutateAsync.mockResolvedValue({});
  });

  it("canCreate=false（新規作成）→ createEstimate は呼ばれない", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({
        mode: "create",
        permissions: { canCreate: false, canEdit: true },
      }),
    );

    act(() => {
      result.current.handleChange("title", "権限なし新規見積書");
    });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.formState.success).toBe(false);
  });

  it("canEdit=false（編集）→ updateEstimate は呼ばれない", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({
        mode: "edit",
        estimate: mockEstimate,
        permissions: { canCreate: true, canEdit: false },
      }),
    );

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.formState.success).toBe(false);
  });

  it("permissions 未指定（既定 deny）→ create/update いずれも呼ばれない", async () => {
    const { result } = renderHook(() => useEstimateForm({ mode: "create" }));

    act(() => {
      result.current.handleChange("title", "permissions未指定の見積書");
    });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.formState.success).toBe(false);
  });
});

// ──────────────────────────────────────────────────────────
// FE-RC-002/004: 死亡ペットへの新規見積書作成を callback 側でも fail-closed に拒否する。
// render 側の fieldset disabled + banner と同じ判定の二重防壁。
// ──────────────────────────────────────────────────────────

describe("useEstimateForm blockCreateReason (FE-RC-002/004 死亡ペット二重防壁)", () => {
  const DECEASED_REASON = "死亡したペットの見積書は作成できません";

  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateMutateAsync.mockResolvedValue({});
    mockUpdateMutateAsync.mockResolvedValue({});
  });

  it("blockCreateReason あり（新規作成）→ createEstimate は呼ばれず toast.error が理由付きで呼ばれる", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({
        mode: "create",
        permissions: ALLOW_ALL_PERMISSIONS,
        blockCreateReason: DECEASED_REASON,
      }),
    );

    act(() => {
      result.current.handleChange("title", "死亡ペットの見積書");
    });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(DECEASED_REASON);
    expect(result.current.formState.success).toBe(false);
  });

  it("blockCreateReason なし（新規作成）→ createEstimate は通常通り呼ばれる", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );

    act(() => {
      result.current.handleChange("title", "生存ペットの見積書");
    });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(mockCreateMutateAsync).toHaveBeenCalledTimes(1);
    expect(mockToast.error).not.toHaveBeenCalled();
  });

  it("blockCreateReason は編集モードには影響しない（updateEstimate は呼ばれる）", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({
        mode: "edit",
        estimate: mockEstimate,
        permissions: ALLOW_ALL_PERMISSIONS,
        blockCreateReason: DECEASED_REASON,
      }),
    );

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(mockUpdateMutateAsync).toHaveBeenCalledTimes(1);
    expect(mockToast.error).not.toHaveBeenCalledWith(DECEASED_REASON);
  });
});

// ──────────────────────────────────────────────────────────
// BUG-019: edit mode は route param 由来。estimate 未取得時に create へ落とさない
// ──────────────────────────────────────────────────────────

describe("useEstimateForm BUG-019 mode vs found", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateMutateAsync.mockResolvedValue({});
    mockUpdateMutateAsync.mockResolvedValue({});
  });

  it("mode=edit で estimate なし → update/create mutation 0 回", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "edit", permissions: ALLOW_ALL_PERMISSIONS }),
    );

    act(() => {
      result.current.handleChange("title", "空白編集の誤保存を防ぐ");
    });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(result.current.formState.success).toBe(false);
    expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
  });

  it("mode=create は estimate なしでも create を呼ぶ", async () => {
    const { result } = renderHook(() => useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }));
    act(() => {
      result.current.handleChange("title", "新規見積");
    });
    await act(async () => {
      await result.current.formAction(new FormData());
    });
    expect(mockCreateMutateAsync).toHaveBeenCalledTimes(1);
    expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
  });

  it("mode=edit + estimate あり → update を呼ぶ", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "edit", estimate: mockEstimate, permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "更新タイトル");
    });
    await act(async () => {
      await result.current.formAction(new FormData());
    });
    expect(mockUpdateMutateAsync).toHaveBeenCalledTimes(1);
    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
  });
});
