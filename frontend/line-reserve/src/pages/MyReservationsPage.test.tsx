import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";

import { MyReservationsPage } from "./MyReservationsPage";
import type { Reservation } from "../types/models";

const BASE_PROPS = {
  clinicId: "1",
  idToken: "test-id-token",
};

const confirmedReservation: Reservation = {
  id: 1,
  course_name: "一般診察",
  staff_name: "山田先生",
  date: "2026-08-01",
  start_time: "10:00",
  end_time: "10:30",
  status: "confirmed",
  notes: "R-20260801-0001",
  created_at: "2026-07-01T00:00:00Z",
};

describe("MyReservationsPage（R-F25: ネイティブダイアログ→インラインUI置換）", () => {
  it("予約一覧を表示する", async () => {
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([confirmedReservation]),
      ),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} />);

    expect(await screen.findByText("一般診察")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "キャンセルする" })).toBeInTheDocument();
  });

  it("「キャンセルする」クリックでネイティブダイアログではなくインライン確認UIを表示する", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([confirmedReservation]),
      ),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} />);

    await user.click(await screen.findByRole("button", { name: "キャンセルする" }));

    expect(screen.getByText("本当にキャンセルしますか？")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "はい、キャンセルする" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "いいえ" })).toBeInTheDocument();
  });

  it("確認UIで「いいえ」を選ぶとキャンセル操作を中断し、元のボタンに戻る", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([confirmedReservation]),
      ),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} />);

    await user.click(await screen.findByRole("button", { name: "キャンセルする" }));
    await user.click(screen.getByRole("button", { name: "いいえ" }));

    expect(screen.queryByText("本当にキャンセルしますか？")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "キャンセルする" })).toBeInTheDocument();
  });

  it("確認UIで「はい」を選び成功すると予約がキャンセル済に更新される", async () => {
    const user = userEvent.setup();
    let deleteCalled = false;
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([confirmedReservation]),
      ),
      http.delete("/api/liff/:clinicId/my-reservations/:id", () => {
        deleteCalled = true;
        return new HttpResponse(null, { status: 204 });
      }),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} />);

    await user.click(await screen.findByRole("button", { name: "キャンセルする" }));
    await user.click(screen.getByRole("button", { name: "はい、キャンセルする" }));

    await waitFor(() => {
      expect(deleteCalled).toBe(true);
    });
    expect(await screen.findByText("キャンセル済")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "キャンセルする" })).not.toBeInTheDocument();
  });

  it("確認UIで「はい」を選び失敗するとネイティブalertではなくインラインバナーを表示し、再試行可能にする", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([confirmedReservation]),
      ),
      http.delete("/api/liff/:clinicId/my-reservations/:id", () =>
        HttpResponse.json(null, { status: 500 }),
      ),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} />);

    await user.click(await screen.findByRole("button", { name: "キャンセルする" }));
    await user.click(screen.getByRole("button", { name: "はい、キャンセルする" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("キャンセルに失敗しました");
    expect(screen.getByRole("button", { name: "キャンセルする" })).not.toBeDisabled();
  });

  it("一覧取得API失敗(5xx)時はサーバーエラーメッセージと再試行ボタンを表示する（R-F22/R-F23）", async () => {
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json(null, { status: 500 }),
      ),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} />);

    expect(await screen.findByRole("alert")).toHaveTextContent("サーバーエラーが発生しました");
    expect(screen.getByRole("button", { name: "再試行" })).toBeInTheDocument();
  });

  it("一覧取得API失敗(401)時は再ログインメッセージを表示し、再試行ボタンは出さない（R-F22/R-F23）", async () => {
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json(null, { status: 401 }),
      ),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} />);

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "LINEアプリを再起動して開き直してください",
    );
    expect(screen.queryByRole("button", { name: "再試行" })).not.toBeInTheDocument();
  });

  it("一覧の再試行ボタンをクリックすると予約一覧を再取得する（R-F22/R-F23）", async () => {
    const user = userEvent.setup();
    let callCount = 0;
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () => {
        callCount += 1;
        if (callCount === 1) return HttpResponse.json(null, { status: 500 });
        return HttpResponse.json([confirmedReservation]);
      }),
    );

    render(<MyReservationsPage {...BASE_PROPS} onBack={vi.fn()} />);

    await user.click(await screen.findByRole("button", { name: "再試行" }));

    expect(await screen.findByText("一般診察")).toBeInTheDocument();
    expect(callCount).toBe(2);
  });
});
