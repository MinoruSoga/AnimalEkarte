import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";

import { ExaminationHistorySection } from "./ExaminationHistorySection";

// usePermission と取得フックをモックして、検査項目テーブルの描画を決定的に検証する。
const hooks = vi.hoisted(() => ({
  usePermission: vi.fn(),
  useGetPetExaminations: vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({ usePermission: hooks.usePermission }));
vi.mock("../api/get-pet-examinations", () => ({
  useGetPetExaminations: hooks.useGetPetExaminations,
}));

function ok<T>(data: T) {
  return { data, isLoading: false, isError: false };
}

const examWithItems = {
  id: "e1",
  testType: "血液検査",
  date: "2026-05-10",
  status: "確定",
  resultSummary: "",
  items: [
    {
      id: "i1",
      name: "WBC",
      inspectionValue: "120",
      result: "",
      unit: "U/L",
      referenceValue: "60-120",
      status: "high",
    },
    {
      id: "i2",
      name: "RBC",
      inspectionValue: "700",
      result: "",
      unit: "万/μL",
      referenceValue: "550-850",
      status: "normal",
    },
  ],
};

beforeEach(() => {
  vi.clearAllMocks();
  hooks.usePermission.mockReturnValue({ canView: true });
  hooks.useGetPetExaminations.mockReturnValue(ok([examWithItems]));
});

describe("ExaminationHistorySection", () => {
  it("検査項目テーブルに列見出し(項目/結果/基準値)を表示する", () => {
    render(<ExaminationHistorySection petId="1" />);

    expect(screen.getByRole("columnheader", { name: "項目" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "結果" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "基準値" })).toBeInTheDocument();
  });

  it("列見出しの下に項目名・結果(単位込み)・基準値が対応して並ぶ", () => {
    render(<ExaminationHistorySection petId="1" />);

    const wbcRow = screen.getByText("WBC").closest("tr");
    expect(wbcRow).not.toBeNull();
    expect(within(wbcRow as HTMLElement).getByText("120 U/L")).toBeInTheDocument();
    expect(within(wbcRow as HTMLElement).getByText("60-120")).toBeInTheDocument();

    // 検査種別・日付などのカードヘッダーは従来どおり表示される。
    expect(screen.getByText("血液検査")).toBeInTheDocument();
  });

  it("検査項目が無い検査は『検査項目なし』を表示し列見出しを出さない", () => {
    hooks.useGetPetExaminations.mockReturnValue(
      ok([{ id: "e2", testType: "尿検査", date: "2026-05-11", status: "確定", resultSummary: "", items: [] }]),
    );
    render(<ExaminationHistorySection petId="1" />);

    expect(screen.getByText("検査項目なし")).toBeInTheDocument();
    expect(screen.queryByRole("columnheader")).not.toBeInTheDocument();
  });

  it("閲覧権限が無ければセクションを縮退表示する", () => {
    hooks.usePermission.mockReturnValue({ canView: false });
    hooks.useGetPetExaminations.mockReturnValue({ data: undefined, isLoading: false, isError: false });
    render(<ExaminationHistorySection petId="1" />);

    expect(screen.getByText("閲覧権限がありません")).toBeInTheDocument();
    expect(screen.queryByRole("columnheader")).not.toBeInTheDocument();
  });
});
