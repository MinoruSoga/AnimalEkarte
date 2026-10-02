import { describe, expect, it } from "vitest";
import { settingsRoute } from "./settings-routes";

describe("/settings/bug-reports — 全スタッフ公開", () => {
  it("権限ゲート（RequirePermission）なしの lazy ルートである", () => {
    // バグ報告は全医院共有の製品フィードバック基盤として意図的に権限フリー —
    // element ベースのガードを持たず、lazy でページを直接解決する。
    const bugReportsRoute = settingsRoute.children?.find((route) => route.path === "bug-reports");

    expect(bugReportsRoute).toBeDefined();
    expect(bugReportsRoute?.lazy).toBeTypeOf("function");
    expect(bugReportsRoute?.element).toBeUndefined();
    expect(bugReportsRoute?.children).toBeUndefined();
  });
});
