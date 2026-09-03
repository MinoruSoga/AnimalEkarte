import { render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ErrorBoundary } from "./ErrorBoundary";

function ThrowingChild(): never {
  throw new Error("boom");
}

describe("ErrorBoundary（FE6-4: liff / line-reserve 白紙化防止）", () => {
  let consoleErrorSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    // React はエラーバウンダリでキャッチされた例外についても console.error を出すため抑制する
    consoleErrorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
  });

  afterEach(() => {
    consoleErrorSpy.mockRestore();
  });

  it("子がthrowした場合にfallbackを表示する", () => {
    render(
      <ErrorBoundary fallback={<div>フォールバック表示</div>}>
        <ThrowingChild />
      </ErrorBoundary>,
    );

    expect(screen.getByText("フォールバック表示")).toBeInTheDocument();
  });

  it("正常時はchildrenを表示する", () => {
    render(
      <ErrorBoundary fallback={<div>フォールバック表示</div>}>
        <div>正常なコンテンツ</div>
      </ErrorBoundary>,
    );

    expect(screen.getByText("正常なコンテンツ")).toBeInTheDocument();
    expect(screen.queryByText("フォールバック表示")).not.toBeInTheDocument();
  });
});
