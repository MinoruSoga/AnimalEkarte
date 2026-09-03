import { describe, expect, it } from "vitest";

import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { normalizeKana } from "@/lib/normalize-kana";
import type { Occupation } from "../api/occupations";
import type { PermissionGroup } from "../api/permission-groups";
import type { Staff } from "../api/staffs";
import type { StaffFormData } from "../lib/staff-side-panel-model";
import {
  buildGroupsByStaffId,
  buildStaffCreateRequest,
  buildStaffFilterProperties,
  buildStaffIds,
  buildStaffUpdateRequest,
  filterStaffByMasterFilters,
  searchStaff,
} from "./staff-settings-model";

const GROUPS = [
  {
    id: "group-1",
    clinicId: "1",
    name: "診療",
    description: "",
    color: "",
    isActive: true,
    sortOrder: 1,
    rules: [],
    createdAt: "",
    updatedAt: "",
  },
  {
    id: "group-2",
    clinicId: "1",
    name: "会計",
    description: "",
    color: "",
    isActive: true,
    sortOrder: 2,
    rules: [],
    createdAt: "",
    updatedAt: "",
  },
  {
    id: "group-3",
    clinicId: "1",
    name: "管理",
    description: "",
    color: "",
    isActive: true,
    sortOrder: 3,
    rules: [],
    createdAt: "",
    updatedAt: "",
  },
] satisfies PermissionGroup[];

const STAFF = {
  id: "staff-1",
  clinicId: "1",
  name: "やまだ",
  isActive: true,
  occupationId: "occupation-1",
  occupationName: "獣医師",
  licenseNumber: "SYN-001",
  sortOrder: 1,
  email: "staff@example.invalid",
  createdAt: "",
  updatedAt: "",
  staffType: "doctor",
  reservationDisplayName: "山田",
  reservationVisible: true,
  reservationComment: "",
  reservationImageUrl: "",
} satisfies Staff;

const FORM_DATA = {
  name: "合成スタッフ",
  jobTitleId: "occupation-1",
  licenseNumber: "SYN-002",
  isActive: false,
  email: "synthetic-staff@example.invalid",
  password: "test-password",
  staffType: "doctor",
  reservationDisplayName: "合成表示名",
  reservationVisible: false,
  reservationComment: "合成コメント",
  reservationImageUrl: "https://example.invalid/staff.png",
} satisfies StaffFormData;

describe("buildGroupsByStaffId", () => {
  it("group ID lookupを入力groups順で行い、未知IDを無視する", () => {
    const result = buildGroupsByStaffId({
      staffGroupMap: new Map([["staff-1", ["group-3", "unknown", "group-1"]]]),
      groups: GROUPS,
    });

    expect(result.get("staff-1")?.map((group) => group.id)).toEqual(["group-1", "group-3"]);
  });

  it("mapping未取得時は空Mapを返す", () => {
    expect(buildGroupsByStaffId({ staffGroupMap: undefined, groups: GROUPS }).size).toBe(0);
  });
});

describe("staff settings derived model", () => {
  it("staff IDを入力順で構築し、未取得は空配列にする", () => {
    expect(buildStaffIds([STAFF, { ...STAFF, id: "staff-2" }])).toEqual(["staff-1", "staff-2"]);
    expect(buildStaffIds(undefined)).toEqual([]);
  });

  it("職種filterはactive項目だけを元順序で公開する", () => {
    const occupations = [
      {
        id: "occupation-1",
        name: "獣医師",
        description: "",
        isActive: true,
        sortOrder: 2,
        createdAt: "",
        updatedAt: "",
      },
      {
        id: "occupation-2",
        name: "停止職種",
        description: "",
        isActive: false,
        sortOrder: 1,
        createdAt: "",
        updatedAt: "",
      },
    ] satisfies Occupation[];

    const properties = buildStaffFilterProperties(occupations);

    expect(properties.map((property) => property.key)).toEqual(["status", "occupationId"]);
    expect(properties[1]?.options).toEqual([{ value: "occupation-1", label: "獣医師" }]);
    expect(properties[1]?.conditions).toEqual(["is", "is_not"]);
  });

  it("氏名・職種をひらがな/カタカナ同一視して検索する", () => {
    expect(searchStaff(STAFF, normalizeKana("ヤマダ").toLowerCase())).toBe(true);
    expect(searchStaff(STAFF, "獣医")).toBe(true);
    expect(searchStaff(STAFF, "該当なし")).toBe(false);
  });

  it("statusとoccupationのis/is_notをANDで評価する", () => {
    const matches = [
      { key: "status", condition: "is", value: "active", displayValue: "有効" },
      {
        key: "occupationId",
        condition: "is_not",
        value: "occupation-2",
        displayValue: "停止職種以外",
      },
    ] satisfies ActiveFilter[];
    const missesStatus = [
      { key: "status", condition: "is_not", value: "active", displayValue: "無効" },
    ] satisfies ActiveFilter[];
    const missesOccupation = [
      { key: "occupationId", condition: "is", value: "occupation-2", displayValue: "停止職種" },
    ] satisfies ActiveFilter[];

    expect(filterStaffByMasterFilters(STAFF, matches)).toBe(true);
    expect(filterStaffByMasterFilters(STAFF, missesStatus)).toBe(false);
    expect(filterStaffByMasterFilters(STAFF, missesOccupation)).toBe(false);
  });

  it("create/update requestへ予約表示項目とoptional値を正規化する", () => {
    expect(buildStaffCreateRequest(FORM_DATA)).toEqual({
      name: "合成スタッフ",
      email: "synthetic-staff@example.invalid",
      password: "test-password",
      license_number: "SYN-002",
      occupation_id: "occupation-1",
      staff_type: "doctor",
      reservation_display_name: "合成表示名",
      reservation_visible: false,
      reservation_comment: "合成コメント",
      reservation_image_url: "https://example.invalid/staff.png",
    });
    expect(
      buildStaffUpdateRequest({
        ...FORM_DATA,
        jobTitleId: null,
        licenseNumber: "",
        password: "",
        reservationDisplayName: "",
        reservationComment: "",
        reservationImageUrl: "",
      }),
    ).toEqual({
      name: "合成スタッフ",
      license_number: undefined,
      is_active: false,
      occupation_id: undefined,
      password: undefined,
      staff_type: "doctor",
      reservation_display_name: undefined,
      reservation_visible: false,
      reservation_comment: undefined,
      reservation_image_url: undefined,
    });
  });
});
