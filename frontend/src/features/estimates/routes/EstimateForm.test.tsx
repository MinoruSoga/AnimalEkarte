import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router";
import type { Estimate } from "../types";
import {
  CREATE_STATUS_OPTIONS,
  EDIT_STATUS_OPTIONS,
  EstimateForm,
} from "./EstimateForm";

const { mockEstimate, mockNavigate, mockToast } = vi.hoisted(() => ({
  mockEstimate: { current: null as Estimate | null },
  mockNavigate: vi.fn(),
  mockToast: { info: vi.fn(), success: vi.fn(), error: vi.fn() },
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
    isLoading: false,
    isError: false,
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
    formState: { success: false, timestamp: 0 },
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

describe("EstimateForm status options", () => {
  it("Create 用選択肢は draft / sent のみ", () => {
    expect(CREATE_STATUS_OPTIONS.map((o) => o.value)).toEqual([
      "draft",
      "sent",
    ]);
  });

  it("Edit 用選択肢は draft / sent / approved / rejected の 4 値", () => {
    expect(EDIT_STATUS_OPTIONS.map((o) => o.value)).toEqual([
      "draft",
      "sent",
      "approved",
      "rejected",
    ]);
  });

  it("Create 用選択肢に approved / rejected は含まれない", () => {
    const values = CREATE_STATUS_OPTIONS.map((o) => o.value);
    expect(values).not.toContain("approved");
    expect(values).not.toContain("rejected");
  });
});

describe("EstimateForm locked edit redirect", () => {
  beforeEach(() => {
    vi.clearAllMocks();
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
