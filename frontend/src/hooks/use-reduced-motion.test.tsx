import { act, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { useReducedMotion } from "./use-reduced-motion";

function ReducedMotionValue({ label }: { label: string }) {
  const reduced = useReducedMotion();
  return (
    <span>
      {label}:{reduced ? "reduce" : "normal"}
    </span>
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("useReducedMotion", () => {
  it("複数consumerでMediaQuery listenerを1本だけ共有し最後にcleanupする", () => {
    let listener: ((event: MediaQueryListEvent) => void) | undefined;
    const addEventListener = vi.fn(
      (_type: string, callback: (event: MediaQueryListEvent) => void) => {
        listener = callback;
      },
    );
    const removeEventListener = vi.fn();
    const matchMedia = vi.fn(() => ({
      matches: false,
      media: "(prefers-reduced-motion: reduce)",
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener,
      removeEventListener,
      dispatchEvent: vi.fn(),
    }));
    vi.stubGlobal("matchMedia", matchMedia);

    const { unmount } = render(
      <>
        <ReducedMotionValue label="first" />
        <ReducedMotionValue label="second" />
      </>,
    );

    expect(matchMedia).toHaveBeenCalledTimes(1);
    expect(addEventListener).toHaveBeenCalledTimes(1);

    act(() => {
      listener?.(
        Object.assign(new Event("change"), {
          matches: true,
          media: "(prefers-reduced-motion: reduce)",
        }),
      );
    });
    expect(screen.getByText("first:reduce")).toBeInTheDocument();
    expect(screen.getByText("second:reduce")).toBeInTheDocument();

    unmount();
    expect(removeEventListener).toHaveBeenCalledTimes(1);
  });
});
