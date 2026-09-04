import { describe, expect, it, vi } from "vitest";
import { renderHook } from "@testing-library/react";
import { useUrlPageSync } from "./use-url-page-sync";

describe("useUrlPageSync", () => {
  it("clamps urlPage above totalPages when not loading", () => {
    const setSearchParams = vi.fn();
    renderHook(() =>
      useUrlPageSync({
        urlPage: 9,
        totalPages: 3,
        isLoading: false,
        setSearchParams,
      }),
    );
    expect(setSearchParams).toHaveBeenCalledTimes(1);
  });

  it("does not update while loading", () => {
    const setSearchParams = vi.fn();
    renderHook(() =>
      useUrlPageSync({
        urlPage: 9,
        totalPages: 3,
        isLoading: true,
        setSearchParams,
      }),
    );
    expect(setSearchParams).not.toHaveBeenCalled();
  });

  it("does not update when page is already in range", () => {
    const setSearchParams = vi.fn();
    renderHook(() =>
      useUrlPageSync({
        urlPage: 2,
        totalPages: 5,
        isLoading: false,
        setSearchParams,
      }),
    );
    expect(setSearchParams).not.toHaveBeenCalled();
  });
});
