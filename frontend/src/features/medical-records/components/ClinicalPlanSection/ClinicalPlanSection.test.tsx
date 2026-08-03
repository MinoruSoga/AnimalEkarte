import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { ClinicalPlanSection } from "./ClinicalPlanSection";

vi.mock("../../api/get-diagnosis-options", () => ({
  useGetDiagnosisTypes: () => ({ data: [], isLoading: false }),
  useGetDiagnosisNames: () => ({ data: [], isLoading: false }),
}));

vi.mock("@/components/shared/MasterLink", () => ({
  MasterLink: () => null,
}));

vi.mock("@/components/ui/searchable-select", () => ({
  SearchableSelect: ({
    value,
    disabled,
    placeholder,
  }: {
    value: string;
    disabled?: boolean;
    placeholder?: string;
  }) => (
    <select aria-label={placeholder ?? "select"} value={value} disabled={disabled} readOnly>
      <option value={value}>{value}</option>
    </select>
  ),
}));

describe("ClinicalPlanSection BUG-010 controlled inputs", () => {
  it("親から渡された3欄の値を表示し、入力変更は親 setter のみを呼ぶ（独自保存 mutation を持たない）", async () => {
    const user = userEvent.setup();
    const onPhysicalExamChange = vi.fn();
    const onDiagnosisDetailsChange = vi.fn();
    const onTreatmentPolicyChange = vi.fn();

    render(
      <ClinicalPlanSection
        medicalRecordId="42"
        canEdit
        physicalExam="身体検査の初期値"
        onPhysicalExamChange={onPhysicalExamChange}
        diagnosisDetails="診断詳細の初期値"
        onDiagnosisDetailsChange={onDiagnosisDetailsChange}
        treatmentPolicy="治療方針の初期値"
        onTreatmentPolicyChange={onTreatmentPolicyChange}
        diagnosisTypeId={null}
        onDiagnosisTypeIdChange={vi.fn()}
        diagnosisNameId={null}
        onDiagnosisNameIdChange={vi.fn()}
      />,
    );

    expect(screen.getByDisplayValue("身体検査の初期値")).toBeInTheDocument();
    expect(screen.getByDisplayValue("診断詳細の初期値")).toBeInTheDocument();
    expect(screen.getByDisplayValue("治療方針の初期値")).toBeInTheDocument();

    await user.type(screen.getByDisplayValue("身体検査の初期値"), "X");
    expect(onPhysicalExamChange).toHaveBeenCalled();

    // onRegisterSave / 独自 save 経路は props に存在しない（controlled only）
    expect(screen.queryByRole("button", { name: /保存/ })).not.toBeInTheDocument();
  });

  it("canEdit=false のとき3欄は disabled", () => {
    render(
      <ClinicalPlanSection
        medicalRecordId="42"
        canEdit={false}
        physicalExam="所見"
        onPhysicalExamChange={vi.fn()}
        diagnosisDetails="診断"
        onDiagnosisDetailsChange={vi.fn()}
        treatmentPolicy="方針"
        onTreatmentPolicyChange={vi.fn()}
        diagnosisTypeId={null}
        onDiagnosisTypeIdChange={vi.fn()}
        diagnosisNameId={null}
        onDiagnosisNameIdChange={vi.fn()}
      />,
    );

    expect(screen.getByDisplayValue("所見")).toBeDisabled();
    expect(screen.getByDisplayValue("診断")).toBeDisabled();
    expect(screen.getByDisplayValue("方針")).toBeDisabled();
  });
});
