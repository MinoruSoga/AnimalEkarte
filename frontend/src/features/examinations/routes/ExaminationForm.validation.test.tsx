import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ExaminationForm } from "./ExaminationForm";

const mocks = vi.hoisted(() => ({
  createMutateAsync: vi.fn(),
  updateMutateAsync: vi.fn(),
  axiosRequest: vi.fn(),
  navigate: vi.fn(),
  setSearchParams: vi.fn(),
  searchParams: "petId=42",
}));

vi.mock("react-router", () => ({
  useNavigate: () => mocks.navigate,
  useLocation: () => ({ state: undefined }),
  useParams: () => ({ id: undefined }),
  useSearchParams: () => [new URLSearchParams(mocks.searchParams), mocks.setSearchParams],
  MemoryRouter: ({ children }: { children: ReactNode }) => children,
  Link: ({ children, to }: { children: ReactNode; to: string }) => <a href={to}>{children}</a>,
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canCreate: true,
    canEdit: true,
    canDelete: true,
    canView: true,
  }),
}));

vi.mock("@/hooks/use-master-items", () => ({
  useGetMasterItems: (category: string) => {
    if (category === "examination") {
      return {
        data: [{ id: 5, name: "血液検査（院内）" }],
        isLoading: false,
      };
    }
    return { data: [], isLoading: false };
  },
}));

vi.mock("@/hooks/use-staffs", () => ({
  useGetStaffs: () => ({
    data: [
      {
        id: "3",
        name: "林文明",
        staffType: "doctor",
        isActive: true,
      },
      {
        id: "99",
        name: "お手入れ・オゾン療法",
        staffType: "resource",
        isActive: true,
      },
    ],
    isLoading: false,
  }),
}));

// ExaminationForm は未紐付け受信バナーを描画する。バナーは useQuery を使うため
// QueryClientProvider の無いフォーム単体テストでは null に差し替える。
vi.mock("@/components/shared/LabDeviceUnlinkedBanner/LabDeviceUnlinkedBanner", () => ({
  LabDeviceUnlinkedBanner: () => null,
}));

vi.mock("@/hooks/use-unsaved-changes", () => ({
  useUnsavedChanges: () => ({
    isDirty: false,
    markDirty: vi.fn(),
    markClean: vi.fn(),
  }),
}));

vi.mock("@/hooks/use-pet-selection", () => ({
  usePetSelection: () => ({
    selectedPets: [
      {
        id: "42",
        name: "ポチ",
        ownerName: "田中",
        ownerId: "5",
        species: "犬",
        status: "生存",
      },
    ],
    setSelectedPets: vi.fn(),
  }),
}));

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: () => ({
    data: {
      id: "42",
      name: "ポチ",
      ownerName: "田中",
      ownerId: "5",
      species: "犬",
      status: "生存",
    },
    isLoading: false,
  }),
}));

vi.mock("../api/get-examination", () => ({
  useGetExamination: () => ({ data: null }),
}));

vi.mock("../api/get-examinations", () => ({
  useGetExaminations: () => ({ data: [] }),
}));

vi.mock("../api/get-examination-items", () => ({
  useGetExaminationItems: () => ({
    data: undefined,
    isSuccess: false,
    isError: false,
  }),
}));

vi.mock("../api/get-exam-type-fields", () => ({
  useGetExamTypeFields: () => ({ data: undefined }),
}));

vi.mock("../api/create-examination", () => ({
  useCreateExamination: () => ({
    mutateAsync: mocks.createMutateAsync,
    isPending: false,
  }),
}));

vi.mock("../api/update-examination", () => ({
  useUpdateExamination: () => ({
    mutateAsync: mocks.updateMutateAsync,
    isPending: false,
  }),
}));

vi.mock("../api/delete-examination", () => ({
  useDeleteExamination: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock("../api/unconfirm-examination", () => ({
  useUnconfirmExamination: () => ({
    mutateAsync: vi.fn(),
    isPending: false,
  }),
}));

vi.mock("@/lib/axios", () => ({
  axios: {
    get: (...args: unknown[]) => mocks.axiosRequest("get", ...args),
    post: (...args: unknown[]) => mocks.axiosRequest("post", ...args),
    put: (...args: unknown[]) => mocks.axiosRequest("put", ...args),
    patch: (...args: unknown[]) => mocks.axiosRequest("patch", ...args),
    delete: (...args: unknown[]) => mocks.axiosRequest("delete", ...args),
  },
}));

vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn(), info: vi.fn() },
}));

vi.mock("@/components/shared/PageLayout/PageLayout", () => ({
  PageLayout: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

vi.mock("@/components/shared/NavigationBlocker", () => ({
  NavigationBlocker: () => null,
}));

vi.mock("@/components/shared/ConfirmDialog", () => ({
  ConfirmDialog: () => null,
}));

vi.mock("@/components/shared/PatientInfoCard", () => ({
  PatientInfoCard: () => null,
}));

vi.mock("../components/ExamItemsTable", () => ({
  ExamItemsTable: () => null,
}));

vi.mock("../components/ExaminationHistoryPanel", () => ({
  ExaminationHistoryPanel: () => null,
}));

vi.mock("../components/ExaminationUnconfirmDialog", () => ({
  ExaminationUnconfirmDialog: () => null,
}));

vi.mock("../components/ExaminationPatientChangeDialog", () => ({
  ExaminationPatientChangeDialog: () => null,
}));

function renderForm() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={client}>
      <ExaminationForm />
    </QueryClientProvider>,
  );
}

describe("ExaminationForm — empty submit validation (BUG-017 Mode3)", () => {
  beforeEach(() => {
    mocks.createMutateAsync.mockReset();
    mocks.updateMutateAsync.mockReset();
    mocks.axiosRequest.mockReset();
    mocks.createMutateAsync.mockResolvedValue({});
    mocks.updateMutateAsync.mockResolvedValue({});
  });

  it("実クリック保存で検査種別・担当医の日本語エラー・ARIA・先頭focus・request 0 を同時に証明する", async () => {
    const user = userEvent.setup();
    renderForm();

    const save = screen.getByRole("button", { name: "保存" });
    expect(save).toBeEnabled();

    await user.click(save);

    await waitFor(() => {
      expect(screen.getAllByRole("alert")).toHaveLength(2);
    });

    expect(screen.getByText("検査種別を選択してください")).toBeInTheDocument();
    expect(screen.getByText("担当医を選択してください")).toBeInTheDocument();

    const testType = screen.getByRole("combobox", { name: "検査種別" });
    const doctor = screen.getByRole("combobox", { name: "担当医" });
    expect(testType).toHaveAttribute("aria-invalid", "true");
    expect(doctor).toHaveAttribute("aria-invalid", "true");
    expect(testType).toHaveAttribute("aria-describedby", "testTypeId-error");
    expect(doctor).toHaveAttribute("aria-describedby", "doctorId-error");
    expect(document.getElementById("testTypeId-error")).toHaveTextContent(
      "検査種別を選択してください",
    );
    expect(document.getElementById("doctorId-error")).toHaveTextContent("担当医を選択してください");

    await waitFor(() => {
      expect(testType).toHaveFocus();
    });

    expect(mocks.createMutateAsync).not.toHaveBeenCalled();
    expect(mocks.updateMutateAsync).not.toHaveBeenCalled();
    expect(mocks.axiosRequest).not.toHaveBeenCalled();
  });
});
