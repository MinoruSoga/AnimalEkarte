import { renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { useFilterExaminationRecords } from "./use-examination-records";

vi.mock("../api/get-examinations", () => ({
  useGetExaminations: vi.fn(() => ({ data: [], isLoading: false, error: null })),
}));

import { useGetExaminations } from "../api/get-examinations";

const mockRecords = [
  { ownerName: "ヤマダタロウ", petName: "ポチ", testType: "血液検査", status: "done", doctor: "田中" },
  { ownerName: "さとうけんじ", petName: "たまごろう", testType: "尿検査", status: "done", doctor: "佐藤" },
];

function setup(searchTerm: string) {
  vi.mocked(useGetExaminations).mockReturnValue({ data: mockRecords as never, isLoading: false, error: null });
  return renderHook(() => useFilterExaminationRecords(searchTerm));
}

describe("useFilterExaminationRecords かな正規化", () => {
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

  it("空文字の場合は全件返す", () => {
    const { result } = setup("");
    expect(result.current.data).toHaveLength(2);
  });
});
