import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";

import { CheckupHistorySection } from "./CheckupHistorySection";

// usePermission と取得フックをモックして、checkup 結果のグルーピング描画を決定的に検証する。
const hooks = vi.hoisted(() => ({
  usePermission: vi.fn(),
  useGetPetCheckupResults: vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({ usePermission: hooks.usePermission }));
vi.mock("@/hooks/use-pet-checkup-results", () => ({
  useGetPetCheckupResults: hooks.useGetPetCheckupResults,
}));

function ok<T>(data: T) {
  return { data, isLoading: false, isError: false };
}

const dentalResults = [
  {
    id: 1, checkupId: 100, date: "2026-05-10", checkupTypeName: "歯科検診",
    fieldName: "歯石除去必要の有無", fieldType: "boolean", unit: "",
    valueNumber: undefined, valueText: "", valueBool: true, valueList: [], isAbnormal: false, status: "normal",
  },
  {
    id: 2, checkupId: 100, date: "2026-05-10", checkupTypeName: "歯科検診",
    fieldName: "歯石付着度スコア", fieldType: "number", unit: "点",
    valueNumber: 5, valueText: "", valueBool: undefined, valueList: [], isAbnormal: true, status: "high",
  },
  {
    id: 3, checkupId: 100, date: "2026-05-10", checkupTypeName: "歯科検診",
    fieldName: "歯科ケアアドバイス", fieldType: "multi_select", unit: "",
    valueNumber: undefined, valueText: "", valueBool: undefined, valueList: ["毎日の歯磨き", "定期スケーリング"], isAbnormal: false, status: "normal",
  },
];

beforeEach(() => {
  vi.clearAllMocks();
  hooks.usePermission.mockReturnValue({ canView: true });
  hooks.useGetPetCheckupResults.mockReturnValue(ok(dentalResults));
});

describe("CheckupHistorySection", () => {
  it("checkup 単位でパッケージ名・日付・フィールド結果を表示する", () => {
    render(<CheckupHistorySection petId="1" />);

    expect(screen.getByText("歯科検診")).toBeInTheDocument();
    expect(screen.getByText("2026-05-10")).toBeInTheDocument();

    // boolean → はい
    const boolRow = screen.getByText("歯石除去必要の有無").closest("tr");
    expect(within(boolRow as HTMLElement).getByText("はい")).toBeInTheDocument();

    // number → 値 + 単位
    const numRow = screen.getByText("歯石付着度スコア").closest("tr");
    expect(within(numRow as HTMLElement).getByText("5 点")).toBeInTheDocument();

    // multi_select → 選択肢を「、」で連結
    const multiRow = screen.getByText("歯科ケアアドバイス").closest("tr");
    expect(within(multiRow as HTMLElement).getByText("毎日の歯磨き、定期スケーリング")).toBeInTheDocument();
  });

  it("結果が無ければ空表示に縮退する", () => {
    hooks.useGetPetCheckupResults.mockReturnValue(ok([]));
    render(<CheckupHistorySection petId="1" />);

    expect(screen.getByText("健診結果の履歴はありません")).toBeInTheDocument();
    expect(screen.queryByRole("columnheader")).not.toBeInTheDocument();
  });

  it("閲覧権限が無ければセクションを縮退表示する", () => {
    hooks.usePermission.mockReturnValue({ canView: false });
    hooks.useGetPetCheckupResults.mockReturnValue({ data: undefined, isLoading: false, isError: false });
    render(<CheckupHistorySection petId="1" />);

    expect(screen.getByText("閲覧権限がありません")).toBeInTheDocument();
    expect(screen.queryByRole("columnheader")).not.toBeInTheDocument();
  });
});
