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
    server.use(
      http.get("/api/liff/:clinicId/settings", () =>
        HttpResponse.json({ clinicName: "テスト病院" }),
      ),
    );

    const settings = await liffApi.getSettings("1");

    expect(settings).toEqual({ clinicName: "テスト病院" });
  });
});
