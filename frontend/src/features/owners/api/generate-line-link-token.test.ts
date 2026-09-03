import { describe, it, expect, afterEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { CURRENT_CLINIC_STORAGE_KEY } from "@/lib/current-clinic";
import { useGenerateLineLinkToken } from "./generate-line-link-token";

afterEach(() => {
  localStorage.clear();
});

describe("useGenerateLineLinkToken (SD-14)", () => {
  it("liff_url を含むレスポンスを camelCase の LineLinkTokenResult に変換して返す", async () => {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");
    server.use(
      http.post("*/v1/owners/:id/line/link-token", () =>
        HttpResponse.json(
          {
            token: "abc123",
            expires_at: "2026-07-17T00:00:00+09:00",
            liff_url: "https://liff.line.me/1234567-abcdefgh?token=abc123&clinic_id=1",
          },
          { status: 201 },
        ),
      ),
    );

    const { result } = renderHook(() => useGenerateLineLinkToken("42"), {
      wrapper: createTestWrapper(),
    });

    result.current.mutate();

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual({
      token: "abc123",
      expiresAt: "2026-07-17T00:00:00+09:00",
      liffUrl: "https://liff.line.me/1234567-abcdefgh?token=abc123&clinic_id=1",
    });
  });

  it("X-Clinic-ID ヘッダーを送信する", async () => {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");
    let receivedClinicHeader: string | null = null;
    server.use(
      http.post("*/v1/owners/:id/line/link-token", ({ request }) => {
        receivedClinicHeader = request.headers.get("X-Clinic-ID");
        return HttpResponse.json(
          { token: "abc123", expires_at: "2026-07-17T00:00:00+09:00", liff_url: "https://liff.line.me/x?token=abc123&clinic_id=1" },
          { status: 201 },
        );
      }),
    );

    const { result } = renderHook(() => useGenerateLineLinkToken("42"), {
      wrapper: createTestWrapper(),
    });

    result.current.mutate();

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(receivedClinicHeader).toBe("1");
  });

  it("liff_url が空文字のレスポンスを liffUrl: '' に変換する（LIFF ID 未設定クリニック）", async () => {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");
    server.use(
      http.post("*/v1/owners/:id/line/link-token", () =>
        HttpResponse.json(
          { token: "abc123", expires_at: "2026-07-17T00:00:00+09:00" },
          { status: 201 },
        ),
      ),
    );

    const { result } = renderHook(() => useGenerateLineLinkToken("42"), {
      wrapper: createTestWrapper(),
    });

    result.current.mutate();

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.liffUrl).toBe("");
  });
});
