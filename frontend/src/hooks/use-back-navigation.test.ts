import { renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useBackNavigation, useBackPath } from "./use-back-navigation";

const navigate = vi.fn();
let locationState: unknown = null;

vi.mock("react-router", () => ({
  useNavigate: () => navigate,
  useLocation: () => ({ state: locationState }),
}));

describe("useBackNavigation", () => {
  beforeEach(() => {
    navigate.mockClear();
    locationState = null;
  });

  it("state.from の内部パスへ戻る", () => {
    locationState = { from: "/?date=2026-06-05" };
    const { result } = renderHook(() => useBackNavigation("/owners"));

    result.current();

    expect(navigate).toHaveBeenCalledWith("/?date=2026-06-05");
  });

  it("state が無ければ fallbackPath へ戻る", () => {
    const { result } = renderHook(() => useBackNavigation("/owners"));

    result.current();

    expect(navigate).toHaveBeenCalledWith("/owners");
  });

  it("from が無ければ fallbackPath へ戻る", () => {
    locationState = { appointmentId: "88" };
    const { result } = renderHook(() => useBackNavigation("/owners"));

    result.current();

    expect(navigate).toHaveBeenCalledWith("/owners");
  });

  it("from が外部 URL なら fallbackPath へ戻る（open redirect 防止）", () => {
    locationState = { from: "https://evil.example/path" };
    const { result } = renderHook(() => useBackNavigation("/owners"));

    result.current();

    expect(navigate).toHaveBeenCalledWith("/owners");
  });

  it("from が // 始まりなら fallbackPath へ戻る（protocol-relative 拒否）", () => {
    locationState = { from: "//evil.example" };
    const { result } = renderHook(() => useBackNavigation("/owners"));

    result.current();

    expect(navigate).toHaveBeenCalledWith("/owners");
  });

  it("from が文字列でなければ fallbackPath へ戻る", () => {
    locationState = { from: 42 };
    const { result } = renderHook(() => useBackNavigation("/owners"));

    result.current();

    expect(navigate).toHaveBeenCalledWith("/owners");
  });
});

describe("useBackPath", () => {
  beforeEach(() => {
    navigate.mockClear();
    locationState = null;
  });

  it("state.from の内部パスを返す（保存後遷移用）", () => {
    locationState = { from: "/?date=2026-06-05" };
    const { result } = renderHook(() => useBackPath("/owners"));

    expect(result.current).toBe("/?date=2026-06-05");
    expect(navigate).not.toHaveBeenCalled();
  });

  it("from が無効なら fallbackPath を返す", () => {
    locationState = { from: "javascript:alert(1)" };
    const { result } = renderHook(() => useBackPath("/owners"));

    expect(result.current).toBe("/owners");
  });
});
