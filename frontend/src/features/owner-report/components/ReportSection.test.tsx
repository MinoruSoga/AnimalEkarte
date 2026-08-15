import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { ReportSection } from "./ReportSection";

const baseProps = {
  title: "投薬履歴",
  canView: true,
  isLoading: false,
  isError: false,
  isEmpty: false,
};

function renderSection(overrides: Partial<React.ComponentProps<typeof ReportSection>> = {}) {
  return render(
    <ReportSection {...baseProps} {...overrides}>
      <div>本文コンテンツ</div>
    </ReportSection>,
  );
}

describe("ReportSection", () => {
  it("タイトルは常に表示される", () => {
    renderSection();
    expect(screen.getByRole("heading", { name: "投薬履歴" })).toBeInTheDocument();
  });

  it("通常時は children を描画する", () => {
    renderSection();
    expect(screen.getByText("本文コンテンツ")).toBeInTheDocument();
  });

  it("権限なしは縮退表示し children を出さない", () => {
    renderSection({ canView: false });
    expect(screen.getByTestId("section-no-permission")).toHaveTextContent("閲覧権限がありません");
    expect(screen.queryByText("本文コンテンツ")).not.toBeInTheDocument();
  });

  it("ローディング中は読み込み表示にする", () => {
    renderSection({ isLoading: true });
    expect(screen.getByRole("status")).toHaveTextContent("読み込み中...");
    expect(screen.queryByText("本文コンテンツ")).not.toBeInTheDocument();
  });

  it("エラー時はエラー表示にする", () => {
    renderSection({ isError: true });
    expect(screen.getByRole("alert")).toHaveTextContent("読み込みに失敗しました");
    expect(screen.queryByText("本文コンテンツ")).not.toBeInTheDocument();
  });

  it("空時は emptyMessage を表示する", () => {
    renderSection({ isEmpty: true, emptyMessage: "投薬の履歴はありません" });
    expect(screen.getByText("投薬の履歴はありません")).toBeInTheDocument();
    expect(screen.queryByText("本文コンテンツ")).not.toBeInTheDocument();
  });

  it("権限なしはローディング/エラー/空より優先される", () => {
    renderSection({ canView: false, isLoading: true, isError: true, isEmpty: true });
    expect(screen.getByTestId("section-no-permission")).toBeInTheDocument();
    expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
  });

  it("SD-18: isTruncated=true のとき打ち切り注記を children の前に表示する", () => {
    renderSection({ isTruncated: true });
    expect(screen.getByTestId("section-truncation-notice")).toHaveTextContent("直近100件を表示しています");
    expect(screen.getByText("本文コンテンツ")).toBeInTheDocument();
  });

  it("isTruncated 未指定（既定 false）では打ち切り注記を出さない", () => {
    renderSection();
    expect(screen.queryByTestId("section-truncation-notice")).not.toBeInTheDocument();
  });

  it("isTruncated=true でも空/エラー/権限なしのときは注記を出さない", () => {
    renderSection({ isTruncated: true, isEmpty: true, emptyMessage: "履歴はありません" });
    expect(screen.queryByTestId("section-truncation-notice")).not.toBeInTheDocument();
  });
});
