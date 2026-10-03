import { describe, it, expect, vi, beforeEach } from "vitest";
import { render } from "@testing-library/react";
import type { ReactNode } from "react";

import { MedicalRecordServiceTabs } from "./MedicalRecordServiceTabs";
import type { MedicalRecordTabsAreaProps } from "../lib/medical-record-tabs-types";
import type { Pet } from "@/types";

// ロック境界の wrapper と兄弟タブは本テストの対象外。MedicalRecordImage への
// isLocked 伝播（EMR-216）だけを検証するため子要素はキャプチャスタブに置き換える。
vi.mock("./MedicalRecordTabsShared", () => ({
  MedicalRecordMountedTab: ({ children }: { children?: ReactNode }) => <div>{children}</div>,
  MedicalRecordSaveRequired: ({ children }: { children?: ReactNode }) => <div>{children}</div>,
}));
vi.mock("./CheckupsTab/CheckupsTab", () => ({ CheckupsTab: () => null }));
vi.mock("./MedicalRecordBillCheck", () => ({ MedicalRecordBillCheck: () => null }));
vi.mock("./MedicalRecordEstimate", () => ({ MedicalRecordEstimate: () => null }));
vi.mock("./MedicalRecordExamination", () => ({ MedicalRecordExamination: () => null }));
vi.mock("./MedicalRecordVaccination", () => ({ MedicalRecordVaccination: () => null }));

const captured = vi.hoisted(() => ({
  imageIsLocked: undefined as boolean | undefined,
}));
vi.mock("./MedicalRecordImage", () => ({
  MedicalRecordImage: (props: { isLocked?: boolean }) => {
    captured.imageIsLocked = props.isLocked;
    return null;
  },
}));

const selectedPet = {
  id: "1",
  status: "生存",
  species: "犬",
} as unknown as Pet;

const baseProps: MedicalRecordTabsAreaProps & { isFinalized: boolean } = {
  activeTab: "画像",
  mountedTabs: new Set(["画像"]),
  isNewRecord: false,
  recordId: "123",
  selectedPet,
  chiefComplaint: "",
  chiefComplaintTypeId: null,
  treatmentPolicy: "",
  historyItems: [],
  physicalExam: "",
  plan: "",
  assessment: "",
  diagnosis1CategoryId: null,
  diagnosis1NameId: null,
  diagnosis2CategoryId: null,
  diagnosis2NameId: null,
  ownerDiscountRate: 0,
  nextVisitDate: "",
  hasLineIntegration: false,
  recommendationReason: null,
  lstepStatus: undefined,
  recordStatus: "",
  diagnosis1NameIdError: undefined,
  isLocked: false,
  isFinalized: false,
  onChiefComplaintChange: () => {},
  onChiefComplaintTypeIdChange: () => {},
  onTreatmentPolicyChange: () => {},
  onPhysicalExamChange: () => {},
  onPlanChange: () => {},
  onAssessmentChange: () => {},
  onDiagnosis1CategoryIdChange: () => {},
  onDiagnosis1NameIdChange: () => {},
  onDiagnosis2CategoryIdChange: () => {},
  onDiagnosis2NameIdChange: () => {},
  onNextVisitDateChange: () => {},
  onNextVisitDateValidChange: () => {},
  onRecommendationReasonChange: () => {},
  onRegisterEstimateSave: () => {},
};

beforeEach(() => {
  captured.imageIsLocked = undefined;
});

describe("MedicalRecordServiceTabs — isLocked 伝播 (EMR-216)", () => {
  it("isLocked=true を画像タブの MedicalRecordImage へ伝播する", () => {
    render(<MedicalRecordServiceTabs {...baseProps} isLocked />);

    expect(captured.imageIsLocked).toBe(true);
  });

  it("isLocked=false のとき MedicalRecordImage へ false を渡す", () => {
    render(<MedicalRecordServiceTabs {...baseProps} isLocked={false} />);

    expect(captured.imageIsLocked).toBe(false);
  });
});
