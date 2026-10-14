import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { startTransition } from "react";
import {
  useEstimateForm,
  type EstimateItemDraft,
  type EstimateMutationPermissions,
} from "./use-estimate-form";
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

const { mockNavigate, mockToast, mockCreateMutateAsync, mockUpdateMutateAsync } = vi.hoisted(
  () => ({
    mockNavigate: vi.fn(),
    mockToast: { error: vi.fn(), success: vi.fn(), info: vi.fn() },
    mockCreateMutateAsync: vi.fn().mockResolvedValue({}),
    mockUpdateMutateAsync: vi.fn().mockResolvedValue({}),
  }),
);

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
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );
      expect(result.current.form.title).toBe("");
    });

    it("estimate 指定時 → form.title は estimate.title の値", () => {
      const { result } = renderHook(() =>
        useEstimateForm({
          mode: "edit",
          estimate: mockEstimate,
          permissions: ALLOW_ALL_PERMISSIONS,
        }),
      );
      expect(result.current.form.title).toBe("テスト見積書");
    });

    it("estimate 未指定時 → form.status は 'draft'", () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );
      expect(result.current.form.status).toBe("draft");
    });

    it("formState の初期値は success: false", () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );
      expect(result.current.formState.success).toBe(false);
    });

    it("isPending の初期値は false", () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );
      expect(result.current.isPending).toBe(false);
    });
  });

  // ──────────────────────────
  // バリデーション: title が空の場合
  // ──────────────────────────
  describe("バリデーション: title が空の場合", () => {
    it("title が空文字 → success: false かつ fieldErrors.title が定義される", async () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );
      // form.title は初期値空文字

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(result.current.formState.success).toBe(false);
      expect(result.current.formState.fieldErrors?.title).toBe("タイトルを入力してください");
    });

    it("title が空文字 → createEstimate.mutateAsync は呼ばれない", async () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("title がスペースのみ → success: false かつ fieldErrors.title が定義される（trim チェック）", async () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );

      // title をスペースのみに変更
      act(() => {
        result.current.handleChange("title", "   ");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(result.current.formState.success).toBe(false);
      expect(result.current.formState.fieldErrors?.title).toBe("タイトルを入力してください");
    });

    it("title がスペースのみ → createEstimate.mutateAsync は呼ばれない", async () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );

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
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );

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
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );

      act(() => {
        result.current.handleChange("title", "新規見積書");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockCreateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ title: "新規見積書" }),
      );
    });

    it("initialPetId あり → create payload に pet_id を載せる（BUG-009）", async () => {
      const { result } = renderHook(() =>
        useEstimateForm({
          mode: "create",
          initialPetId: "8",
          initialOwnerId: "5",
          permissions: ALLOW_ALL_PERMISSIONS,
        }),
      );

      act(() => {
        result.current.handleChange("title", "ペット紐付き見積");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockCreateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ pet_id: 8, owner_id: 5 }),
      );
    });

    it("成功時 → toast.success はhook側から呼ばれない（FE-RC-005: useCreateEstimateのonSuccessに一元化済み）", async () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );

      act(() => {
        result.current.handleChange("title", "新規見積書");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockToast.success).not.toHaveBeenCalled();
    });

    it("成功時 → formState.success = true", async () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );

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

      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );

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

      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );

      act(() => {
        result.current.handleChange("title", "新規見積書");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockToast.error).not.toHaveBeenCalled();
    });

    it("status が approved → create は呼ばれず fieldErrors.status が返る", async () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );

      act(() => {
        result.current.handleChange("title", "新規見積書");
        result.current.handleChange("status", "approved");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(result.current.formState.success).toBe(false);
      expect(result.current.formState.fieldErrors?.status).toBe(
        "作成時は下書きまたは送付済みのみ選択できます",
      );
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("status が rejected → create は呼ばれず fieldErrors.status が返る", async () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );

      act(() => {
        result.current.handleChange("title", "新規見積書");
        result.current.handleChange("status", "rejected");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(result.current.formState.success).toBe(false);
      expect(result.current.formState.fieldErrors?.status).toBe(
        "作成時は下書きまたは送付済みのみ選択できます",
      );
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("status が sent → create に status: 'sent' が渡る", async () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );

      act(() => {
        result.current.handleChange("title", "新規見積書");
        result.current.handleChange("status", "sent");
      });

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockCreateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ title: "新規見積書", status: "sent" }),
      );
    });
  });

  // ──────────────────────────
  // 編集モード
  // ──────────────────────────
  describe("編集モード（estimate あり）", () => {
    it("title あり & estimate あり → updateEstimate.mutateAsync が呼ばれる", async () => {
      const { result } = renderHook(() =>
        useEstimateForm({
          mode: "edit",
          estimate: mockEstimate,
          permissions: ALLOW_ALL_PERMISSIONS,
        }),
      );

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockUpdateMutateAsync).toHaveBeenCalledTimes(1);
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("updateEstimate.mutateAsync に id と title が渡される", async () => {
      const { result } = renderHook(() =>
        useEstimateForm({
          mode: "edit",
          estimate: mockEstimate,
          permissions: ALLOW_ALL_PERMISSIONS,
        }),
      );

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockUpdateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          id: "1",
          data: expect.objectContaining({ title: "テスト見積書" }),
        }),
      );
    });

    it("成功時 → toast.success はhook側から呼ばれない（FE-RC-005: useUpdateEstimateのonSuccessに一元化済み）", async () => {
      const { result } = renderHook(() =>
        useEstimateForm({
          mode: "edit",
          estimate: mockEstimate,
          permissions: ALLOW_ALL_PERMISSIONS,
        }),
      );

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockToast.success).not.toHaveBeenCalled();
    });

    it("成功時 → formState.success = true", async () => {
      const { result } = renderHook(() =>
        useEstimateForm({
          mode: "edit",
          estimate: mockEstimate,
          permissions: ALLOW_ALL_PERMISSIONS,
        }),
      );

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(result.current.formState.success).toBe(true);
    });

    it("updateEstimate.mutateAsync が失敗した場合 → formState.success = false", async () => {
      mockUpdateMutateAsync.mockRejectedValueOnce(new Error("API Error"));

      const { result } = renderHook(() =>
        useEstimateForm({
          mode: "edit",
          estimate: mockEstimate,
          permissions: ALLOW_ALL_PERMISSIONS,
        }),
      );

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(result.current.formState.success).toBe(false);
    });

    it("updateEstimate.mutateAsync が失敗した場合 → toast.error はhook側から呼ばれない（FE-RC-005: onErrorに一元化済み）", async () => {
      mockUpdateMutateAsync.mockRejectedValueOnce(new Error("API Error"));

      const { result } = renderHook(() =>
        useEstimateForm({
          mode: "edit",
          estimate: mockEstimate,
          permissions: ALLOW_ALL_PERMISSIONS,
        }),
      );

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockToast.error).not.toHaveBeenCalled();
    });

    it("status が approved → update に status: 'approved' が渡る", async () => {
      const { result } = renderHook(() =>
        useEstimateForm({
          mode: "edit",
          estimate: mockEstimate,
          permissions: ALLOW_ALL_PERMISSIONS,
        }),
      );

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
        }),
      );
      expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    });

    it("既存 status が approved → update は呼ばれず toast.info を出す", async () => {
      const locked: Estimate = { ...mockEstimate, status: "approved" };
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "edit", estimate: locked, permissions: ALLOW_ALL_PERMISSIONS }),
      );

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
      expect(mockToast.info).toHaveBeenCalledWith("承認済みまたは却下済みの見積書は編集できません");
      expect(result.current.formState.success).toBe(false);
    });

    it("既存 status が rejected → update は呼ばれず toast.info を出す", async () => {
      const locked: Estimate = { ...mockEstimate, status: "rejected" };
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "edit", estimate: locked, permissions: ALLOW_ALL_PERMISSIONS }),
      );

      await act(async () => {
        await result.current.formAction(new FormData());
      });

      expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
      expect(mockToast.info).toHaveBeenCalledWith("承認済みまたは却下済みの見積書は編集できません");
      expect(result.current.formState.success).toBe(false);
    });
  });

  // ──────────────────────────
  // handleChange
  // ──────────────────────────
  describe("handleChange", () => {
    it("handleChange('title', ...) でタイトルを更新できる", () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );

      act(() => {
        result.current.handleChange("title", "更新タイトル");
      });

      expect(result.current.form.title).toBe("更新タイトル");
    });

    it("handleChange('status', ...) でステータスを更新できる", () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );

      act(() => {
        result.current.handleChange("status", "sent");
      });

      expect(result.current.form.status).toBe("sent");
    });

    it("小計/税/合計は明細行から導出される（ヘッダー金額の直接上書きなし）", () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );

      act(() => {
        result.current.handleChange("items", [makeItem({ unitPrice: 1100, quantity: 2 })]);
      });

      expect(result.current.totals.subtotal).toBe(2200);
      expect(result.current.totals.taxTotal).toBe(200);
      expect(result.current.totals.totalAmount).toBe(2200);
    });
  });

  // ──────────────────────────
  // handleCancel
  // ──────────────────────────
  describe("handleCancel", () => {
    it("estimate なし → 見積一覧ページへナビゲートする", () => {
      const { result } = renderHook(() =>
        useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
      );

      act(() => {
        result.current.handleCancel();
      });

      expect(mockNavigate).toHaveBeenCalledWith("/estimates");
    });

    it("estimate あり → 見積詳細ページへナビゲートする", () => {
      const { result } = renderHook(() =>
        useEstimateForm({
          mode: "edit",
          estimate: mockEstimate,
          permissions: ALLOW_ALL_PERMISSIONS,
        }),
      );

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
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
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

function makeItem(overrides: Partial<EstimateItemDraft> = {}): EstimateItemDraft {
  return {
    id: "",
    localId: "local-1",
    name: "検査",
    category: "other",
    unitPrice: 1100,
    quantity: 1,
    taxType: "included",
    taxRate: 0.1,
    discountRate: 0,
    discountAmount: 0,
    isInsuranceApplicable: false,
    sortOrder: 0,
    ...overrides,
  };
}

const estimateWithItems: Estimate = {
  ...mockEstimate,
  items: [
    {
      id: "11",
      estimateId: "1",
      name: "薬剤A",
      category: "medicine",
      unitPrice: 1000,
      quantity: 1,
      taxType: "excluded",
      taxRate: 0.1,
      discountRate: 0,
      discountAmount: 100,
      isInsuranceApplicable: true,
      sortOrder: 0,
      createdAt: "2026-01-01T00:00:00Z",
    },
  ],
};

async function submitEstimateAction(formAction: (data: FormData) => void): Promise<void> {
  await act(async () => {
    startTransition(() => formAction(new FormData()));
  });
}

describe("useEstimateForm 明細行 (EMR-401)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateMutateAsync.mockResolvedValue({});
    mockUpdateMutateAsync.mockResolvedValue({});
  });

  it("hydrate: estimate.items の税区分・税率・割引・id を保持する", () => {
    const { result } = renderHook(() =>
      useEstimateForm({
        mode: "edit",
        estimate: estimateWithItems,
        permissions: ALLOW_ALL_PERMISSIONS,
      }),
    );

    expect(result.current.form.items).toHaveLength(1);
    expect(result.current.form.items[0]).toMatchObject({
      id: "11",
      name: "薬剤A",
      category: "medicine",
      unitPrice: 1000,
      quantity: 1,
      taxType: "excluded",
      taxRate: 0.1,
      discountAmount: 100,
      isInsuranceApplicable: true,
    });
    expect(result.current.form.items[0]!.localId).toBeTruthy();
  });

  it("createEmptyItem: 内税・標準税率・数量1・単価0・割引0・カテゴリother", () => {
    const { result } = renderHook(() =>
      useEstimateForm({
        mode: "create",
        permissions: ALLOW_ALL_PERMISSIONS,
        standardTaxRate: 0.1,
      }),
    );

    const row = result.current.createEmptyItem();
    expect(row).toMatchObject({
      name: "",
      category: "other",
      unitPrice: 0,
      quantity: 1,
      discountRate: 0,
      discountAmount: 0,
      taxType: "included",
      taxRate: 0.1,
      isInsuranceApplicable: false,
    });
    expect(row.localId).toBeTruthy();
  });

  it("totals: 内税 1100×2 → subtotal2200 / tax200 / total2200", () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("items", [makeItem({ unitPrice: 1100, quantity: 2 })]);
    });
    expect(result.current.totals).toEqual({ subtotal: 2200, taxTotal: 200, totalAmount: 2200 });
  });

  it("totals: 外税 1000 → tax100 / total1100", () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("items", [
        makeItem({ taxType: "excluded", unitPrice: 1000, quantity: 1 }),
      ]);
    });
    expect(result.current.totals).toEqual({ subtotal: 1000, taxTotal: 100, totalAmount: 1100 });
  });

  it("totals: 非課税 1000 → tax0 / total1000", () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("items", [
        makeItem({ taxType: "exempt", unitPrice: 1000, quantity: 1 }),
      ]);
    });
    expect(result.current.totals).toEqual({ subtotal: 1000, taxTotal: 0, totalAmount: 1000 });
  });

  it("totals: 軽減内税 1080 (.08) → tax80 / total1080", () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("items", [
        makeItem({ taxType: "included", taxRate: 0.08, unitPrice: 1080, quantity: 1 }),
      ]);
    });
    expect(result.current.totals).toEqual({ subtotal: 1080, taxTotal: 80, totalAmount: 1080 });
  });

  it("totals: 外税 46×0.1 → subtotal5 / tax0 / total5（raw税ベースで計算）", () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("items", [
        makeItem({ taxType: "excluded", unitPrice: 46, quantity: 0.1 }),
      ]);
    });
    expect(result.current.totals).toEqual({ subtotal: 5, taxTotal: 0, totalAmount: 5 });
  });

  it("totals: 内税 50×0.1 → subtotal5 / tax0 / total5", () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("items", [
        makeItem({ taxType: "included", unitPrice: 50, quantity: 0.1 }),
      ]);
    });
    expect(result.current.totals).toEqual({ subtotal: 5, taxTotal: 0, totalAmount: 5 });
  });

  it("create payload: items に tax_type/tax_rate/discount 等の全フィールドを載せる", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "明細付き見積");
      result.current.handleChange("items", [
        makeItem({ name: "検査", unitPrice: 1100, quantity: 2 }),
      ]);
    });
    await submitEstimateAction(result.current.formAction);
    expect(mockCreateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        subtotal: 2200,
        tax_total: 200,
        total_amount: 2200,
        items: [
          expect.objectContaining({
            name: "検査",
            category: "other",
            unit_price: 1100,
            quantity: 2,
            tax_type: "included",
            tax_rate: 0.1,
            discount_rate: 0,
            discount_amount: 0,
            is_insurance_applicable: false,
            sort_order: 0,
          }),
        ],
      }),
    );
  });

  it("update payload: タイトルのみ変更で明細無変更 → items と金額フィールドを送らない（ReplaceItems で行が再作成されない）", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({
        mode: "edit",
        estimate: estimateWithItems,
        permissions: ALLOW_ALL_PERMISSIONS,
      }),
    );
    act(() => {
      result.current.handleChange("title", "タイトルだけ変更");
    });
    await submitEstimateAction(result.current.formAction);
    const data = mockUpdateMutateAsync.mock.calls[0]![0].data;
    expect(data.title).toBe("タイトルだけ変更");
    expect(data).not.toHaveProperty("items");
    expect(data).not.toHaveProperty("subtotal");
    expect(data).not.toHaveProperty("tax_total");
    expect(data).not.toHaveProperty("total_amount");
  });

  it("update payload: 同一idの再フェッチで行が変わってもタイトルのみ変更なら items/totals を送らない", async () => {
    const initial: Estimate = {
      ...estimateWithItems,
      items: [{ ...estimateWithItems.items[0]!, unitPrice: 1100 }],
    };
    const refetched: Estimate = {
      ...initial,
      items: [{ ...initial.items[0]!, unitPrice: 2200 }],
    };
    const { result, rerender } = renderHook(
      ({ estimate }: { estimate: Estimate }) =>
        useEstimateForm({
          mode: "edit",
          estimate,
          permissions: ALLOW_ALL_PERMISSIONS,
        }),
      { initialProps: { estimate: initial } },
    );
    expect(result.current.form.items[0]?.unitPrice).toBe(1100);
    rerender({ estimate: refetched });
    expect(result.current.form.items[0]?.unitPrice).toBe(1100);
    act(() => {
      result.current.handleChange("title", "タイトルだけ変更");
    });
    await submitEstimateAction(result.current.formAction);
    const data = mockUpdateMutateAsync.mock.calls[0]![0].data;
    expect(data.title).toBe("タイトルだけ変更");
    expect(data).not.toHaveProperty("items");
    expect(data).not.toHaveProperty("subtotal");
    expect(data).not.toHaveProperty("tax_total");
    expect(data).not.toHaveProperty("total_amount");
  });

  it("update payload: 無変更のレガシー不正行はヘッダー保存をブロックしない", async () => {
    const legacyInvalid: Estimate = {
      ...estimateWithItems,
      items: [{ ...estimateWithItems.items[0]!, quantity: 0.15 }],
    };
    const { result } = renderHook(() =>
      useEstimateForm({
        mode: "edit",
        estimate: legacyInvalid,
        permissions: ALLOW_ALL_PERMISSIONS,
      }),
    );
    act(() => {
      result.current.handleChange("title", "ヘッダーだけ変更");
    });
    await submitEstimateAction(result.current.formAction);
    expect(mockUpdateMutateAsync).toHaveBeenCalled();
    const data = mockUpdateMutateAsync.mock.calls[0]![0].data;
    expect(data).not.toHaveProperty("items");
  });

  it("update payload: 行を編集した場合は明示的な税フィールドと導出合計を送る", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({
        mode: "edit",
        estimate: estimateWithItems,
        permissions: ALLOW_ALL_PERMISSIONS,
      }),
    );
    act(() => {
      result.current.handleChange("items", [{ ...result.current.form.items[0]!, unitPrice: 1200 }]);
    });
    await submitEstimateAction(result.current.formAction);
    expect(mockUpdateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "1",
        data: expect.objectContaining({
          subtotal: 1100,
          tax_total: 110,
          total_amount: 1210,
          items: [
            expect.objectContaining({
              name: "薬剤A",
              unit_price: 1200,
              tax_type: "excluded",
              tax_rate: 0.1,
              discount_amount: 100,
              is_insurance_applicable: true,
            }),
          ],
        }),
      }),
    );
  });

  it("update payload: 最終行を削除 → items: [] を送る", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({
        mode: "edit",
        estimate: estimateWithItems,
        permissions: ALLOW_ALL_PERMISSIONS,
      }),
    );
    act(() => {
      result.current.handleChange("items", []);
    });
    await submitEstimateAction(result.current.formAction);
    const data = mockUpdateMutateAsync.mock.calls[0]![0].data;
    expect(data.items).toEqual([]);
  });

  it("0行の既存見積を明細無変更で保存 → items/合計を送らず表示は保存済みの金額", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({
        mode: "edit",
        estimate: mockEstimate,
        permissions: ALLOW_ALL_PERMISSIONS,
      }),
    );
    expect(result.current.totals).toEqual({
      subtotal: 1000,
      taxTotal: 100,
      totalAmount: 1100,
    });
    act(() => {
      result.current.handleChange("title", "ヘッダーのみ変更");
    });
    await submitEstimateAction(result.current.formAction);
    const data = mockUpdateMutateAsync.mock.calls[0]![0].data;
    expect(data).not.toHaveProperty("items");
    expect(data).not.toHaveProperty("subtotal");
    expect(data).not.toHaveProperty("tax_total");
    expect(data).not.toHaveProperty("total_amount");
  });

  it("0行ドラフトは有効（items: [] で create を呼ぶ）", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "空明細ドラフト");
    });
    await submitEstimateAction(result.current.formAction);
    expect(mockCreateMutateAsync).toHaveBeenCalledWith(expect.objectContaining({ items: [] }));
  });

  it("品目名が空の行 → mutation せず fieldErrors.items", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "不完全明細");
      result.current.handleChange("items", [makeItem({ name: " " })]);
    });
    await submitEstimateAction(result.current.formAction);
    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.formState.fieldErrors?.items).toBeTruthy();
  });

  it("負の単価の行 → mutation しない", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "負単価");
      result.current.handleChange("items", [makeItem({ unitPrice: -100 })]);
    });
    await submitEstimateAction(result.current.formAction);
    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.formState.fieldErrors?.items).toBeTruthy();
  });

  it("数量 0 の行 → mutation しない", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "数量0");
      result.current.handleChange("items", [makeItem({ quantity: 0 })]);
    });
    await submitEstimateAction(result.current.formAction);
    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.formState.fieldErrors?.items).toBeTruthy();
  });

  it("値引額 > 単価×数量 の行 → mutation しない", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "値引過大");
      result.current.handleChange("items", [
        makeItem({ unitPrice: 100, quantity: 1, discountAmount: 200 }),
      ]);
    });
    await submitEstimateAction(result.current.formAction);
    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.formState.fieldErrors?.items).toBeTruthy();
  });

  it.each([
    ["NaN の単価", { unitPrice: Number.NaN }],
    ["Infinity の単価", { unitPrice: Number.POSITIVE_INFINITY }],
    ["小数の単価", { unitPrice: 100.5 }],
    ["NaN の数量", { quantity: Number.NaN }],
    ["Infinity の数量", { quantity: Number.POSITIVE_INFINITY }],
    ["0.1 刻みでない数量", { quantity: 0.15 }],
    ["上限超過の数量", { quantity: 1000000000 }],
    ["NaN の値引額", { discountAmount: Number.NaN }],
    ["Infinity の値引額", { discountAmount: Number.POSITIVE_INFINITY }],
    ["小数の値引額", { discountAmount: 50.5 }],
    ["100 超の割引率", { discountRate: 101 }],
    ["NaN の割引率", { discountRate: Number.NaN }],
    ["不正な課税区分", { taxType: "bogus" as EstimateItemDraft["taxType"] }],
    ["小数第3位の税率", { taxRate: 0.075 }],
    ["1 超の税率", { taxRate: 1.1 }],
    ["NaN の税率", { taxRate: Number.NaN }],
    ["Infinity の税率", { taxRate: Number.POSITIVE_INFINITY }],
  ])("不正行 %s → mutation せず fieldErrors.items", async (_label, overrides) => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "不正行");
      result.current.handleChange("items", [makeItem(overrides)]);
    });
    await submitEstimateAction(result.current.formAction);
    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.formState.fieldErrors?.items).toBeTruthy();
  });

  it.each([
    ["単価が安全整数上限超過", { unitPrice: Number.MAX_SAFE_INTEGER + 1 }],
    ["値引額が安全整数上限超過", { discountAmount: Number.MAX_SAFE_INTEGER + 1 }],
    ["単価×数量が安全整数上限超過", { unitPrice: Number.MAX_SAFE_INTEGER, quantity: 2 }],
  ])("金額入力 %s → mutation せず fieldErrors.items", async (_label, overrides) => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "上限境界");
      result.current.handleChange("items", [makeItem(overrides)]);
    });

    await submitEstimateAction(result.current.formAction);

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.formState.fieldErrors?.items).toBeTruthy();
  });

  it("外税の行合計が上限を超える → mutation せず fieldErrors.items", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "外税行合計上限");
      result.current.handleChange("items", [
        makeItem({
          unitPrice: Number.MAX_SAFE_INTEGER,
          quantity: 1,
          taxType: "excluded",
          taxRate: 0.1,
        }),
      ]);
    });

    await submitEstimateAction(result.current.formAction);

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.formState.fieldErrors?.items).toBeTruthy();
  });

  it("非課税行の合計が上限を超える → mutation せず fieldErrors.items", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "複数行合計上限");
      result.current.handleChange("items", [
        makeItem({ unitPrice: Number.MAX_SAFE_INTEGER, taxType: "exempt" }),
        makeItem({ localId: "local-2", unitPrice: 1, taxType: "exempt" }),
      ]);
    });

    await submitEstimateAction(result.current.formAction);

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(result.current.formState.fieldErrors?.items).toBeTruthy();
  });

  it("安全整数上限の非課税行を正確な合計で送信する", async () => {
    const { result } = renderHook(() =>
      useEstimateForm({ mode: "create", permissions: ALLOW_ALL_PERMISSIONS }),
    );
    act(() => {
      result.current.handleChange("title", "安全整数境界");
      result.current.handleChange("items", [
        makeItem({
          unitPrice: Number.MAX_SAFE_INTEGER,
          quantity: 1,
          taxType: "exempt",
          discountAmount: 0,
        }),
      ]);
    });

    await submitEstimateAction(result.current.formAction);

    expect(mockCreateMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        subtotal: Number.MAX_SAFE_INTEGER,
        tax_total: 0,
        total_amount: Number.MAX_SAFE_INTEGER,
        items: [expect.objectContaining({ unit_price: Number.MAX_SAFE_INTEGER })],
      }),
    );
  });
});
