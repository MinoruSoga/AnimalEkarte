import { render, screen } from "@testing-library/react";
import { StrictMode, type ReactNode } from "react";
import { createMemoryRouter, RouterProvider } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { RootErrorBoundary, RouteErrorBoundary } from "./RouteErrorBoundary";
import { reloadForStaleChunk, shouldRecoverByReload } from "@/lib/stale-chunk-recovery";

vi.mock("@/lib/stale-chunk-recovery", () => ({
  shouldRecoverByReload: vi.fn(() => false),
  reloadForStaleChunk: vi.fn(() => true),
}));

const mockedShouldRecover = vi.mocked(shouldRecoverByReload);
const mockedReload = vi.mocked(reloadForStaleChunk);

function renderFailingRoute(error: unknown, boundary: ReactNode = <RouteErrorBoundary />) {
  const router = createMemoryRouter(
    [
      {
        path: "/",
        errorElement: boundary,
        children: [{ index: true, lazy: () => Promise.reject(error) }],
      },
    ],
    { initialEntries: ["/"] },
  );
  render(
    <StrictMode>
      <RouterProvider router={router} />
    </StrictMode>,
  );
}

beforeEach(() => {
  mockedShouldRecover.mockReset().mockReturnValue(false);
  mockedReload.mockReset().mockReturnValue(true);
});

afterEach(() => {
  vi.restoreAllMocks();
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

  it.each([
    ["route", RouteErrorBoundary, "link", "ダッシュボードへ戻る"],
    ["root", RootErrorBoundary, "button", "再読み込み"],
  ] as const)(
    "%s: マーカー書込み失敗時は通常の復帰操作へ戻る",
    async (_name, Boundary, role, label) => {
      const actual = await vi.importActual<typeof import("@/lib/stale-chunk-recovery")>(
        "@/lib/stale-chunk-recovery",
      );
      sessionStorage.clear();
      vi.spyOn(navigator, "onLine", "get").mockReturnValue(true);
      vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
        throw new DOMException("quota exceeded", "QuotaExceededError");
      });
      mockedShouldRecover.mockImplementation(actual.shouldRecoverByReload);
      mockedReload.mockImplementation(actual.reloadForStaleChunk);

      renderFailingRoute(new TypeError("Importing a module script failed."), <Boundary />);

      expect(await screen.findByRole(role, { name: label })).toBeInTheDocument();
      expect(screen.queryByText("画面を再読み込みしています")).not.toBeInTheDocument();
      expect(mockedReload).toHaveBeenCalledTimes(1);
    },
  );
});

describe("stale-chunk-recovery 実装統合 (EMR-271)", () => {
  const NOW = 1_760_000_000_000;
  const TARGET_PATH = "/medical-records/new";
  const TARGET_SEARCH = "?petId=10&appointmentId=101&visitDate=2026-05-29";
  const FULL_HREF = `https://karte.example.test${TARGET_PATH}${TARGET_SEARCH}`;
  const RELOAD_MARKER_KEY = "emr:stale-chunk-reload-at";

  const CHUNK_ERRORS = [
    [
      "Chrome",
      () =>
        new TypeError(
          "Failed to fetch dynamically imported module: https://karte.example.test/assets/medical-records-OLD.js",
        ),
    ],
    [
      "Firefox",
      () =>
        new TypeError(
          "error loading dynamically imported module: https://karte.example.test/assets/x.js",
        ),
    ],
    ["Safari", () => new TypeError("Importing a module script failed.")],
    [
      "MIME",
      () =>
        new TypeError(
          'Failed to load module script: Expected a JavaScript-or-Wasm module script but the server responded with a MIME type of "text/html".',
        ),
    ],
  ] as const;

  const BOUNDARIES = [
    ["route", RouteErrorBoundary, "link", "ダッシュボードへ戻る"],
    ["root", RootErrorBoundary, "button", "再読み込み"],
  ] as const;

  let reloadMock: ReturnType<typeof vi.fn>;
  let assignMock: ReturnType<typeof vi.fn>;
  let replaceMock: ReturnType<typeof vi.fn>;

  function stubBrowserLocation() {
    reloadMock = vi.fn();
    assignMock = vi.fn();
    replaceMock = vi.fn();
    const location = {
      href: FULL_HREF,
      reload: reloadMock,
      assign: assignMock,
      replace: replaceMock,
    };
    vi.stubGlobal(
      "window",
      new Proxy(window, {
        get(target, prop) {
          if (prop === "location") return location;
          const value = Reflect.get(target, prop);
          return typeof value === "function" ? value.bind(target) : value;
        },
      }),
    );
  }

  function renderLazyChartRejection(error: unknown, boundary: ReactNode) {
    const router = createMemoryRouter(
      [
        {
          path: TARGET_PATH,
          errorElement: boundary,
          children: [{ index: true, lazy: () => Promise.reject(error) }],
        },
      ],
      { initialEntries: [`${TARGET_PATH}${TARGET_SEARCH}`] },
    );
    const rendered = render(
      <StrictMode>
        <RouterProvider router={router} />
      </StrictMode>,
    );
    return { router, ...rendered };
  }

  beforeEach(async () => {
    const actual = await vi.importActual<typeof import("@/lib/stale-chunk-recovery")>(
      "@/lib/stale-chunk-recovery",
    );
    sessionStorage.clear();
    vi.spyOn(Date, "now").mockReturnValue(NOW);
    stubBrowserLocation();
    mockedShouldRecover.mockImplementation(actual.shouldRecoverByReload);
    mockedReload.mockImplementation(actual.reloadForStaleChunk);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  describe.each(BOUNDARIES)("%s boundary", (_name, Boundary, role, label) => {
    it.each(CHUNK_ERRORS)(
      "%s の動的 import 失敗は同じルートのまま 1 度だけフルリロードし、マーカー保持中の再失敗はエラー画面へ落ちる",
      async (_signature, makeError) => {
        const first = renderLazyChartRejection(makeError(), <Boundary />);

        expect(await screen.findByText("画面を再読み込みしています")).toBeInTheDocument();
        expect(reloadMock).toHaveBeenCalledTimes(1);
        expect(sessionStorage.getItem(RELOAD_MARKER_KEY)).toBe(String(NOW));
        expect(assignMock).not.toHaveBeenCalled();
        expect(replaceMock).not.toHaveBeenCalled();
        expect(first.router.state.location.pathname).toBe(TARGET_PATH);
        expect(first.router.state.location.search).toBe(TARGET_SEARCH);
        expect(window.location.href).toBe(FULL_HREF);
        expect(screen.queryByText("エラーが発生しました")).not.toBeInTheDocument();

        first.unmount();

        const second = renderLazyChartRejection(makeError(), <Boundary />);

        expect(
          await screen.findByRole("heading", { name: "エラーが発生しました" }),
        ).toBeInTheDocument();
        expect(screen.getByRole(role, { name: label })).toBeInTheDocument();
        expect(reloadMock).toHaveBeenCalledTimes(1);
        expect(second.router.state.location.pathname).toBe(TARGET_PATH);
        expect(second.router.state.location.search).toBe(TARGET_SEARCH);
        expect(window.location.href).toBe(FULL_HREF);
        expect(screen.queryByText("画面を再読み込みしています")).not.toBeInTheDocument();
      },
    );

    it("オフライン時は実装上リロードせずエラー復帰導線を出す", async () => {
      vi.spyOn(navigator, "onLine", "get").mockReturnValue(false);
      renderLazyChartRejection(CHUNK_ERRORS[0][1](), <Boundary />);

      expect(
        await screen.findByRole("heading", { name: "エラーが発生しました" }),
      ).toBeInTheDocument();
      expect(screen.getByRole(role, { name: label })).toBeInTheDocument();
      expect(reloadMock).not.toHaveBeenCalled();
      expect(screen.queryByText("画面を再読み込みしています")).not.toBeInTheDocument();
    });

    it("sessionStorage が読めない環境では実装上リロードせずエラー復帰導線を出す", async () => {
      vi.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
        throw new DOMException("denied");
      });
      renderLazyChartRejection(CHUNK_ERRORS[0][1](), <Boundary />);

      expect(
        await screen.findByRole("heading", { name: "エラーが発生しました" }),
      ).toBeInTheDocument();
      expect(screen.getByRole(role, { name: label })).toBeInTheDocument();
      expect(reloadMock).not.toHaveBeenCalled();
      expect(screen.queryByText("画面を再読み込みしています")).not.toBeInTheDocument();
    });

    it("lazy 失敗以外の描画例外は通常のエラー画面で、リロードもマーカー書込みもしない", async () => {
      const router = createMemoryRouter(
        [
          {
            path: TARGET_PATH,
            errorElement: <Boundary />,
            children: [
              {
                index: true,
                lazy: () =>
                  Promise.resolve({
                    Component: () => {
                      throw new Error("unrelated render boom");
                    },
                  }),
              },
            ],
          },
        ],
        { initialEntries: [`${TARGET_PATH}${TARGET_SEARCH}`] },
      );
      render(
        <StrictMode>
          <RouterProvider router={router} />
        </StrictMode>,
      );

      expect(
        await screen.findByRole("heading", { name: "エラーが発生しました" }),
      ).toBeInTheDocument();
      expect(screen.getByRole(role, { name: label })).toBeInTheDocument();
      expect(reloadMock).not.toHaveBeenCalled();
      expect(sessionStorage.getItem(RELOAD_MARKER_KEY)).toBeNull();
      expect(screen.queryByText("画面を再読み込みしています")).not.toBeInTheDocument();
    });
  });
});
