import { describe, expect, it } from "vitest";
import { jstDateStartISOString, todayJSTISO } from "@/lib/jst-date";
import { buildCreateExaminationRequest } from "./use-examination-form-model";

describe("buildCreateExaminationRequest — date は RFC3339 (PR665 #4272420944)", () => {
  const base = { medicalRecordId: "", petId: "42", items: [] };

  it("既定（日付未指定）は当日 JST の RFC3339 開始時刻", () => {
    const req = buildCreateExaminationRequest({ ...base, current: {} });
    expect(req.date).toBe(jstDateStartISOString(todayJSTISO()));
    expect(req.date).toMatch(/^\d{4}-\d{2}-\d{2}T00:00:00\+09:00$/);
  });

  it("YYYY-MM-DD のスナップショットは JST 開始 RFC3339 へ変換される", () => {
    const req = buildCreateExaminationRequest({ ...base, current: { date: "2026-10-15" } });
    expect(req.date).toBe("2026-10-15T00:00:00+09:00");
  });

  it("明示的な RFC3339 はそのまま保持される", () => {
    const req = buildCreateExaminationRequest({
      ...base,
      current: { date: "2026-10-15T00:00:00+09:00" },
    });
    expect(req.date).toBe("2026-10-15T00:00:00+09:00");
  });

  it("explicitly cleared dates are not silently replaced with today", () => {
    const req = buildCreateExaminationRequest({ ...base, current: { date: "" } });
    expect(req.date).toBe("");
  });
});
