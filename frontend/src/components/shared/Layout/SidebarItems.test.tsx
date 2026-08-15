import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import type { MenuItem } from "@/types";

import { SidebarItemWithPermission } from "./SidebarItems";

// String constants avoid generated/models imports in this layout test (TASK-444-S1 allowlist).
const ResourceHospitalSettings = "hospital-settings" as NonNullable<MenuItem["resource"]>;
const ResourceLstepAnalytics = "lstep-analytics" as NonNullable<MenuItem["resource"]>;

const defaultHasPermission = vi.fn(() => true);

function authValue(
  hasPermission: AuthContextValue["hasPermission"] = defaultHasPermission,
): AuthContextValue {
  return {
    user: null,
    currentClinicId: "clinic-test-1",
    isAuthenticated: true,
    isLoading: false,
    login: async () => {},
    logout: async () => {},
    switchClinic: () => {},
    hasPermission,
    refreshPermissions: async () => {},
  };
}

function renderSidebarItem(
  item: MenuItem,
  collapsed = false,
  initialEntry = "/",
  hasPermission: AuthContextValue["hasPermission"] = defaultHasPermission,
) {
  return render(
    <AuthContext.Provider value={authValue(hasPermission)}>
      <MemoryRouter initialEntries={[initialEntry]}>
        <SidebarItemWithPermission item={item} collapsed={collapsed} />
      </MemoryRouter>
    </AuthContext.Provider>,
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
    expect(screen.getByRole("button", { name: "LINE予約管理を展開" })).toHaveClass(
      "min-h-11",
      "min-w-11",
    );
  });

  it("折りたたみ時も親メニューbuttonにaccessible nameと44px以上の操作領域を保つ", () => {
    renderSidebarItem({
      label: "マスタ設定",
      path: "/settings/master",
      subItems: [{ label: "スタッフ", path: "/settings/master/staffs" }],
    }, true, "/settings/master/staffs");

    const parentButton = screen.getByRole("button", { name: "マスタ設定" });
    expect(parentButton).toHaveClass(
      "min-h-11",
      "min-w-11",
    );
    expect(parentButton).not.toHaveAttribute("aria-expanded");
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
    expect(screen.getByRole("button", { name: "LINE予約管理を折りたたむ" })).toHaveClass(
      "min-h-11",
      "min-w-11",
    );
    expect(screen.getByRole("link", { name: "基本設定" })).toBeInTheDocument();
  });

  it("親が HospitalSettings でも Analytics 子が view 可ならグループを表示する (R-06/R-07)", () => {
    const hasPermission = vi.fn<AuthContextValue["hasPermission"]>(
      (resource, action) =>
        resource === ResourceLstepAnalytics && action === "view",
    );
    renderSidebarItem(
      {
        label: "Lステップ連携",
        path: "/settings/integrations/lstep",
        resource: ResourceHospitalSettings,
        subItems: [
          {
            label: "連携設定",
            path: "/settings/integrations/lstep",
            resource: ResourceHospitalSettings,
          },
          {
            label: "配信監視",
            path: "/lstep/delivery-monitor",
            resource: ResourceLstepAnalytics,
          },
        ],
      },
      false,
      "/",
      hasPermission,
    );

    fireEvent.click(screen.getByRole("button", { name: "Lステップ連携を展開" }));
    expect(screen.getByRole("link", { name: "配信監視" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "連携設定" })).not.toBeInTheDocument();
  });

  it("Analytics-only では親 path へ navigate しない (expand-only shell)", () => {
    const hasPermission = vi.fn<AuthContextValue["hasPermission"]>(
      (resource, action) =>
        resource === ResourceLstepAnalytics && action === "view",
    );
    renderSidebarItem(
      {
        label: "Lステップ連携",
        path: "/settings/integrations/lstep",
        resource: ResourceHospitalSettings,
        subItems: [
          {
            label: "配信監視",
            path: "/lstep/delivery-monitor",
            resource: ResourceLstepAnalytics,
          },
        ],
      },
      false,
      "/",
      hasPermission,
    );

    // Parent click expands only; unauthorized settings path must not become current route label via Link.
    fireEvent.click(screen.getByRole("button", { name: "Lステップ連携" }));
    expect(screen.getByRole("link", { name: "配信監視" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Lステップ連携" })).not.toBeInTheDocument();
  });

  it("neither 権限では Lステップ group を出さない", () => {
    const hasPermission = vi.fn<AuthContextValue["hasPermission"]>(() => false);
    renderSidebarItem(
      {
        label: "Lステップ連携",
        path: "/settings/integrations/lstep",
        resource: ResourceHospitalSettings,
        subItems: [
          {
            label: "配信監視",
            path: "/lstep/delivery-monitor",
            resource: ResourceLstepAnalytics,
          },
        ],
      },
      false,
      "/",
      hasPermission,
    );
    expect(screen.queryByRole("button", { name: "Lステップ連携" })).not.toBeInTheDocument();
  });
});
