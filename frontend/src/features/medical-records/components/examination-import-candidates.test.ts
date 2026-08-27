import { describe, expect, it } from "vitest";
import {
  filterImportableExaminations,
  isExaminationImportable,
  type ExaminationImportCandidate,
} from "./examination-import-candidates";

const MR = "10";

function exam(
  overrides: Partial<ExaminationImportCandidate> & { id: string },
): ExaminationImportCandidate {
  return {
    medicalRecordId: undefined,
    status: "依頼中",
    currentRevisionVersion: undefined,
    ...overrides,
  };
}

describe("isExaminationImportable / filterImportableExaminations (BUG-014)", () => {
  it("excludes confirmed (確定) exams — not selectable import candidates", () => {
    expect(
      isExaminationImportable(exam({ id: "c1", status: "確定" }), MR),
    ).toBe(false);
    expect(
      isExaminationImportable(exam({ id: "c2", status: "confirmed" }), MR),
    ).toBe(false);
  });

  it("keeps unconfirmed / unlinked importable exams selectable", () => {
    expect(
      isExaminationImportable(exam({ id: "u1", status: "依頼中" }), MR),
    ).toBe(true);
    expect(
      isExaminationImportable(
        exam({ id: "u2", status: "完了", medicalRecordId: undefined }),
        MR,
      ),
    ).toBe(true);
    expect(
      isExaminationImportable(
        exam({ id: "u3", status: "結果入力済み", medicalRecordId: MR }),
        MR,
      ),
    ).toBe(true);
  });

  it("excludes exams already linked to another medical record", () => {
    expect(
      isExaminationImportable(
        exam({ id: "other", medicalRecordId: "99", status: "依頼中" }),
        MR,
      ),
    ).toBe(false);
  });

  it("excludes revisioned exams (BE blocks medical_record_id change)", () => {
    expect(
      isExaminationImportable(
        exam({ id: "rev", status: "完了", currentRevisionVersion: 1 }),
        MR,
      ),
    ).toBe(false);
  });

  it("fail-closed on unknown status", () => {
    expect(
      isExaminationImportable(exam({ id: "x", status: "謎" }), MR),
    ).toBe(false);
    expect(
      isExaminationImportable(exam({ id: "empty", status: "" }), MR),
    ).toBe(false);
  });

  it("filterImportableExaminations returns only importable rows (immutable)", () => {
    const source = [
      exam({ id: "ok", status: "検査中" }),
      exam({ id: "confirmed", status: "確定" }),
      exam({ id: "linked-elsewhere", medicalRecordId: "7", status: "依頼中" }),
      exam({ id: "rev", status: "依頼中", currentRevisionVersion: 2 }),
    ];
    const filtered = filterImportableExaminations(source, MR);
    expect(filtered.map((e) => e.id)).toEqual(["ok"]);
    expect(source).toHaveLength(4);
  });
});
