import { beforeEach, describe, expect, it, vi } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { ownersLoader } from "./loaders";

// EMR-463: このファイルは活性時（OWNERS_CLINIC_SCOPE_ACTIVE=true）の経路を検証する。
vi.mock("./lib/owners-clinic-scope-activation", () => ({ OWNERS_CLINIC_SCOPE_ACTIVE: true }));

// EMR-463: technical-spec#owners-list-scope-contract-emr463 の FE 受入ケース（契約 mock）。
// GET /v1/me の owners_view_clinic_ids は runtime 未実装（x-contract-status:
// selected-not-implemented）のため、共有 handlers.ts を変えずにこのファイル内の server.use で代替する。
// 値はすべて合成 ID・合成名称（実患者・実飼主の値は含めない）。

const USER_ID = "10";
const STORAGE_KEY = `owners-list.clinics.v1.${USER_ID}`;

type Call = { kind: "me" } | { kind: "pets"; clinicIds: string | null };

let calls: Call[] = [];

function meBody(permitted: unknown, userId = USER_ID) {
  const permittedIds = Array.isArray(permitted)
    ? permitted.filter((id): id is string => typeof id === "string")
    : [];
  return {
    id: userId,
    clinics: permittedIds.map((id, index) => ({
      clinic_id: id,
      clinic_name: `合成医院${id}`,
      is_main: index === 0,
    })),
    owners_view_clinic_ids: permitted,
  };
}

function petsBody() {
  return { data: [], total: 0, page: 1, limit: 20 };
}

/** scopeResponses を順に返す GET /v1/me と、clinic_ids を記録する GET /v1/pets。 */
function useHandlers({
  scopeResponses,
  petsStatuses = [],
}: {
  scopeResponses: (() => Response)[];
  petsStatuses?: number[];
}) {
  let meIndex = 0;
  let petsIndex = 0;
  server.use(
    http.get("*/v1/me", () => {
      calls.push({ kind: "me" });
      const respond = scopeResponses[Math.min(meIndex, scopeResponses.length - 1)];
      meIndex += 1;
      return respond();
    }),
    http.get("*/v1/pets", ({ request }) => {
      calls.push({
        kind: "pets",
        clinicIds: new URL(request.url).searchParams.get("clinic_ids"),
      });
      const status = petsStatuses[petsIndex] ?? 200;
      petsIndex += 1;
      return status === 200
        ? HttpResponse.json(petsBody())
        : HttpResponse.json({ error: "forbidden" }, { status });
    }),
  );
}

const scopeOk = (permitted: unknown, userId?: string) => () =>
  HttpResponse.json(meBody(permitted, userId));

const petsCalls = () =>
  calls.filter((c): c is Extract<Call, { kind: "pets" }> => c.kind === "pets");

function load(search = "") {
  return ownersLoader({ request: new Request(`http://localhost/owners${search}`) });
}

beforeEach(() => {
  calls = [];
  window.localStorage.clear();
});

describe("ownersLoader — scope 先行と有効選択", () => {
  it('初回・許可 ["1","2"] は全許可を選択し、scope 完了後にだけ /pets を呼ぶ', async () => {
    useHandlers({ scopeResponses: [scopeOk(["1", "2"])] });

    const data = await load();

    expect(calls).toEqual([{ kind: "me" }, { kind: "pets", clinicIds: "1,2" }]);
    expect(data.listState).toEqual({ kind: "ready", scopeRechecked: false });
    expect(data.clinicScope?.selectedClinicIds).toEqual(["1", "2"]);
    expect(data.clinicScope?.clinics.map((c) => c.clinicId)).toEqual(["1", "2"]);
  });

  it('保存 ["2","3"]・許可 ["1","2"] は交差 ["2"] だけを送り、保存値から失権 ID を除く', async () => {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(["2", "3"]));
    useHandlers({ scopeResponses: [scopeOk(["1", "2"])] });

    const data = await load();

    expect(petsCalls()).toEqual([{ kind: "pets", clinicIds: "2" }]);
    expect(data.clinicScope?.selectedClinicIds).toEqual(["2"]);
    expect(window.localStorage.getItem(STORAGE_KEY)).toBe('["2"]');
  });

  it("保存の全件が失効したら現在の全許可へ戻す（未検証の 3 を補わない）", async () => {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(["3"]));
    useHandlers({ scopeResponses: [scopeOk(["1", "2"])] });

    await load();

    expect(petsCalls()).toEqual([{ kind: "pets", clinicIds: "1,2" }]);
    expect(window.localStorage.getItem(STORAGE_KEY)).toBeNull();
  });

  it("URL に許可外 ID が混入しても交差外を送らない", async () => {
    useHandlers({ scopeResponses: [scopeOk(["1", "2"])] });

    await load("?clinics=2,99");

    expect(petsCalls()).toEqual([{ kind: "pets", clinicIds: "2" }]);
  });

  it("URL が許可外 ID だけなら clinic_ids を省略せず全許可を送る", async () => {
    useHandlers({ scopeResponses: [scopeOk(["1", "2"])] });

    await load("?clinics=99");

    expect(petsCalls()).toEqual([{ kind: "pets", clinicIds: "1,2" }]);
  });

  it("別ユーザーの保存 key は読まない", async () => {
    window.localStorage.setItem("owners-list.clinics.v1.11", JSON.stringify(["2"]));
    useHandlers({ scopeResponses: [scopeOk(["1", "2"])] });

    await load();

    expect(petsCalls()).toEqual([{ kind: "pets", clinicIds: "1,2" }]);
    expect(window.localStorage.getItem("owners-list.clinics.v1.11")).toBe('["2"]');
  });

  it('"9007199254740993" は scope → 選択保存 → clinic_ids まで文字列のまま', async () => {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(["9007199254740993", "3"]));
    useHandlers({ scopeResponses: [scopeOk(["9007199254740993", "1"])] });

    const data = await load();

    expect(petsCalls()).toEqual([{ kind: "pets", clinicIds: "9007199254740993" }]);
    expect(data.clinicScope?.selectedClinicIds).toEqual(["9007199254740993"]);
    expect(window.localStorage.getItem(STORAGE_KEY)).toBe('["9007199254740993"]');
  });
});

describe("ownersLoader — 許可ゼロ・上限", () => {
  it("許可 [] は /pets を呼ばず、取得失敗と区別した許可ゼロ状態を返す", async () => {
    useHandlers({ scopeResponses: [scopeOk([])] });

    const data = await load("?clinics=1");

    expect(petsCalls()).toEqual([]);
    expect(data.listState).toEqual({ kind: "no-permitted-clinics" });
    expect(data.total).toBe(0);
  });

  it("選択 50 件は上限内で 50 ID を送る", async () => {
    const ids = Array.from({ length: 50 }, (_, i) => String(i + 1));
    useHandlers({ scopeResponses: [scopeOk(ids)] });

    const data = await load();

    expect(data.listState.kind).toBe("ready");
    expect(petsCalls()).toHaveLength(1);
    expect(petsCalls()[0].clinicIds?.split(",")).toHaveLength(50);
  });

  it("選択 51 件は超過として /pets を呼ばない（切捨て・分割取得なし）", async () => {
    const ids = Array.from({ length: 51 }, (_, i) => String(i + 1));
    useHandlers({ scopeResponses: [scopeOk(ids)] });

    const data = await load();

    expect(petsCalls()).toEqual([]);
    expect(data.listState).toEqual({ kind: "selection-over-limit", selectedCount: 51, limit: 50 });
    expect(data.clinicScope?.selectedClinicIds).toHaveLength(51);
  });
});

describe("ownersLoader — scope 取得失敗", () => {
  it.each([
    ["503", () => HttpResponse.json({ error: "unavailable" }, { status: 503 })],
    ["500", () => HttpResponse.json({ error: "internal" }, { status: 500 })],
    ["field 欠落", () => HttpResponse.json({ id: USER_ID, clinics: [] })],
    ["NULL", () => HttpResponse.json({ id: USER_ID, clinics: [], owners_view_clinic_ids: null })],
    ["数値 ID", () => HttpResponse.json(meBody([1, 2]))],
    ['"0"', () => HttpResponse.json(meBody(["0"]))],
    ['"01"', () => HttpResponse.json(meBody(["01"]))],
    ["重複", () => HttpResponse.json(meBody(["1", "1"]))],
  ])("%s は一覧エラー。/pets を呼ばず、保存選択を消さない", async (_label, respond) => {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(["2"]));
    useHandlers({ scopeResponses: [respond] });

    const data = await load();

    expect(petsCalls()).toEqual([]);
    expect(data.listState).toEqual({ kind: "scope-error" });
    // 所属全件・現在医院・旧 scope へ fallback しない
    expect(data.clinicScope).toBeNull();
    expect(window.localStorage.getItem(STORAGE_KEY)).toBe('["2"]');
  });
});

describe("ownersLoader — /pets 403 後の再照合", () => {
  it("403 は scope を再取得し、照合し直した選択で 1 回だけ再送する", async () => {
    useHandlers({
      scopeResponses: [scopeOk(["1", "2"]), scopeOk(["1"])],
      petsStatuses: [403, 200],
    });

    const data = await load();

    expect(calls).toEqual([
      { kind: "me" },
      { kind: "pets", clinicIds: "1,2" },
      { kind: "me" },
      { kind: "pets", clinicIds: "1" },
    ]);
    expect(data.listState).toEqual({ kind: "ready", scopeRechecked: true });
    expect(data.clinicScope?.selectedClinicIds).toEqual(["1"]);
  });

  it("再取得しても同じ選択なら失敗した選択で再送せず 403 を返す", async () => {
    useHandlers({
      scopeResponses: [scopeOk(["1", "2"]), scopeOk(["2", "1"])],
      petsStatuses: [403, 200],
    });

    const thrown = await load().then(
      () => undefined,
      (err: unknown) => err,
    );

    expect(thrown).toBeInstanceOf(Response);
    expect((thrown as Response).status).toBe(403);
    expect(petsCalls()).toEqual([{ kind: "pets", clinicIds: "1,2" }]);
  });

  it("再取得で許可ゼロになったら /pets を再送せず許可ゼロ状態にする", async () => {
    useHandlers({ scopeResponses: [scopeOk(["1"]), scopeOk([])], petsStatuses: [403] });

    const data = await load();

    expect(petsCalls()).toHaveLength(1);
    expect(data.listState).toEqual({ kind: "no-permitted-clinics" });
  });

  it("再取得が失敗したら旧 scope で再送せず scope 取得エラーにする（保存値は消さない）", async () => {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(["1"]));
    useHandlers({
      scopeResponses: [
        scopeOk(["1", "2"]),
        () => HttpResponse.json({ error: "unavailable" }, { status: 503 }),
      ],
      petsStatuses: [403],
    });

    const data = await load();

    expect(petsCalls()).toHaveLength(1);
    expect(data.listState).toEqual({ kind: "scope-error" });
    expect(window.localStorage.getItem(STORAGE_KEY)).toBe('["1"]');
  });

  it("再送も 403 なら 2 回目で止めて 403 を返す（無限再照合しない）", async () => {
    useHandlers({
      scopeResponses: [scopeOk(["1", "2"]), scopeOk(["1"]), scopeOk(["2"])],
      petsStatuses: [403, 403, 200],
    });

    const thrown = await load().then(
      () => undefined,
      (err: unknown) => err,
    );

    expect((thrown as Response).status).toBe(403);
    expect(calls.filter((c) => c.kind === "me")).toHaveLength(2);
    expect(petsCalls()).toHaveLength(2);
  });
});

describe("ownersLoader — 旧要求の無効化", () => {
  it("中断済み request では /pets を呼ばず、保存値にも触れない", async () => {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(["2", "3"]));
    useHandlers({ scopeResponses: [scopeOk(["1", "2"])] });
    const controller = new AbortController();
    controller.abort();

    await expect(
      // loader が使うのは url と signal だけ。Request 実装差（jsdom/undici）を避けて最小形で渡す。
      ownersLoader({
        request: { url: "http://localhost/owners", signal: controller.signal } as unknown as Request,
      }),
    ).rejects.toBeDefined();

    expect(petsCalls()).toEqual([]);
    expect(window.localStorage.getItem(STORAGE_KEY)).toBe('["2","3"]');
  });
});
