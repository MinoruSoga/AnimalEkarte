import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/lib/axios", () => ({
  axios: { get: vi.fn() },
}));

import { axios } from "@/lib/axios";
import { ownersLoader } from "./loaders";

const mockedGet = vi.mocked(axios.get);

// #266: owners-pets-list-plan.md の PO 決定により GET /v1/pets（ペット行粒度）へ再設計。
// owner 情報は petListResponse.owner サマリから直接埋め込まれる（飼主→ペットのフラット化は不要）。
describe("ownersLoader — #266 pets API (ペット行粒度)", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it("pet 行を Pet 型へ変換する（owner サマリ埋め込み）", async () => {
    mockedGet.mockResolvedValue({
      data: {
        data: [
          {
            id: 1,
            clinic_id: 1,
            owner_id: 5,
            animal_species_id: 2,
            pet_number: "P-001",
            name: "ポチ",
            pet_name_kana: "ポチ",
            gender: "male",
            status: "alive",
            breed: "柴犬",
            color: "茶",
            danger_level: "low",
            food: "ドライ",
            environment: "室内",
            remarks: "",
            owner: {
              id: 5,
              owner_number: 5,
              name: "山田太郎",
              name_kana: "ヤマダ タロウ",
              phone: "090-0000-0000",
              is_dangerous: false,
            },
            animal_species: { id: 2, name: "犬", sort_order: 1 },
          },
        ],
        total: 1,
        page: 1,
        limit: 20,
      },
    });

    const result = await ownersLoader({
      request: new Request("http://localhost/owners"),
    });

    expect(result.pets).toHaveLength(1);
    const pet = result.pets[0];
    expect(pet.id).toBe("1");
    expect(pet.clinicId).toBe("1");
    expect(pet.ownerId).toBe("5");
    expect(pet.ownerNumber).toBe(5);
    expect(pet.ownerName).toBe("山田太郎");
    expect(pet.ownerNameKana).toBe("ヤマダ タロウ");
    expect(pet.phone).toBe("090-0000-0000");
    expect(pet.name).toBe("ポチ");
    expect(pet.species).toBe("犬");
    expect(pet.animalSpeciesId).toBe("2");
    expect(pet.status).toBe("生存");
    expect(pet.deceasedAt).toBeUndefined();
  });

  it("owner が無い pet 行でもクラッシュせず安全な既定値になる", async () => {
    mockedGet.mockResolvedValue({
      data: {
        data: [
          {
            id: 2,
            clinic_id: 1,
            owner_id: 9,
            animal_species_id: 1,
            pet_number: "",
            name: "タマ",
            pet_name_kana: "",
            gender: "female",
            status: "alive",
            breed: "",
            color: "",
            danger_level: "low",
            food: "",
            environment: "",
            remarks: "",
          },
        ],
        total: 1,
        page: 1,
        limit: 20,
      },
    });

    const result = await ownersLoader({ request: new Request("http://localhost/owners") });

    expect(result.pets[0].ownerName).toBe("");
    expect(result.pets[0].ownerNumber).toBeUndefined();
    expect(result.pets[0].phone).toBe("");
  });
});

// #266: 白画面バグの根治確認 — 旧実装は total からページ数を計算して全ページを
// Promise.all で並列取得していた（10,370件規模で103並列リクエストのスラッシングが発生）。
// 新実装は1ページだけを取得すること、および URL の page/search/species/include_deceased を
// そのまま backend に転送することを保証する。
describe("ownersLoader — #266 サーバサイドページネーション", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it("total が大きくてもリクエストは1回だけ発火する（全ページ取得の回帰防止）", async () => {
    mockedGet.mockResolvedValue({
      data: { data: [], total: 10370, page: 1, limit: 20 },
    });

    await ownersLoader({ request: new Request("http://localhost/owners") });

    expect(mockedGet).toHaveBeenCalledTimes(1);
  });

  it("URL の page/search/species/include_deceased/clinics を backend にそのまま転送する", async () => {
    mockedGet.mockResolvedValue({
      data: { data: [], total: 0, page: 3, limit: 20 },
    });

    await ownersLoader({
      request: new Request(
        "http://localhost/owners?page=3&search=%E7%94%B0%E4%B8%AD&species=2&include_deceased=true&clinics=1,2"
      ),
    });

    expect(mockedGet).toHaveBeenCalledWith("/v1/pets", {
      params: {
        clinic_ids: "1,2",
        page: 3,
        limit: 20,
        search: "田中",
        species: "2",
        include_deceased: "true",
      },
    });
  });

  it("include_deceased が未指定の場合は backend にパラメータを送らない（既定=生存のみ）", async () => {
    mockedGet.mockResolvedValue({
      data: { data: [], total: 0, page: 1, limit: 20 },
    });

    await ownersLoader({ request: new Request("http://localhost/owners") });

    expect(mockedGet).toHaveBeenCalledWith("/v1/pets", {
      params: { page: 1, limit: 20 },
    });
  });

  it("page/limit/total を loader data として返す", async () => {
    mockedGet.mockResolvedValue({
      data: { data: [], total: 42, page: 2, limit: 20 },
    });

    const result = await ownersLoader({
      request: new Request("http://localhost/owners?page=2"),
    });

    expect(result.page).toBe(2);
    expect(result.limit).toBe(20);
    expect(result.total).toBe(42);
  });
});
