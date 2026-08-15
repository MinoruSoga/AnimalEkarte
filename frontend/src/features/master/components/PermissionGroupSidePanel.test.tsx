import { useState, type ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { PermissionGroup } from "../api/permission-groups";
import { PermissionGroupSidePanel } from "./PermissionGroupSidePanel";

vi.mock("@/components/shared/SidePeek", () => ({
  MasterSidePanel: ({
    action,
    children,
    isDirty,
  }: {
    action?: () => Promise<void> | void;
    children: ReactNode;
    isDirty: boolean;
  }) => (
    <div>
      <output aria-label="編集状態">{isDirty ? "dirty" : "clean"}</output>
      <button type="button" onClick={() => void action?.()}>
        保存
      </button>
      {children}
    </div>
  ),
  PropertyRow: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  StatusToggleButton: () => null,
}));

vi.mock("./PermissionRuleTable", () => ({
  PermissionRuleTable: () => null,
}));

const EXISTING_GROUP: PermissionGroup = {
  id: "1",
  clinicId: "1",
  name: "受付",
  description: "受付担当",
  color: "#000000",
  isActive: true,
  sortOrder: 1,
  rules: [],
  createdAt: "2026-07-26T00:00:00Z",
  updatedAt: "2026-07-26T00:00:00Z",
};

describe("PermissionGroupSidePanel dirty state", () => {
  it("saveがfalseを返した場合はdirty状態を維持する", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn().mockResolvedValue(false);
    const onSaveSuccess = vi.fn();

    render(
      <PermissionGroupSidePanel
        item={EXISTING_GROUP}
        onClose={vi.fn()}
        onSave={onSave}
        onSaveSuccess={onSaveSuccess}
      />,
    );

    await user.type(screen.getByPlaceholderText("グループの説明"), "変更");
    expect(screen.getByRole("status", { name: "編集状態" })).toHaveTextContent("dirty");

    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(onSave).toHaveBeenCalledTimes(1);
    expect(onSaveSuccess).not.toHaveBeenCalled();
    expect(screen.getByRole("status", { name: "編集状態" })).toHaveTextContent("dirty");
  });

  it("save成功時はdirty guardをcleanにしてからpanelを閉じる", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn().mockResolvedValue(true);
    const onDirtyChange = vi.fn();

    function Harness() {
      const [open, setOpen] = useState(true);
      return open ? (
        <PermissionGroupSidePanel
          item={EXISTING_GROUP}
          onClose={vi.fn()}
          onSave={onSave}
          onSaveSuccess={() => setOpen(false)}
          onDirtyChange={onDirtyChange}
        />
      ) : (
        <output aria-label="panel状態">closed</output>
      );
    }

    render(<Harness />);

    await user.type(screen.getByPlaceholderText("グループの説明"), "変更");
    expect(screen.getByRole("status", { name: "編集状態" })).toHaveTextContent("dirty");

    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(onSave).toHaveBeenCalledTimes(1);
    expect(onDirtyChange).toHaveBeenLastCalledWith(false);
    expect(screen.getByRole("status", { name: "panel状態" })).toHaveTextContent("closed");
  });
});
