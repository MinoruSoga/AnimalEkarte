import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes, useLocation } from "react-router";

import { ExaminationGroup } from "./ExaminationGroup";
import type { ExamGroup } from "../api/get-record-examinations";
import type { ExamResult } from "@/lib/transforms/examination";
import { C, STYLE } from "@/lib/design-tokens";

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

function renderGroup(group = makeGroup()) {
  return render(
    <MemoryRouter>
      <ExaminationGroup group={group} petId="7" />
    </MemoryRouter>,
  );
}

describe("ExaminationGroup", () => {
  it("ヘッダ列（項目名・結果値・単位・基準値・判定）を描画する", () => {
    renderGroup();
    expect(screen.getByText("項目名")).toBeInTheDocument();
    expect(screen.getByText("結果値")).toBeInTheDocument();
    expect(screen.getByText("単位")).toBeInTheDocument();
    expect(screen.getByText("基準値")).toBeInTheDocument();
    expect(screen.getByText("判定")).toBeInTheDocument();
  });

  it("ヘッダ行が DESIGN.md ex-data-table-cell（sectionLabel）で表示される", () => {
    renderGroup();
    const headerRow = screen.getByText("項目名").parentElement;
    expect(headerRow).not.toBeNull();
    for (const cls of STYLE.sectionLabel.split(" ")) {
      expect(headerRow?.className).toContain(cls);
    }
  });

  it("date と machine をヘッダに表示する", () => {
    renderGroup(makeGroup({ date: "2026-04-29 10:00", machine: "DRI-CHEM" }));
    expect(screen.getByText("2026-04-29 10:00")).toBeInTheDocument();
    expect(screen.getByText("DRI-CHEM")).toBeInTheDocument();
  });

  it("項目名・結果値・単位・基準値（referenceValue）を表示する", () => {
    renderGroup(
      makeGroup({
        items: [
          makeItem({
            name: "GLU",
            result: "95",
            unit: "mg/dL",
            referenceValue: "70-110",
          }),
        ],
      }),
    );
    expect(screen.getByText("GLU")).toBeInTheDocument();
    expect(screen.getByText("95")).toBeInTheDocument();
    expect(screen.getByText("mg/dL")).toBeInTheDocument();
    expect(screen.getByText("70-110")).toBeInTheDocument();
  });

  // BUG-456: active writer は inspectionValue / normalValue を主に書く一方、
  // 表示が result / referenceValue のみだと通常検査の保存値が空になる。
  // 新field優先・旧field fallback・両方空の3系統を固定する。
  it("新fieldがあるとき inspectionValue / referenceValue を結果値・基準値に表示する", () => {
    renderGroup(
      makeGroup({
        items: [
          makeItem({
            name: "GLU",
            // 新旧不一致: 旧fieldは stale でも新fieldが canonical
            inspectionValue: "120",
            result: "99",
            referenceValue: "70-110",
            normalValue: "0-999",
          }),
        ],
      }),
    );
    expect(screen.getByText("120")).toBeInTheDocument();
    expect(screen.getByText("70-110")).toBeInTheDocument();
    expect(screen.queryByText("99")).not.toBeInTheDocument();
    expect(screen.queryByText("0-999")).not.toBeInTheDocument();
  });

  it("新fieldが空のとき result / normalValue にフォールバックする", () => {
    renderGroup(
      makeGroup({
        items: [
          makeItem({
            name: "BUN",
            inspectionValue: "",
            result: "18",
            referenceValue: "",
            normalValue: "6-25",
          }),
        ],
      }),
    );
    expect(screen.getByText("18")).toBeInTheDocument();
    expect(screen.getByText("6-25")).toBeInTheDocument();
  });

  it("結果値・基準値の新旧fieldが両方空なら '-' を表示する", () => {
    renderGroup(
      makeGroup({
        items: [
          makeItem({
            name: "EMPTY",
            inspectionValue: "",
            result: "",
            referenceValue: "",
            normalValue: "",
          }),
        ],
      }),
    );
    // 結果値列・基準値列それぞれに placeholder "-"
    const dashes = screen.getAllByText("-");
    expect(dashes.length).toBeGreaterThanOrEqual(2);
  });

  // BUG-456 residual: unit is a single field end-to-end; empty master/persisted
  // unit must show "-" like ExamItemsTable / ExamPivotTable (parity).
  it("unit が空のとき単位列に '-' を表示する", () => {
    renderGroup(
      makeGroup({
        items: [
          makeItem({
            name: "WBC",
            inspectionValue: "12.3",
            unit: "",
            referenceValue: "5-15",
          }),
        ],
      }),
    );
    expect(screen.getByText("WBC")).toBeInTheDocument();
    expect(screen.getByText("12.3")).toBeInTheDocument();
    expect(screen.getByText("-")).toBeInTheDocument();
    expect(screen.getByText("5-15")).toBeInTheDocument();
  });

  it("status=high のとき HIGH バッジを表示する", () => {
    renderGroup(
      makeGroup({
        items: [makeItem({ status: "high", isAbnormal: true })],
      }),
    );
    expect(screen.getByText("HIGH")).toHaveClass(C.bgDanger, C.hoverBgDanger90);
    expect(screen.getByText("95")).toHaveClass(C.danger, "font-bold");
    expect(screen.queryByText("LOW")).not.toBeInTheDocument();
  });

  // BUG-450: 基準値マスタが空の間、未評価の項目が「基準値内」と同じ緑チェックで
  // 描画されると、獣医師が未評価を正常と読む。この3 case がその誤読を塞ぐ。
  it("isAssessed=false のとき未判定バッジを表示し基準値内アイコンを出さない", () => {
    renderGroup(
      makeGroup({
        items: [makeItem({ status: "normal", isAssessed: false })],
      }),
    );
    expect(screen.getByText("未判定")).toHaveClass(
      C.textWarning,
      C.borderWarning20,
      C.bgWarning50,
    );
    expect(
      screen.getByText("（基準値未設定のため判定していない）"),
    ).toHaveClass("sr-only");
    expect(screen.queryByLabelText("基準値内")).not.toBeInTheDocument();
  });

  it("isAssessed=true かつ status=normal のとき基準値内の accessible name を持つ", () => {
    renderGroup(
      makeGroup({
        items: [makeItem({ status: "normal", isAssessed: true })],
      }),
    );
    expect(screen.getByLabelText("基準値内")).toBeInTheDocument();
    expect(screen.queryByText("未判定")).not.toBeInTheDocument();
  });

  it("isAssessed=false でも status=high なら異常警告を優先する", () => {
    renderGroup(
      makeGroup({
        items: [makeItem({ status: "high", isAssessed: false, isAbnormal: true })],
      }),
    );
    expect(screen.getByText("HIGH")).toBeInTheDocument();
    expect(screen.queryByText("未判定")).not.toBeInTheDocument();
  });

  it("status=low のとき LOW バッジを表示する", () => {
    renderGroup(
      makeGroup({
        items: [makeItem({ status: "low", isAbnormal: true })],
      }),
    );
    expect(screen.getByText("LOW")).toHaveClass(
      C.textStatusBlue,
      C.borderBlue400,
      C.bgStatusBlueLight,
    );
    expect(screen.getByText("95")).toHaveClass(C.textStatusBlue, "font-bold");
    expect(screen.queryByText("HIGH")).not.toBeInTheDocument();
  });

  it("複数行を全て描画する", () => {
    renderGroup(
      makeGroup({
        items: [
          makeItem({ id: "1", name: "GLU" }),
          makeItem({ id: "2", name: "BUN" }),
          makeItem({ id: "3", name: "ALT" }),
        ],
      }),
    );
    expect(screen.getByText("GLU")).toBeInTheDocument();
    expect(screen.getByText("BUN")).toBeInTheDocument();
    expect(screen.getByText("ALT")).toBeInTheDocument();
  });

  it("検歴を表示の1操作で対象ペットの時系列ピボットへ遷移する", async () => {
    const user = userEvent.setup();

    function LocationProbe() {
      const location = useLocation();
      return (
        <output>
          {location.pathname}
          {location.search}
          {JSON.stringify(location.state)}
        </output>
      );
    }

    render(
      <MemoryRouter initialEntries={["/medical-records/55?tab=examinations"]}>
        <Routes>
          <Route
            path="/medical-records/:id"
            element={<ExaminationGroup group={makeGroup()} petId="7" />}
          />
          <Route path="/examinations/:id" element={<LocationProbe />} />
        </Routes>
      </MemoryRouter>,
    );

    await user.click(
      screen.getByRole("link", {
        name: "2026-04-29 10:00の検歴を表示",
      }),
    );

    expect(screen.getByRole("status")).toHaveTextContent(
      "/examinations/100?petId=7&historyView=pivot",
    );
    expect(screen.getByRole("status")).toHaveTextContent(
      '{"from":"/medical-records/55?tab=examinations"}',
    );
  });
});
