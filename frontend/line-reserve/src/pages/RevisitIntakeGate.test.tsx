import { render, screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { describe, expect, it, vi } from "vitest";
import { server } from "@/testing/mocks/node";

import { CompletePage } from "./CompletePage";
import { MyReservationsPage } from "./MyReservationsPage";
import { ReservationPageSwitch } from "../app-page-switch";
import type { ReservationFlow, Reservation } from "../types/models";

// EMR-474: LIFF_REVISIT_INTAKE_ACTIVE=false（既定）の間は、問診の入口・ページ・API をすべて止める。
// このファイルは flag を mock しない。有効時の挙動は各 page の既存テストが flag=true を mock して確認する。

const BASE_PROPS = { clinicId: "1", idToken: "token" };

const flow = {
  customerInfo: { name: "テスト", phone: "000", ownerName: "", pets: [] },
  date: "2026-08-01",
  startTime: "10:00",
  endTime: "10:30",
  courseName: "一般診察",
  staffId: 1,
  staffName: "スタッフA",
} as unknown as ReservationFlow;

function reservation(status: Reservation["status"], id: string): Reservation {
  return {
    id,
    pet_id: null,
    pet_name: null,
    course_name: "一般診察",
    staff_name: "スタッフA",
    date: "2026-08-01",
    start_time: "10:00",
    end_time: "10:30",
    status,
    notes: "R-20260801-0001",
    created_at: "2026-07-01T00:00:00Z",
  };
}

describe("再診問診の入口抑止（flag=false）", () => {
  it("予約一覧: 全状態で問診の入口・unsafe ID の案内を出さない（取消・通常表示は維持）", async () => {
    server.use(
      http.get("/api/liff/:clinicId/my-reservations", () =>
        HttpResponse.json([
          reservation("confirmed", "1"),
          reservation("pending", "2"),
          reservation("completed", "3"),
          reservation("cancelled", "4"),
          reservation("confirmed", "9007199254740993"),
        ]),
      ),
    );
    const draftStore = {
      findPending: vi.fn(() => ({ pending: {} })),
    } as unknown as Parameters<typeof MyReservationsPage>[0]["draftStore"];

    render(
      <MyReservationsPage
        {...BASE_PROPS}
        onBack={vi.fn()}
        onOpenRevisitIntake={vi.fn()}
        draftStore={draftStore}
      />,
    );

    expect((await screen.findAllByText(/一般診察/)).length).toBeGreaterThan(0);
    expect(screen.queryByRole("button", { name: "問診を回答する／訂正する" })).toBeNull();
    expect(screen.queryByRole("button", { name: "回答を見る" })).toBeNull();
    expect(screen.queryByRole("button", { name: "送信結果を確認する" })).toBeNull();
    expect(screen.queryByText("この予約の問診は現在開けません。")).toBeNull();
    // 取消の導線は従来どおり（confirmed の予約に出る）。
    expect(screen.getAllByRole("button", { name: /キャンセル/ }).length).toBeGreaterThan(0);
  });

  it("完了画面: unsafe ID の案内も含め問診ブロック全体を出さず、通常の導線は維持する", () => {
    const onRevisitIntake = vi.fn();
    const { rerender } = render(
      <CompletePage
        reservationId="1"
        notes="R-1"
        flow={flow}
        onMyReservations={vi.fn()}
        onNewReservation={vi.fn()}
        onRevisitIntake={onRevisitIntake}
      />,
    );
    expect(screen.queryByText(/事前問診に回答する/)).toBeNull();
    expect(screen.getByRole("button", { name: "予約確認・キャンセル" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "新しい予約をする" })).toBeInTheDocument();

    rerender(
      <CompletePage
        reservationId="9223372036854775808"
        notes="R-1"
        flow={flow}
        onMyReservations={vi.fn()}
        onNewReservation={vi.fn()}
        onRevisitIntake={onRevisitIntake}
      />,
    );
    expect(screen.queryByText("この予約の問診は現在開けません。")).toBeNull();
    expect(onRevisitIntake).not.toHaveBeenCalled();
  });

  it("問診ページ: 直接 page=revisit-intake にしても mount せず、問診 API を呼ばない", () => {
    let intakeRequests = 0;
    server.use(
      http.get("/api/liff/:clinicId/my-reservations/:id/revisit-intake", () => {
        intakeRequests += 1;
        return HttpResponse.json({});
      }),
    );
    const draftStore = {
      get: vi.fn(),
      set: vi.fn(),
      clear: vi.fn(),
      findPending: vi.fn(),
      hasDraft: vi.fn(() => false),
    } as unknown as Parameters<typeof MyReservationsPage>[0]["draftStore"];

    const props = {
      page: "revisit-intake",
      clinicId: "1",
      idToken: "token",
      revisitIntake: { appointmentId: "10", returnTo: "my-reservations" },
      revisitIntakeDraftStore: draftStore,
      goTo: vi.fn(),
    } as unknown as Parameters<typeof ReservationPageSwitch>[0];

    render(<ReservationPageSwitch {...props} />);

    expect(intakeRequests).toBe(0);
    expect(screen.queryByText(/問診/)).toBeNull();
  });
});
