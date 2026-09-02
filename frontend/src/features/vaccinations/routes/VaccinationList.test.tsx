import { useLayoutEffect } from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, useLocation } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { C } from "@/lib/design-tokens";
import { useGetPet } from "@/hooks/use-pet";
import type { VaccinationRecord } from "../api/transforms";
import { VaccinationList } from "./VaccinationList";

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
}));

vi.mock("@/components/shared/ConfirmDialog/ConfirmDialog", () => ({
  ConfirmDialog: ({
    open,
    onConfirm,
  }: {
    open: boolean;
    onConfirm: () => void;
  }) => {
    listMocks.confirmDelete = onConfirm;
    return open ? <button onClick={onConfirm}>確認削除</button> : null;
  },
}));

const vaccination: VaccinationRecord = {
  id: "vac-1",
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
};

function LocationProbe() {
  const { pathname, search } = useLocation();
  return <output data-testid="location">{`${pathname}${search}`}</output>;
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
    const view = renderPage(
      <DeleteRevocationHarness confirmAfterRender={false} />,
    );

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

    expect(screen.getByTestId("location")).toHaveTextContent(
      /^\/vaccinations\/vac-1$/,
    );
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
    return row?.querySelectorAll("td")[4] ?? null;
  }

  it("過去日だけをdanger表現で期限超過表示する", () => {
    renderWithNextDate("2026-07-23");

    const cell = getNextDateCell();
    expect(cell).toHaveClass("font-mono", C.danger);
    expect(cell).toHaveTextContent("2026-07-23");
    expect(cell).toHaveTextContent("（期限超過）");
  });

  it("未来日は現状の通常表示を維持する", () => {
    renderWithNextDate("2026-07-25");

    const cell = getNextDateCell();
    expect(cell).toHaveClass("font-mono", C.text);
    expect(cell).not.toHaveClass(C.danger);
    expect(cell).toHaveTextContent("2026-07-25");
    expect(cell).not.toHaveTextContent("期限超過");
  });

  it("空欄は現状の空セル表示を維持する", () => {
    renderWithNextDate("");

    const cell = getNextDateCell();
    expect(cell).toHaveClass("font-mono", C.text);
    expect(cell).not.toHaveClass(C.danger);
    expect(cell).toHaveTextContent(/^\s*$/);
    expect(cell).not.toHaveTextContent("期限超過");
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
});
