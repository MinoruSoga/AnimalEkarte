import { describe, it, expect } from "vitest";
import { transformExamination } from "./examination";
import type {
  Examination as BackendExamination,
  ExamResult as BackendExamResult,
} from "@/types/generated/models";

const minimal: BackendExamination = {
  id: 1,
  clinic_id: 1,
  exam_type_id: 2,
  doctor_id: 3,
  date: "2026-03-25T00:00:00Z",
  status: "pending",
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

describe("transformExamination", () => {
  it("id を string に変換する", () => {
    expect(transformExamination({ ...minimal, id: 42 }).id).toBe("42");
  });

  it("id が未設定のとき '0' を返す", () => {
    expect(transformExamination({ ...minimal, id: undefined as unknown as number }).id).toBe("0");
  });

  it("date をスライスして日付部分のみ返す", () => {
    expect(transformExamination({ ...minimal, date: "2026-03-25T10:00:00Z" }).date).toBe(
      "2026-03-25",
    );
  });

  it("date が未設定のとき空文字を返す", () => {
    expect(transformExamination({ ...minimal, date: undefined as unknown as string }).date).toBe(
      "",
    );
  });

  it("status: pending → '依頼中'", () => {
    expect(transformExamination({ ...minimal, status: "pending" }).status).toBe("依頼中");
  });

  it("status: in_progress → '検査中'", () => {
    expect(transformExamination({ ...minimal, status: "in_progress" }).status).toBe("検査中");
  });

  it("status: completed → '完了'", () => {
    expect(transformExamination({ ...minimal, status: "completed" }).status).toBe("完了");
  });

  it("未知の status は '依頼中' にフォールバックする", () => {
    expect(transformExamination({ ...minimal, status: "unknown" as "pending" }).status).toBe(
      "依頼中",
    );
  });

  it("exam_type.name を testType にマップする", () => {
    const result = transformExamination({
      ...minimal,
      exam_type: { id: 2, clinic_id: 1, name: "血液検査" } as BackendExamination["exam_type"],
    });
    expect(result.testType).toBe("血液検査");
  });

  it("doctor.name を doctor にマップする", () => {
    const result = transformExamination({
      ...minimal,
      doctor: { id: 3, clinic_id: 1, name: "山田医師" } as BackendExamination["doctor"],
    });
    expect(result.doctor).toBe("山田医師");
  });

  it("doctor が未設定のとき doctor_id の文字列表現を使う", () => {
    expect(transformExamination({ ...minimal, doctor: undefined, doctor_id: 7 }).doctor).toBe("7");
  });

  it("items を変換して返す", () => {
    const item: BackendExamResult = {
      id: 10,
      name: "白血球",
      result: "5.0",
      unit: "10^3/μL",
      reference_value: "4.0-10.0",
    } as BackendExamResult;
    const result = transformExamination({ ...minimal, items: [item] });
    expect(result.items).toHaveLength(1);
    expect(result.items![0].id).toBe("10");
    expect(result.items![0].name).toBe("白血球");
    expect(result.items![0].referenceValue).toBe("4.0-10.0");
  });

  it("pet.owner.name を ownerName にマップする", () => {
    const result = transformExamination({
      ...minimal,
      pet: {
        id: 1,
        clinic_id: 1,
        owner: { id: 1, clinic_id: 1, name: "田中太郎" } as BackendExamination["pet"]["owner"],
      } as BackendExamination["pet"],
    });
    expect(result.ownerName).toBe("田中太郎");
  });
});
