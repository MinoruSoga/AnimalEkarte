import { renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { useFilterVaccinations } from "./use-vaccinations";

vi.mock("../api/get-vaccinations", () => ({
  useGetVaccinations: vi.fn(() => ({ data: [], isLoading: false, error: null })),
}));

import { useGetVaccinations } from "../api/get-vaccinations";

const mockRecords = [
  { ownerName: "ヤマダタロウ", petName: "ポチ", vaccineName: "狂犬病ワクチン", doctor: "田中" },
  {
    ownerName: "さとうけんじ",
    petName: "たまごろう",
    vaccineName: "フィラリア予防",
    doctor: "佐藤",
  },
];

function setup(searchTerm: string) {
  vi.mocked(useGetVaccinations).mockReturnValue({
    data: mockRecords as never,
    isLoading: false,
    error: null,
  });
  return renderHook(() => useFilterVaccinations(searchTerm));
}

describe("useFilterVaccinations かな正規化", () => {
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

  it("空文字の場合は全件返す", () => {
    const { result } = setup("");
    expect(result.current.data).toHaveLength(2);
  });
});
