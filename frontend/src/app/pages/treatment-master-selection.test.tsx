import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { axios } from "@/lib/axios";
import { createTestWrapper } from "@/testing/TestUtils";
import {
  TreatmentSearchDialog,
  type TreatmentMasterItem,
} from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";
// Composition-level contract test: raw HTTP response -> real read hooks/transform -> Dialog -> create payload.
// eslint-disable-next-line no-restricted-imports -- Composition-only regression exercises the existing internal builder without expanding the public feature API.
import { buildMasterSelectionPayload } from "@/features/medical-records/lib/treatments-tab-model";
// eslint-disable-next-line no-restricted-imports -- The same composition test verifies master-edit and clinical-selection adoption stay aligned.
import { buildTreatmentTabConfigs } from "@/features/master/routes/treatment-plan-master-model";
// Medicine already has a non-nullable boolean wire contract; nullable migration applies to the four new masters.
vi.mock("@/lib/axios", () => ({ axios: { get: vi.fn() } }));
const configs = buildTreatmentTabConfigs({
  consultationData: [],
  examinationData: [],
  procedureData: [],
  vaccineData: [],
  checkupData: [],
  onReorderConsultations: () => {},
  onReorderExaminations: () => {},
  onReorderProcedures: () => {},
  onReorderVaccines: () => {},
  onReorderCheckups: () => {},
});
beforeEach(() => {
  cleanup();
  vi.mocked(axios.get).mockReset();
});
describe("clinical selection follows the adopted master contract", () => {
  it.each(
    [
      { kind: "consultation", endpoint: "consultations", adopted: true },
      { kind: "procedure", endpoint: "procedures", adopted: true },
      { kind: "medicine", endpoint: "medicines", adopted: true },
      { kind: "vaccine", endpoint: "vaccines", adopted: !!configs.vaccine.insuranceTriState },
      { kind: "checkup", endpoint: "checkup-types", adopted: !!configs.checkup.insuranceTriState },
    ].flatMap((row) =>
      (row.kind === "medicine" ? [true, false] : [undefined, null, true, false]).map((setting) => ({
        ...row,
        setting,
      })),
    ),
  )("$kind setting=$setting", async ({ kind, endpoint, adopted, setting }) => {
    const raw = {
      id: 7,
      name: `selection-${kind}`,
      price: 123,
      is_active: true,
      parent_id: 1,
      ...(adopted ? { version: 3, is_non_insurance: setting } : {}),
    };
    vi.mocked(axios.get).mockImplementation(async (url) => ({
      data: String(url).endsWith(`/masters/${endpoint}`) ? [raw] : [],
    }));
    let selected: TreatmentMasterItem | undefined;
    render(
      <TreatmentSearchDialog
        open
        onOpenChange={() => {}}
        onSelect={(item) => {
          selected = item;
        }}
      />,
      { wrapper: createTestWrapper() },
    );
    await userEvent.setup().click(await screen.findByText(raw.name));
    expect(selected).toBeDefined();
    const build = () =>
      buildMasterSelectionPayload({
        item: selected!,
        quantity: 1,
        sortOrder: 0,
        insuranceFallback: true,
      });
    if (adopted) {
      expect(selected?.insuranceContract).toBe("master");
      if (typeof setting === "boolean") expect(build().is_insurance).toBe(!setting);
      else expect(build).toThrow("保険設定");
    } else {
      expect(selected).not.toHaveProperty("insuranceContract");
      expect(build().is_insurance).toBe(true);
    }
  });
});
