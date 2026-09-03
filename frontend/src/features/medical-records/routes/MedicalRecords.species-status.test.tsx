import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { useAnimalSpecies } from "@/hooks/use-animal-species";

import { MedicalRecords } from "./MedicalRecords";

type AnimalSpeciesState = ReturnType<typeof useAnimalSpecies>;

const mocks = vi.hoisted(() => ({
  useAnimalSpecies: vi.fn<() => AnimalSpeciesState>(),
}));

vi.mock("@/hooks/use-animal-species", () => ({
  useAnimalSpecies: mocks.useAnimalSpecies,
}));

vi.mock("@/hooks/use-staffs", () => ({
  useGetStaffs: () => ({ data: [] }),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  }),
}));

vi.mock("@/hooks/use-clinic-scope", () => ({
  useClinicScope: () => ({
    assignedClinics: [],
    selectedClinicIds: ["clinic-1"],
    isMultiClinic: false,
    clinicNameById: new Map<string, string>(),
    currentClinicId: "clinic-1",
    handleToggleClinic: vi.fn(),
  }),
}));

vi.mock("@/hooks/use-staff-validation", () => ({
  useStaffValidation: () => ({ isValidStaff: () => true }),
}));

vi.mock("../hooks/use-medical-records", () => ({
  useMedicalRecordsList: () => ({
    records: [],
    total: 0,
    isLoading: false,
    isError: false,
  }),
}));

vi.mock("../api/delete-medical-record", () => ({
  useDeleteMedicalRecord: () => ({ mutate: vi.fn() }),
}));

const animalSpecies = [
  {
    id: 1,
    name: "犬",
    is_active: true,
    sort_order: 1,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    label: "犬",
    isInactive: false,
  },
] satisfies AnimalSpeciesState["activeSpecies"];

function createSpeciesState(overrides: Partial<AnimalSpeciesState> = {}): AnimalSpeciesState {
  return {
    allSpecies: animalSpecies,
    activeSpecies: animalSpecies,
    isLoading: false,
    isError: false,
    error: null,
    ...overrides,
  };
}

function renderPage() {
  render(
    <MemoryRouter initialEntries={["/medical-records"]}>
      <MedicalRecords />
    </MemoryRouter>,
  );
}

async function openSpeciesFilter(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole("button", { name: "フィルタを追加" }));
  await user.click(screen.getByRole("option", { name: "種" }));
  await user.click(screen.getByRole("button", { name: "次と一致" }));
}

describe("MedicalRecords animal species status", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useAnimalSpecies.mockReturnValue(createSpeciesState());
  });

  it("取得失敗を最優先のalertで示し、生エラーを隠して新規登録と種フィルタを維持する", async () => {
    const rawError = "GET /v1/masters/animal-species: database timeout";
    mocks.useAnimalSpecies.mockReturnValue(
      createSpeciesState({
        isLoading: true,
        isError: true,
        error: new Error(rawError),
      }),
    );
    const user = userEvent.setup();

    renderPage();

    const alert = screen.getByRole("alert");
    expect(alert).toHaveTextContent("動物種の取得に失敗しました。");
    expect(alert).toHaveAttribute("aria-atomic", "true");
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(screen.queryByText(rawError)).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "新規カルテ登録" })).toBeEnabled();

    await openSpeciesFilter(user);

    expect(screen.getByText("選択肢がありません")).toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "犬" })).not.toBeInTheDocument();
  });

  it("読み込み中を候補より優先してpoliteなstatusで示す", async () => {
    mocks.useAnimalSpecies.mockReturnValue(
      createSpeciesState({
        isLoading: true,
      }),
    );
    const user = userEvent.setup();

    renderPage();

    const status = screen.getByRole("status");
    expect(status).toHaveTextContent("動物種を読み込み中です。");
    expect(status).toHaveAttribute("aria-live", "polite");
    expect(status).toHaveAttribute("aria-atomic", "true");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByText("動物種マスタが登録されていません。")).not.toBeInTheDocument();

    await openSpeciesFilter(user);

    expect(screen.getByText("選択肢がありません")).toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "犬" })).not.toBeInTheDocument();
  });

  it("取得成功かつ0件をdistinctなpolite statusで示す", () => {
    mocks.useAnimalSpecies.mockReturnValue(
      createSpeciesState({
        allSpecies: [],
        activeSpecies: [],
      }),
    );

    renderPage();

    const status = screen.getByRole("status");
    expect(status).toHaveTextContent("動物種マスタが登録されていません。");
    expect(status).toHaveAttribute("aria-live", "polite");
    expect(status).toHaveAttribute("aria-atomic", "true");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByText("動物種を読み込み中です。")).not.toBeInTheDocument();
  });

  it("取得成功かつ候補ありでは状態表示を消して種を選択肢にする", async () => {
    const user = userEvent.setup();

    renderPage();

    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByRole("status")).not.toBeInTheDocument();

    await openSpeciesFilter(user);

    expect(screen.getByRole("option", { name: "犬" })).toBeInTheDocument();
    expect(screen.queryByText("選択肢がありません")).not.toBeInTheDocument();
  });
});
