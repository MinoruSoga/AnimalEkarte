import { renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";

import { createTestWrapper } from "@/testing/TestUtils";
import type { OwnerSearchItem, PetSearchItem } from "@/types/generated/identitylink-responses";

import { useOwnerSearchQuery, usePetSearchQuery } from "./use-identity-link-search";

const searchOwnersForLink = vi.fn();
const searchPetsForLink = vi.fn();

vi.mock("../api/identity-links-api", () => ({
  searchOwnersForLink: (...args: unknown[]) => searchOwnersForLink(...args),
  searchPetsForLink: (...args: unknown[]) => searchPetsForLink(...args),
}));

const ownerHit: OwnerSearchItem = {
  clinic_id: 1,
  owner_id: 10,
  name: "佐藤太郎",
  name_kana: "サトウタロウ",
  phone: "09011112222",
};

const petHit: PetSearchItem = {
  clinic_id: 1,
  pet_id: 5,
  owner_id: 10,
  name: "ポチ",
  name_kana: "ポチ",
  pet_number: "P-001",
};

describe("useOwnerSearchQuery", () => {
  beforeEach(() => {
    searchOwnersForLink.mockReset();
    searchPetsForLink.mockReset();
  });

  it("空文字のときは検索 API を呼ばない", () => {
    const { result } = renderHook(() => useOwnerSearchQuery(""), {
      wrapper: createTestWrapper(),
    });

    expect(searchOwnersForLink).not.toHaveBeenCalled();
    expect(result.current.data).toBeUndefined();
  });

  it("1文字以上で検索 API を呼び、結果を返す", async () => {
    searchOwnersForLink.mockResolvedValue([ownerHit]);

    const { result } = renderHook(() => useOwnerSearchQuery("佐藤"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => {
      expect(result.current.data).toEqual([ownerHit]);
    });
    expect(searchOwnersForLink).toHaveBeenCalledWith("佐藤");
  });

  it("API が失敗すると isError になる", async () => {
    searchOwnersForLink.mockRejectedValue(new Error("network error"));

    const { result } = renderHook(() => useOwnerSearchQuery("佐藤"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
  });
});

describe("usePetSearchQuery", () => {
  beforeEach(() => {
    searchOwnersForLink.mockReset();
    searchPetsForLink.mockReset();
  });

  it("空文字のときは検索 API を呼ばない", () => {
    const { result } = renderHook(() => usePetSearchQuery(""), {
      wrapper: createTestWrapper(),
    });

    expect(searchPetsForLink).not.toHaveBeenCalled();
    expect(result.current.data).toBeUndefined();
  });

  it("1文字以上で検索 API を呼び、結果を返す", async () => {
    searchPetsForLink.mockResolvedValue([petHit]);

    const { result } = renderHook(() => usePetSearchQuery("ポチ"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => {
      expect(result.current.data).toEqual([petHit]);
    });
    expect(searchPetsForLink).toHaveBeenCalledWith("ポチ");
  });
});
