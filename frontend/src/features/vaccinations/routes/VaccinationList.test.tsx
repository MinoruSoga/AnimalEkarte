import { act, useLayoutEffect } from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import {
  createMemoryRouter,
  Link,
  MemoryRouter,
  Outlet,
  RouterProvider,
  useLocation,
} from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { C } from "@/lib/design-tokens";
import { useBackNavigation } from "@/hooks/use-back-navigation";
import { useGetPet } from "@/hooks/use-pet";
import type { Pet } from "@/types";
import type { VaccinationRecord } from "../api/transforms";
import { VaccinationForm } from "./VaccinationForm";
import { VaccinationList } from "./VaccinationList";
import { VaccinationPetSelection } from "./VaccinationPetSelection";

const selectionPet = vi.hoisted(
  () =>
    ({
      id: "pet-9",
      ownerId: "owner-1",
      ownerName: "山田太郎",
      name: "ポチ",
      species: "犬",
      status: "生存",
      birthDate: "2020-01-01",
      weight: "5",
      environment: "室内",
      lastVisit: "2026-05-01",
      petNumber: "P-9",
    }) as Pet,
);

const mockUseFilterVaccinations = vi.hoisted(() => vi.fn());
const listMocks = vi.hoisted(() => ({
  canDelete: true,
  deleteVaccination: vi.fn(),
  confirmDelete: undefined as (() => void) | undefined,
}));

vi.mock("../hooks/use-vaccinations", () => ({
  useFilterVaccinations: mockUseFilterVaccinations,
}));

vi.mock("../api/delete-vaccination", () => ({
  useDeleteVaccination: vi.fn(() => ({ mutate: listMocks.deleteVaccination })),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(() => ({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: listMocks.canDelete,
  })),
}));

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: vi.fn(() => ({
    data: { id: "pet-1", status: "生存" },
    isLoading: false,
    isError: false,
  })),
  useGetPets: vi.fn(() => ({
    data: [selectionPet],
    total: 1,
    page: 1,
    limit: 20,
    error: null,
    isLoading: false,
    isPlaceholderData: false,
  })),
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: { clinics: [{ clinicId: "clinic-1" }] },
    currentClinicId: "clinic-1",
  }),
}));

vi.mock("@/hooks/use-unsaved-changes", () => ({
  useUnsavedChanges: () => ({
    isDirty: false,
    markDirty: vi.fn(),
    markClean: vi.fn(),
  }),
}));

vi.mock("../api/get-vaccinations", () => ({
  useGetVaccinations: () => ({ data: [] }),
}));

vi.mock("../hooks/use-vaccination-form", async () => {
  const { useState } = await import("react");
  return {
    useVaccinationForm: () => {
      const [success, setSuccess] = useState(false);
      return {
        isEdit: false,
        isReadLoading: false,
        isEditPetReady: true,
        isReadNotFound: false,
        isReadError: false,
        retryRead: vi.fn(),
        petSelection: {
          selectedPets: [{ id: "pet-9", ownerId: "owner-1", name: "ポチ", status: "生存" }],
        },
        form: {
          doctorName: "",
          date: "",
          setDate: vi.fn(),
          vaccineId: "",
          setVaccineId: vi.fn(),
          vaccineOptions: [],
          nextScheduleType: "1year",
          setNextScheduleType: vi.fn(),
          nextDate: "",
          setNextDate: vi.fn(),
          supplemental: "",
          setSupplemental: vi.fn(),
          lot1: "",
          setLot1: vi.fn(),
          lot2: "",
          setLot2: vi.fn(),
          lot3: "",
          setLot3: vi.fn(),
          lot4: "",
          setLot4: vi.fn(),
          remarks: "",
          setRemarks: vi.fn(),
        },
        formAction: () => setSuccess(true),
        formState: { success, timestamp: success ? 1 : 0, fieldErrors: {} },
        isSaving: false,
        fieldErrors: {},
        handleDelete: vi.fn(),
        isDeleting: false,
        historyFilter: {
          filterStartDate: "",
          setFilterStartDate: vi.fn(),
          filterEndDate: "",
          setFilterEndDate: vi.fn(),
          historySearchTerm: "",
          setHistorySearchTerm: vi.fn(),
          sortOrder: "desc" as const,
          setSortOrder: vi.fn(),
          handleClearHistoryFilter: vi.fn(),
        },
      };
    },
  };
});

vi.mock("../components/VaccinationFormPanels", () => ({
  VaccinationFieldsPanel: () => null,
  VaccinationHistoryPanel: () => null,
}));

vi.mock("@/components/shared/PatientInfoCard", () => ({
  PatientInfoCard: () => null,
  formatPatientPetDetails: () => "",
}));

vi.mock("@/components/shared/NavigationBlocker", () => ({
  NavigationBlocker: () => null,
}));

vi.mock("@/components/shared/PetSelection/PetSelectionSearchForm", () => ({
  PetSelectionSearchForm: () => null,
}));

vi.mock("@/components/shared/ConfirmDialog/ConfirmDialog", () => ({
  ConfirmDialog: ({ open, onConfirm }: { open: boolean; onConfirm: () => void }) => {
    listMocks.confirmDelete = onConfirm;
    return open ? <button onClick={onConfirm}>確認削除</button> : null;
  },
}));

const vaccination: VaccinationRecord = {
  id: "vac-1",
  medicalRecordId: undefined,
  petId: "pet-1",
  ownerName: "山田太郎",
  petName: "ポチ",
  vaccineId: "vaccine-1",
  vaccineName: "混合ワクチン",
  doctor: "佐藤医師",
  date: "2026-07-13",
  nextDate: "2027-07-13",
  nextScheduleType: undefined,
  lot1: undefined,
  lot2: undefined,
  lot3: undefined,
  lot4: undefined,
  supplemental: undefined,
  remarks: undefined,
  version: 1,
};

function LocationProbe() {
  const { pathname, search } = useLocation();
  return <output data-testid="location">{`${pathname}${search}`}</output>;
}

function RegistrationShell() {
  const location = useLocation();
  return (
    <>
      <output data-testid="location">{`${location.pathname}${location.search}`}</output>
      <output data-testid="nav-state">{JSON.stringify(location.state ?? null)}</output>
      <Outlet />
    </>
  );
}

function renderRegistrationFlow() {
  const router = createMemoryRouter(
    [
      {
        element: <RegistrationShell />,
        children: [
          { path: "/vaccinations", element: <VaccinationList /> },
          { path: "/vaccinations/select-pet", element: <VaccinationPetSelection /> },
          { path: "/vaccinations/new", element: <VaccinationForm /> },
        ],
      },
    ],
    {
      initialEntries: [
        {
          pathname: "/vaccinations",
          search: "?petId=10",
          state: { from: "/?date=2026-05-29" },
        },
      ],
    },
  );
  render(<RouterProvider router={router} />);
}

function DetailReturnHarness() {
  const location = useLocation();
  const back = useBackNavigation("/vaccinations");
  return (
    <>
      <VaccinationList />
      <output data-testid="location">{`${location.pathname}${location.search}`}</output>
      <output data-testid="nav-state">{JSON.stringify(location.state)}</output>
      <button type="button" onClick={back}>
        詳細から戻る
      </button>
    </>
  );
}

function DeleteRevocationHarness({ confirmAfterRender }: { confirmAfterRender: boolean }) {
  useLayoutEffect(() => {
    if (confirmAfterRender) {
      listMocks.confirmDelete?.();
    }
  }, [confirmAfterRender]);

  return <VaccinationList />;
}

function renderPage(content = <VaccinationList />) {
  return render(
    <MemoryRouter initialEntries={["/vaccinations"]}>
      {content}
      <LocationProbe />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  listMocks.canDelete = true;
  listMocks.deleteVaccination.mockReset();
  listMocks.confirmDelete = undefined;
  mockUseFilterVaccinations.mockReturnValue({
    data: [vaccination],
    allVaccinations: [vaccination],
    isLoading: false,
    error: null,
  });
  vi.mocked(useGetPet).mockReturnValue({
    data: { id: "pet-1", status: "生存" },
    isLoading: false,
    isError: false,
  } as ReturnType<typeof useGetPet>);
});

describe("VaccinationList mutation permission boundary", () => {
  it("削除権限剥奪をcommitした直後のlayout phaseで確認してもdelete mutationを発行しない", async () => {
    const user = userEvent.setup();
    const view = renderPage(<DeleteRevocationHarness confirmAfterRender={false} />);

    await user.click(screen.getByRole("button", { name: /vac-1/ }));
    await user.click(screen.getByRole("menuitem", { name: "削除" }));
    expect(screen.getByRole("button", { name: "確認削除" })).toBeInTheDocument();

    listMocks.canDelete = false;
    view.rerender(
      <MemoryRouter initialEntries={["/vaccinations"]}>
        <DeleteRevocationHarness confirmAfterRender />
        <LocationProbe />
      </MemoryRouter>,
    );

    expect(listMocks.deleteVaccination).not.toHaveBeenCalled();
  });

  it("削除対象が死亡petならdelete mutationを発行しない", async () => {
    vi.mocked(useGetPet).mockReturnValue({
      data: { id: "pet-1", status: "死亡" },
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);
    const user = userEvent.setup();
    renderPage();

    await user.click(screen.getByRole("button", { name: /vac-1/ }));
    await user.click(screen.getByRole("menuitem", { name: "削除" }));
    await user.click(screen.getByRole("button", { name: "確認削除" }));

    expect(listMocks.deleteVaccination).not.toHaveBeenCalled();
  });

  it("編集対象が死亡petならdetailへ遷移しない", async () => {
    vi.mocked(useGetPet).mockReturnValue({
      data: { id: "pet-1", status: "死亡" },
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);
    const user = userEvent.setup();
    renderPage();

    await user.click(screen.getByRole("button", { name: /vac-1/ }));

    expect(screen.queryByRole("menuitem", { name: "編集" })).not.toBeInTheDocument();
    expect(screen.getByTestId("location")).toHaveTextContent(/^\/vaccinations$/);
  });

  it("カルテ紐付け済みの行はカルテ予防接種タブへ遷移する", async () => {
    mockUseFilterVaccinations.mockReturnValue({
      data: [{ ...vaccination, medicalRecordId: "mr-1" }],
      allVaccinations: [{ ...vaccination, medicalRecordId: "mr-1" }],
      isLoading: false,
      error: null,
    });
    const user = userEvent.setup();
    renderPage();

    await user.click(screen.getByRole("button", { name: /vac-1/ }));
    await user.click(screen.getByRole("menuitem", { name: "編集" }));

    expect(screen.getByTestId("location")).toHaveTextContent(
      "/medical-records/mr-1?tab=%E4%BA%88%E9%98%B2%E6%8E%A5%E7%A8%AE&vaccinationId=vac-1",
    );
  });

  it("編集対象が生存petなら従来どおりdetailへ遷移する", async () => {
    const user = userEvent.setup();
    renderPage();

    await user.click(screen.getByRole("button", { name: /vac-1/ }));
    await user.click(screen.getByRole("menuitem", { name: "編集" }));

    expect(screen.getByTestId("location")).toHaveTextContent(/^\/vaccinations\/vac-1$/);
  });

  it("削除対象petのlookup失敗時はdelete mutationを発行しない", async () => {
    vi.mocked(useGetPet).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
    } as ReturnType<typeof useGetPet>);
    const user = userEvent.setup();
    renderPage();

    await user.click(screen.getByRole("button", { name: /vac-1/ }));
    await user.click(screen.getByRole("menuitem", { name: "削除" }));
    await user.click(screen.getByRole("button", { name: "確認削除" }));

    expect(listMocks.deleteVaccination).not.toHaveBeenCalled();
  });
});

describe("VaccinationList row navigation accessibility", () => {
  it("ペット名・実施日を含む44px以上のnative detail linkを行内に表示する", () => {
    renderPage();

    const detailLink = screen.getByRole("link", { name: /ポチ/ });
    expect(detailLink).toHaveAttribute("href", "/vaccinations/vac-1");
    expect(detailLink).toHaveAccessibleName(/ポチ/);
    expect(detailLink).toHaveAccessibleName(/2026-07-13/);
    expect(detailLink).toHaveAccessibleName(/vac-1/);
    expect(detailLink).toHaveClass("min-h-11", "min-w-11");
  });

  it("detail link以外のセルclickでは行遷移しない", () => {
    renderPage();

    fireEvent.click(screen.getByText("山田太郎"));

    expect(screen.getByTestId("location")).toHaveTextContent(/^\/vaccinations$/);
  });

  it("行操作buttonのaccessible nameに予防接種IDを含める", () => {
    renderPage();

    expect(screen.getByRole("button", { name: /vac-1/ })).toBeInTheDocument();
  });
});

describe("VaccinationList 次回予定の期限超過表示", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-07-23T15:30:00Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  function renderWithNextDate(nextDate: string) {
    mockUseFilterVaccinations.mockReturnValue({
      data: [{ ...vaccination, nextDate }],
      allVaccinations: [{ ...vaccination, nextDate }],
      isLoading: false,
      error: null,
    });
    renderPage();
  }

  function getNextDateCell() {
    const row = screen.getByText(vaccination.petName).closest("tr");
    return row?.querySelectorAll("td")[1] ?? null;
  }

  it("過去日の次回予定に期限切れバッジを出し、（期限超過）の文言を表示しない (Q2)", () => {
    renderWithNextDate("2026-07-23");

    const cell = getNextDateCell();
    expect(cell).toHaveTextContent("2026/07/23");
    expect(cell).toHaveTextContent("期限切れ");
    expect(cell).not.toHaveTextContent("期限超過");
    expect(cell).not.toHaveClass(C.danger);
  });

  it("当日は期限切れにならず、30日以内なら期限間近バッジ (Q2)", () => {
    renderWithNextDate("2026-07-24");

    const cell = getNextDateCell();
    expect(cell).toHaveTextContent("2026/07/24");
    expect(cell).toHaveTextContent("期限間近");
    expect(cell).not.toHaveTextContent("期限切れ");
  });

  it("30日超の未来日はバッジなしの通常表示 (Q2)", () => {
    renderWithNextDate("2026-12-01");

    const cell = getNextDateCell();
    expect(cell).toHaveTextContent("2026/12/01");
    expect(cell).not.toHaveTextContent("期限切れ");
    expect(cell).not.toHaveTextContent("期限間近");
    expect(cell).not.toHaveClass(C.danger);
  });

  it("空欄は未設定プレースホルダーを表示する", () => {
    renderWithNextDate("");

    const cell = getNextDateCell();
    expect(cell).not.toHaveClass(C.danger);
    expect(cell).toHaveTextContent(/^-$/);
    expect(cell).not.toHaveTextContent("期限切れ");
  });

  it("日付+期限バッジは1行表示する (EMR-378)", () => {
    renderWithNextDate("2026-07-23");

    const cell = getNextDateCell();
    expect(cell).toHaveClass("whitespace-nowrap");
  });
});

describe("VaccinationList 列レイアウト (EMR-378)", () => {
  it("次回予定ヘッダは日付+バッジ分の最小幅を持ち折り返さない", () => {
    renderPage();

    const header = screen.getByRole("columnheader", { name: /次回予定/ });
    expect(header.className).toContain("whitespace-nowrap");
    expect(header.className).toMatch(/min-w-\[/);
  });

  it("飼主名・ペット名・予防接種名ヘッダは折り返さず読める最小幅を持つ (T002)", () => {
    renderPage();

    const ownerHeader = screen.getByRole("columnheader", { name: /飼主名/ });
    expect(ownerHeader.className).toContain("whitespace-nowrap");
    expect(ownerHeader.className).toContain("min-w-[140px]");

    const petHeader = screen.getByRole("columnheader", { name: /ペット名/ });
    expect(petHeader.className).toContain("whitespace-nowrap");
    expect(petHeader.className).toContain("min-w-[160px]");

    const vaccineHeader = screen.getByRole("columnheader", { name: /予防接種名/ });
    expect(vaccineHeader.className).toContain("whitespace-nowrap");
    expect(vaccineHeader.className).toContain("min-w-[160px]");
  });

  it("識別子セルはヘッダと同じ最小幅契約を持ち、氏名は折り返せる", () => {
    renderPage();

    const ownerCell = screen.getByText("山田太郎").closest("td");
    expect(ownerCell?.className).toContain("min-w-[140px]");

    const petCell = screen.getByRole("link", { name: /ポチ/ }).closest("td");
    expect(petCell?.className).toContain("min-w-[160px]");

    const vaccineCell = screen.getByText("混合ワクチン").closest("td");
    expect(vaccineCell?.className).toContain("min-w-[160px]");
  });
});

describe("VaccinationList petId query", () => {
  it("URL の petId を既存 GET フィルタ引数へ渡し、新規登録は残す", () => {
    render(
      <MemoryRouter initialEntries={["/vaccinations?petId=pet-9"]}>
        <VaccinationList />
      </MemoryRouter>,
    );

    const filters = mockUseFilterVaccinations.mock.calls.at(-1)?.[1] as { petId?: string };
    expect(filters.petId).toBe("pet-9");
    expect(screen.getByRole("button", { name: "新規予防接種登録" })).toBeInTheDocument();
  });

  it("petId が無い一覧は GET フィルタに petId を載せない", () => {
    renderPage();

    const filters = mockUseFilterVaccinations.mock.calls.at(-1)?.[1] as { petId?: string };
    expect(filters).not.toHaveProperty("petId");
  });

  it("ペット指定の 2 ページ目から一覧全体へ戻ると 1 ページ目を表示する", async () => {
    const user = userEvent.setup();
    mockUseFilterVaccinations.mockImplementation(
      (_search: string, filters?: { petId?: string }) => {
        const label = filters?.petId ? `pet-${filters.petId}` : "all";
        const rows = Array.from({ length: 21 }, (_, index) => ({
          ...vaccination,
          id: `${label}-${index + 1}`,
          petName: `${label}-${index + 1}`,
        }));
        return { data: rows, allVaccinations: rows, isLoading: false, error: null };
      },
    );

    render(
      <MemoryRouter initialEntries={["/vaccinations?petId=10"]}>
        <Link to="/vaccinations">予防接種管理</Link>
        <VaccinationList />
        <LocationProbe />
      </MemoryRouter>,
    );

    await user.click(screen.getByRole("button", { name: "2" }));
    expect(screen.getByTestId("location")).toHaveTextContent("petId=10");
    expect(screen.getByTestId("location")).toHaveTextContent("page=2");
    expect(screen.getByText("pet-10-21")).toBeInTheDocument();

    await user.click(screen.getByRole("link", { name: "予防接種管理" }));

    expect(screen.getByTestId("location")).toHaveTextContent(/^\/vaccinations$/);
    expect(screen.getByRole("button", { name: "1" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByText("all-1")).toBeInTheDocument();
    expect(screen.queryByText("all-21")).not.toBeInTheDocument();
  });

  it("URL の page=2 を直接開くと 2 ページ目を表示する", () => {
    mockUseFilterVaccinations.mockImplementation(
      (_search: string, filters?: { petId?: string }) => {
        const label = filters?.petId ? `pet-${filters.petId}` : "all";
        const rows = Array.from({ length: 21 }, (_, index) => ({
          ...vaccination,
          id: `${label}-${index + 1}`,
          petName: `${label}-${index + 1}`,
        }));
        return { data: rows, allVaccinations: rows, isLoading: false, error: null };
      },
    );

    render(
      <MemoryRouter initialEntries={["/vaccinations?petId=10&page=2"]}>
        <VaccinationList />
        <LocationProbe />
      </MemoryRouter>,
    );

    expect(screen.getByRole("button", { name: "2" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByText("pet-10-21")).toBeInTheDocument();
  });

  it("全体一覧から戻るとペット指定の page を表示に戻す", async () => {
    mockUseFilterVaccinations.mockImplementation(
      (_search: string, filters?: { petId?: string }) => {
        const label = filters?.petId ? `pet-${filters.petId}` : "all";
        const rows = Array.from({ length: 21 }, (_, index) => ({
          ...vaccination,
          id: `${label}-${index + 1}`,
          petName: `${label}-${index + 1}`,
        }));
        return { data: rows, allVaccinations: rows, isLoading: false, error: null };
      },
    );
    const router = createMemoryRouter(
      [
        {
          path: "/vaccinations",
          element: (
            <>
              <VaccinationList />
              <LocationProbe />
            </>
          ),
        },
      ],
      {
        initialEntries: ["/vaccinations?petId=10&page=2", "/vaccinations"],
        initialIndex: 1,
      },
    );
    render(<RouterProvider router={router} />);

    expect(screen.getByRole("button", { name: "1" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByText("all-1")).toBeInTheDocument();

    await act(async () => {
      await router.navigate(-1);
    });

    expect(screen.getByTestId("location")).toHaveTextContent("petId=10");
    expect(screen.getByTestId("location")).toHaveTextContent("page=2");
    expect(screen.getByRole("button", { name: "2" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByText("pet-10-21")).toBeInTheDocument();

    await act(async () => {
      await router.navigate(1);
    });

    expect(screen.getByTestId("location")).toHaveTextContent(/^\/vaccinations$/);
    expect(screen.getByRole("button", { name: "1" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByText("all-1")).toBeInTheDocument();
  });
});

describe("VaccinationList URL page 同期（FE-RC-028: useUrlPageSync）", () => {
  it("totalPagesを超えるURL pageは読み込み後にクランプされる", async () => {
    render(
      <MemoryRouter initialEntries={["/vaccinations?page=99"]}>
        <VaccinationList />
        <LocationProbe />
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("location")).toHaveTextContent(/^\/vaccinations$/);
    });
  });

  it("ページのクランプ後も受付の戻り先を残す", async () => {
    render(
      <MemoryRouter
        initialEntries={[
          {
            pathname: "/vaccinations",
            search: "?petId=10&page=99",
            state: { from: "/?date=2026-05-29" },
          },
        ]}
      >
        <DetailReturnHarness />
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("location")).toHaveTextContent(/^\/vaccinations\?petId=10$/);
    });
    expect(screen.getByTestId("nav-state")).toHaveTextContent("/?date=2026-05-29");
  });
});

describe("VaccinationList detail return", () => {
  it("ペット名から開いた詳細は一覧URLへ戻り、受付の戻り先を一覧へ戻す", async () => {
    const user = userEvent.setup();
    render(
      <MemoryRouter
        initialEntries={[
          {
            pathname: "/vaccinations",
            search: "?petId=10",
            state: { from: "/?date=2026-05-29" },
          },
        ]}
      >
        <DetailReturnHarness />
      </MemoryRouter>,
    );

    await user.click(screen.getByRole("link", { name: /ポチ/ }));

    expect(screen.getByTestId("location")).toHaveTextContent(/^\/vaccinations\/vac-1$/);
    expect(screen.getByTestId("nav-state")).toHaveTextContent('"/vaccinations?petId=10"');
    expect(screen.getByTestId("nav-state")).toHaveTextContent("/?date=2026-05-29");

    await user.click(screen.getByRole("button", { name: "詳細から戻る" }));

    expect(screen.getByTestId("location")).toHaveTextContent(/^\/vaccinations\?petId=10$/);
    expect(screen.getByTestId("nav-state")).toHaveTextContent('{"from":"/?date=2026-05-29"}');
  });

  it("カルテ付きの編集も同じ一覧URLを戻り先にする", async () => {
    const user = userEvent.setup();
    mockUseFilterVaccinations.mockReturnValue({
      data: [{ ...vaccination, medicalRecordId: "mr-9" }],
      allVaccinations: [{ ...vaccination, medicalRecordId: "mr-9" }],
      isLoading: false,
      error: null,
    });
    render(
      <MemoryRouter
        initialEntries={[
          {
            pathname: "/vaccinations",
            search: "?petId=10",
            state: { from: "/?date=2026-05-29" },
          },
        ]}
      >
        <DetailReturnHarness />
      </MemoryRouter>,
    );

    await user.click(screen.getByRole("button", { name: /vac-1/ }));
    await user.click(screen.getByRole("menuitem", { name: "編集" }));

    expect(screen.getByTestId("location")).toHaveTextContent("/medical-records/mr-9");
    expect(screen.getByTestId("nav-state")).toHaveTextContent('"/vaccinations?petId=10"');
    expect(screen.getByTestId("nav-state")).toHaveTextContent("/?date=2026-05-29");
  });

  it("検索でpageを外しても受付の戻り先を残す", () => {
    render(
      <MemoryRouter
        initialEntries={[
          {
            pathname: "/vaccinations",
            search: "?petId=10&page=2",
            state: { from: "/?date=2026-05-29" },
          },
        ]}
      >
        <DetailReturnHarness />
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByPlaceholderText("飼主名、ペット名、予防接種名..."), {
      target: { value: "混合" },
    });

    expect(screen.getByTestId("location")).toHaveTextContent(/^\/vaccinations\?petId=10$/);
    expect(screen.getByTestId("nav-state")).toHaveTextContent("/?date=2026-05-29");
  });
});

describe("VaccinationList registration return", () => {
  it("新規登録をやめるとペット指定の一覧と受付の戻り先へ戻る", async () => {
    const user = userEvent.setup();
    renderRegistrationFlow();

    await user.click(screen.getByRole("button", { name: "新規予防接種登録" }));

    expect(screen.getByTestId("location")).toHaveTextContent(/^\/vaccinations\/select-pet$/);
    expect(screen.getByTestId("nav-state")).toHaveTextContent('"/vaccinations?petId=10"');
    expect(screen.getByTestId("nav-state")).toHaveTextContent("/?date=2026-05-29");

    await user.click(screen.getByRole("button", { name: "戻る" }));

    expect(screen.getByTestId("location")).toHaveTextContent(/^\/vaccinations\?petId=10$/);
    expect(screen.getByTestId("nav-state")).toHaveTextContent('{"from":"/?date=2026-05-29"}');
  });

  it("選択後に戻ると同じ一覧と受付の戻り先へ戻る", async () => {
    const user = userEvent.setup();
    renderRegistrationFlow();

    await user.click(screen.getByRole("button", { name: "新規予防接種登録" }));
    await user.click(screen.getByRole("button", { name: "選択: ポチ (ID pet-9)" }));

    expect(screen.getByTestId("location")).toHaveTextContent(/^\/vaccinations\/new\?petId=pet-9$/);
    expect(screen.getByTestId("nav-state")).toHaveTextContent('"/vaccinations?petId=10"');
    expect(screen.getByTestId("nav-state")).toHaveTextContent("/?date=2026-05-29");

    await user.click(screen.getByRole("button", { name: "戻る" }));

    expect(screen.getByTestId("location")).toHaveTextContent(/^\/vaccinations\?petId=10$/);
    expect(screen.getByTestId("nav-state")).toHaveTextContent('{"from":"/?date=2026-05-29"}');
  });

  it("選択後に保存すると同じ一覧と受付の戻り先へ戻る", async () => {
    const user = userEvent.setup();
    renderRegistrationFlow();

    await user.click(screen.getByRole("button", { name: "新規予防接種登録" }));
    await user.click(screen.getByRole("button", { name: "選択: ポチ (ID pet-9)" }));
    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(screen.getByTestId("location")).toHaveTextContent(/^\/vaccinations\?petId=10$/);
    expect(screen.getByTestId("nav-state")).toHaveTextContent('{"from":"/?date=2026-05-29"}');
  });
});

describe("VaccinationList 列順 (EMR-415 V4)", () => {
  it("列は 実施日→次回予定→飼主名→ペット名→予防接種名→操作 の順", () => {
    renderPage();

    const headerTexts = screen
      .getAllByRole("columnheader")
      .map((cell) => cell.textContent ?? "");
    expect(headerTexts).toEqual([
      "実施日",
      "次回予定",
      "飼主名",
      "ペット名",
      "予防接種名",
      "操作",
    ]);
  });
});
