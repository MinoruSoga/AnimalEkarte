import { describe, it, expect } from "vitest";

import {
  mapPetStatusLabel,
  transformBackendPetToFrontend,
  transformCreatePetRequest,
  transformUpdatePetRequest,
} from "./pet";
import type { PetResponse } from "@/types/generated/pet-responses";

// makeBackendPet は transformBackendPetToFrontend に渡す最小の PetResponse を組み立てる。
function makeBackendPet(overrides: Partial<PetResponse> = {}): PetResponse {
  return {
    id: 7,
    version: 1,
    clinic_id: 1,
    owner_id: 42,
    animal_species_id: 1,
    pet_number: "42-1",
    name: "ポチ",
    pet_name_kana: "ぽち",
    gender: "male",
    status: "alive",
    breed: "",
    color: "",
    danger_level: "low",
    food: "",
    environment: "",
    phone: "",
    remarks: "",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    ...overrides,
  };
}

describe("transformBackendPetToFrontend", () => {
  it("last_visit を birth_date / neutered_date と同じ『日付のみ』形式へ正規化する", () => {
    const pet = transformBackendPetToFrontend(
      makeBackendPet({
        birth_date: "2015-04-14T00:00:00Z",
        neutered_date: "2016-05-20T00:00:00Z",
        last_visit: "2024-08-25T00:00:00Z",
      }),
    );

    expect(pet.birthDate).toBe("2015-04-14");
    expect(pet.neuteredDate).toBe("2016-05-20");
    // #158: 以前は ISO datetime をそのまま通す非対称があった。兄弟の日付フィールドと揃える。
    expect(pet.lastVisit).toBe("2024-08-25");
  });

  it("日付フィールドが未設定なら undefined のまま（" - " 化やゼロ値で潰さない）", () => {
    const pet = transformBackendPetToFrontend(
      makeBackendPet({ birth_date: undefined, neutered_date: undefined, last_visit: undefined }),
    );

    expect(pet.birthDate).toBeUndefined();
    expect(pet.neuteredDate).toBeUndefined();
    expect(pet.lastVisit).toBeUndefined();
  });

  it("血液型 / マイクロチップ番号をマッピングする", () => {
    const pet = transformBackendPetToFrontend(
      makeBackendPet({ blood_type: "DEA1.1陽性", microchip_number: "392140000123456" }),
    );

    expect(pet.bloodType).toBe("DEA1.1陽性");
    expect(pet.microchipNumber).toBe("392140000123456");
  });

  it("血液型 / マイクロチップ番号 未設定は undefined（捏造しない）", () => {
    const pet = transformBackendPetToFrontend(makeBackendPet());

    expect(pet.bloodType).toBeUndefined();
    expect(pet.microchipNumber).toBeUndefined();
  });

  it("作成リクエストへ血液型 / マイクロチップ番号を含める", () => {
    const request = transformCreatePetRequest({
      ownerId: "42",
      name: "ポチ",
      animalSpeciesId: "1",
      bloodType: "DEA1.1陽性",
      microchipNumber: "392140000123456",
    });

    expect(request.blood_type).toBe("DEA1.1陽性");
    expect(request.microchip_number).toBe("392140000123456");
  });

  it("更新リクエストへ血液型 / マイクロチップ番号を含める", () => {
    const request = transformUpdatePetRequest({
      bloodType: "B",
      microchipNumber: "900000000000001",
    });

    expect(request.blood_type).toBe("B");
    expect(request.microchip_number).toBe("900000000000001");
  });

  // PR#186 P2-2 Bug#1 回帰テスト: deceased_at は response DTO への追加のみで
  // transform 層の配線が漏れると、値が API から届いても UI に渡らない。
  // BUG-003: deceased_reason も staff PetResponse 経由で deceasedReason へ。
  it("deceased_at を deceasedAt へマッピングする", () => {
    const pet = transformBackendPetToFrontend(
      makeBackendPet({
        status: "deceased",
        deceased_at: "2026-07-10T12:00:00+09:00",
        deceased_reason: "老衰",
      }),
    );

    expect(pet.status).toBe("死亡");
    expect(pet.deceasedAt).toBe("2026-07-10T12:00:00+09:00");
    expect(pet.deceasedReason).toBe("老衰");
  });

  it("deceased_at 未設定（生存中）は undefined のまま（捏造しない）", () => {
    const pet = transformBackendPetToFrontend(makeBackendPet());

    expect(pet.deceasedAt).toBeUndefined();
    expect(pet.deceasedReason).toBeUndefined();
  });

  it("未知statusは生存へ推測せず不明にする", () => {
    const pet = transformBackendPetToFrontend(makeBackendPet({ status: "unexpected" }));

    expect(pet.status).toBe("不明");
  });

  it.each([
    ["unexpected", "未知値"],
    ["constructor", "Object prototype key"],
    ["toString", "Object prototype method"],
    ["__proto__", "Object prototype accessor"],
    ["", "空文字"],
    [null, "null"],
    [undefined, "未指定"],
  ])("API境界のstatus %s は不明にする（%s）", (status) => {
    expect(mapPetStatusLabel(status)).toBe("不明");
  });

  // BUG-415: generic PATCH /pets/:id 経由の status 書込は deceased_at・監査ログと
  // 無結合のため除去した。status 変更は監査付きの死亡登録/取消
  // (PetDeceasedRecordButton → /:id/death)に一本化済み。
  // このテストは修正前は落ちていた(RED): 旧実装は status を無条件送信していたため
  // request.status が "alive" になり toBeUndefined() は失敗していた。
  it("更新リクエストは status を送信しない（死亡/復活は /:id/death に一本化）", () => {
    const request = transformUpdatePetRequest({
      name: "ポチ",
      status: "alive",
    });

    expect(request.status).toBeUndefined();
  });

  it("pet_name_kana を petNameKana へマッピングする（models.Pet の name_kana ではない）", () => {
    const pet = transformBackendPetToFrontend(makeBackendPet({ pet_name_kana: "ぽちたろう" }));
    expect(pet.petNameKana).toBe("ぽちたろう");
  });
});

describe("transformUpdatePetRequest", () => {
  it("既存の危険理由をクリアすると danger_reason を null として送信する", () => {
    const request = transformUpdatePetRequest({
      dangerReason: "",
      originalDangerReason: "咬傷歴あり",
    });

    expect(Object.prototype.hasOwnProperty.call(request, "danger_reason")).toBe(true);
    expect(request.danger_reason).toBeNull();
  });

  it("危険理由が未変更なら danger_reason を送信しない", () => {
    const request = transformUpdatePetRequest({
      dangerReason: "咬傷歴あり",
      originalDangerReason: "咬傷歴あり",
    });

    expect(Object.prototype.hasOwnProperty.call(request, "danger_reason")).toBe(false);
  });
});
