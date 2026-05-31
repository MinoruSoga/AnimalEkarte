import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import { SidebarItemWithPermission } from "./SidebarItems";
import type { MenuItem } from "@/types";

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  }),
}));

function renderSidebarItem(item: MenuItem) {
  return render(
    <MemoryRouter>
      <SidebarItemWithPermission item={item} />
    </MemoryRouter>,
  );
}

describe("SidebarItemWithPermission", () => {
  it("サブメニュー付き項目で button の入れ子を作らない", () => {
    const { container } = renderSidebarItem({
      label: "LINE予約管理",
      path: "/line-reservation",
      subItems: [
        { label: "基本設定", path: "/line-reservation/settings" },
      ],
    });

    expect(container.querySelector("button button")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "LINE予約管理" })).toHaveAttribute("aria-expanded", "false");
    expect(screen.getByRole("button", { name: "LINE予約管理を展開" })).toBeInTheDocument();
  });

  it("展開ボタンからサブメニューを開閉できる", () => {
    renderSidebarItem({
      label: "LINE予約管理",
      path: "/line-reservation",
      subItems: [
        { label: "基本設定", path: "/line-reservation/settings" },
      ],
    });

    fireEvent.click(screen.getByRole("button", { name: "LINE予約管理を展開" }));

    expect(screen.getByRole("button", { name: "LINE予約管理" })).toHaveAttribute("aria-expanded", "true");
    expect(screen.getByRole("link", { name: "基本設定" })).toBeInTheDocument();
  });
});
