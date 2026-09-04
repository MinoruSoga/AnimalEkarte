import { describe, expect, it } from "vitest";

import type { BackendExamination, ExamItemsResponse } from "../types";
import { transformExamination, transformExamResult } from "./transforms";

function makeBackendItem(
  overrides: Partial<ExamItemsResponse["items"][number]> = {},
): ExamItemsResponse["items"][number] {
  return {
    id: 10,
    exam_id: 20,
    name: "白血球",
    inspection_value: "5.0",
    normal_value: "4.0-10.0",
    result: "5.0",
    unit: "10^3/μL",
    reference_value: "4.0-10.0",
    is_assessed: true,
    is_abnormal: false,
    status: "normal",
    sort_order: 1,
    created_at: "2026-07-27T00:00:00Z",
    updated_at: "2026-07-27T00:00:00Z",
    ...overrides,
  } satisfies ExamItemsResponse["items"][number];
}

describe("transformExamResult", () => {
  it.each([{ isAssessed: false }, { isAssessed: true }])(
    "server-computed assessed state $isAssessed を行データへ写像する",
    ({ isAssessed }) => {
      expect(transformExamResult(makeBackendItem({ is_assessed: isAssessed })).isAssessed).toBe(
        isAssessed,
      );
    },
  );

  it("qualitative_min/max を qualitativeMin/Max に写像する", () => {
    const result = transformExamResult(
      makeBackendItem({
        reference_value: "",
        qualitative_min: "(-)",
        qualitative_max: "(+)",
      }),
    );

    expect(result.qualitativeMin).toBe("(-)");
    expect(result.qualitativeMax).toBe("(+)");
  });

  it("qualitative_min/max が省略されたとき undefined を返す", () => {
    const result = transformExamResult(makeBackendItem());

    expect(result.qualitativeMin).toBeUndefined();
    expect(result.qualitativeMax).toBeUndefined();
  });
});

describe("transformExamination", () => {
  const minimalExamination = {
    id: 1,
    clinic_id: 1,
    exam_type_id: 2,
    date: "2026-08-03T00:00:00Z",
    result_summary: "",
    machine: "",
    status: "completed",
    created_at: "2026-08-03T00:00:00Z",
    updated_at: "2026-08-03T00:00:00Z",
  } satisfies BackendExamination;

  it("revision pointer を患者変更ロック用の record field に写像する", () => {
    expect(
      transformExamination({
        ...minimalExamination,
        current_revision_version: 2,
      }).currentRevisionVersion,
    ).toBe(2);
  });

  it("revision pointer がない未確定記録では undefined を維持する", () => {
    expect(transformExamination(minimalExamination).currentRevisionVersion).toBeUndefined();
  });
});
