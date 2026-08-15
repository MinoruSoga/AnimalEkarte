import { describe, it, expect, vi, beforeEach } from "vitest";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router";
import type { Estimate } from "../types";
import {
  CREATE_STATUS_OPTIONS,
  EDIT_STATUS_OPTIONS,
} from "../constants/estimate-status-options";
import { EstimateForm } from "./EstimateForm";

const { mockEstimate, mockGetState, mockNavigate, mockToast, mockFormState } = vi.hoisted(() => ({
  mockEstimate: { current: null as Estimate | null },
  mockGetState: {
    current: {
      isLoading: false,
      isError: false,
      error: null as unknown,
      refetch: vi.fn(),
    },
  },
  mockNavigate: vi.fn(),
  mockToast: { info: vi.fn(), success: vi.fn(), error: vi.fn() },
  mockFormState: {
    current: {
      success: false,
      timestamp: 0,
      fieldErrors: undefined as Record<string, string> | undefined,
    },
  },
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
      subtotal: 0,
      taxTotal: 0,
      totalAmount: 0,
      insuranceAmount: 0,
      discountAmount: 0,
      validUntil: "",
      comment: "",
      notes: "",
    },
    handleChange: vi.fn(),
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
  return render(
    <MemoryRouter initialEntries={["/estimates/1/edit"]}>
      <Routes>
        <Route path="/estimates/:id/edit" element={<EstimateForm />} />
      </Routes>
    </MemoryRouter>,
  );
}

function renderCreateForm() {
  return render(
    <MemoryRouter initialEntries={["/estimates/new"]}>
      <Routes>
        <Route path="/estimates/new" element={<EstimateForm />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("EstimateForm status options re-export", () => {
  it("Create 用選択肢は draft / sent のみ（utils 単一ソース）", () => {
    expect(CREATE_STATUS_OPTIONS.map((o) => o.value)).toEqual([
      "draft",
      "sent",
    ]);
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
  });

  it("statusError があるとき SelectTrigger に aria-describedby=status-error を接続する", () => {
    mockFormState.current = {
      success: false,
      timestamp: Date.now(),
      fieldErrors: { status: "作成時は下書きまたは送付済みのみ選択できます" },
    };
    renderCreateForm();

    const trigger = screen.getByRole("combobox");
    expect(trigger).toHaveAttribute("aria-describedby", "status-error");
    expect(screen.getByRole("alert")).toHaveAttribute("id", "status-error");
    expect(screen.getByRole("alert")).toHaveTextContent(
      "作成時は下書きまたは送付済みのみ選択できます",
    );
  });

  it("statusError がないとき aria-describedby を付けない", () => {
    renderCreateForm();

    const trigger = screen.getByRole("combobox");
    expect(trigger).not.toHaveAttribute("aria-describedby");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
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
      expect(mockToast.info).toHaveBeenCalledWith(
        "承認済みまたは却下済みの見積書は編集できません",
      );
      expect(mockNavigate).toHaveBeenCalledWith("/estimates/1", { replace: true });
    });
  });

  it("rejected の edit 直アクセス → detail へ redirect + toast", async () => {
    mockEstimate.current = makeEstimate("rejected");
    renderEditForm();

    await waitFor(() => {
      expect(mockToast.info).toHaveBeenCalledWith(
        "承認済みまたは却下済みの見積書は編集できません",
      );
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
    const amountGrid = screen.getByLabelText("小計（税抜）").closest('[class*="grid-cols"]');

    expect(statusTrigger).toHaveClass("w-full", "sm:w-[200px]");
    expect(validUntilInput.parentElement).toHaveClass("w-full", "sm:w-[220px]");
    expect(amountGrid).toHaveClass("grid-cols-1", "sm:grid-cols-2");
  });
});


describe("EstimateForm BUG-019 not-found / network gate", () => {
  function axiosError(status: number | undefined) {
    const config = { headers: new AxiosHeaders() } as InternalAxiosRequestConfig;
    if (status === undefined) {
      return new AxiosError("Network Error", AxiosError.ERR_NETWORK, config, undefined, undefined);
    }
    return new AxiosError(
      "request failed",
      AxiosError.ERR_BAD_RESPONSE,
      config,
      undefined,
      {
        config,
        data: { error: "not found" },
        headers: new AxiosHeaders(),
        status,
        statusText: "Error",
      },
    );
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
