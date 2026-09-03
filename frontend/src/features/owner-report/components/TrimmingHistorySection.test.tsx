import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";

import { TrimmingHistorySection } from "./TrimmingHistorySection";

// usePermission と取得フックをモックして、トリミング履歴テーブルの描画を決定的に検証する。
const hooks = vi.hoisted(() => ({
  usePermission: vi.fn(),
  useGetPetTrimmingHistory: vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({ usePermission: hooks.usePermission }));
vi.mock("../api/get-pet-trimming-history", () => ({
  useGetPetTrimmingHistory: hooks.useGetPetTrimmingHistory,
}));

function ok<T>(data: T) {
  return { data, isLoading: false, isError: false };
}

// 実施履歴セクションは取得フックが「完了」のみを返す（フィルタは api 層 selectCompletedTrimmingHistory）。
// コンポーネントは渡された行を描画するだけなので、テストデータも完了のみで構成する。
const trimmings = [
  { id: "t1", date: "2026-02-01", status: "完了", courseName: "シャンプー＆カット", staff: "鈴木" },
  { id: "t2", date: "2026-03-10", status: "完了", courseName: "シャンプー", staff: "" },
];

beforeEach(() => {
  vi.clearAllMocks();
  hooks.usePermission.mockReturnValue({ canView: true });
  hooks.useGetPetTrimmingHistory.mockReturnValue(ok({ items: trimmings, isTruncated: false }));
});

describe("TrimmingHistorySection", () => {
  it("列見出し(実施日/状態/コース/担当)を表示する", () => {
    render(<TrimmingHistorySection petId="7" />);

    expect(screen.getByRole("columnheader", { name: "実施日" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "状態" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "コース" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "担当" })).toBeInTheDocument();
  });

  it("各行に実施日・状態・コース・担当が対応して並ぶ", () => {
    render(<TrimmingHistorySection petId="7" />);

    const doneRow = screen.getByText("シャンプー＆カット").closest("tr");
    expect(doneRow).not.toBeNull();
    expect(within(doneRow as HTMLElement).getByText("2026-02-01")).toBeInTheDocument();
    expect(within(doneRow as HTMLElement).getByText("完了")).toBeInTheDocument();
    expect(within(doneRow as HTMLElement).getByText("鈴木")).toBeInTheDocument();
  });

  it("担当が未設定の行は捏造せず '-' を表示する", () => {
    render(<TrimmingHistorySection petId="7" />);

    const noStaffRow = screen.getByText("シャンプー", { selector: "td" }).closest("tr");
    expect(noStaffRow).not.toBeNull();
    expect(within(noStaffRow as HTMLElement).getByText("完了")).toBeInTheDocument();
    // staff = "" → "-"
    expect(within(noStaffRow as HTMLElement).getByText("-")).toBeInTheDocument();
  });

  it("履歴ゼロのとき空状態を表示する", () => {
    hooks.useGetPetTrimmingHistory.mockReturnValue(ok({ items: [], isTruncated: false }));
    render(<TrimmingHistorySection petId="7" />);

    expect(screen.getByText("トリミングの履歴はありません")).toBeInTheDocument();
    expect(screen.queryByRole("columnheader")).not.toBeInTheDocument();
  });

  it("閲覧権限が無ければセクションを縮退表示し pet_id を取得に渡さない", () => {
    hooks.usePermission.mockReturnValue({ canView: false });
    hooks.useGetPetTrimmingHistory.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
    });
    render(<TrimmingHistorySection petId="7" />);

    expect(screen.getByText("閲覧権限がありません")).toBeInTheDocument();
    expect(screen.queryByRole("columnheader")).not.toBeInTheDocument();
    // canView=false のとき hook 引数は undefined（fetch を起こさない）
    expect(hooks.useGetPetTrimmingHistory).toHaveBeenCalledWith(undefined);
  });

  it("SD-18: isTruncated=true のとき打ち切り注記を表示する", () => {
    hooks.useGetPetTrimmingHistory.mockReturnValue(ok({ items: trimmings, isTruncated: true }));
    render(<TrimmingHistorySection petId="7" />);

    expect(screen.getByTestId("section-truncation-notice")).toBeInTheDocument();
  });
});
