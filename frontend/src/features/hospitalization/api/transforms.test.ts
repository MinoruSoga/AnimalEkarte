import { describe, it, expect } from "vitest";
import { transformHospitalization } from "./transforms";
import type { BackendHospitalization } from "./types";

/** HospitalizationResponse wire — no treatment_plans / full model relations. */
const minimal: BackendHospitalization = {
  id: 1,
  clinic_id: 1,
  pet_id: 10,
  owner_id: 20,
  start_date: "2026-03-25T00:00:00Z",
  end_date: "2026-03-28T00:00:00Z",
  status: "admitted",
  hospitalization_type: "hospitalization",
  memo: "",
  owner_request: "",
  staff_notes: "",
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

describe("transformHospitalization (HospitalizationResponse wire)", () => {
  it("id を string に変換する", () => {
    expect(transformHospitalization({ ...minimal, id: 99 }).id).toBe("99");
  });

  it("status: admitted → '入院中'", () => {
    expect(transformHospitalization({ ...minimal, status: "admitted" }).status).toBe("入院中");
  });

  it("status: discharged → '退院済'", () => {
    expect(transformHospitalization({ ...minimal, status: "discharged" }).status).toBe("退院済");
  });

  it("status: reserved → '予約'", () => {
    expect(transformHospitalization({ ...minimal, status: "reserved" }).status).toBe("予約");
  });

  it("未知の status は '予約' へ推測せず '不明' で fail-closed する (BUG-009)", () => {
    expect(transformHospitalization({ ...minimal, status: "unknown" }).status).toBe("不明");
  });

  it("空文字 status も '不明' で fail-closed する (BUG-009)", () => {
    expect(transformHospitalization({ ...minimal, status: "" }).status).toBe("不明");
  });

  it("hospitalization_type: hospitalization → '入院'", () => {
    expect(transformHospitalization({ ...minimal, hospitalization_type: "hospitalization" }).hospitalizationType).toBe("入院");
  });

  it("hospitalization_type: hotel → 'ホテル'", () => {
    expect(transformHospitalization({ ...minimal, hospitalization_type: "hotel" }).hospitalizationType).toBe("ホテル");
  });

  it("未知の hospitalization_type は '入院' にフォールバックする", () => {
    expect(transformHospitalization({ ...minimal, hospitalization_type: "other" }).hospitalizationType).toBe("入院");
  });

  it("start_date を整形する", () => {
    const result = transformHospitalization({ ...minimal, start_date: "2026-03-25T00:00:00Z" });
    expect(result.startDate).toBeTruthy();
  });

  it("cage_id を string に変換して cageId にマップする", () => {
    expect(transformHospitalization({ ...minimal, cage_id: 5 }).cageId).toBe("5");
  });

  it("cage_id が未設定のとき cageId は undefined", () => {
    expect(transformHospitalization({ ...minimal, cage_id: undefined }).cageId).toBeUndefined();
  });

  it("PetSummaryResponse.name を petName にマップする", () => {
    const result = transformHospitalization({
      ...minimal,
      pet: { id: 10, name: "ポチ", pet_number: "P-10" },
    });
    expect(result.petName).toBe("ポチ");
  });

  it("OwnerSummaryResponse.name を ownerName にマップする", () => {
    const result = transformHospitalization({
      ...minimal,
      owner: { id: 20, name: "田中" },
    });
    expect(result.ownerName).toBe("田中");
  });

  it("wire に treatment_plans フィールドが無い（models.Hospitalization リーク防止）", () => {
    // satisfies BackendHospitalization で treatment_plans を載せると type error になることを
    // 実行時でも「transform 入力に treatment_plans が無い」ことを固定する。
    expect("treatment_plans" in minimal).toBe(false);
    expect(transformHospitalization(minimal).id).toBe("1");
  });

  it("insurance_company_name / insurance_number は wire 上 optional で保持される", () => {
    const withIns: BackendHospitalization = {
      ...minimal,
      insurance_company_name: "ペット保険",
      insurance_number: "INS-1",
    };
    // transform は list UI 用で保険フィールドを載せないが、raw 型としては存在する
    expect(withIns.insurance_company_name).toBe("ペット保険");
    expect(withIns.insurance_number).toBe("INS-1");
  });
});
