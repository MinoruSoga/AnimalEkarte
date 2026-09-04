import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { ExamStatusBadge } from "./ExamStatusBadge";

// FE-RC-027: ExamPivotTable / ExamItemsTable の完全コピペ StatusBadge を統合した
// 共有コンポーネント。両呼び出し元の既存表示（compact 有無）を固定する回帰テスト。
describe("ExamStatusBadge", () => {
  it("status=high でHIGHバッジを表示する（compact有無で共通）", () => {
    render(<ExamStatusBadge status="high" />);
    expect(screen.getByText("HIGH")).toBeInTheDocument();
  });

  it("status=low でLOWバッジを表示する", () => {
    render(<ExamStatusBadge status="low" />);
    expect(screen.getByText("LOW")).toBeInTheDocument();
  });

  it("isAssessed=false で未判定バッジを表示する（status問わず優先）", () => {
    render(<ExamStatusBadge status="normal" isAssessed={false} />);
    expect(screen.getByText("未判定")).toBeInTheDocument();
  });

  it("既定（非compact）status=normal はCheckCircleアイコンを表示する", () => {
    render(<ExamStatusBadge status="normal" />);
    expect(screen.getByRole("img", { name: "基準値内" })).toBeInTheDocument();
  });

  it("既定（非compact）status未設定は '-' を表示する（保存前の未判定）", () => {
    render(<ExamStatusBadge />);
    expect(screen.getByText("-")).toBeInTheDocument();
  });

  it("compact=true の status=normal は何も表示しない（pivotの既存挙動）", () => {
    const { container } = render(<ExamStatusBadge status="normal" compact />);
    expect(container).toBeEmptyDOMElement();
  });

  it("compact=true の status未設定は何も表示しない（pivotの既存挙動）", () => {
    const { container } = render(<ExamStatusBadge compact />);
    expect(container).toBeEmptyDOMElement();
  });
});
