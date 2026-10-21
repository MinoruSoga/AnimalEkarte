// EMR-463: /owners 一覧 index の入口判定。活性時は loader の listState が判定するため
// guard は子をそのまま描画し、非活性時（テスト専用の退避経路）は従来どおり
// RequirePermission(owners, view) を経由することを固定する。
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { OwnersListRouteGuard } from "./OwnersListRouteGuard";

const guardState = vi.hoisted(() => ({ active: true, canView: false }));

vi.mock("../lib/owners-clinic-scope-activation", () => ({
  get OWNERS_CLINIC_SCOPE_ACTIVE() {
    return guardState.active;
  },
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    hasPermission: () => guardState.canView,
  }),
}));

afterEach(() => {
  cleanup();
  guardState.active = true;
  guardState.canView = false;
});

describe("OwnersListRouteGuard", () => {
  it("活性時は owners:view なしでも子を描画する（入口判定は loader の listState が担う）", () => {
    guardState.active = true;
    guardState.canView = false;

    render(
      <OwnersListRouteGuard>
        <div>飼主一覧</div>
      </OwnersListRouteGuard>,
    );

    expect(screen.getByText("飼主一覧")).toBeInTheDocument();
    expect(screen.queryByText("アクセス権限がありません")).not.toBeInTheDocument();
  });

  it("非活性時は従来どおり owners:view を要求し、view なしは AccessDenied", () => {
    guardState.active = false;
    guardState.canView = false;

    render(
      <OwnersListRouteGuard>
        <div>飼主一覧</div>
      </OwnersListRouteGuard>,
    );

    expect(screen.getByText("アクセス権限がありません")).toBeInTheDocument();
    expect(screen.queryByText("飼主一覧")).not.toBeInTheDocument();
  });

  it("非活性時でも owners:view があれば子を描画する", () => {
    guardState.active = false;
    guardState.canView = true;

    render(
      <OwnersListRouteGuard>
        <div>飼主一覧</div>
      </OwnersListRouteGuard>,
    );

    expect(screen.getByText("飼主一覧")).toBeInTheDocument();
    expect(screen.queryByText("アクセス権限がありません")).not.toBeInTheDocument();
  });
});
