import { describe, it, expect } from "vitest";
import { ResourceExaminationUnconfirm } from "@/types/generated/models";
import {
  ALL_PERMISSION_RESOURCES,
  RESOURCE_LABELS,
  buildPermissionRuleMap,
  createEmptyPermissionRule,
  type PermissionRule,
} from "../lib/permission-rule-table-model";

describe("createEmptyPermissionRule", () => {
  it("指定した resource で全アクション false の初期ルールを返す", () => {
    expect(createEmptyPermissionRule("owners")).toEqual({
      resource: "owners",
      canView: false,
      canCreate: false,
      canEdit: false,
      canDelete: false,
    });
  });
});

describe("buildPermissionRuleMap", () => {
  it("resource をキーとした Map を構築する", () => {
    const rules: PermissionRule[] = [
      { resource: "owners", canView: true, canCreate: false, canEdit: false, canDelete: false },
      { resource: "reservations", canView: true, canCreate: true, canEdit: true, canDelete: false },
    ];

    const map = buildPermissionRuleMap(rules);

    expect(map.size).toBe(2);
    expect(map.get("owners")).toEqual(rules[0]);
    expect(map.get("reservations")).toEqual(rules[1]);
  });

  it("空配列からは空の Map を返す（未登録リソースは呼び出し側で createEmptyPermissionRule 補完が必要）", () => {
    const map = buildPermissionRuleMap([]);
    expect(map.size).toBe(0);
    expect(map.get("owners")).toBeUndefined();
  });

  it("同一 resource が重複する場合は配列内で後の要素が Map に残る（Map の仕様通り）", () => {
    const first: PermissionRule = {
      resource: "owners",
      canView: true,
      canCreate: false,
      canEdit: false,
      canDelete: false,
    };
    const second: PermissionRule = {
      resource: "owners",
      canView: false,
      canCreate: true,
      canEdit: true,
      canDelete: true,
    };

    const map = buildPermissionRuleMap([first, second]);

    expect(map.size).toBe(1);
    expect(map.get("owners")).toEqual(second);
  });
});

describe("ALL_PERMISSION_RESOURCES / RESOURCE_LABELS 整合性", () => {
  it("backend-generated examination unconfirm resource has an explicit label", () => {
    expect(ALL_PERMISSION_RESOURCES).toContain(ResourceExaminationUnconfirm);
    expect(RESOURCE_LABELS[ResourceExaminationUnconfirm]).toBe("検査確定解除");
  });

  it("ALL_PERMISSION_RESOURCES は RESOURCE_LABELS の全キーと一致する（欠落・重複がない）", () => {
    expect(ALL_PERMISSION_RESOURCES).toEqual(Object.keys(RESOURCE_LABELS));
    expect(new Set(ALL_PERMISSION_RESOURCES).size).toBe(ALL_PERMISSION_RESOURCES.length);
  });

  it("各リソースに空でないラベルが定義されている", () => {
    for (const resource of ALL_PERMISSION_RESOURCES) {
      expect(RESOURCE_LABELS[resource]).toBeTruthy();
    }
  });
});
