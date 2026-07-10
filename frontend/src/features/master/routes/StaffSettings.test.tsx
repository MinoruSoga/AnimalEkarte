import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, act } from "@testing-library/react";
import { StaffSettings } from "./StaffSettings";
import type { Staff } from "../api/staffs";
import type { StaffFormData } from "../components/staff-side-panel-model";

// ──────────────────────────────────────────────────────────
// R-F18: StaffSettings の validate() は「新規=email/password 必須」
// 「編集=email/password 任意」を editTarget（"new" か既存レコードか）で判定する。
// この判定が反転・欠落すると、新規登録がノーガードで通ってしまう、または
// 編集のたびにパスワード再入力を強制してしまう。
// useMasterSave.handleSave は validate 失敗時に createMutation/updateMutation を
// 呼ばずに黙って return するだけ（トースト等のユーザー通知は行われない — 既存の
// サイレント失敗。本テストはこの現状挙動を固定する）。
// ──────────────────────────────────────────────────────────

const { mockCreateMutate, mockUpdateMutate } = vi.hoisted(() => ({
  mockCreateMutate: vi.fn(),
  mockUpdateMutate: vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canView: true, canCreate: true, canEdit: true, canDelete: true }),
}));

vi.mock("../api/occupations", () => ({
  useGetAllOccupations: () => ({ data: [] }),
}));

vi.mock("../api/reservation-types", () => ({
  useGetReservationTypes: () => ({ data: [] }),
}));

vi.mock("../api/permission-groups", () => ({
  useGetPermissionGroups: () => ({ data: [] }),
}));

function makeStaff(overrides: Partial<Staff> = {}): Staff {
  return {
    id: "1",
    clinicId: "1",
    name: "既存 太郎",
    isActive: true,
    occupationId: null,
    occupationName: null,
    licenseNumber: "",
    sortOrder: 1,
    email: "existing@example.com",
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
    staffType: "doctor",
    reservationDisplayName: "",
    reservationVisible: true,
    reservationComment: "",
    reservationImageUrl: "",
    ...overrides,
  };
}

vi.mock("../api/staffs", () => ({
  useGetStaffs: () => ({ data: [makeStaff()] }),
  useCreateStaff: () => ({ mutate: mockCreateMutate }),
  useUpdateStaff: () => ({ mutate: mockUpdateMutate }),
  useDeleteStaff: () => ({ mutate: vi.fn() }),
  useUpdateStaffClinics: () => ({ mutate: vi.fn() }),
  useGetClinicsList: () => ({ data: [] }),
  useGetAllStaffPermissionGroupMap: () => ({ data: new Map() }),
  useUpdateStaffCapableReservationTypes: () => ({ mutate: vi.fn() }),
  useUpdateStaffPermissionGroups: () => ({ mutate: vi.fn() }),
}));

// MasterCRUDPage は PageLayout/DataTable/PropertyFilter 等の重い依存を持つため、
// ここでは薄くモックして handleSave / crud の呼び出しだけを検証する。
// renderRow / renderSidePanel は本テストでは呼ばれない（描画しない）。
let latestProps: {
  crud: { setEditTarget: (t: Staff | "new" | null) => void };
  handleSave: (data: StaffFormData) => void;
} | null = null;

vi.mock("../components/MasterCRUDPage", () => ({
  MasterCRUDPage: (props: typeof latestProps extends infer T ? NonNullable<T> : never) => {
    latestProps = props;
    return null;
  },
}));

function baseFormData(overrides: Partial<StaffFormData> = {}): StaffFormData {
  return {
    name: "テスト太郎",
    jobTitleId: null,
    licenseNumber: "",
    isActive: true,
    email: "",
    password: "",
    staffType: "doctor",
    reservationDisplayName: "",
    reservationVisible: true,
    reservationComment: "",
    reservationImageUrl: "",
    ...overrides,
  };
}

describe("StaffSettings validate() — 新規/編集判定", () => {
  beforeEach(() => {
    mockCreateMutate.mockClear();
    mockUpdateMutate.mockClear();
    latestProps = null;
  });

  it("新規登録: email が空の場合 createMutation を呼ばない", () => {
    render(<StaffSettings />);
    act(() => latestProps!.crud.setEditTarget("new"));
    act(() => latestProps!.handleSave(baseFormData({ email: "", password: `password123` })));

    expect(mockCreateMutate).not.toHaveBeenCalled();
  });

  it("新規登録: password が8文字未満の場合 createMutation を呼ばない", () => {
    render(<StaffSettings />);
    act(() => latestProps!.crud.setEditTarget("new"));
    act(() => latestProps!.handleSave(baseFormData({ email: "new@example.com", password: "short1" })));

    expect(mockCreateMutate).not.toHaveBeenCalled();
  });

  it("新規登録: email + 8文字以上の password が揃えば createMutation を呼ぶ", () => {
    render(<StaffSettings />);
    act(() => latestProps!.crud.setEditTarget("new"));
    act(() =>
      latestProps!.handleSave(baseFormData({ email: "new@example.com", password: `password123` })),
    );

    expect(mockCreateMutate).toHaveBeenCalledTimes(1);
    expect(mockUpdateMutate).not.toHaveBeenCalled();
  });

  it("編集: 既存スタッフを編集する場合 email/password が空でも updateMutation を呼ぶ（新規判定に誤って倒れない）", () => {
    render(<StaffSettings />);
    act(() => latestProps!.crud.setEditTarget(makeStaff()));
    act(() => latestProps!.handleSave(baseFormData({ email: "", password: "" })));

    expect(mockUpdateMutate).toHaveBeenCalledTimes(1);
    expect(mockCreateMutate).not.toHaveBeenCalled();
  });

  it("編集: name が空の場合は新規/編集を問わず updateMutation を呼ばない", () => {
    render(<StaffSettings />);
    act(() => latestProps!.crud.setEditTarget(makeStaff()));
    act(() => latestProps!.handleSave(baseFormData({ name: "   ", email: "", password: "" })));

    expect(mockUpdateMutate).not.toHaveBeenCalled();
  });
});
