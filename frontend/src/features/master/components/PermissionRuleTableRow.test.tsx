import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { PermissionRuleTableRow } from "./PermissionRuleTableRow";

describe("PermissionRuleTableRow", () => {
  it("各checkboxをresource表示名とaction表示名を含む一意なaccessible nameで識別できる", () => {
    render(
      <table>
        <tbody>
          <PermissionRuleTableRow
            resource="master-permission"
            rule={{
              resource: "master-permission",
              canView: true,
              canCreate: false,
              canEdit: true,
              canDelete: false,
            }}
            onRuleChange={vi.fn()}
          />
        </tbody>
      </table>,
    );

    const expectedNames = [
      "権限グループ 表示",
      "権限グループ 作成",
      "権限グループ 編集",
      "権限グループ 削除",
    ];

    for (const name of expectedNames) {
      expect(screen.getByRole("checkbox", { name })).toBeInTheDocument();
    }
    expect(new Set(expectedNames).size).toBe(expectedNames.length);
  });
});
