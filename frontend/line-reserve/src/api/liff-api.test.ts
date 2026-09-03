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
        return HttpResponse.json({ id: 1, notes: "R-20260801-0001" });
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
    server.use(
      http.get("/api/liff/:clinicId/settings", () => HttpResponse.json(settingsFixture)),
    );

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

    await liffApi.cancelReservation("a/b", 1, "test-id-token");

    expect(capturedClinicId).toBe("a/b");
  });
});
