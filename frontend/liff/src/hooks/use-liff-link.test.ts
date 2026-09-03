import { StrictMode } from "react";
import { renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";

// リンク導線のスモークテスト（FE-refactor.md R-F4: liffは規模が小さく健全なため smoke 1本）。
// useLiff（frontend/src/shared-liff/use-liff.ts）自体は use-liff.test.ts で個別に検証済みのため、
// ここでは useLiffLink 自身のロジック（isReady/idToken/initError の検証 → token/clinic_id の
// 読み取り → リンク処理）に集中するため useLiff をモックして isReady の遷移を直接制御する。
//
// SD-14: useLiffLink() は token/clinic_id を引数ではなく window.location.search から
// isReady（liff.init() 完了）後に読み取る。LIFF はログインリダイレクトを経由する場合、
// 元のクエリを liff.state に包んで戻し、liff.init() 完了までに history.replaceState で
// 元の URL（?token=...&clinic_id=...）へ復元する。isReady より前に読むと liff.state に
// 包まれたままの URL を掴んでしまい欠落する — この race を防ぐタイミング仕様を検証する。

const linkLineAccount = vi.fn();
vi.mock("../api/liff-api", async () => {
  const actual = await vi.importActual<typeof import("../api/liff-api")>("../api/liff-api");
  return {
    ...actual,
    linkLineAccount: (...args: unknown[]) => linkLineAccount(...args),
  };
});

const useLiffMock = vi.fn();
vi.mock("@/shared-liff/use-liff", () => ({
  useLiff: (...args: unknown[]) => useLiffMock(...args),
}));

function setLocationSearch(search: string) {
  window.history.pushState({}, "", `/${search}`);
}

describe("useLiffLink（R-F4 smoke: LINEアカウント連携導線）", () => {
  beforeEach(() => {
    vi.resetModules();
    vi.clearAllMocks();
    vi.stubEnv("VITE_LIFF_MOCK", "false");
    setLocationSearch("");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    setLocationSearch("");
  });

  it("isReady=true でも clinicId/linkToken が空のとき、無効なURLエラーになる", async () => {
    useLiffMock.mockReturnValue({ idToken: "real-id-token", isReady: true, initError: false });

    const { useLiffLink } = await import("./use-liff-link");
    const { result } = renderHook(() => useLiffLink());

    await waitFor(() => {
      expect(result.current.status).toBe("error");
    });
    expect(result.current.errorMessage).toBe("無効なURLです。QRコードを再度読み取ってください");
    expect(linkLineAccount).not.toHaveBeenCalled();
  });

  it("clinicId/linkToken が URL に揃っているとき、linkLineAccount を呼び success になる", async () => {
    setLocationSearch("?token=link-token-abc&clinic_id=1");
    linkLineAccount.mockResolvedValue(undefined);
    useLiffMock.mockReturnValue({ idToken: "real-id-token", isReady: true, initError: false });

    const { useLiffLink } = await import("./use-liff-link");
    const { result } = renderHook(() => useLiffLink());

    await waitFor(() => {
      expect(result.current.status).toBe("success");
    });
    expect(linkLineAccount).toHaveBeenCalledWith("1", "link-token-abc", "real-id-token");
  });

  it("StrictMode でも linkLineAccount は1回だけ呼ばれる", async () => {
    setLocationSearch("?token=link-token-abc&clinic_id=1");
    linkLineAccount.mockResolvedValue(undefined);
    useLiffMock.mockReturnValue({ idToken: "real-id-token", isReady: true, initError: false });

    const { useLiffLink } = await import("./use-liff-link");
    const { result } = renderHook(() => useLiffLink(), { wrapper: StrictMode });

    await waitFor(() => {
      expect(result.current.status).toBe("success");
    });
    expect(linkLineAccount).toHaveBeenCalledTimes(1);
  });
});

describe("useLiffLink — liff.state 経由のログインリダイレクト後もクエリを取得できる（SD-14）", () => {
  beforeEach(() => {
    vi.resetModules();
    vi.clearAllMocks();
    vi.stubEnv("VITE_LIFF_MOCK", "false");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    setLocationSearch("");
  });

  it("isReady=false の間は URL が liff.state に包まれていても連携処理を開始せず、isReady=true 時点の URL から正しく読み取る", async () => {
    // マウント時点は LINE ログインリダイレクト直後を模し、token/clinic_id が
    // liff.state に包まれたまま URL に直接現れていない状態にする。
    setLocationSearch(
      "?liff.state=%2F%3Ftoken%3Dlink-token-abc%26clinic_id%3D1&code=xxx&state=yyy",
    );
    linkLineAccount.mockResolvedValue(undefined);
    useLiffMock.mockReturnValue({ idToken: null, isReady: false, initError: false });

    const { useLiffLink } = await import("./use-liff-link");
    const { result, rerender } = renderHook(() => useLiffLink());

    // isReady=false の間は loading のまま — この時点で window.location.search を
    // 読んでいたら token/clinic_id は liff.state に包まれたままで欠落するはず。
    expect(result.current.status).toBe("loading");
    expect(linkLineAccount).not.toHaveBeenCalled();

    // 実 LIFF SDK が liff.init() 完了までに history.replaceState で行う
    // クエリ復元、および isReady=true への遷移を模する。
    setLocationSearch("?token=link-token-abc&clinic_id=1");
    useLiffMock.mockReturnValue({ idToken: "real-id-token", isReady: true, initError: false });
    rerender();

    await waitFor(() => {
      expect(linkLineAccount).toHaveBeenCalledWith("1", "link-token-abc", "real-id-token");
    });
    expect(result.current.status).not.toBe("error");
  });
});
