import { render, screen } from "@testing-library/react";
import { createMemoryRouter, RouterProvider } from "react-router";
import { afterEach, describe, expect, it, vi } from "vitest";

import { RouteErrorBoundary } from "./RouteErrorBoundary";
import { reloadForStaleChunk, shouldRecoverByReload } from "@/lib/stale-chunk-recovery";

vi.mock("@/lib/stale-chunk-recovery", () => ({
  shouldRecoverByReload: vi.fn(() => false),
  reloadForStaleChunk: vi.fn(() => true),
}));

const mockedShouldRecover = vi.mocked(shouldRecoverByReload);
const mockedReload = vi.mocked(reloadForStaleChunk);

function renderFailingRoute(error: unknown) {
  const router = createMemoryRouter(
    [
      {
        path: "/",
        errorElement: <RouteErrorBoundary />,
        children: [{ index: true, lazy: () => Promise.reject(error) }],
      },
    ],
    { initialEntries: ["/"] },
  );
  render(<RouterProvider router={router} />);
}

afterEach(() => {
  vi.clearAllMocks();
});

describe("RouteErrorBoundary stale-chunk recovery (EMR-271)", () => {
  it("チャンク読み込み失敗と判定されたらエラー画面ではなくリロードを試みる", async () => {
    mockedShouldRecover.mockReturnValue(true);
    renderFailingRoute(
      new TypeError("Failed to fetch dynamically imported module: /assets/old.js"),
    );

    expect(await screen.findByText("画面を再読み込みしています")).toBeInTheDocument();
    expect(mockedReload).toHaveBeenCalledTimes(1);
    expect(screen.queryByText("エラーが発生しました")).not.toBeInTheDocument();
    expect(screen.queryByText("ダッシュボードへ戻る")).not.toBeInTheDocument();
  });

  it("回復対象外のエラーは従来どおりエラー画面を表示しリロードしない", async () => {
    mockedShouldRecover.mockReturnValue(false);
    renderFailingRoute(new Error("boom"));

    expect(await screen.findByText("エラーが発生しました")).toBeInTheDocument();
    expect(screen.getByText("ダッシュボードへ戻る")).toBeInTheDocument();
    expect(mockedReload).not.toHaveBeenCalled();
  });

  it("連続失敗(リロード済み)も回復対象外ならエラー画面へフォールバックする", async () => {
    mockedShouldRecover.mockReturnValue(false);
    renderFailingRoute(new TypeError("Importing a module script failed."));

    expect(await screen.findByText("エラーが発生しました")).toBeInTheDocument();
    expect(mockedReload).not.toHaveBeenCalled();
  });
});
