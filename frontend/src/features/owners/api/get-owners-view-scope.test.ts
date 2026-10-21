import { describe, expect, it } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import {
  OwnersViewScopeContractError,
  getOwnersViewScope,
  parseOwnersViewScope,
} from "./get-owners-view-scope";

// EMR-463: GET /me owners_view_clinic_ids の必須 parse（technical-spec#owners-list-scope-contract-emr463
// fixture 表「field欠落／NULL／数値ID／"0"／"01"／重複」「"9007199254740993"」）。
// 値はすべて合成 fixture。
function meBody(overrides: Record<string, unknown> = {}): Record<string, unknown> {
  return {
    id: "10",
    clinics: [
      { clinic_id: "1", clinic_name: "合成本院", is_main: true },
      { clinic_id: "2", clinic_name: "合成分院", is_main: false },
      { clinic_id: "4", clinic_name: "合成第三院", is_main: false },
    ],
    owners_view_clinic_ids: ["1", "2"],
    ...overrides,
  };
}

describe("parseOwnersViewScope — 必須 field の契約", () => {
  it("許可 ID を文字列のまま返し、表示名は同じ応答の clinics から許可医院だけ引く", () => {
    const scope = parseOwnersViewScope(meBody());
    expect(scope.userId).toBe("10");
    expect(scope.permittedClinicIds).toEqual(["1", "2"]);
    // clinics に 4 があっても許可集合へ追加しない
    expect(scope.clinics).toEqual([
      { clinicId: "1", clinicName: "合成本院", isMain: true },
      { clinicId: "2", clinicName: "合成分院", isMain: false },
    ]);
  });

  it("許可ゼロ [] は取得成功として空集合を返す（エラーと区別）", () => {
    const scope = parseOwnersViewScope(meBody({ owners_view_clinic_ids: [] }));
    expect(scope.permittedClinicIds).toEqual([]);
    expect(scope.clinics).toEqual([]);
  });

  it('"9007199254740993" を Number 化せず精度を保つ', () => {
    const scope = parseOwnersViewScope(
      meBody({
        clinics: [{ clinic_id: "9007199254740993", clinic_name: "合成大ID院", is_main: true }],
        owners_view_clinic_ids: ["9007199254740993"],
      }),
    );
    expect(scope.permittedClinicIds).toEqual(["9007199254740993"]);
    expect(scope.clinics[0].clinicId).toBe("9007199254740993");
  });

  it("clinics に名称の無い許可 ID は ID 表示で残し、許可集合から落とさない", () => {
    const scope = parseOwnersViewScope(meBody({ owners_view_clinic_ids: ["1", "7"] }));
    expect(scope.permittedClinicIds).toEqual(["1", "7"]);
    expect(scope.clinics.map((c) => c.clinicId)).toEqual(["1", "7"]);
  });

  it.each([
    ["field 欠落", { owners_view_clinic_ids: undefined }],
    ["NULL", { owners_view_clinic_ids: null }],
    ["非配列", { owners_view_clinic_ids: "1,2" }],
    ["数値 ID", { owners_view_clinic_ids: [1, 2] }],
    ['"0"', { owners_view_clinic_ids: ["0"] }],
    ['"01"', { owners_view_clinic_ids: ["01"] }],
    ["負数", { owners_view_clinic_ids: ["-1"] }],
    ["空白入り", { owners_view_clinic_ids: [" 1"] }],
    ["重複", { owners_view_clinic_ids: ["1", "1"] }],
    ["id 欠落", { id: undefined }],
  ])("%s は契約エラー（[] へ補完しない）", (_label, overrides) => {
    // undefined の上書きは JSON 上の欠落と同じ扱い（wire では key が無い）
    expect(() => parseOwnersViewScope(meBody(overrides))).toThrow(OwnersViewScopeContractError);
  });

  it("応答が object でなければ契約エラー", () => {
    expect(() => parseOwnersViewScope(null)).toThrow(OwnersViewScopeContractError);
    expect(() => parseOwnersViewScope([])).toThrow(OwnersViewScopeContractError);
  });
});

describe("getOwnersViewScope — owners 専有の GET /v1/me 取得", () => {
  it("呼ぶたびに GET /v1/me を取得する（全体 /me cache を流用しない）", async () => {
    let calls = 0;
    server.use(
      http.get("*/v1/me", () => {
        calls += 1;
        return HttpResponse.json(meBody());
      }),
    );
    await getOwnersViewScope();
    await getOwnersViewScope();
    expect(calls).toBe(2);
  });

  it("field 欠落の応答は取得エラーにする", async () => {
    const body = meBody();
    delete body.owners_view_clinic_ids;
    server.use(http.get("*/v1/me", () => HttpResponse.json(body)));
    await expect(getOwnersViewScope()).rejects.toBeInstanceOf(OwnersViewScopeContractError);
  });
});
