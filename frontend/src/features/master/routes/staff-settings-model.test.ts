import { describe, expect, it } from "vitest";

import type { PermissionGroup } from "../api/permission-groups";
import { buildGroupsByStaffId } from "./staff-settings-model";

const GROUPS = [
  { id: "group-1", clinicId: "1", name: "診療", description: "", color: "", isActive: true, sortOrder: 1, rules: [], createdAt: "", updatedAt: "" },
  { id: "group-2", clinicId: "1", name: "会計", description: "", color: "", isActive: true, sortOrder: 2, rules: [], createdAt: "", updatedAt: "" },
  { id: "group-3", clinicId: "1", name: "管理", description: "", color: "", isActive: true, sortOrder: 3, rules: [], createdAt: "", updatedAt: "" },
] satisfies PermissionGroup[];

describe("buildGroupsByStaffId", () => {
  it("group ID lookupを入力groups順で行い、未知IDを無視する", () => {
    const result = buildGroupsByStaffId({
      staffGroupMap: new Map([["staff-1", ["group-3", "unknown", "group-1"]]]),
      groups: GROUPS,
    });

    expect(result.get("staff-1")?.map((group) => group.id)).toEqual(["group-1", "group-3"]);
  });
});
