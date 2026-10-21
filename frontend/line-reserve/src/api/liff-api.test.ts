import { describe, expect, it } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { liffApi } from "./liff-api";
import type { CreateReservationBody } from "../types/models";

const NULL_BYTE = String.fromCharCode(0);

describe("liffApi（R-F20: NULL バイトサニタイズ）", () => {
  it("createReservation の POST ボディから NULL バイトを除去して送信する", async () => {
    let capturedBody: unknown;
    server.use(
      http.post("/api/liff/:clinicId/reservations", async ({ request }) => {
        capturedBody = await request.json();
        return HttpResponse.json({ id: "1", notes: "R-20260801-0001" });
      }),
    );

    const body: CreateReservationBody = {
      course_id: 10,
      staff_id: 0,
      date: "2026-08-01",
      start_time: "1000",
      end_time: "1030",
      customer_fields: {
        customer_name: `山田${NULL_BYTE}花子`,
        phone: "090-1234-5678",
        owner_name: `山田${NULL_BYTE}太郎`,
        pets: [{ name: `ポチ${NULL_BYTE}`, type: "柴犬", is_new: false }],
      },
      request_text: `爪切り${NULL_BYTE}希望`,
    };

    await liffApi.createReservation("1", body, "test-id-token");

    expect(capturedBody).toEqual({
      course_id: 10,
      staff_id: 0,
      date: "2026-08-01",
      start_time: "1000",
      end_time: "1030",
      customer_fields: {
        customer_name: "山田花子",
        phone: "090-1234-5678",
        owner_name: "山田太郎",
        pets: [{ name: "ポチ", type: "柴犬", is_new: false }],
      },
      request_text: "爪切り希望",
    });
  });

  it("GET リクエストのボディサニタイズ処理は影響しない（GET には body が無い）", async () => {
    // FE5-18: 実行時検証（liffSettingsSchema）導入に伴い、models.ts 準拠の完全な
    // フィクスチャに是正（旧フィクスチャは GET がサニタイズの影響を受けないことのみを
    // 確認する仮データで、実際の LiffSettings 契約を表していなかった）。
    const settingsFixture = {
      liff_id: "123",
      header_text: "テスト病院",
      phone_number: "",
      status: "running",
      request_example: "",
      reservation_notice: "",
      cancel_notice: "",
      privacy_policy: "",
      show_no_staff_option: false,
      booking_window: 30,
    };
    server.use(http.get("/api/liff/:clinicId/settings", () => HttpResponse.json(settingsFixture)));

    const settings = await liffApi.getSettings("1");

    expect(settings).toEqual(settingsFixture);
  });
});

describe("liffApi（FE-RC-079: clinicId は URL パスセグメントとして encode される）", () => {
  it("clinicId に '/' を含む値を渡しても、単一の :clinicId セグメントとして解決される", async () => {
    // clinicId が encode されていないと "/api/liff/a/b/settings" のように
    // パスセグメントが分裂し、":clinicId/settings" にマッチしなくなる。
    let capturedClinicId: string | undefined;
    server.use(
      http.get("/api/liff/:clinicId/settings", ({ params }) => {
        capturedClinicId = params.clinicId as string;
        return HttpResponse.json({
          liff_id: "123",
          header_text: "テスト病院",
          phone_number: "",
          status: "running",
          request_example: "",
          reservation_notice: "",
          cancel_notice: "",
          privacy_policy: "",
          show_no_staff_option: false,
          booking_window: 30,
        });
      }),
    );

    await liffApi.getSettings("a/b");

    expect(capturedClinicId).toBe("a/b");
  });

  it("cancelReservation も clinicId を encode してリクエストする", async () => {
    let capturedClinicId: string | undefined;
    server.use(
      http.delete("/api/liff/:clinicId/my-reservations/:id", ({ params }) => {
        capturedClinicId = params.clinicId as string;
        return new HttpResponse(null, { status: 204 });
      }),
    );

    await liffApi.cancelReservation("a/b", "1", "test-id-token");

    expect(capturedClinicId).toBe("a/b");
  });
});

describe("liffApi（EMR-519: 予約 ID は正の int64 十進文字列）", () => {
  const reservationFixture = (overrides: Record<string, unknown>) => ({
    id: "1",
    pet_id: null,
    pet_name: null,
    course_name: "一般診察",
    staff_name: "佐藤",
    date: "2026-08-01",
    start_time: "10:00",
    end_time: "10:30",
    status: "confirmed",
    notes: "",
    created_at: "2026-07-01T00:00:00Z",
    ...overrides,
  });
  const createBody: CreateReservationBody = {
    course_id: 10,
    staff_id: 0,
    date: "2026-08-01",
    start_time: "1000",
    end_time: "1030",
    customer_fields: {},
    request_text: "",
  };

  it("一覧の旧応答（pet_id/pet_name 両方省略）は受理し両方 null に正規化する", async () => {
    const { pet_id: _a, pet_name: _b, ...legacy } = reservationFixture({});
    server.use(http.get("/api/liff/:clinicId/my-reservations", () => HttpResponse.json([legacy])));
    const list = await liffApi.getMyReservations("1", "test-id-token");
    expect(list[0].pet_id).toBeNull();
    expect(list[0].pet_name).toBeNull();
  });

  it("一覧の大きな十進文字列 ID（9007199254740993）は丸めずそのまま保持される", async () => {
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([reservationFixture({ id: "9007199254740993" })]),
      ),
    );
    const list = await liffApi.getMyReservations("1", "test-id-token");
    expect(list[0].id).toBe("9007199254740993");
  });

  it.each([
    ["pet_id のみ", { pet_id: "2" }],
    ["pet_name のみ", { pet_name: "ポチ" }],
    ["pet_id だけ値", { pet_id: "2", pet_name: null }],
    ["pet_name だけ値", { pet_id: null, pet_name: "ポチ" }],
  ])("一覧の pet ペアが片方だけの %s は reject される", async (_label, pair) => {
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([reservationFixture(pair)]),
      ),
    );
    await expect(liffApi.getMyReservations("1", "test-id-token")).rejects.toThrow();
  });

  it("一覧の数値 ID（旧 wire 形式）は新契約として受理しない", async () => {
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([reservationFixture({ id: 42 })]),
      ),
    );
    await expect(liffApi.getMyReservations("1", "test-id-token")).rejects.toThrow();
  });

  it("201 応答の id は十進文字列を受理し、数値形式は reject する", async () => {
    server.use(
      http.post("/api/liff/:clinicId/reservations", () =>
        HttpResponse.json({ id: "42", notes: "R-1" }, { status: 201 }),
      ),
    );
    const created = await liffApi.createReservation("1", createBody, "test-id-token");
    expect(created.id).toBe("42");

    server.use(
      http.post("/api/liff/:clinicId/reservations", () =>
        HttpResponse.json({ id: 42, notes: "R-1" }, { status: 201 }),
      ),
    );
    await expect(liffApi.createReservation("1", createBody, "test-id-token")).rejects.toThrow();
  });

  it("非正規形の ID では cancelReservation が request を発行しない", async () => {
    let deleteCalls = 0;
    server.use(
      http.delete("/api/liff/:clinicId/my-reservations/:id", () => {
        deleteCalls += 1;
        return new HttpResponse(null, { status: 204 });
      }),
    );
    for (const bad of ["0", "01", "-1", "+1", "1.0", "1e3", " 1", "9223372036854775808"]) {
      await expect(liffApi.cancelReservation("1", bad, "test-id-token")).rejects.toThrow();
    }
    expect(deleteCalls).toBe(0);
    await liffApi.cancelReservation("1", "9007199254740993", "test-id-token");
    expect(deleteCalls).toBe(1);
  });
});
