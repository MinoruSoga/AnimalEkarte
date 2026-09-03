import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import { C } from "@/lib/design-tokens";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";

import type { ExaminationRecord } from "../api/transforms";
import { ExamPivotTable } from "./ExamPivotTable";

interface ExamItemFixture {
  id: number;
  exam_id: number;
  exam_type_field_id: number | null;
  name: string;
  inspection_value: string;
  normal_value: string;
  result: string;
  unit: string;
  reference_value: string;
  ref_min: number | null;
  ref_max: number | null;
  qualitative_min?: string | null;
  qualitative_max?: string | null;
  is_assessed?: boolean;
  is_abnormal: boolean;
  status: "normal" | "high" | "low";
  sort_order: number;
}

function makeExamination(
  id: string,
  date: string,
  testTypeId = "10",
): ExaminationRecord {
  return {
    id,
    date,
    ownerName: "飼主",
    petName: "ポチ",
    petId: "7",
    medicalRecordId: undefined,
    testType: "血液検査",
    testTypeId,
    doctor: "獣医師",
    doctorId: "3",
    status: "確定",
    resultSummary: undefined,
    machine: "DRI-CHEM",
    items: undefined,
  };
}

function makeItem(
  overrides: Partial<ExamItemFixture> = {},
): ExamItemFixture {
  return {
    id: 1,
    exam_id: 1,
    exam_type_field_id: 100,
    name: "GLU",
    inspection_value: "95",
    normal_value: "",
    result: "result-field-is-not-the-source",
    unit: "mg/dL",
    reference_value: "70-110",
    ref_min: 70,
    ref_max: 110,
    is_abnormal: false,
    status: "normal",
    sort_order: 1,
    ...overrides,
  };
}

function renderPivot(
  examinations: ExaminationRecord[],
  itemsByExaminationId: Record<string, ExamItemFixture[]>,
  requestedIds?: string[],
) {
  server.use(
    http.get("/api/v1/examinations/:id/items", ({ params }) => {
      const id = String(params.id);
      requestedIds?.push(id);
      return HttpResponse.json({ items: itemsByExaminationId[id] ?? [] });
    }),
  );

  return render(
    <ExamPivotTable examinations={examinations} sortOrder="desc" />,
    { wrapper: createTestWrapper() },
  );
}

describe("ExamPivotTable", () => {
  it("同日検査を別列に保ち、field ID行を転置して欠測セルを「-」で表示する", async () => {
    renderPivot(
      [
        makeExamination("2", "2026-07-20"),
        makeExamination("1", "2026-07-20"),
      ],
      {
        "2": [
          makeItem({
            id: 2,
            exam_id: 2,
            exam_type_field_id: 100,
            inspection_value: "120",
            is_assessed: false,
            status: "high",
            is_abnormal: true,
          }),
          makeItem({
            id: 3,
            exam_id: 2,
            exam_type_field_id: 200,
            name: "ALT",
            inspection_value: "40",
            unit: "U/L",
          }),
        ],
        "1": [
          makeItem({
            id: 1,
            exam_id: 1,
            exam_type_field_id: 100,
            inspection_value: "60",
            is_assessed: false,
            status: "low",
            is_abnormal: true,
          }),
        ],
      },
    );

    expect(
      await screen.findByRole("columnheader", { name: "2026-07-20 1件目" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("columnheader", { name: "2026-07-20 2件目" }),
    ).toBeInTheDocument();

    const gluRow = screen.getByRole("row", { name: /GLU.*120.*60/ });
    expect(within(gluRow).getByText("120").closest("td")).toHaveClass(
      C.bgDanger8,
    );
    expect(within(gluRow).getByText("60").closest("td")).toHaveClass(
      C.bgStatusBlueLight,
    );
    expect(within(gluRow).getByText("HIGH")).toBeInTheDocument();
    expect(within(gluRow).getByText("LOW")).toBeInTheDocument();
    expect(within(gluRow).queryByText("未判定")).not.toBeInTheDocument();

    const altRow = screen.getByRole("row", { name: /ALT.*40.*-/ });
    expect(within(altRow).getByText("-")).toBeInTheDocument();
  });

  it("未評価のnormal cellに「未判定」cueを表示する", async () => {
    renderPivot(
      [makeExamination("1", "2026-07-20")],
      {
        "1": [makeItem({ is_assessed: false })],
      },
    );

    const row = await screen.findByRole("row", { name: /GLU.*95/ });
    expect(within(row).getByText("未判定")).toBeInTheDocument();
    expect(
      within(row).getByText("（基準値未設定のため判定していない）"),
    ).toHaveClass("sr-only");
  });

  it("評価済みnormal cellに「未判定」を表示しない", async () => {
    renderPivot(
      [makeExamination("1", "2026-07-20")],
      {
        "1": [makeItem({ is_assessed: true })],
      },
    );

    const row = await screen.findByRole("row", { name: /GLU.*95/ });
    expect(within(row).queryByText("未判定")).not.toBeInTheDocument();
  });

  it("評価情報未指定のnormal cellは現行どおりbadge無しで表示する", async () => {
    renderPivot(
      [makeExamination("1", "2026-07-20")],
      {
        "1": [makeItem()],
      },
    );

    const row = await screen.findByRole("row", { name: /GLU.*95/ });
    expect(within(row).queryByText("未判定")).not.toBeInTheDocument();
    expect(within(row).queryByText("HIGH")).not.toBeInTheDocument();
    expect(within(row).queryByText("LOW")).not.toBeInTheDocument();
  });

  it("同一検査内でlegacy表示キーが衝突しても行を失わない", async () => {
    renderPivot(
      [makeExamination("1", "2026-07-20")],
      {
        "1": [
          makeItem({
            id: 1,
            exam_id: 1,
            exam_type_field_id: null,
            name: " ＧＬＵ ",
            inspection_value: "95",
          }),
          makeItem({
            id: 2,
            exam_id: 1,
            exam_type_field_id: null,
            name: "GLU",
            inspection_value: "105",
          }),
        ],
      },
    );

    expect(await screen.findByRole("row", { name: /GLU.*95/ })).toBeInTheDocument();
    expect(screen.getByRole("row", { name: /GLU.*105/ })).toBeInTheDocument();
    expect(screen.getAllByText("未マッピング")).toHaveLength(2);
  });

  it("同名でも異なるfield IDは別行にし、inspection_valueだけを値として表示する", async () => {
    renderPivot(
      [makeExamination("1", "2026-07-20")],
      {
        "1": [
          makeItem({
            id: 1,
            exam_type_field_id: 100,
            inspection_value: "95",
          }),
          makeItem({
            id: 2,
            exam_type_field_id: 101,
            inspection_value: "105",
          }),
        ],
      },
    );

    expect(await screen.findAllByRole("row", { name: /GLU/ })).toHaveLength(2);
    expect(screen.getByText("95")).toBeInTheDocument();
    expect(screen.getByText("105")).toBeInTheDocument();
    expect(
      screen.queryByText("result-field-is-not-the-source"),
    ).not.toBeInTheDocument();
  });

  it("field IDなし行をlegacy snapshot keyで表示グループ化し、未マッピングを可視化する", async () => {
    renderPivot(
      [
        makeExamination("2", "2026-07-20", "10"),
        makeExamination("1", "2026-07-19", "10"),
        makeExamination("3", "2026-07-18", "11"),
        makeExamination("4", "2026-07-17", "10"),
      ],
      {
        "2": [
          makeItem({
            id: 2,
            exam_id: 2,
            exam_type_field_id: null,
            name: " ＧＬＵ ",
            inspection_value: "110",
          }),
        ],
        "1": [
          makeItem({
            id: 1,
            exam_id: 1,
            exam_type_field_id: null,
            name: "GLU",
            inspection_value: "90",
          }),
        ],
        "3": [
          makeItem({
            id: 3,
            exam_id: 3,
            exam_type_field_id: null,
            name: "GLU",
            inspection_value: "80",
          }),
        ],
        "4": [
          makeItem({
            id: 4,
            exam_id: 4,
            exam_type_field_id: null,
            name: "GLU",
            unit: "mmol/L",
            inspection_value: "5.2",
          }),
        ],
      },
    );

    await screen.findByText("110");
    const groupedLegacyRow = screen.getByRole("row", {
      name: /GLU.*未マッピング.*110.*90/,
    });
    expect(within(groupedLegacyRow).getByText("未マッピング")).toBeInTheDocument();

    const legacyRows = screen.getAllByRole("row", {
      name: /GLU.*未マッピング/,
    });
    expect(legacyRows).toHaveLength(3);
    expect(screen.getAllByText("未マッピング")).toHaveLength(3);
  });

  it("inspection_valueが空の未実施項目をデフォルト行から除外する", async () => {
    renderPivot(
      [makeExamination("1", "2026-07-20")],
      {
        "1": [
          makeItem({ name: "実施済み", inspection_value: "0" }),
          makeItem({
            id: 2,
            exam_type_field_id: 200,
            name: "未実施",
            inspection_value: "",
          }),
        ],
      },
    );

    expect(await screen.findByText("実施済み")).toBeInTheDocument();
    expect(screen.getByText("0")).toBeInTheDocument();
    expect(screen.queryByText("未実施")).not.toBeInTheDocument();
  });

  it("直近10検査だけを取得・表示する", async () => {
    const examinations = Array.from({ length: 11 }, (_, index) => {
      const day = String(index + 1).padStart(2, "0");
      return makeExamination(String(index + 1), `2026-07-${day}`);
    });
    const itemsByExaminationId = Object.fromEntries(
      examinations.map((examination) => [
        examination.id,
        [
          makeItem({
            id: Number(examination.id),
            exam_id: Number(examination.id),
            inspection_value: examination.id,
          }),
        ],
      ]),
    );
    const requestedIds: string[] = [];

    renderPivot(examinations, itemsByExaminationId, requestedIds);

    expect(
      await screen.findByRole("columnheader", { name: "2026-07-11" }),
    ).toBeInTheDocument();
    await waitFor(() => expect(requestedIds).toHaveLength(10));
    expect(
      screen.queryByRole("columnheader", { name: "2026-07-01" }),
    ).not.toBeInTheDocument();
    expect(requestedIds).not.toContain("1");
  });

  it("項目名で例外的に行を絞り込める", async () => {
    const user = userEvent.setup();
    renderPivot(
      [makeExamination("1", "2026-07-20")],
      {
        "1": [
          makeItem({ name: "GLU" }),
          makeItem({
            id: 2,
            exam_type_field_id: 200,
            name: "ALT",
            inspection_value: "40",
          }),
        ],
      },
    );

    await screen.findByText("GLU");
    await user.type(screen.getByRole("searchbox", { name: "検査項目" }), "ALT");

    expect(screen.queryByText("GLU")).not.toBeInTheDocument();
    expect(screen.getByText("ALT")).toBeInTheDocument();
  });

  it("保存済みreference_valueを優先し、無ければ保存済みref_min/ref_maxを表示する", async () => {
    renderPivot(
      [makeExamination("1", "2026-07-20")],
      {
        "1": [
          makeItem({ reference_value: "70-110" }),
          makeItem({
            id: 2,
            exam_type_field_id: 200,
            name: "ALT",
            unit: "U/L",
            inspection_value: "40",
            reference_value: "",
            ref_min: 10,
            ref_max: 50,
          }),
          makeItem({
            id: 3,
            exam_type_field_id: 300,
            name: "BUN",
            inspection_value: "20",
            reference_value: "",
            ref_min: null,
            ref_max: null,
          }),
          makeItem({
            id: 4,
            exam_type_field_id: 400,
            name: "CRE",
            inspection_value: "1.0",
            reference_value: "",
            ref_min: 0,
            ref_max: null,
          }),
        ],
      },
    );

    expect(await screen.findByText("70-110")).toBeInTheDocument();
    expect(screen.getByText("10-50")).toBeInTheDocument();
    expect(
      within(screen.getByRole("row", { name: /BUN/ })).getByText("-"),
    ).toBeInTheDocument();
    expect(screen.getByText("0-")).toBeInTheDocument();
  });

  it("保存済み qualitative_min/max を基準値列に表示し、reference_value・数値基準の優先順位を守る", async () => {
    renderPivot(
      [makeExamination("1", "2026-07-20")],
      {
        "1": [
          makeItem({
            id: 1,
            exam_type_field_id: 100,
            name: "PRO",
            inspection_value: "(+)",
            reference_value: "",
            ref_min: null,
            ref_max: null,
            qualitative_min: "(-)",
            qualitative_max: "(+)",
          }),
          makeItem({
            id: 2,
            exam_type_field_id: 200,
            name: "BLO",
            inspection_value: "(-)",
            reference_value: "",
            ref_min: null,
            ref_max: null,
            qualitative_min: "(-)",
            qualitative_max: null,
          }),
          makeItem({
            id: 3,
            exam_type_field_id: 300,
            name: "KET",
            inspection_value: "(±)",
            reference_value: "",
            ref_min: null,
            ref_max: null,
            qualitative_min: null,
            qualitative_max: "(+)",
          }),
          makeItem({
            id: 4,
            exam_type_field_id: 400,
            name: "GLU-RV",
            inspection_value: "95",
            reference_value: "70-110",
            ref_min: null,
            ref_max: null,
            qualitative_min: "(-)",
            qualitative_max: "(+)",
          }),
          makeItem({
            id: 5,
            exam_type_field_id: 500,
            name: "ALT",
            unit: "U/L",
            inspection_value: "40",
            reference_value: "",
            ref_min: 10,
            ref_max: 50,
            qualitative_min: "(-)",
            qualitative_max: "(+)",
          }),
          makeItem({
            id: 6,
            exam_type_field_id: 600,
            name: "BUN",
            inspection_value: "20",
            reference_value: "",
            ref_min: null,
            ref_max: null,
          }),
          makeItem({
            id: 7,
            exam_type_field_id: 700,
            name: "未実施定性",
            inspection_value: "",
            reference_value: "",
            ref_min: null,
            ref_max: null,
            qualitative_min: "(-)",
            qualitative_max: "(+)",
          }),
        ],
      },
    );

    expect(await screen.findByText("(-)-(+)")).toBeInTheDocument();
    expect(screen.getByText("(-)-")).toBeInTheDocument();
    expect(screen.getByText("-(+)")).toBeInTheDocument();
    expect(screen.getByText("70-110")).toBeInTheDocument();
    expect(screen.getByText("10-50")).toBeInTheDocument();
    expect(
      within(screen.getByRole("row", { name: /BUN/ })).getByText("-"),
    ).toBeInTheDocument();
    expect(screen.queryByText("未実施定性")).not.toBeInTheDocument();
  });

  it("検査履歴が無い場合は空状態を表示する", () => {
    render(
      <ExamPivotTable examinations={[]} sortOrder="desc" />,
      { wrapper: createTestWrapper() },
    );

    expect(screen.getByText("検査記録がありません")).toBeInTheDocument();
  });

  it("項目取得失敗を欠測表示にせず明示する", async () => {
    server.use(
      http.get(
        "/api/v1/examinations/:id/items",
        () => new HttpResponse(null, { status: 500 }),
      ),
    );

    render(
      <ExamPivotTable
        examinations={[makeExamination("1", "2026-07-20")]}
        sortOrder="desc"
      />,
      { wrapper: createTestWrapper() },
    );

    expect(
      await screen.findByRole("alert", {
        name: "検査項目の取得に失敗しました",
      }),
    ).toBeInTheDocument();
  });
});
