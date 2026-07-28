import { describe, it, expect, vi, beforeEach } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, useLocation } from "react-router";
import { C, STYLE } from "@/lib/design-tokens";
import { ResourceAccounting, type Resource } from "@/types/generated/models";
import type { MedicalRecord } from "../api/transforms";
import { MedicalRecords } from "./MedicalRecords";

// ─────────────────────────────────────────────────────────────
// MedicalRecords はルーティング・複数マスタ取得に依存する大きなルートのため、
// DESIGN.md 準拠（テーブルヘッダー・新規登録ボタン）の検証に必要な最小限のフックのみモックする。
// ─────────────────────────────────────────────────────────────

const mockUseGetMedicalRecords = vi.hoisted(() => vi.fn());
const mockIsValidStaff = vi.hoisted(() => vi.fn(() => true));
const mockUseClinicScope = vi.hoisted(() => vi.fn());
const mockUsePermission = vi.hoisted(() => vi.fn());
const mockDeleteRecord = vi.hoisted(() => vi.fn());

vi.mock("../api/get-medical-records", () => ({
  useGetMedicalRecords: mockUseGetMedicalRecords,
}));

vi.mock("../api/delete-medical-record", () => ({
  useDeleteMedicalRecord: vi.fn(() => ({ mutate: mockDeleteRecord })),
}));

vi.mock("@/hooks/use-staffs", () => ({
  useGetStaffs: vi.fn(() => ({ data: [] })),
}));

vi.mock("@/hooks/use-animal-species", () => ({
  useAnimalSpecies: vi.fn(() => ({ activeSpecies: [], allSpecies: [] })),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: mockUsePermission,
}));

vi.mock("@/hooks/use-staff-validation", () => ({
  useStaffValidation: vi.fn(() => ({ isValidStaff: mockIsValidStaff })),
}));

vi.mock("@/hooks/use-clinic-scope", () => ({
  useClinicScope: mockUseClinicScope,
}));

function defaultClinicScope() {
  return {
    assignedClinics: [],
    selectedClinicIds: ["clinic-1"],
    isMultiClinic: false,
    clinicNameById: new Map(),
    currentClinicId: "clinic-1",
    handleToggleClinic: vi.fn(),
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  document.body.style.pointerEvents = "";
  mockUseGetMedicalRecords.mockReturnValue({
    data: { data: [], total: 0, page: 1, limit: 20 },
    isLoading: false,
    isError: false,
  });
  mockIsValidStaff.mockReturnValue(true);
  mockUseClinicScope.mockReturnValue(defaultClinicScope());
  mockUsePermission.mockReturnValue({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  });
});

function makeMedicalRecord(overrides: Partial<MedicalRecord> = {}): MedicalRecord {
  return {
    id: "mr-1",
    recordNo: "MR-1",
    date: "2026-07-13",
    ownerId: "owner-1",
    ownerName: "山田太郎",
    petId: "pet-1",
    petName: "ポチ",
    species: "犬",
    chiefComplaint: "定期診察",
    chiefComplaintTypeId: null,
    doctor: "佐藤医師",
    visitType: undefined,
    nextVisitRecommendedDate: "",
    subjective: undefined,
    objective: undefined,
    assessment: undefined,
    plan: undefined,
    surgeryNotes: undefined,
    diagnosis: undefined,
    treatment: undefined,
    prescription: undefined,
    diagnosis1CategoryId: null,
    diagnosis1NameId: null,
    diagnosis2CategoryId: null,
    diagnosis2NameId: null,
    notes: undefined,
    accountingId: "acct-1",
    visitCount: 1,
    version: 1,
    recommendationReason: null,
    clinicId: "clinic-1",
    status: "作成中",
    petIsDeceased: false,
    ...overrides,
  };
}

function mockMedicalRecords(records: MedicalRecord[]) {
  mockUseGetMedicalRecords.mockReturnValue({
    data: { data: records, total: records.length, page: 1, limit: 20 },
    isLoading: false,
    isError: false,
  });
}

function LocationProbe() {
  const { pathname, state } = useLocation();
  const from =
    state &&
    typeof state === "object" &&
    "from" in state &&
    typeof state.from === "string"
      ? state.from
      : "";
  return (
    <>
      <output data-testid="location">{pathname}</output>
      <output data-testid="location-from">{from}</output>
    </>
  );
}

function renderPage(initialEntry = "/medical-records") {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <MedicalRecords />
      <LocationProbe />
    </MemoryRouter>,
  );
}

describe("MedicalRecords pet_id route filter", () => {
  it("URL の pet_id をAPIフィルタへ渡す", () => {
    renderPage("/medical-records?pet_id=22");

    expect(mockUseGetMedicalRecords).toHaveBeenLastCalledWith(
      expect.objectContaining({ petId: "22" }),
    );
  });

  it("pet_id が無いURLでは従来どおりpetIdを未指定にする", () => {
    renderPage();

    expect(mockUseGetMedicalRecords).toHaveBeenLastCalledWith(
      expect.objectContaining({ petId: undefined }),
    );
  });
});

describe("MedicalRecords 一覧テーブル (DESIGN.md ex-data-table-cell)", () => {
  it("テーブルヘッダーが sectionLabel（eyebrow 相当）で表示される", () => {
    renderPage();
    const header = screen.getByRole("columnheader", { name: "飼主名" });
    for (const cls of STYLE.sectionLabel.split(" ")) {
      expect(header.className).toContain(cls);
    }
  });

  it("『新規カルテ登録』が brand と同じ primary teal + pill を使う", () => {
    renderPage();
    const button = screen.getByRole("button", { name: /新規カルテ登録/ });
    expect(button.className).toContain("rounded-full");
    expect(button.className).toContain(C.bgActionPrimary);
  });
});

describe("MedicalRecords row navigation accessibility", () => {
  it("ペット名・診療日を含む44px以上のnative detail linkを行内に表示する", () => {
    mockMedicalRecords([makeMedicalRecord()]);
    renderPage();

    const detailLink = screen.getByRole("link", { name: /ポチ/ });
    expect(detailLink).toHaveAttribute("href", "/medical-records/mr-1");
    expect(detailLink).toHaveAccessibleName(/ポチ/);
    expect(detailLink).toHaveAccessibleName(/2026-07-13/);
    expect(detailLink).toHaveAccessibleName(/mr-1/);
    expect(detailLink).toHaveClass("min-h-11", "min-w-11");
    fireEvent.click(detailLink);
    expect(screen.getByTestId("location-from")).toHaveTextContent(
      /^\/medical-records$/,
    );
  });

  it("detail link以外のセルclickでは行遷移しない", () => {
    mockMedicalRecords([makeMedicalRecord()]);
    renderPage();

    fireEvent.click(screen.getByText("山田太郎"));

    expect(screen.getByTestId("location")).toHaveTextContent(/^\/medical-records$/);
  });
});

describe("MedicalRecords contextual actions", () => {
  it("会計buttonを44px以上にしペット名・診療日をaccessible nameに含める", () => {
    mockMedicalRecords([makeMedicalRecord()]);
    renderPage();

    const accountingButton = screen.getByRole("button", { name: /会計/ });
    expect(accountingButton).toHaveClass("min-h-11", "min-w-11");
    expect(accountingButton).toHaveAccessibleName(/ポチ/);
    expect(accountingButton).toHaveAccessibleName(/2026-07-13/);
    expect(accountingButton).toHaveAccessibleName(/mr-1/);
  });

  it("行操作buttonのaccessible nameにカルテIDを含める", () => {
    mockMedicalRecords([makeMedicalRecord()]);
    renderPage();

    expect(
      screen.getByRole("button", { name: /カルテ操作:.*mr-1/ }),
    ).toBeInTheDocument();
  });

  it("無効な担当医の警告をscreen readerへ文脈付きで説明する", () => {
    mockIsValidStaff.mockReturnValue(false);
    mockMedicalRecords([makeMedicalRecord({ doctor: "退職医" })]);
    renderPage();

    const warning = screen.getByLabelText(/無効/);
    expect(warning).toHaveAccessibleName(/退職医/);
  });

  it("他院のカルテ行では会計導線を表示しない", () => {
    mockUseClinicScope.mockReturnValue({
      ...defaultClinicScope(),
      isMultiClinic: true,
      selectedClinicIds: ["clinic-1", "clinic-2"],
      clinicNameById: new Map([
        ["clinic-1", "本院"],
        ["clinic-2", "分院"],
      ]),
    });
    mockMedicalRecords([
      makeMedicalRecord({ clinicId: "clinic-2", accountingId: "acct-other" }),
    ]);

    renderPage();

    expect(screen.queryByRole("button", { name: /会計/ })).not.toBeInTheDocument();
  });

  it("accounting:view権限がない場合は会計導線を表示しない", () => {
    mockUsePermission.mockImplementation((resource: Resource) =>
      resource === ResourceAccounting
        ? {
            canView: false,
            canCreate: false,
            canEdit: false,
            canDelete: false,
          }
        : {
            canView: true,
            canCreate: true,
            canEdit: true,
            canDelete: true,
          },
    );
    mockMedicalRecords([makeMedicalRecord({ accountingId: "acct-1" })]);

    renderPage();

    expect(mockUsePermission).toHaveBeenCalledWith(ResourceAccounting);
    expect(screen.queryByRole("button", { name: /会計/ })).not.toBeInTheDocument();
  });

  it("削除確認中にdelete権限を失った場合は削除せず、カルテの閲覧導線は維持する", async () => {
    const user = userEvent.setup();
    mockMedicalRecords([makeMedicalRecord()]);
    const view = renderPage();

    await user.click(screen.getByRole("button", { name: /カルテ操作:.*mr-1/ }));
    await user.click(await screen.findByRole("menuitem", { name: "削除" }));
    expect(screen.getByRole("alertdialog")).toBeInTheDocument();

    mockUsePermission.mockImplementation((resource: Resource) =>
      resource === ResourceAccounting
        ? {
            canView: true,
            canCreate: true,
            canEdit: true,
            canDelete: true,
          }
        : {
            canView: true,
            canCreate: true,
            canEdit: true,
            canDelete: false,
          },
    );
    view.rerender(
      <MemoryRouter initialEntries={["/medical-records"]}>
        <MedicalRecords />
        <LocationProbe />
      </MemoryRouter>,
    );

    await user.click(screen.getByRole("button", { name: "削除" }));

    expect(mockDeleteRecord).not.toHaveBeenCalled();
    expect(screen.getByRole("link", { name: /カルテ詳細:.*mr-1/ })).toBeInTheDocument();
  });

  it("死亡ペットのカルテは編集・削除actionを表示せず、閲覧導線は維持する", () => {
    mockMedicalRecords([makeMedicalRecord({ petIsDeceased: true })]);

    renderPage();

    expect(screen.queryByRole("button", { name: /カルテ操作:.*mr-1/ })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: /カルテ詳細:.*mr-1/ })).toBeInTheDocument();
  });

  it("削除確認中に対象ペットが死亡へ変わった場合は最新状態でdelete mutationを拒否する", async () => {
    const user = userEvent.setup();
    mockMedicalRecords([makeMedicalRecord()]);
    const view = renderPage();

    await user.click(screen.getByRole("button", { name: /カルテ操作:.*mr-1/ }));
    await user.click(await screen.findByRole("menuitem", { name: "削除" }));
    expect(screen.getByRole("alertdialog")).toBeInTheDocument();

    mockMedicalRecords([makeMedicalRecord({ petIsDeceased: true })]);
    view.rerender(
      <MemoryRouter initialEntries={["/medical-records"]}>
        <MedicalRecords />
        <LocationProbe />
      </MemoryRouter>,
    );

    await user.click(screen.getByRole("button", { name: "削除" }));

    expect(mockDeleteRecord).not.toHaveBeenCalled();
    expect(screen.getByRole("link", { name: /カルテ詳細:.*mr-1/ })).toBeInTheDocument();
  });
});
