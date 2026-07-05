import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";

import { ExaminationGroup } from "./ExaminationGroup";
import type { ExamGroup } from "../api/get-record-examinations";
import type { ExamResult } from "@/features/examinations";
import { STYLE } from "@/lib/design-tokens";

// shared transformExamResult が返す ViewModel 形状をフィクスチャに使う。
// ここで snake_case フィールドを使ってしまうと型エラーになり、共有型のズレを検知できる。
const makeItem = (overrides: Partial<ExamResult> = {}): ExamResult => ({
  id: "1",
  examTypeFieldId: 1,
  name: "GLU",
  result: "95",
  inspectionValue: "95",
  normalValue: "70-110",
  unit: "mg/dL",
  referenceValue: "70-110",
  refMin: 70,
  refMax: 110,
  isAbnormal: false,
  status: "normal",
  sortOrder: 0,
  ...overrides,
});

const makeGroup = (overrides: Partial<ExamGroup> = {}): ExamGroup => ({
  id: 100,
  date: "2026-04-29 10:00",
  machine: "DRI-CHEM",
  items: [makeItem()],
  ...overrides,
});

describe("ExaminationGroup", () => {
  it("ヘッダ列（項目名・結果値・単位・基準値・判定）を描画する", () => {
    render(<ExaminationGroup group={makeGroup()} />);
    expect(screen.getByText("項目名")).toBeInTheDocument();
    expect(screen.getByText("結果値")).toBeInTheDocument();
    expect(screen.getByText("単位")).toBeInTheDocument();
    expect(screen.getByText("基準値")).toBeInTheDocument();
    expect(screen.getByText("判定")).toBeInTheDocument();
  });

  it("ヘッダ行が DESIGN.md ex-data-table-cell（sectionLabel）で表示される", () => {
    render(<ExaminationGroup group={makeGroup()} />);
    const headerRow = screen.getByText("項目名").parentElement;
    expect(headerRow).not.toBeNull();
    for (const cls of STYLE.sectionLabel.split(" ")) {
      expect(headerRow?.className).toContain(cls);
    }
  });

  it("date と machine をヘッダに表示する", () => {
    render(
      <ExaminationGroup
        group={makeGroup({ date: "2026-04-29 10:00", machine: "DRI-CHEM" })}
      />,
    );
    expect(screen.getByText("2026-04-29 10:00")).toBeInTheDocument();
    expect(screen.getByText("DRI-CHEM")).toBeInTheDocument();
  });

  it("項目名・結果値・単位・基準値（referenceValue）を表示する", () => {
    render(
      <ExaminationGroup
        group={makeGroup({
          items: [
            makeItem({
              name: "GLU",
              result: "95",
              unit: "mg/dL",
              referenceValue: "70-110",
            }),
          ],
        })}
      />,
    );
    expect(screen.getByText("GLU")).toBeInTheDocument();
    expect(screen.getByText("95")).toBeInTheDocument();
    expect(screen.getByText("mg/dL")).toBeInTheDocument();
    expect(screen.getByText("70-110")).toBeInTheDocument();
  });

  it("status=high のとき HIGH バッジを表示する", () => {
    render(
      <ExaminationGroup
        group={makeGroup({
          items: [makeItem({ status: "high", isAbnormal: true })],
        })}
      />,
    );
    expect(screen.getByText("HIGH")).toBeInTheDocument();
    expect(screen.queryByText("LOW")).not.toBeInTheDocument();
  });

  it("status=low のとき LOW バッジを表示する", () => {
    render(
      <ExaminationGroup
        group={makeGroup({
          items: [makeItem({ status: "low", isAbnormal: true })],
        })}
      />,
    );
    expect(screen.getByText("LOW")).toBeInTheDocument();
    expect(screen.queryByText("HIGH")).not.toBeInTheDocument();
  });

  it("複数行を全て描画する", () => {
    render(
      <ExaminationGroup
        group={makeGroup({
          items: [
            makeItem({ id: "1", name: "GLU" }),
            makeItem({ id: "2", name: "BUN" }),
            makeItem({ id: "3", name: "ALT" }),
          ],
        })}
      />,
    );
    expect(screen.getByText("GLU")).toBeInTheDocument();
    expect(screen.getByText("BUN")).toBeInTheDocument();
    expect(screen.getByText("ALT")).toBeInTheDocument();
  });
});
