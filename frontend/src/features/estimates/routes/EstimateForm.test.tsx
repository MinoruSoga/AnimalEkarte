import { describe, it, expect, vi, beforeEach } from "vitest";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter, Route, Routes } from "react-router";
import type { Estimate } from "../types";
import { CREATE_STATUS_OPTIONS, EDIT_STATUS_OPTIONS } from "../constants/estimate-status-options";
import { EstimateForm } from "./EstimateForm";

const {
  mockEstimate,
  mockGetState,
  mockNavigate,
  mockHandleChange,
  mockToast,
  mockFormState,
  mockPetQuery,
  mockAmounts,
  mockPermissions,
} = vi.hoisted(() => ({
  mockEstimate: { current: null as Estimate | null },
  mockGetState: {
    current: {
      isLoading: false,
      isError: false,
      error: null as unknown,
      refetch: vi.fn(),
    },
  },
  mockPetQuery: {
    current: null as {
      data: { status?: string; deceasedAt?: string | null } | null;
      isPending: boolean;
      isSuccess: boolean;
      isError: boolean;
    } | null,
  },
  mockNavigate: vi.fn(),
  mockHandleChange: vi.fn(),
  mockToast: { info: vi.fn(), success: vi.fn(), error: vi.fn() },
  mockFormState: {
    current: {
      success: false,
      timestamp: 0,
      fieldErrors: undefined as Record<string, string> | undefined,
    },
  },
  mockAmounts: {
    current: { insuranceAmount: 0, discountAmount: 0 },
  },
  mockPermissions: {
    current: {
      canCreateEstimates: true,
      canCreateDiscount: true,
      canEditDiscount: true,
      withItem: false,
    },
  },
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: (resource: string) => ({
    canView: true,
    canCreate:
      resource === "discount"
        ? mockPermissions.current.canCreateDiscount
        : mockPermissions.current.canCreateEstimates,
    canEdit: resource === "discount" ? mockPermissions.current.canEditDiscount : true,
    canDelete: true,
  }),
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: {
      clinics: [{ clinicId: "1", clinicName: "テスト動物病院", isMain: true }],
      clinic: { name: "テスト動物病院" },
    },
    currentClinicId: "1",
    hasPermission: () => true,
  }),
}));

vi.mock("react-router", async () => {
  const actual = await vi.importActual<typeof import("react-router")>("react-router");
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock("sonner", () => ({ toast: mockToast }));

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: (id: string) =>
    mockPetQuery.current ?? { data: null, isPending: !id, isSuccess: false, isError: false },
}));

vi.mock("../api/get-estimate", () => ({
  useGetEstimate: () => ({
    data: mockEstimate.current,
    isLoading: mockGetState.current.isLoading,
    isError: mockGetState.current.isError,
    error: mockGetState.current.error,
    refetch: mockGetState.current.refetch,
  }),
}));

vi.mock("../hooks/use-estimate-form", () => ({
  useEstimateForm: () => ({
    form: {
      title: "テスト見積書",
      status: mockEstimate.current?.status ?? "draft",
      ownerId: "",
      medicalRecordId: "",
      insuranceAmount: mockAmounts.current.insuranceAmount,
      discountAmount: mockAmounts.current.discountAmount,
      validUntil: "",
      comment: "",
      notes: "",
      items: mockPermissions.current.withItem
        ? [
            {
              id: "1",
              localId: "row-1",
              name: "診察",
              category: "other",
              unitPrice: 1000,
              quantity: 1,
              taxType: "excluded",
              taxRate: 0.1,
              discountRate: 0,
              discountAmount: 0,
              isInsuranceApplicable: false,
              sortOrder: 0,
            },
          ]
        : [],
    },
    totals: { subtotal: 0, taxTotal: 0, totalAmount: 0 },
    createEmptyItem: vi.fn(() => ({
      id: "",
      localId: "mock-new",
      name: "",
      category: "other",
      unitPrice: 0,
      quantity: 1,
      taxType: "included",
      taxRate: 0.1,
      discountRate: 0,
      discountAmount: 0,
      isInsuranceApplicable: false,
      sortOrder: 0,
    })),
    handleChange: mockHandleChange,
    formAction: vi.fn(),
    formState: mockFormState.current,
    handleCancel: vi.fn(),
    isPending: false,
  }),
}));

vi.mock("@/hooks/use-unsaved-changes", () => ({
  useUnsavedChanges: () => ({
    isDirty: false,
    markDirty: vi.fn(),
    markClean: vi.fn(),
  }),
}));

function makeEstimate(status: Estimate["status"]): Estimate {
  return {
    id: "1",
    clinicId: "1",
    estimateNo: "EST-001",
    title: "テスト見積書",
    status,
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
}

function renderEditForm() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={["/estimates/1/edit"]}>
        <Routes>
          <Route path="/estimates/:id/edit" element={<EstimateForm />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

function renderCreateForm() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={["/estimates/new"]}>
        <Routes>
          <Route path="/estimates/new" element={<EstimateForm />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

function renderCreateFormWithPet(entry = "/estimates/new?petId=p1") {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[entry]}>
        <Routes>
          <Route path="/estimates/new" element={<EstimateForm />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("EstimateForm status options re-export", () => {
  it("Create 用選択肢は draft / sent のみ（utils 単一ソース）", () => {
    expect(CREATE_STATUS_OPTIONS.map((o) => o.value)).toEqual(["draft", "sent"]);
  });

  it("Edit 用選択肢は draft / sent / approved / rejected の 4 値（utils 単一ソース）", () => {
    expect(EDIT_STATUS_OPTIONS.map((o) => o.value)).toEqual([
      "draft",
      "sent",
      "approved",
      "rejected",
    ]);
  });
});

describe("EstimateForm statusError aria-describedby", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockEstimate.current = null;
    mockFormState.current = { success: false, timestamp: 0, fieldErrors: undefined };
    mockAmounts.current = { insuranceAmount: 0, discountAmount: 0 };
    mockPermissions.current = {
      canCreateEstimates: true,
      canCreateDiscount: true,
      canEditDiscount: true,
      withItem: false,
    };
  });

  it.each([
    [false, true, false, true],
    [false, false, true, false],
    [true, false, true, true],
    [true, true, false, false],
  ])(
    "割引のヘッダーと明細は保存モードの権限に一致する edit=%s create=%s editPermission=%s",
    (edit, createPermission, editPermission, allowed) => {
      mockPermissions.current.canCreateDiscount = createPermission;
      mockPermissions.current.canEditDiscount = editPermission;
      mockPermissions.current.withItem = true;
      if (edit) {
        mockEstimate.current = makeEstimate("draft");
        renderEditForm();
      } else {
        renderCreateForm();
      }
      const header = screen.getByRole("spinbutton", { name: "割引額" });
      expect(header).toHaveProperty("disabled", !allowed);
      expect(screen.queryByRole("spinbutton", { name: "値引（円）" }) !== null).toBe(allowed);
    },
  );

  it("statusError があるとき SelectTrigger に aria-describedby=status-error を接続する", () => {
    mockFormState.current = {
      success: false,
      timestamp: Date.now(),
      fieldErrors: { status: "作成時は下書きまたは送付済みのみ選択できます" },
    };
    renderCreateForm();

    const trigger = screen.getByRole("combobox");
    expect(trigger).toHaveAttribute("aria-describedby", "status-error");
    // EMR-319: ステータス必須エラーで SelectTrigger が aria-invalid を持つ
    expect(trigger).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByRole("alert")).toHaveAttribute("id", "status-error");
    expect(screen.getByRole("alert")).toHaveTextContent(
      "作成時は下書きまたは送付済みのみ選択できます",
    );
  });

  it("statusError がないとき aria-describedby を付けない", () => {
    renderCreateForm();

    const trigger = screen.getByRole("combobox");
    expect(trigger).not.toHaveAttribute("aria-describedby");
    expect(trigger).not.toHaveAttribute("aria-invalid");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  // EMR-319: タイトル必須エラーで Input が aria-invalid+describedby を持ちエラー文が出る。
  it("titleError があるときタイトル Input に aria-invalid と title-error 参照を接続する", () => {
    mockFormState.current = {
      success: false,
      timestamp: Date.now(),
      fieldErrors: { title: "タイトルを入力してください" },
    };
    renderCreateForm();

    const titleInput = screen.getByRole("textbox", { name: /タイトル/ });
    expect(titleInput).toHaveAttribute("aria-invalid", "true");
    expect(titleInput).toHaveAttribute("aria-describedby", "title-error");
    expect(document.getElementById("title-error")).toHaveTextContent("タイトルを入力してください");
  });
});

describe("EstimateForm locked edit redirect", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockFormState.current = { success: false, timestamp: 0, fieldErrors: undefined };
  });

  it("approved の edit 直アクセス → detail へ redirect + toast", async () => {
    mockEstimate.current = makeEstimate("approved");
    renderEditForm();

    await waitFor(() => {
      expect(mockToast.info).toHaveBeenCalledWith("承認済みまたは却下済みの見積書は編集できません");
      expect(mockNavigate).toHaveBeenCalledWith("/estimates/1", { replace: true });
    });
  });

  it("rejected の edit 直アクセス → detail へ redirect + toast", async () => {
    mockEstimate.current = makeEstimate("rejected");
    renderEditForm();

    await waitFor(() => {
      expect(mockToast.info).toHaveBeenCalledWith("承認済みまたは却下済みの見積書は編集できません");
      expect(mockNavigate).toHaveBeenCalledWith("/estimates/1", { replace: true });
    });
  });

  it("draft の edit では redirect しない", async () => {
    mockEstimate.current = makeEstimate("draft");
    renderEditForm();

    await waitFor(() => {
      expect(mockNavigate).not.toHaveBeenCalled();
      expect(mockToast.info).not.toHaveBeenCalled();
    });
  });
});

describe("EstimateForm mobile-first layout", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockEstimate.current = null;
    mockFormState.current = { success: false, timestamp: 0, fieldErrors: undefined };
  });

  it("固定幅 fields と金額2列を mobile では全幅・単一列にする", () => {
    renderCreateForm();

    const statusTrigger = screen.getByRole("combobox");
    const validUntilInput = screen.getByLabelText("有効期限");
    const amountGrid = screen.getByText("小計（明細合計）").closest('[class*="grid-cols"]');

    expect(statusTrigger).toHaveClass("w-full", "sm:w-[200px]");
    expect(validUntilInput.parentElement).toHaveClass("w-full", "sm:w-[220px]");
    expect(amountGrid).toHaveClass("grid-cols-1", "sm:grid-cols-2");
  });

  it("ヘッダー金額サマリ: 小計/税/合計は読み取り専用、保険適用額と割引額のみ入力可 (PR665)", () => {
    renderCreateForm();

    expect(screen.getByText("小計（明細合計）")).toBeInTheDocument();
    expect(screen.getByText("消費税（内税分を含む）")).toBeInTheDocument();
    expect(screen.getByText("合計金額")).toBeInTheDocument();
    // 導出金額は spinbutton を持たない
    for (const label of ["小計（明細合計）", "消費税（内税分を含む）", "合計金額"]) {
      expect(screen.queryByRole("spinbutton", { name: label })).not.toBeInTheDocument();
    }
    // 保険適用額・割引額のみ独立入力
    expect(screen.getByRole("spinbutton", { name: "保険適用額" })).toBeInTheDocument();
    expect(screen.getByRole("spinbutton", { name: "割引額" })).toBeInTheDocument();
    expect(screen.getAllByRole("spinbutton")).toHaveLength(2);
  });

  it("PR665: 保険適用額・割引額の変更とクリアが form state へ流れる", () => {
    renderCreateForm();

    const insuranceInput = screen.getByRole("spinbutton", { name: "保険適用額" });
    fireEvent.change(insuranceInput, { target: { value: "500" } });
    expect(mockHandleChange).toHaveBeenCalledWith("insuranceAmount", 500);
    fireEvent.change(insuranceInput, { target: { value: "" } });
    expect(mockHandleChange).toHaveBeenCalledWith("insuranceAmount", 0);

    const discountInput = screen.getByRole("spinbutton", { name: "割引額" });
    fireEvent.change(discountInput, { target: { value: "200" } });
    expect(mockHandleChange).toHaveBeenCalledWith("discountAmount", 200);
    fireEvent.change(discountInput, { target: { value: "" } });
    expect(mockHandleChange).toHaveBeenCalledWith("discountAmount", 0);
  });

  it("PR665: discount:create 権限なしでは割引額は disabled で既存値を保持し、保険適用額は入力可", () => {
    mockPermissions.current.canCreateDiscount = false;
    mockAmounts.current.discountAmount = 300;
    renderCreateForm();

    expect(screen.getByRole("spinbutton", { name: "割引額" })).toBeDisabled();
    expect((screen.getByRole("spinbutton", { name: "割引額" }) as HTMLInputElement).value).toBe(
      "300",
    );
    expect(screen.getByRole("spinbutton", { name: "保険適用額" })).not.toBeDisabled();
  });

  it("明細行エディタを表示する（行追加アクションあり）", () => {
    renderCreateForm();
    expect(screen.getByRole("button", { name: "明細行を追加" })).toBeInTheDocument();
  });

  it("明細行を追加クリック → createEmptyItem の行が handleChange('items') で渡る", () => {
    renderCreateForm();
    fireEvent.click(screen.getByRole("button", { name: "明細行を追加" }));
    expect(mockHandleChange).toHaveBeenCalledWith("items", [
      expect.objectContaining({
        localId: "mock-new",
        name: "",
        category: "other",
        quantity: 1,
        taxType: "included",
        taxRate: 0.1,
      }),
    ]);
  });
});

describe("EstimateForm BUG-019 not-found / network gate", () => {
  function axiosError(status: number | undefined) {
    const config = { headers: new AxiosHeaders() } as InternalAxiosRequestConfig;
    if (status === undefined) {
      return new AxiosError("Network Error", AxiosError.ERR_NETWORK, config, undefined, undefined);
    }
    return new AxiosError("request failed", AxiosError.ERR_BAD_RESPONSE, config, undefined, {
      config,
      data: { error: "not found" },
      headers: new AxiosHeaders(),
      status,
      statusText: "Error",
    });
  }

  beforeEach(() => {
    vi.clearAllMocks();
    mockEstimate.current = null;
    mockGetState.current = {
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
    mockFormState.current = { success: false, timestamp: 0, fieldErrors: undefined };
  });

  it("404 相当: 見積書が見つかりません を表示し、更新/作成ボタンを出さない", () => {
    mockGetState.current = {
      isLoading: false,
      isError: true,
      error: axiosError(404),
      refetch: vi.fn(),
    };
    renderEditForm();
    expect(screen.getByText("見積書が見つかりません")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "更新" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "作成" })).not.toBeInTheDocument();
  });

  it("403 相当: 404 と同一の非開示メッセージ（見積書が見つかりません）", () => {
    mockGetState.current = {
      isLoading: false,
      isError: true,
      error: axiosError(403),
      refetch: vi.fn(),
    };
    renderEditForm();
    expect(screen.getByText("見積書が見つかりません")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "更新" })).not.toBeInTheDocument();
  });

  it("network error: 取得失敗 + 再試行、blank form ではない", () => {
    const refetch = vi.fn();
    mockGetState.current = {
      isLoading: false,
      isError: true,
      error: axiosError(undefined),
      refetch,
    };
    renderEditForm();
    expect(screen.getByText("見積書の取得に失敗しました")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "再試行" })).toBeInTheDocument();
    screen.getByRole("button", { name: "再試行" }).click();
    expect(refetch).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole("button", { name: "更新" })).not.toBeInTheDocument();
  });

  it("create route /estimates/new は従来どおり新規フォームを開く", () => {
    renderCreateForm();
    expect(screen.getByText("新規見積書作成")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "作成" })).toBeInTheDocument();
  });

  it("正常 edit: draft 見積で編集フォームを開く", async () => {
    mockEstimate.current = makeEstimate("draft");
    mockGetState.current = {
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
    renderEditForm();
    await waitFor(() => {
      expect(screen.getByText("見積書編集")).toBeInTheDocument();
    });
    expect(screen.queryByText("見積書が見つかりません")).not.toBeInTheDocument();
  });
});

// PR#614 指摘: render gate が status のみを見て deceasedAt を無視していたため、
// status=生存 + deceasedAt 非 null の個体はメッセージは出るがフォームが有効のままだった。
describe("EstimateForm deceased pet gate (deceasedAt)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockPetQuery.current = null;
    mockFormState.current = { success: false, timestamp: 0, fieldErrors: undefined };
  });

  it("status=生存 でも deceasedAt があればメッセージ表示かつ fieldset を disabled にする", () => {
    mockPetQuery.current = {
      data: { status: "生存", deceasedAt: "2026-01-10T00:00:00Z" },
      isPending: false,
      isSuccess: true,
      isError: false,
    };
    const { container } = renderCreateFormWithPet();

    expect(screen.getByText("死亡したペットの見積書は作成できません")).toBeInTheDocument();
    expect(container.querySelector("fieldset")).toHaveProperty("disabled", true);
    expect(screen.getByPlaceholderText("見積書タイトルを入力")).toBeDisabled();
  });

  it("status=生存 かつ deceasedAt=null なら通常どおりフォームが有効", () => {
    mockPetQuery.current = {
      data: { status: "生存", deceasedAt: null },
      isPending: false,
      isSuccess: true,
      isError: false,
    };
    const { container } = renderCreateFormWithPet();

    expect(screen.queryByText("死亡したペットの見積書は作成できません")).not.toBeInTheDocument();
    expect(container.querySelector("fieldset")).toHaveProperty("disabled", false);
  });
});
