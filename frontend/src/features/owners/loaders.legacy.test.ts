import { beforeEach, describe, expect, it } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { OWNERS_CLINIC_SCOPE_ACTIVE } from "./lib/owners-clinic-scope-activation";
import { ownersLoader } from "./loaders";

// EMR-463: 非活性時（OWNERS_CLINIC_SCOPE_ACTIVE=false・既定）は PR 前（main 05da02f3）と同じ挙動。
// GET /me を呼ばず、?clinics= をそのまま clinic_ids へ透過する。値は合成のみ。

let meCalls = 0;
let petsClinicIds: (string | null)[] = [];

beforeEach(() => {
  meCalls = 0;
  petsClinicIds = [];
  window.localStorage.clear();
  server.use(
    http.get("*/v1/me", () => {
      meCalls += 1;
      return HttpResponse.json({ id: "10", clinics: [], owners_view_clinic_ids: ["1"] });
    }),
    http.get("*/v1/pets", ({ request }) => {
      petsClinicIds.push(new URL(request.url).searchParams.get("clinic_ids"));
      return HttpResponse.json({ data: [], total: 0, page: 1, limit: 20 });
    }),
  );
});

describe("ownersLoader — EMR-463 非活性時（PR 前と同じ）", () => {
  it("activation の既定値は false（463-BE 実装前は接続しない）", () => {
    expect(OWNERS_CLINIC_SCOPE_ACTIVE).toBe(false);
  });

  it("GET /me を呼ばず、?clinics= をそのまま clinic_ids へ透過する", async () => {
    const data = await ownersLoader({
      request: new Request("http://localhost/owners?clinics=1,99"),
    });

    expect(meCalls).toBe(0);
    // 許可外かもしれない 99 も含めて透過（所属検証は従来どおりサーバ側）
    expect(petsClinicIds).toEqual(["1,99"]);
    expect(data.listState).toEqual({ kind: "ready", scopeRechecked: false });
    expect(data.clinicScope).toBeNull();
  });

  it("?clinics= 未指定は clinic_ids を送らない（現在の医院のみ・従来挙動）", async () => {
    window.localStorage.setItem("owners-list.clinics.v1.10", JSON.stringify(["2"]));

    await ownersLoader({ request: new Request("http://localhost/owners") });

    expect(meCalls).toBe(0);
    expect(petsClinicIds).toEqual([null]);
    // 保存値は読まず、書き換えもしない
    expect(window.localStorage.getItem("owners-list.clinics.v1.10")).toBe('["2"]');
  });

  it("/pets の 403 は scope 再照合をせず、そのまま 403 Response にする", async () => {
    server.use(
      http.get("*/v1/pets", () => HttpResponse.json({ error: "forbidden" }, { status: 403 })),
    );

    const thrown = await ownersLoader({ request: new Request("http://localhost/owners") }).then(
      () => undefined,
      (err: unknown) => err,
    );

    expect((thrown as Response).status).toBe(403);
    expect(meCalls).toBe(0);
  });
});
