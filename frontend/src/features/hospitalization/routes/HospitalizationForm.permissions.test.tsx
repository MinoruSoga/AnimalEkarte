import type { ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { HospitalizationForm } from "./HospitalizationForm";

const mocks = vi.hoisted(() => ({
  canDelete: true,
  petIsDeceased: false,
  deleteHospitalization: vi.fn(),
  useHospitalizationForm: vi.fn(),
  navigate: vi.fn(),
}));

vi.mock("react-router", () => ({
  useNavigate: () => mocks.navigate,
  useLocation: () => ({ state: undefined }),
  useParams: () => ({ id: "hospitalization-1" }),
  useSearchParams: () => [new URLSearchParams(), vi.fn()],
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ user: { displayName: "担当者" } }),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canCreate: true, canEdit: true, canDelete: mocks.canDelete }),
}));

vi.mock("@/hooks/use-master-items", () => ({
  useMasterItems: () => ({ data: [] }),
}));

vi.mock("@/hooks/use-unsaved-changes", () => ({
  useUnsavedChanges: () => ({ isDirty: false, markDirty: vi.fn(), markClean: vi.fn() }),
}));

vi.mock("../hooks/use-hospitalization-form", () => ({
  useHospitalizationForm: mocks.useHospitalizationForm,
}));

vi.mock("../api/delete-hospitalization", () => ({
  useDeleteHospitalization: () => ({ mutate: mocks.deleteHospitalization }),
}));

vi.mock("@/components/shared/PageLayout/PageLayout", () => ({
  PageLayout: ({ children, headerAction }: { children: ReactNode; headerAction: ReactNode }) => (
    <div>
      {headerAction}
      {children}
    </div>
  ),
}));

vi.mock("@/components/ui/button", () => ({
  Button: ({ children, ...props }: { children: ReactNode } & Record<string, unknown>) => (
    <button {...props}>{children}</button>
  ),
}));

vi.mock("@/components/shared/Form/SubmitButton", () => ({
  SubmitButton: ({ children }: { children: ReactNode }) => <button type="submit">{children}</button>,
}));

vi.mock("@/components/shared/ConfirmDialog/ConfirmDialog", () => ({
  ConfirmDialog: ({ open, onConfirm }: { open: boolean; onConfirm: () => void }) =>
    open ? <button aria-label="confirm-delete" onClick={onConfirm}>確認削除</button> : null,
}));

vi.mock("@/components/shared/PatientInfoCard/PatientInfoCard", () => ({ PatientInfoCard: () => null }));
vi.mock("../components/HospitalizationBasicInfo", () => ({ HospitalizationBasicInfo: () => null }));
vi.mock("../components/HospitalizationNoteCard", () => ({ HospitalizationNoteCard: () => null }));
vi.mock("../components/HospitalizationTreatmentTable", () => ({ HospitalizationTreatmentTable: () => null }));
vi.mock("../components/HospitalizationCostSummary", () => ({ HospitalizationCostSummary: () => null }));
vi.mock("@/components/shared/NavigationBlocker", () => ({ NavigationBlocker: () => null }));
vi.mock("@/components/shared/DataStates", () => ({
  LoadingFallback: () => <div>loading</div>,
  ErrorFallback: ({ message }: { message?: string }) => <div role="alert">{message}</div>,
}));
vi.mock("@/components/shared/FormFieldError", () => ({ FormFieldError: () => null }));

function renderForm() {
  return render(<HospitalizationForm />);
}

beforeEach(() => {
  mocks.canDelete = true;
  mocks.petIsDeceased = false;
  mocks.deleteHospitalization.mockReset();
  mocks.useHospitalizationForm.mockReset();
  mocks.useHospitalizationForm.mockImplementation(() => ({
    isEdit: true,
    isReadLoading: false,
    isReadNotFound: false,
    isReadError: false,
    retryRead: undefined,
    formData: {
      hospitalizationType: "入院",
      ownerRequest: "",
      staffNotes: "",
    },
    handleFormDataChange: vi.fn(),
    treatmentPlans: [],
    addTreatmentPlan: vi.fn(),
    removeTreatmentPlan: vi.fn(),
    updateTreatmentPlan: vi.fn(),
    calculateTotals: () => ({
      subtotalBeforeDiscount: 0,
      discountAmount: 0,
      subtotalAfterDiscount: 0,
      consumptionTax: 0,
      total: 0,
    }),
    petSelection: {
      selectedPets: [{
        id: "pet-1",
        name: "ポチ",
        ownerName: "飼主",
        species: "犬",
        status: mocks.petIsDeceased ? "死亡" : "生存",
      }],
    },
    formAction: vi.fn(),
    formState: { success: false },
  }));
});

describe("HospitalizationForm — mutation permission boundary", () => {
  it("permission wiring passes the resolved mode permission to the form hook", () => {
    renderForm();

    expect(mocks.useHospitalizationForm).toHaveBeenCalledWith("hospitalization-1", true);
  });

  it("permission revoked after opening confirmation prevents delete mutation", () => {
    const view = renderForm();

    fireEvent.click(screen.getByRole("button", { name: "削除" }));
    mocks.canDelete = false;
    view.rerender(<HospitalizationForm />);
    fireEvent.click(screen.getByRole("button", { name: "confirm-delete" }));

    expect(mocks.deleteHospitalization).not.toHaveBeenCalled();
  });

  it("petが死亡している場合は確認後もdelete mutationを拒否する", () => {
    mocks.petIsDeceased = true;
    renderForm();

    fireEvent.click(screen.getByRole("button", { name: "削除" }));
    fireEvent.click(screen.getByRole("button", { name: "confirm-delete" }));

    expect(mocks.deleteHospitalization).not.toHaveBeenCalled();
  });

  it("delete mutation keeps the existing id payload when permission is granted", () => {
    renderForm();

    fireEvent.click(screen.getByRole("button", { name: "削除" }));
    fireEvent.click(screen.getByRole("button", { name: "confirm-delete" }));

    expect(mocks.deleteHospitalization).toHaveBeenCalledWith("hospitalization-1", expect.any(Object));
  });

  it("hides delete button and shows honesty when child treatment plans exist", () => {
    mocks.useHospitalizationForm.mockImplementation(() => ({
      isEdit: true,
      formData: {
        hospitalizationType: "入院",
        ownerRequest: "",
        staffNotes: "",
      },
      handleFormDataChange: vi.fn(),
      treatmentPlans: [{ id: "tp-1", treatmentContent: "点滴" }],
      addTreatmentPlan: vi.fn(),
      removeTreatmentPlan: vi.fn(),
      updateTreatmentPlan: vi.fn(),
      calculateTotals: () => ({
        subtotalBeforeDiscount: 0,
        discountAmount: 0,
        subtotalAfterDiscount: 0,
        consumptionTax: 0,
        total: 0,
      }),
      petSelection: {
        selectedPets: [{
          id: "pet-1",
          name: "ポチ",
          ownerName: "飼主",
          species: "犬",
          status: "生存",
        }],
      },
      formAction: vi.fn(),
      formState: { success: false },
    }));

    renderForm();

    expect(screen.queryByRole("button", { name: "削除" })).not.toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent("治療プランが紐付いているため、この入院は削除できません。");
  });
});


describe("HospitalizationForm BUG-016 not-found gate", () => {
  it("isReadNotFound 時は ErrorFallback を出し保存ボタンを出さない", () => {
    mocks.useHospitalizationForm.mockImplementation(() => ({
      isEdit: true,
      isReadLoading: false,
      isReadNotFound: true,
      isReadError: false,
      retryRead: undefined,
      formData: {
        hospitalizationType: "入院",
        ownerRequest: "",
        staffNotes: "",
      },
      handleFormDataChange: vi.fn(),
      treatmentPlans: [],
      addTreatmentPlan: vi.fn(),
      removeTreatmentPlan: vi.fn(),
      updateTreatmentPlan: vi.fn(),
      calculateTotals: () => ({
        subtotalBeforeDiscount: 0,
        discountAmount: 0,
        subtotalAfterDiscount: 0,
        consumptionTax: 0,
        total: 0,
      }),
      petSelection: { selectedPets: [] },
      formAction: vi.fn(),
      formState: { success: false },
    }));

    renderForm();
    expect(screen.getByRole("alert")).toHaveTextContent("入院情報が見つかりません");
    expect(screen.queryByRole("button", { name: "更新" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "登録" })).not.toBeInTheDocument();
  });

  it("isReadError 時は retry 導線を出し保存ボタンを出さない", () => {
    const retry = vi.fn();
    mocks.useHospitalizationForm.mockImplementation(() => ({
      isEdit: true,
      isReadLoading: false,
      isReadNotFound: false,
      isReadError: true,
      retryRead: retry,
      formData: {
        hospitalizationType: "入院",
        ownerRequest: "",
        staffNotes: "",
      },
      handleFormDataChange: vi.fn(),
      treatmentPlans: [],
      addTreatmentPlan: vi.fn(),
      removeTreatmentPlan: vi.fn(),
      updateTreatmentPlan: vi.fn(),
      calculateTotals: () => ({
        subtotalBeforeDiscount: 0,
        discountAmount: 0,
        subtotalAfterDiscount: 0,
        consumptionTax: 0,
        total: 0,
      }),
      petSelection: { selectedPets: [] },
      formAction: vi.fn(),
      formState: { success: false },
    }));

    renderForm();
    expect(screen.getByRole("alert")).toHaveTextContent("入院情報の取得に失敗しました");
    expect(screen.getByRole("button", { name: "再試行" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "再試行" }));
    expect(retry).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole("button", { name: "更新" })).not.toBeInTheDocument();
  });
});
