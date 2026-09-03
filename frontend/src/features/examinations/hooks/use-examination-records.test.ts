import { renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { useFilterExaminationRecords } from "./use-examination-records";

vi.mock("@/hooks/use-examinations", () => ({
  useGetExaminationsPage: vi.fn(() => ({
    data: { data: [], total: 0, page: 1, limit: 20 },
    isLoading: false,
    error: null,
  })),
}));

import { useGetExaminationsPage } from "@/hooks/use-examinations";

const mockRecords = [
  {
    ownerName: "ヤマダタロウ",
    petName: "ポチ",
    testType: "血液検査",
    status: "done",
    doctor: "田中",
  },
  {
    ownerName: "さとうけんじ",
    petName: "たまごろう",
    testType: "尿検査",
    status: "done",
    doctor: "佐藤",
  },
];

function setup(searchTerm: string, page = 1, limit = 20) {
  vi.mocked(useGetExaminationsPage).mockReturnValue({
    data: { data: mockRecords as never, total: mockRecords.length, page, limit },
    isLoading: false,
    error: null,
  } as never);
  return renderHook(() => useFilterExaminationRecords(searchTerm, { page, limit }));
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

// BUG-411: サーバサイドページネーション配線の回帰テスト。
describe("useFilterExaminationRecords BUG-411 サーバページネーション配線", () => {
  it("page/limit を useGetExaminationsPage にそのまま転送する", () => {
    setup("", 3, 20);
    expect(useGetExaminationsPage).toHaveBeenCalledWith(
      expect.objectContaining({ page: 3, limit: 20 }),
    );
  });

  it("サーバの total/page/limit をそのまま返す（クライアント側で再計算・再スライスしない）", () => {
    vi.mocked(useGetExaminationsPage).mockReturnValue({
      data: { data: mockRecords as never, total: 14533, page: 5, limit: 20 },
      isLoading: false,
      error: null,
    } as never);
    const { result } = renderHook(() => useFilterExaminationRecords("", { page: 5, limit: 20 }));
    expect(result.current.total).toBe(14533);
    expect(result.current.page).toBe(5);
  });

  it("反証: 20件しか無いように見えても total は真の総件数を示す（旧偽ページネーションの再発検知）", () => {
    // 旧実装は total を一切見ず data.length（<=20）を全件数とみなしていた。
    // 本テストは total フィールドが data.length と独立に伝播することを固定する。
    vi.mocked(useGetExaminationsPage).mockReturnValue({
      data: { data: mockRecords as never, total: 14533, page: 1, limit: 20 },
      isLoading: false,
      error: null,
    } as never);
    const { result } = renderHook(() => useFilterExaminationRecords("", { page: 1, limit: 20 }));
    expect(result.current.data.length).toBeLessThanOrEqual(20);
    expect(result.current.total).toBe(14533);
    expect(result.current.total).not.toBe(result.current.data.length);
  });
});
