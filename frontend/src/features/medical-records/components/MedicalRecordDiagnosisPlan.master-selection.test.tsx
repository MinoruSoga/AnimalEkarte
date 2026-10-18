import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { toast } from "sonner";

import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { MedicalRecordDiagnosisPlan } from "./MedicalRecordDiagnosisPlan";

const { mockCreateMutate, mockDialogProps } = vi.hoisted(() => ({
  mockCreateMutate: vi.fn(),
  mockDialogProps: { current: {} as Record<string, unknown> },
}));

const { mockPerm } = vi.hoisted(() => ({
  mockPerm: { current: { canView: true, canCreate: true, canEdit: true, canDelete: true } },
}));
vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => mockPerm.current,
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ user: { clinic: {} } }),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock("../api/treatments", () => ({
  useGetTreatments: () => ({ data: [{ id: "1", sort_order: 4 }] }),
  useCreateTreatment: () => ({ mutate: mockCreateMutate, isPending: false }),
  useUpdateTreatment: () => ({ mutate: vi.fn() }),
  useDeleteTreatment: () => ({ mutate: vi.fn() }),
}));

vi.mock("../api/clinical-plan", () => ({
  useGetClinicalPlan: () => ({ data: undefined }),
}));

vi.mock("./ClinicalPlanSection/ClinicalPlanSection", () => ({
  ClinicalPlanSection: () => <div data-testid="clinical-plan" />,
}));

vi.mock("./DiagnosisHeader", () => ({
  DiagnosisHeader: () => <div data-testid="diagnosis-header" />,
}));

vi.mock("./TreatmentTable", () => ({
  TreatmentTable: ({ onOpenSearch }: { onOpenSearch?: () => void }) => (
    <button type="button" onClick={onOpenSearch}>
      マスタ検索を開く
    </button>
  ),
}));

vi.mock("./TreatmentDetailedSummary", () => ({
  TreatmentDetailedSummary: () => <div data-testid="summary" />,
}));

vi.mock("@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog", () => ({
  TreatmentSearchDialog: (props: {
    open?: boolean;
    clinicId?: string;
    onSelect: (item: {
      id: string;
      name: string;
      unitPrice: number;
      category: string;
      medicineId?: string;
    }) => void;
  }) => {
    mockDialogProps.current = props as unknown as Record<string, unknown>;
    if (!props.open) return null;
    return (
      <>
        <button
          type="button"
          onClick={() =>
            props.onSelect({ id: "7", name: "初診", unitPrice: 3307, category: "診察" })
          }
        >
          診察Aを選択
        </button>
        <button
          type="button"
          onClick={() =>
            props.onSelect({ id: "8", name: "処置X", unitPrice: 3306, category: "処置" })
          }
        >
          処置Xを選択
        </button>
        <button
          type="button"
          onClick={() =>
            props.onSelect({
              id: "42",
              name: "抗生剤A",
              unitPrice: 3303,
              category: "薬剤",
              medicineId: "42",
            })
          }
        >
          薬剤Aを選択
        </button>
      </>
    );
  },
}));

beforeEach(() => {
  mockCreateMutate.mockClear();
  mockDialogProps.current = {};
  mockPerm.current = { canView: true, canCreate: true, canEdit: true, canDelete: true };
  vi.mocked(toast.error).mockClear();
});

afterEach(() => {
  server.resetHandlers();
});

const BASE_PROPS = {
  isNewRecord: false,
  chiefComplaint: "",
  physicalExam: "",
  setPhysicalExam: vi.fn(),
  plan: "",
  setPlan: vi.fn(),
  assessment: "",
  setAssessment: vi.fn(),
  diagnosis1CategoryId: null,
  setDiagnosis1CategoryId: vi.fn(),
  diagnosis1NameId: null,
  setDiagnosis1NameId: vi.fn(),
  diagnosis2CategoryId: null,
  setDiagnosis2CategoryId: vi.fn(),
  diagnosis2NameId: null,
  setDiagnosis2NameId: vi.fn(),
  medicalRecordId: "10",
};

describe("MedicalRecordDiagnosisPlan PR670 マスタ選択", () => {
  it.each([
    { label: "診察Aを選択", ref: { consultation_id: 7 }, memo: "診察" },
    { label: "処置Xを選択", ref: { procedure_id: 8 }, memo: "処置" },
    { label: "薬剤Aを選択", ref: { medicine_id: 42 }, memo: "薬剤" },
  ])(
    "$label → 実ビルダーで数値参照・memo=カテゴリ・is_insurance=true・nextOrder=5 で作成する",
    async ({ label, ref, memo }) => {
      server.use(http.get("*/v1/medical-records/10/clinical-plan", () => HttpResponse.json({})));
      const user = userEvent.setup();
      render(<MedicalRecordDiagnosisPlan {...BASE_PROPS} recordClinicId="2" />, {
        wrapper: createTestWrapper(),
      });

      await user.click(await screen.findByRole("button", { name: /検索|マスタ|追加/ }));
      await user.click(screen.getByRole("button", { name: label }));

      expect(mockCreateMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          ...ref,
          memo,
          is_insurance: true,
          quantity: 1,
          sort_order: 5,
        }),
      );
      expect(mockDialogProps.current.clinicId).toBe("2");
    },
  );

  it.each(["deceased", "permission"] as const)(
    "ガード (%s) はキャプチャした選択コールバックでも create を拒否する",
    async (guard) => {
      server.use(http.get("*/v1/medical-records/10/clinical-plan", () => HttpResponse.json({})));
      if (guard === "permission") mockPerm.current.canCreate = false;
      render(
        <MedicalRecordDiagnosisPlan
          {...BASE_PROPS}
          recordClinicId="2"
          isPetDeceased={guard === "deceased"}
        />,
        { wrapper: createTestWrapper() },
      );
      const select = mockDialogProps.current.onSelect as ((item: unknown) => void) | undefined;
      expect(select).toBeDefined();
      select?.({ id: "7", name: "初診", unitPrice: 3307, category: "診察" });
      expect(mockCreateMutate).not.toHaveBeenCalled();
      if (guard === "deceased") expect(toast.error).toHaveBeenCalled();
    },
  );
});
