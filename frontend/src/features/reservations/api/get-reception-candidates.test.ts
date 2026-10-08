import { beforeEach, describe, expect, it } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { getReceptionCandidates } from "./get-reception-candidates";

const row = (id: number, status = "confirmed") => ({
  id,
  status,
  pet_id: 10,
  owner_id: 20,
  start_time: "2026-10-08T01:00:00Z",
});
describe("reception candidate fresh API reads", () => {
  beforeEach(() =>
    server.use(
      http.get("*/v1/pets/10", () => HttpResponse.json({ status: "alive", owner_id: 20 })),
    ),
  );
  it.each([
    { status: "deceased", owner_id: 20 },
    { status: "alive", owner_id: 20, deceased_at: "2026-10-08" },
    { status: "unknown", owner_id: 20 },
    { status: "alive", owner_id: 21 },
  ])("fails closed when fresh pet state changed", async (pet) => {
    server.use(http.get("*/v1/pets/10", () => HttpResponse.json(pet)));
    await expect(getReceptionCandidates("2026-10-08", "10", "20")).rejects.toThrow("再読み込み");
  });
  it("reads every page and excludes closed visits and tentative entries", async () => {
    const pages: string[] = [];
    server.use(
      http.get("*/v1/reservations", ({ request }) => {
        const params = new URL(request.url).searchParams;
        expect(params.get("pet_id")).toBe("10");
        expect(params.get("owner_id")).toBe("20");
        expect(params.get("date")).toBe("2026-10-08");
        pages.push(params.get("page") ?? "");
        return HttpResponse.json({
          data:
            params.get("page") === "1"
              ? [row(1), row(2, "completed")]
              : [row(3, "tentative"), row(4, "cancelled")],
          total: 4,
        });
      }),
    );
    expect(await getReceptionCandidates("2026-10-08", "10", "20")).toEqual([row(1)]);
    expect(pages).toEqual(["1", "2"]);
  });
  it.each([NaN, -1, 1.5, undefined])("rejects malformed totals", async (total) => {
    server.use(http.get("*/v1/reservations", () => HttpResponse.json({ data: [], total })));
    await expect(getReceptionCandidates("2026-10-08", "10", "20")).rejects.toThrow("件数");
  });
  it.each([
    { ...row(1), pet_id: 11 },
    { ...row(1), owner_id: 21 },
    { ...row(1), start_time: "2026-10-09T01:00:00Z" },
  ])("rejects other date or patient rows", async (other) => {
    server.use(http.get("*/v1/reservations", () => HttpResponse.json({ data: [other], total: 1 })));
    await expect(getReceptionCandidates("2026-10-08", "10", "20")).rejects.toThrow("対象");
  });
  it("does not treat incomplete pagination as no booking", async () => {
    server.use(http.get("*/v1/reservations", () => HttpResponse.json({ data: [], total: 1 })));
    await expect(getReceptionCandidates("2026-10-08", "10", "20")).rejects.toThrow("取得");
  });
  it("propagates a denied lookup instead of creating", async () => {
    server.use(
      http.get("*/v1/reservations", () =>
        HttpResponse.json({ error: "forbidden" }, { status: 403 }),
      ),
    );
    await expect(getReceptionCandidates("2026-10-08", "10", "20")).rejects.toThrow();
  });
});
