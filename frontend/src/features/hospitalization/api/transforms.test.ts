import { describe, it, expect } from "vitest";
import { transformHospitalization } from "./transforms";
import type { BackendHospitalization } from "./types";

const minimal: BackendHospitalization = {
  id: 1,
  clinic_id: 1,
  pet_id: 10,
  owner_id: 20,
  start_date: "2026-03-25T00:00:00Z",
  end_date: "2026-03-28T00:00:00Z",
  status: "admitted",
  hospitalization_type: "hospitalization",
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

describe("transformHospitalization", () => {
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

  it("未知の status は '予約' にフォールバックする", () => {
    expect(transformHospitalization({ ...minimal, status: "unknown" as "admitted" }).status).toBe("予約");
  });

  it("hospitalization_type: hospitalization → '入院'", () => {
    expect(transformHospitalization({ ...minimal, hospitalization_type: "hospitalization" }).hospitalizationType).toBe("入院");
  });

  it("hospitalization_type: hotel → 'ホテル'", () => {
    expect(transformHospitalization({ ...minimal, hospitalization_type: "hotel" }).hospitalizationType).toBe("ホテル");
  });

  it("未知の hospitalization_type は '入院' にフォールバックする", () => {
    expect(transformHospitalization({ ...minimal, hospitalization_type: "other" as "hotel" }).hospitalizationType).toBe("入院");
  });

  it("start_date を整形する", () => {
    // formatDate の実装依存だが、少なくとも空でないことを確認
    const result = transformHospitalization({ ...minimal, start_date: "2026-03-25T00:00:00Z" });
    expect(result.startDate).toBeTruthy();
  });

  it("cage_id を string に変換して cageId にマップする", () => {
    expect(transformHospitalization({ ...minimal, cage_id: 5 }).cageId).toBe("5");
  });

  it("cage_id が未設定のとき cageId は undefined", () => {
    expect(transformHospitalization({ ...minimal, cage_id: undefined }).cageId).toBeUndefined();
  });

  it("pet.name を petName にマップする", () => {
    const result = transformHospitalization({
      ...minimal,
      pet: { id: 10, clinic_id: 1, name: "ポチ" } as BackendHospitalization["pet"],
    });
    expect(result.petName).toBe("ポチ");
  });

  it("owner.name を ownerName にマップする", () => {
    const result = transformHospitalization({
      ...minimal,
      owner: { id: 20, clinic_id: 1, name: "田中" } as BackendHospitalization["owner"],
    });
    expect(result.ownerName).toBe("田中");
  });
});
