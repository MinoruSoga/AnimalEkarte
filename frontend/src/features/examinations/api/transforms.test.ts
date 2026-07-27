import { describe, expect, it } from "vitest";

import type { ExamItemsResponse } from "./types";
import { transformExamResult } from "./transforms";

describe("transformExamResult", () => {
  it.each([
    { isAssessed: false },
    { isAssessed: true },
  ])("server-computed assessed state $isAssessed を行データへ写像する", ({ isAssessed }) => {
    const item = {
      id: 10,
      exam_id: 20,
      name: "白血球",
      inspection_value: "5.0",
      normal_value: "4.0-10.0",
      result: "5.0",
      unit: "10^3/μL",
      reference_value: "4.0-10.0",
      is_assessed: isAssessed,
      is_abnormal: false,
      status: "normal",
      sort_order: 1,
      created_at: "2026-07-27T00:00:00Z",
      updated_at: "2026-07-27T00:00:00Z",
    } satisfies ExamItemsResponse["items"][number];

    expect(transformExamResult(item).isAssessed).toBe(isAssessed);
  });
});
