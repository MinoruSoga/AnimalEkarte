import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";

import { DateSelectPage } from "./DateSelectPage";
import { EXPLICIT_PRIMARY_CTA_LABEL } from "../lib/advance-policy";

const BASE_PROPS = {
  clinicId: "1",
  idToken: "test-id-token",
  courseId: 10,
  staffId: 0,
  selectedDate: "",
  bookingWindow: 30,
  isTrimming: false,
};

describe("DateSelectPage（R-F22/R-F23: 共通フェッチ状態管理・ステータス別エラー）", () => {
  it("API失敗(5xx)時はサーバーエラーメッセージと再試行ボタンを表示する", async () => {
    server.use(
      http.get("/api/liff/:clinicId/available-dates", () =>
        HttpResponse.json(null, { status: 500 }),
      ),
    );

    render(<DateSelectPage {...BASE_PROPS} onSelect={vi.fn()} onNext={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByRole("alert")).toHaveTextContent("サーバーエラーが発生しました");
    expect(screen.getByRole("button", { name: "再試行" })).toBeInTheDocument();
  });

  it("API失敗(401)時は再ログインメッセージを表示し、再試行ボタンは出さない", async () => {
    server.use(
      http.get("/api/liff/:clinicId/available-dates", () =>
        HttpResponse.json(null, { status: 401 }),
      ),
    );

    render(<DateSelectPage {...BASE_PROPS} onSelect={vi.fn()} onNext={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "LINEアプリを再起動して開き直してください",
    );
    expect(screen.queryByRole("button", { name: "再試行" })).not.toBeInTheDocument();
  });

  it("再試行ボタンをクリックすると空き日程を再取得する", async () => {
    const user = userEvent.setup();
    let callCount = 0;
    server.use(
      http.get("/api/liff/:clinicId/available-dates", () => {
        callCount += 1;
        if (callCount === 1) return HttpResponse.json(null, { status: 500 });
        return HttpResponse.json({
          dates: [{ date: "2026-08-01", available: true }],
          window: null,
        });
      }),
    );

    render(<DateSelectPage {...BASE_PROPS} onSelect={vi.fn()} onNext={vi.fn()} onBack={vi.fn()} />);

    await user.click(await screen.findByRole("button", { name: "再試行" }));

    expect(callCount).toBe(2);
  });

  it("BUG-030: explicit-cta のため主CTA「次へ」を表示し、auto-advance ヒントは出さない", async () => {
    server.use(
      http.get("/api/liff/:clinicId/available-dates", () =>
        HttpResponse.json({ dates: [{ date: "2026-08-01", available: true }], window: null }),
      ),
    );

    render(
      <DateSelectPage
        {...BASE_PROPS}
        selectedDate="2026-08-01"
        onSelect={vi.fn()}
        onNext={vi.fn()}
        onBack={vi.fn()}
      />,
    );

    expect(
      await screen.findByRole("button", { name: EXPLICIT_PRIMARY_CTA_LABEL }),
    ).toBeInTheDocument();
    expect(screen.queryByTestId("auto-advance-hint")).not.toBeInTheDocument();
  });
});
