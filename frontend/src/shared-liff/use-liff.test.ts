import { renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";

// R-F21: liff/ と line-reserve/ の重複実装を統合した共有フックの単体テスト。
// LIFF_MOCK はモジュール評価時に VITE_LIFF_MOCK を判定するため、
// vi.stubEnv + vi.resetModules + 動的 import でモジュール評価前に環境変数を確定させる。
const liffInit = vi.fn();
const liffIsLoggedIn = vi.fn();
const liffLogin = vi.fn();
const liffGetIDToken = vi.fn();
const liffGetProfile = vi.fn();

vi.mock("@line/liff", () => ({
  default: {
    init: (...args: unknown[]) => liffInit(...args),
    isLoggedIn: () => liffIsLoggedIn(),
    login: () => liffLogin(),
    getIDToken: () => liffGetIDToken(),
    getProfile: () => liffGetProfile(),
  },
}));

describe("useLiff（共有 LIFF 初期化フック）", () => {
  beforeEach(() => {
    vi.resetModules();
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("LIFF_MOCK=true のとき、liff.init を呼ばずにモック値で即座に isReady になる", async () => {
    vi.stubEnv("VITE_LIFF_MOCK", "true");
    const { useLiff } = await import("./use-liff");

    const { result } = renderHook(() => useLiff("liff-id"));

    expect(result.current.isReady).toBe(true);
    expect(result.current.idToken).toBe("mock-token");
    expect(result.current.displayName).toBe("テストユーザー");
    expect(result.current.initError).toBe(false);
    expect(liffInit).not.toHaveBeenCalled();
  });

  it("liff.init が成功しログイン済みのとき、idToken/displayName/pictureUrl を反映する", async () => {
    vi.stubEnv("VITE_LIFF_MOCK", "false");
    liffInit.mockResolvedValue(undefined);
    liffIsLoggedIn.mockReturnValue(true);
    liffGetIDToken.mockReturnValue("real-token");
    liffGetProfile.mockResolvedValue({
      displayName: "山田太郎",
      pictureUrl: "https://example.com/p.png",
    });

    const { useLiff } = await import("./use-liff");
    const { result } = renderHook(() => useLiff("liff-id"));

    await waitFor(() => {
      expect(result.current.isReady).toBe(true);
    });

    expect(result.current.idToken).toBe("real-token");
    expect(result.current.displayName).toBe("山田太郎");
    expect(result.current.pictureUrl).toBe("https://example.com/p.png");
    expect(result.current.initError).toBe(false);
    expect(liffLogin).not.toHaveBeenCalled();
  });

  it("liff.init が成功し未ログインのとき、liff.login を呼ぶ", async () => {
    vi.stubEnv("VITE_LIFF_MOCK", "false");
    liffInit.mockResolvedValue(undefined);
    liffIsLoggedIn.mockReturnValue(false);

    const { useLiff } = await import("./use-liff");
    const { result } = renderHook(() => useLiff("liff-id"));

    await waitFor(() => {
      expect(liffLogin).toHaveBeenCalledTimes(1);
    });

    expect(result.current.isReady).toBe(false);
  });

  it("liff.init が失敗したとき、console.error でログを残し initError=true / isReady=true になる", async () => {
    vi.stubEnv("VITE_LIFF_MOCK", "false");
    const consoleErrorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    const initFailure = new Error("liff init failed");
    liffInit.mockRejectedValue(initFailure);

    const { useLiff } = await import("./use-liff");
    const { result } = renderHook(() => useLiff("liff-id"));

    await waitFor(() => {
      expect(result.current.initError).toBe(true);
    });

    expect(result.current.isReady).toBe(true);
    expect(consoleErrorSpy).toHaveBeenCalledWith("[useLiff] init failed:", initFailure);

    consoleErrorSpy.mockRestore();
  });
});
