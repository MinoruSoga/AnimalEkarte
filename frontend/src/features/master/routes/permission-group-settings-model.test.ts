import { describe, expect, it } from "vitest";

import type { PermissionGroup } from "../api/permission-groups";
import { ALL_PERMISSION_RESOURCES } from "../lib/permission-rule-table-model";
import {
  assertSavedPermissionRulesMatch,
  buildPermissionGroupUpdateRequest,
  expandPermissionGroupRules,
  permissionRulesMatchRequest,
} from "./permission-group-settings-model";

describe("permission-group-settings-model", () => {
  it("expandPermissionGroupRules fills every UI resource and keeps extras", () => {
    const expanded = expandPermissionGroupRules([
      {
        resource: "reception",
        canView: true,
        canCreate: false,
        canEdit: false,
        canDelete: false,
      },
      {
        resource: "custom-extra",
        canView: false,
        canCreate: true,
        canEdit: false,
        canDelete: false,
      },
    ]);

    expect(expanded.length).toBe(ALL_PERMISSION_RESOURCES.length + 1);
    const reception = expanded.find((rule) => rule.resource === "reception");
    expect(reception).toEqual({
      resource: "reception",
      canView: true,
      canCreate: false,
      canEdit: false,
      canDelete: false,
    });
    const owners = expanded.find((rule) => rule.resource === "owners");
    expect(owners).toEqual({
      resource: "owners",
      canView: false,
      canCreate: false,
      canEdit: false,
      canDelete: false,
    });
    const extra = expanded.find((rule) => rule.resource === "custom-extra");
    expect(extra?.canCreate).toBe(true);
  });

  it("update request always includes explicit false flags for unchecked boxes", () => {
    const req = buildPermissionGroupUpdateRequest({
      name: "執行",
      description: "",
      color: "#6B7280",
      isActive: true,
      rules: [
        {
          resource: "reception",
          canView: false,
          canCreate: true,
          canEdit: false,
          canDelete: false,
        },
      ],
    });

    expect(req.rules).toBeDefined();
    expect(req.rules!.length).toBeGreaterThanOrEqual(ALL_PERMISSION_RESOURCES.length);
    const reception = req.rules!.find((rule) => rule.resource === "reception");
    expect(reception).toEqual({
      resource: "reception",
      can_view: false,
      can_create: true,
      can_edit: false,
      can_delete: false,
    });
  });

  it("permissionRulesMatchRequest detects stale server rules (BUG-024 false success)", () => {
    const requested = [
      {
        resource: "reception",
        can_view: false,
        can_create: true,
        can_edit: true,
        can_delete: true,
      },
    ];
    const staleSaved = [
      {
        id: "1",
        groupId: "1",
        resource: "reception",
        canView: true,
        canCreate: true,
        canEdit: true,
        canDelete: true,
        createdAt: "",
        updatedAt: "",
      },
    ];
    expect(permissionRulesMatchRequest(requested, staleSaved)).toBe(false);

    const matchingSaved = [
      {
        ...staleSaved[0],
        canView: false,
      },
    ];
    expect(permissionRulesMatchRequest(requested, matchingSaved)).toBe(true);
  });

  it("assertSavedPermissionRulesMatch throws on mismatch", () => {
    const saved: PermissionGroup = {
      id: "1",
      clinicId: "1",
      name: "執行",
      description: "",
      color: "#6B7280",
      isActive: true,
      sortOrder: 1,
      rules: [
        {
          id: "1",
          groupId: "1",
          resource: "reception",
          canView: true,
          canCreate: true,
          canEdit: true,
          canDelete: true,
          createdAt: "",
          updatedAt: "",
        },
      ],
      createdAt: "",
      updatedAt: "",
    };

    expect(() =>
      assertSavedPermissionRulesMatch(
        [
          {
            resource: "reception",
            can_view: false,
            can_create: true,
            can_edit: true,
            can_delete: true,
          },
        ],
        saved,
      ),
    ).toThrow(/権限マトリクスの保存結果/);
  });
});
