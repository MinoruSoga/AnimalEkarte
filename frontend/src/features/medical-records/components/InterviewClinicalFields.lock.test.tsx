import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { InterviewChiefComplaint } from "./InterviewChiefComplaint";
import { InterviewTreatmentPolicy } from "./InterviewTreatmentPolicy";
import { MedicalRecordFloatingActions } from "./MedicalRecordFormActions";

const mockUsePermission = vi.hoisted(() => vi.fn());

vi.mock("@/hooks/use-permission", () => ({
  usePermission: mockUsePermission,
}));

vi.mock("../api/get-chief-complaint-types", () => ({
  useGetChiefComplaintTypes: () => ({ data: [], isLoading: false }),
}));

vi.mock("react-router", () => ({
  useNavigate: () => vi.fn(),
}));

function FinalizedFieldset({ children }: { children: ReactNode }) {
  return (
    <fieldset disabled className="border-0 p-0 m-0 min-w-0" data-testid="medical-record-edit-lock">
      {children}
    </fieldset>
  );
}

describe("BUG-035 interview clinical fields lock residual", () => {
  beforeEach(() => {
    mockUsePermission.mockReturnValue({
      canView: true,
      canCreate: true,
      canEdit: true,
      canDelete: true,
    });
  });

  it("isFinalized 時は canEdit=true でも主訴・治療方針 textarea の disabled 属性が true", () => {
    render(
      <FinalizedFieldset>
        <InterviewChiefComplaint
          chiefComplaint="主訴"
          setChiefComplaint={vi.fn()}
          chiefComplaintTypeId={null}
          setChiefComplaintTypeId={vi.fn()}
          templates={[]}
          onInsertTemplate={vi.fn()}
          isFinalized
        />
        <InterviewTreatmentPolicy treatmentPolicy="方針" setTreatmentPolicy={vi.fn()} isFinalized />
      </FinalizedFieldset>,
    );

    const chief = screen.getByLabelText("主訴詳細");
    const policy = screen.getByLabelText("治療方針");
    expect(chief).toBeDisabled();
    expect(policy).toBeDisabled();
    // UAT checked the content attribute / IDL; fieldset alone can leave .disabled false.
    expect(chief).toHaveAttribute("disabled");
    expect(policy).toHaveAttribute("disabled");
  });

  it("isFinalized 時は主操作の保存を出さない（追記は別コンポーネント）", () => {
    render(
      <MedicalRecordFloatingActions
        activeTab="問診"
        canDelete={false}
        canEdit
        canSubmit
        isNewRecord={false}
        isCreating={false}
        isSaving={false}
        isFinalized
        onDeleteClick={vi.fn()}
        onVitalsClick={vi.fn()}
        onPrintClick={vi.fn()}
        onFinalizeClick={vi.fn()}
      />,
    );

    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "確定する" })).not.toBeInTheDocument();
  });
});
