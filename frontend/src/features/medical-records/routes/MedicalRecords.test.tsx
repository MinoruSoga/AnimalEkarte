import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { C, STYLE } from "@/lib/design-tokens";
import { MedicalRecords } from "./MedicalRecords";

// ─────────────────────────────────────────────────────────────
// MedicalRecords はルーティング・複数マスタ取得に依存する大きなルートのため、
// DESIGN.md 準拠（テーブルヘッダー・新規登録ボタン）の検証に必要な最小限のフックのみモックする。
// ─────────────────────────────────────────────────────────────

vi.mock("../api/get-medical-records", () => ({
  useGetMedicalRecords: vi.fn(() => ({
    data: { data: [], total: 0, page: 1, limit: 20 },
    isLoading: false,
    isError: false,
  })),
}));

vi.mock("../api/delete-medical-record", () => ({
  useDeleteMedicalRecord: vi.fn(() => ({ mutate: vi.fn() })),
}));

vi.mock("@/hooks/use-staffs", () => ({
  useGetStaffs: vi.fn(() => ({ data: [] })),
}));

vi.mock("@/hooks/use-animal-species", () => ({
  useAnimalSpecies: vi.fn(() => ({ activeSpecies: [], allSpecies: [] })),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(() => ({ canCreate: true, canEdit: true, canDelete: true })),
}));

vi.mock("@/hooks/use-staff-validation", () => ({
  useStaffValidation: vi.fn(() => ({ isValidStaff: () => true })),
}));

vi.mock("@/hooks/use-clinic-scope", () => ({
  useClinicScope: vi.fn(() => ({
    assignedClinics: [],
    selectedClinicIds: ["clinic-1"],
    isMultiClinic: false,
    clinicNameById: new Map(),
    currentClinicId: "clinic-1",
    handleToggleClinic: vi.fn(),
  })),
}));

beforeEach(() => {
  vi.clearAllMocks();
});

function renderPage() {
  return render(
    <MemoryRouter>
      <MedicalRecords />
    </MemoryRouter>,
  );
}

describe("MedicalRecords 一覧テーブル (DESIGN.md ex-data-table-cell)", () => {
  it("テーブルヘッダーが sectionLabel（eyebrow 相当）で表示される", () => {
    renderPage();
    const header = screen.getByRole("columnheader", { name: "飼主名" });
    for (const cls of STYLE.sectionLabel.split(" ")) {
      expect(header.className).toContain(cls);
    }
  });

  it("『新規カルテ登録』が DESIGN.md button-primary（brand blue + pill）を使う", () => {
    renderPage();
    const button = screen.getByRole("button", { name: /新規カルテ登録/ });
    expect(button.className).toContain("rounded-full");
    expect(button.className).toContain(C.bgBrand);
  });
});
