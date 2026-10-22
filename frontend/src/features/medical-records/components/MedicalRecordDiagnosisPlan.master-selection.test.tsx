import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { act, render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { toast } from "sonner";

import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { queryKeys } from "@/lib/query-keys";
import {
  DOSE_PARAMS_LOOKUP_FAILED_MESSAGE,
  medicineDoseParamsQueryKey,
} from "../api/medicine-dose-lookup";
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
    onSelect: (item: TreatmentMasterItem) => void;
  }) => {
    mockDialogProps.current = props as unknown as Record<string, unknown>;
    if (!props.open) return null;
    return (
      <>
        <button
          type="button"
          onClick={() =>
            props.onSelect({
              id: "7",
              name: "初診",
              unitPrice: 3307,
              category: "診察",
              insuranceContract: "master",
              isNonInsurance: false,
            })
          }
        >
          診察Aを選択
        </button>
        <button
          type="button"
          onClick={() =>
            props.onSelect({
              id: "8",
              name: "処置X",
              unitPrice: 3306,
              category: "処置",
              insuranceContract: "master",
              isNonInsurance: false,
            })
          }
        >
          処置Xを選択
        </button>
        <button
          type="button"
          onClick={() =>
            props.onSelect({
              id: "9",
              name: "自費診察",
              unitPrice: 5000,
              category: "診察",
              insuranceContract: "master",
              isNonInsurance: true,
            })
          }
        >
          保険対象外診察を選択
        </button>
        <button
          type="button"
          onClick={() =>
            props.onSelect({
              id: "42",
              name: "抗生剤A",
              unitPrice: 3303,
              category: "薬剤",
              insuranceContract: "master",
              isNonInsurance: false,
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
  it.each(["予防", "検査"])("未採用 %s は caller の保険対象=true を維持する", async (category) => {
    server.use(
      http.get("*/v1/medical-records/:id/clinical-plan", () => HttpResponse.json({})),
      http.get("*/v1/medical-records/:id/billing-confirmation", () =>
        HttpResponse.json({ id: "1", medical_record_id: "10", status: "pending" }),
      ),
      http.get("*/v1/medical-records/:id/vitals", () => HttpResponse.json([])),
    );
    const user = userEvent.setup();
    render(
      <MedicalRecordDiagnosisPlan {...BASE_PROPS} recordClinicId="2" />,
      { wrapper: createTestWrapper() },
    );
    await user.click(await screen.findByRole("button", { name: /検索|マスタ|追加/ }));
    const select = mockDialogProps.current.onSelect as (item: TreatmentMasterItem) => void;
    await act(async () => {
      select({ id: "71", name: "legacy master", unitPrice: 1000, category });
    });
    await waitFor(() =>
      expect(mockCreateMutate).toHaveBeenCalledWith(
        expect.objectContaining({ item_type: "other", memo: category, is_insurance: true }),
      ),
    );
  });

  it("EMR-454: master が明示的に保険対象外の品目は is_insurance=false で作成する", async () => {
    server.use(
      http.get("*/v1/medical-records/10/clinical-plan", () => HttpResponse.json({})),
      http.get("*/v1/medical-records/10/vitals", () => HttpResponse.json([])),
    );
    const user = userEvent.setup();
    render(<MedicalRecordDiagnosisPlan {...BASE_PROPS} recordClinicId="2" />, {
      wrapper: createTestWrapper(),
    });

    await user.click(await screen.findByRole("button", { name: /検索|マスタ|追加/ }));
    await user.click(screen.getByRole("button", { name: "保険対象外診察を選択" }));

    await waitFor(() =>
      expect(mockCreateMutate).toHaveBeenCalledWith(
        expect.objectContaining({ consultation_id: 9, is_insurance: false }),
      ),
    );
  });

  it.each([
    { label: "診察Aを選択", ref: { consultation_id: 7 }, memo: "診察" },
    { label: "処置Xを選択", ref: { procedure_id: 8 }, memo: "処置" },
    { label: "薬剤Aを選択", ref: { medicine_id: 42 }, memo: "薬剤" },
  ])(
    "$label → 実ビルダーで数値参照・memo=カテゴリ・is_insurance=true・nextOrder=5 で作成する",
    async ({ label, ref, memo }) => {
      server.use(
        http.get("*/v1/medical-records/10/clinical-plan", () => HttpResponse.json({})),
        http.get("*/v1/masters/medicines", () =>
          HttpResponse.json([
            { id: 42, name: "抗生剤A", price: 3303, is_active: true, tax_rate: "0.1" },
          ]),
        ),
        http.get("*/v1/medical-records/10/vitals", () => HttpResponse.json([])),
      );
      const user = userEvent.setup();
      render(<MedicalRecordDiagnosisPlan {...BASE_PROPS} recordClinicId="2" />, {
        wrapper: createTestWrapper(),
      });

      await user.click(await screen.findByRole("button", { name: /検索|マスタ|追加/ }));
      await user.click(screen.getByRole("button", { name: label }));

      await waitFor(() =>
        expect(mockCreateMutate).toHaveBeenCalledWith(
          expect.objectContaining({
            ...ref,
            memo,
            is_insurance: true,
            quantity: 1,
            sort_order: 5,
          }),
        ),
      );
      expect(mockDialogProps.current.clinicId).toBe("2");
    },
  );

  it.each(["lock", "permission", "deceased", "record", "clinic", "species"] as const)(
    "PR700 deferred: %s during lookup prevents create",
    async (change) => {
      let releaseGet: (() => void) | undefined;
      let getSeen = false;
      server.use(
        http.get("*/v1/medical-records/10/clinical-plan", () => HttpResponse.json({})),
        http.get("*/v1/medical-records/:id/billing-confirmation", () =>
          HttpResponse.json({ id: "1", medical_record_id: "10", status: "pending" }),
        ),
        http.get("*/v1/masters/medicines", () => {
          getSeen = true;
          const medicine = { id: 42, name: "薬A", price: 3303, is_active: true, tax_rate: "0.1" };
          return new Promise<ReturnType<typeof HttpResponse.json>>((resolve) => {
            releaseGet = () => resolve(HttpResponse.json([medicine]));
          });
        }),
        http.get("*/v1/medical-records/:id/vitals", () => HttpResponse.json([])),
      );
      const { rerender } = render(
        <MedicalRecordDiagnosisPlan {...BASE_PROPS} recordClinicId="2" />,
        { wrapper: createTestWrapper() },
      );
      let select: ((item: unknown) => void) | undefined;
      await waitFor(() => {
        select = mockDialogProps.current.onSelect as ((item: unknown) => void) | undefined;
        expect(select).toBeDefined();
      });
      let resolveDone = false;
      const promise = Promise.resolve(
        select!({
          id: "42",
          name: "薬A",
          unitPrice: 3303,
          category: "薬剤",
          insuranceContract: "master",
          isNonInsurance: false,
          medicineId: "42",
        }),
      ).finally(() => {
        resolveDone = true;
      });
      await waitFor(() => expect(getSeen).toBe(true), { timeout: 3000 });
      if (change === "permission") mockPerm.current.canCreate = false;
      rerender(
        <MedicalRecordDiagnosisPlan
          {...BASE_PROPS}
          medicalRecordId={change === "record" ? "11" : "10"}
          recordClinicId={change === "clinic" ? "3" : "2"}
          petSpecies={change === "species" ? "猫" : undefined}
          isPetDeceased={change === "deceased"}
          isWriteBlocked={change === "lock"}
        />,
      );
      await act(async () => {
        releaseGet!();
        await promise;
      });
      expect(resolveDone).toBe(true);
      expect(mockCreateMutate).not.toHaveBeenCalled();
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
      await waitFor(() => expect(mockDialogProps.current.onSelect).toBeDefined());
      const select = mockDialogProps.current.onSelect as ((item: unknown) => void) | undefined;
      expect(select).toBeDefined();
      select?.({
        id: "7",
        name: "初診",
        unitPrice: 3307,
        category: "診察",
        insuranceContract: "master",
        isNonInsurance: false,
      });
      expect(mockCreateMutate).not.toHaveBeenCalled();
      if (guard === "deceased") expect(toast.error).toHaveBeenCalled();
    },
  );
});

describe("PR700 DiagnosisPlan real dose selection", () => {
  it.each(["recommended", "cap", "reason", "http", "missing-weight", "missing-params"] as const)(
    "%s uses scoped authority and the actual dose gate",
    async (mode) => {
      const headers: string[] = [];
      const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
      const globalMasters = [{ id: "42", isActive: false, price: 1 }];
      queryClient.setQueryData(queryKeys.masters.category("medicines"), globalMasters);
      server.use(
        http.get("*/v1/masters/medicines", ({ request }) => {
          headers.push(`medicine:${request.headers.get("X-Clinic-ID")}`);
          return HttpResponse.json([
            {
              id: 42,
              name: "薬A",
              price: 3303,
              is_active: true,
              tax_rate: "0.1",
              calculation_type: "per_weight",
              medicine_unit: "per_tablet",
              strength: "100",
            },
          ]);
        }),
        http.get("*/v1/medical-records/10/vitals", ({ request }) => {
          headers.push(`vitals:${request.headers.get("X-Clinic-ID")}`);
          return HttpResponse.json(
            mode === "missing-weight"
              ? []
              : [
                  {
                    id: "v1",
                    medical_record_id: "10",
                    recorded_at: "2026-10-18T12:00:00Z",
                    weight: 10,
                    weight_unit: "Kg",
                  },
                ],
          );
        }),
        http.get("*/v1/masters/medicines/42/dose-params", ({ request }) => {
          headers.push(`dose:${request.headers.get("X-Clinic-ID")}`);
          if (mode === "http") return HttpResponse.json({ error: "private" }, { status: 500 });
          return HttpResponse.json(
            mode === "missing-params"
              ? []
              : [
                  {
                    species: "dog",
                    dose_basis: "per_dose",
                    dose_per_kg: 20,
                    min_mg_per_kg: mode === "reason" ? 20 : 1,
                    max_mg_per_kg: 100,
                    absolute_max_dose: mode === "cap" ? 50 : mode === "reason" ? 150 : 1000,
                    rounding_step: 1,
                    rounding_mode: "up",
                  },
                ],
          );
        }),
      );
      render(
        <QueryClientProvider client={queryClient}>
          <MedicalRecordDiagnosisPlan {...BASE_PROPS} recordClinicId="2" petSpecies="犬" />
        </QueryClientProvider>,
      );
      const user = userEvent.setup();
      await user.click(await screen.findByRole("button", { name: /検索|マスタ|追加/ }));
      await user.click(screen.getByRole("button", { name: "薬剤Aを選択" }));
      if (mode === "cap" || mode === "reason" || mode === "http") {
        await waitFor(() => expect(toast.error).toHaveBeenCalled());
        expect(mockCreateMutate).not.toHaveBeenCalled();
        expect(toast.error).toHaveBeenCalledWith(
          mode === "http"
            ? DOSE_PARAMS_LOOKUP_FAILED_MESSAGE
            : expect.stringContaining(mode === "cap" ? "絶対上限" : "推奨値付近で追加"),
        );
      } else {
        await waitFor(() =>
          expect(mockCreateMutate).toHaveBeenCalledWith(
            expect.objectContaining({
              medicine_id: 42,
              quantity: mode === "recommended" ? 2 : 1,
              memo: "薬剤",
              is_insurance: true,
              sort_order: 5,
            }),
          ),
        );
      }
      expect(headers).toEqual(expect.arrayContaining(["medicine:2", "vitals:2"]));
      if (mode !== "missing-weight") expect(headers).toContain("dose:2");
      expect(queryClient.getQueryData(queryKeys.masters.category("medicines"))).toEqual(
        globalMasters,
      );
      expect(
        queryClient.getQueryData(queryKeys.masters.categoryForClinic("medicines", "2")),
      ).toBeDefined();
      if (mode === "recommended")
        expect(queryClient.getQueryData(medicineDoseParamsQueryKey("42", "2"))).toBeDefined();
    },
  );
});
