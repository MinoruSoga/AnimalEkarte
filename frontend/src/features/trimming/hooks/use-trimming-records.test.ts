import { renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { useFilterTrimmingRecords } from "./use-trimming-records";

vi.mock("../api/get-trimmings", () => ({
  useGetTrimmingsPage: vi.fn(() => ({
    data: { items: [], total: 0, isTruncated: false },
    isLoading: false,
    error: null,
  })),
}));

vi.mock("../api/delete-trimming", () => ({
  useDeleteTrimming: vi.fn(() => ({ mutate: vi.fn() })),
}));

import { useGetTrimmingsPage } from "../api/get-trimmings";

const mockRecords = [
  { ownerName: "ヤマダタロウ", petName: "ポチ", status: "active", species: "犬", staff: "田中", breed: "トイプードル" },
  { ownerName: "さとうけんじ", petName: "たまごろう", status: "active", species: "猫", staff: "佐藤", breed: "" },
];

function setup(searchTerm: string) {
  vi.mocked(useGetTrimmingsPage).mockReturnValue({
    data: { items: mockRecords as never, total: mockRecords.length, isTruncated: false },
    isLoading: false,
    error: null,
  });
  return renderHook(() => useFilterTrimmingRecords(searchTerm));
}

describe("useFilterTrimmingRecords かな正規化", () => {
  it("ひらがな入力でカタカナ ownerName がヒットする", () => {
    const { result } = setup("やまだ");
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data[0].ownerName).toBe("ヤマダタロウ");
  });

  it("カタカナ入力でひらがな ownerName がヒットする", () => {
    const { result } = setup("サトウ");
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data[0].ownerName).toBe("さとうけんじ");
  });

  it("ひらがな入力でカタカナ petName がヒットする", () => {
    const { result } = setup("ぽち");
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data[0].petName).toBe("ポチ");
  });

  it("カタカナ入力でひらがな petName がヒットする", () => {
    const { result } = setup("タマゴロウ");
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data[0].petName).toBe("たまごろう");
  });

  it("#231: 犬種でも検索できる", () => {
    const { result } = setup("トイプードル");
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data[0].ownerName).toBe("ヤマダタロウ");
  });

  it("空文字の場合は全件返す", () => {
    const { result } = setup("");
    expect(result.current.data).toHaveLength(2);
  });

  it("漢字・ASCII 検索は変換されない", () => {
    const { result } = setup("田中");
    // ownerName/petName に「田中」はないのでヒットなし
    expect(result.current.data).toHaveLength(0);
  });

  it("APIの打ち切り状態を一覧UIへ伝播する", () => {
    vi.mocked(useGetTrimmingsPage).mockReturnValue({
      data: { items: mockRecords as never, total: 250, isTruncated: true },
      isLoading: false,
      error: null,
    });

    const { result } = renderHook(() => useFilterTrimmingRecords(""));

    expect(result.current.isTruncated).toBe(true);
  });
});
