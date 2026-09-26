import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/lib/axios", () => ({
  axios: { get: vi.fn() },
}));

import { axios } from "@/lib/axios";
import { ownerLoader, ownersLoader } from "./loaders";

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
            danger_reason: "保定時に噛む",
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
    expect(pet.dangerReason).toBe("保定時に噛む");
    expect(pet.deceasedAt).toBeUndefined();
  });

  it("owner.is_dangerous=true を ownerIsDangerous へマッピングし、false はキーを出さない", async () => {
    const makeRow = (id: number, isDangerous: boolean) => ({
      id,
      clinic_id: 1,
      owner_id: 5 + id,
      animal_species_id: 2,
      pet_number: `P-00${id}`,
      name: `ペット${id}`,
      pet_name_kana: "",
      gender: "male",
      status: "alive",
      breed: "",
      color: "",
      danger_level: "low",
      food: "",
      environment: "",
      remarks: "",
      owner: {
        id: 5 + id,
        owner_number: 5 + id,
        name: `飼主${id}`,
        name_kana: "",
        phone: "",
        is_dangerous: isDangerous,
      },
    });
    mockedGet.mockResolvedValue({
      data: {
        data: [makeRow(1, true), makeRow(2, false)],
        total: 2,
        page: 1,
        limit: 20,
      },
    });

    const result = await ownersLoader({ request: new Request("http://localhost/owners") });

    expect(result.pets[0].ownerIsDangerous).toBe(true);
    // EMR-173: is_dangerous=false は危険人物ではなく、未設定と同じくキー自体を出さない。
    expect(result.pets[1].ownerIsDangerous).toBeUndefined();
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

  it("死亡は死亡のまま、未知・null status は生存へ推測せず不明へ変換する", async () => {
    const makeRow = (id: number, status: string | null) => ({
      id,
      clinic_id: 1,
      owner_id: 9,
      animal_species_id: 1,
      pet_number: `P-${id}`,
      name: `合成ペット${id}`,
      pet_name_kana: "ゴウセイ",
      gender: "unknown",
      status,
      breed: "",
      color: "",
      danger_level: "low",
      food: "",
      environment: "",
      remarks: "",
    });
    mockedGet.mockResolvedValue({
      data: {
        data: [makeRow(1, "deceased"), makeRow(2, "unexpected"), makeRow(3, null)],
        total: 3,
        page: 1,
        limit: 20,
      },
    });

    const result = await ownersLoader({ request: new Request("http://localhost/owners") });

    expect(result.pets.map((pet) => pet.status)).toEqual(["死亡", "不明", "不明"]);
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
        "http://localhost/owners?page=3&search=%E7%94%B0%E4%B8%AD&species=2&include_deceased=true&clinics=1,2",
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

// 回帰防止: 旧実装は `} catch { throw new Response(..., { status: 500 }) }` で
// 上流のHTTPステータスを握り潰していたため、GET /v1/pets の 400（DB スキーマ不整合等）が
// errorElement 側では 500 として見え、原因の切り分けを不可能にしていた。
describe("ownersLoader — 上流ステータスの保全", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it.each([
    [400, "リクエスト不正"],
    [403, "権限不足"],
    [404, "未検出"],
  ])("上流の%dを500へ潰さずそのまま伝播する", async (status, _label) => {
    mockedGet.mockRejectedValue({
      isAxiosError: true,
      response: { status },
    });

    const thrown = await ownersLoader({
      request: new Request("http://localhost/owners"),
    }).then(
      () => undefined,
      (err: unknown) => err,
    );

    expect(thrown).toBeInstanceOf(Response);
    expect((thrown as Response).status).toBe(status);
    expect((thrown as Response).status).not.toBe(500);
    await expect((thrown as Response).text()).resolves.toBe("飼主一覧の取得に失敗しました");
  });

  it("response を持たない通信エラー（ネットワーク断）は500になる", async () => {
    mockedGet.mockRejectedValue({ isAxiosError: true, response: undefined });

    const thrown = await ownersLoader({
      request: new Request("http://localhost/owners"),
    }).then(
      () => undefined,
      (err: unknown) => err,
    );

    expect((thrown as Response).status).toBe(500);
  });

  it("axios 由来でない例外は500になる", async () => {
    mockedGet.mockRejectedValue(new Error("boom"));

    const thrown = await ownersLoader({
      request: new Request("http://localhost/owners"),
    }).then(
      () => undefined,
      (err: unknown) => err,
    );

    expect(thrown).toBeInstanceOf(Response);
    expect((thrown as Response).status).toBe(500);
  });
});

// PERF-E5-N1-PETS: ownerLoader の N+1 解消 — GET /v1/owners/:id に加えて
// GET /v1/pets?owner_id=<id>&include_deceased=true を1回だけ発行する。
// 旧実装は owner.pets の各 id に GET /v1/pets/{id} を並列発行していた（1+N）。
describe("ownerLoader — PERF-E5-N1-PETS ペット一括取得", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  const ownerApiResponse = {
    id: 7,
    clinic_id: 1,
    owner_name: "山田太郎",
    pets: [{ id: 101 }, { id: 102 }],
  };

  const petsListResponse = {
    data: {
      data: [
        {
          id: 101,
          clinic_id: 1,
          owner_id: 7,
          animal_species_id: 2,
          pet_number: "P-101",
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
        },
        {
          id: 102,
          clinic_id: 1,
          owner_id: 7,
          animal_species_id: 2,
          pet_number: "P-102",
          name: "タマ",
          pet_name_kana: "タマ",
          gender: "female",
          status: "deceased",
          breed: "",
          color: "",
          danger_level: "low",
          food: "",
          environment: "",
          remarks: "",
          deceased_at: "2026-07-10T03:00:00+09:00",
          deceased_reason: "老衰",
        },
      ],
      total: 2,
      page: 1,
      limit: 100,
    },
  };

  it("飼主1件の取得でペットは1回の list リクエストのみ発火する（N+1 回帰防止）", async () => {
    mockedGet
      .mockResolvedValueOnce({ data: ownerApiResponse })
      .mockResolvedValueOnce(petsListResponse);

    const result = await ownerLoader({ params: { id: "7" } });

    expect(mockedGet).toHaveBeenCalledTimes(2);
    expect(mockedGet).toHaveBeenNthCalledWith(1, "/v1/owners/7");
    expect(mockedGet).toHaveBeenNthCalledWith(
      2,
      "/v1/pets",
      expect.objectContaining({
        params: expect.objectContaining({
          owner_id: "7",
          include_deceased: "true",
          clinic_ids: "1",
        }),
      }),
    );
    // GET /v1/pets/{id} の個別取得が残っていないことを保証
    const perPetCalls = mockedGet.mock.calls.filter(([url]) =>
      /^\/v1\/pets\/\d+$/.test(url as string),
    );
    expect(perPetCalls).toHaveLength(0);

    expect(result.owner.pets).toHaveLength(2);
    // list 応答順（pets.id ASC）をそのまま保持する
    expect(result.owner.pets?.map((p) => p.id)).toEqual(["101", "102"]);
  });

  it("list 行の deceased_at / deceased_reason を deceasedAt / deceasedReason にマップする", async () => {
    mockedGet
      .mockResolvedValueOnce({ data: ownerApiResponse })
      .mockResolvedValueOnce(petsListResponse);

    const result = await ownerLoader({ params: { id: "7" } });

    const deceased = result.owner.pets?.[1];
    expect(deceased?.status).toBe("死亡");
    expect(deceased?.deceasedAt).toBe("2026-07-10T03:00:00+09:00");
    expect(deceased?.deceasedReason).toBe("老衰");
  });

  // BUG-010 回帰防止: 選択中医院と異なる医院の飼主でも、飼主の所属医院を
  // clinic_ids へ渡してペット一覧が空にならないことを保証する。
  it("他医院所属の飼主では clinic_ids に飼主の clinicId を渡す（拠点横断回帰）", async () => {
    mockedGet
      .mockResolvedValueOnce({ data: { ...ownerApiResponse, clinic_id: 2 } })
      .mockResolvedValueOnce(petsListResponse);

    const result = await ownerLoader({ params: { id: "7" } });

    expect(mockedGet).toHaveBeenNthCalledWith(
      2,
      "/v1/pets",
      expect.objectContaining({
        params: expect.objectContaining({ clinic_ids: "2" }),
      }),
    );
    expect(result.owner.pets).toHaveLength(2);
  });

  // limit(=100) 超過の飼主: total が初回取得件数を超える場合は残ページを
  // 順次取得して全件を保持する（打ち切りでペットを失わない）。
  it("total が1ページを超える飼主は残ページを取得して全件保持する", async () => {
    const makeRow = (id: number) => ({
      id,
      clinic_id: 1,
      owner_id: 7,
      animal_species_id: 2,
      pet_number: `P-${id}`,
      name: `ペット${id}`,
      pet_name_kana: "",
      gender: "unknown",
      status: "alive",
      breed: "",
      color: "",
      danger_level: "low",
      food: "",
      environment: "",
      remarks: "",
    });
    const page1Rows = Array.from({ length: 100 }, (_, i) => makeRow(i + 1));
    const page2Rows = Array.from({ length: 50 }, (_, i) => makeRow(i + 101));
    mockedGet
      .mockResolvedValueOnce({ data: ownerApiResponse })
      .mockResolvedValueOnce({
        data: { data: page1Rows, total: 150, page: 1, limit: 100 },
      })
      .mockResolvedValueOnce({
        data: { data: page2Rows, total: 150, page: 2, limit: 100 },
      });

    const result = await ownerLoader({ params: { id: "7" } });

    expect(mockedGet).toHaveBeenCalledTimes(3);
    expect(mockedGet).toHaveBeenNthCalledWith(
      3,
      "/v1/pets",
      expect.objectContaining({
        params: expect.objectContaining({ owner_id: "7", page: 2, limit: 100 }),
      }),
    );
    expect(result.owner.pets).toHaveLength(150);
    // ページ結合後も list 応答順（id 昇順）を保持する
    expect(result.owner.pets?.[0].id).toBe("1");
    expect(result.owner.pets?.[149].id).toBe("150");
  });

  it("ペット一覧取得の失敗は1回だけ表面化する（per-pet エラーの重複なし）", async () => {
    mockedGet
      .mockResolvedValueOnce({ data: ownerApiResponse })
      .mockRejectedValueOnce({ isAxiosError: true, response: { status: 403 } });

    const thrown = await ownerLoader({ params: { id: "7" } }).then(
      () => undefined,
      (err: unknown) => err,
    );

    expect(mockedGet).toHaveBeenCalledTimes(2);
    expect(thrown).toBeInstanceOf(Response);
    expect((thrown as Response).status).toBe(403);
  });
});

describe("ownerLoader — BUG-010 clinic mismatch 404", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it("GET /owners/:id が 404 のとき明示メッセージの Response を投げる", async () => {
    mockedGet.mockRejectedValue({
      isAxiosError: true,
      response: { status: 404 },
    });

    const thrown = await ownerLoader({ params: { id: "99" } }).then(
      () => undefined,
      (err: unknown) => err,
    );

    expect(thrown).toBeInstanceOf(Response);
    expect((thrown as Response).status).toBe(404);
    await expect((thrown as Response).text()).resolves.toContain("異なる医院");
  });

  it("id 未指定は 400", async () => {
    const thrown = await ownerLoader({ params: {} }).then(
      () => undefined,
      (err: unknown) => err,
    );
    expect((thrown as Response).status).toBe(400);
  });
});
