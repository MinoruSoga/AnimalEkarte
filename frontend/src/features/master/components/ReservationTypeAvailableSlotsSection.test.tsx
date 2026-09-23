import { describe, expect, it, afterEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { ReservationTypeAvailableSlotsSection } from "./ReservationTypeAvailableSlotsSection";

function createWrapper() {
  return createTestWrapper({ router: true });
}

afterEach(() => {
  server.resetHandlers();
});

describe("ReservationTypeAvailableSlotsSection", () => {
  it("予約可能枠を表示する", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types/5/available-slots", () =>
        HttpResponse.json([
          {
            id: 1,
            clinic_id: 1,
            reservation_type_id: 5,
            available_type: "weekly",
            day_of_week: 1,
            start_time: "09:45",
            is_active: true,
            created_at: "2026-05-29T00:00:00Z",
            updated_at: "2026-05-29T00:00:00Z",
          },
          {
            id: 2,
            clinic_id: 1,
            reservation_type_id: 5,
            available_type: "specific",
            specific_date: "2026-06-01",
            start_time: "14:00",
            is_active: true,
            created_at: "2026-05-29T00:00:00Z",
            updated_at: "2026-05-29T00:00:00Z",
          },
        ]),
      ),
    );

    render(<ReservationTypeAvailableSlotsSection clinicId="1" reservationTypeId="5" />, {
      wrapper: createWrapper(),
    });

    expect(await screen.findByText("予約可能枠")).toBeInTheDocument();
    expect(await screen.findByText("毎週月曜日")).toBeInTheDocument();
    expect(screen.getByText("2026-06-01")).toBeInTheDocument();
    expect(screen.getAllByText("09:45").length).toBeGreaterThan(0);
    expect(screen.getAllByText("14:00").length).toBeGreaterThan(0);
  });

  // BUG-MASTER-RESVTYPE-SLOT-FORM-NESTED (EMR-208): 親 <form> 内にネストした
  // <form> は HTML では無効でブラウザが内側を破棄し、「追加」が外側の
  // javascript: action へ submit されて CSP にブロックされる無音失敗だった。
  // 回帰防止: セクション内に <form> を持たず、formAction ボタンで POST を発行する。
  it("親フォーム内で「追加」を押すと POST /available-slots が発行され、外側 action は呼ばれない", async () => {
    const posts: unknown[] = [];
    const outerAction = vi.fn();
    server.use(
      http.get("/api/v1/masters/reservation-types/5/available-slots", () => HttpResponse.json([])),
      http.post("/api/v1/masters/reservation-types/5/available-slots", async ({ request }) => {
        posts.push(await request.json());
        return HttpResponse.json({ id: 3 }, { status: 201 });
      }),
    );

    const user = userEvent.setup();
    const { container } = render(
      <form action={outerAction}>
        <ReservationTypeAvailableSlotsSection clinicId="1" reservationTypeId="5" />
      </form>,
      { wrapper: createWrapper() },
    );

    expect(await screen.findByText("予約可能枠")).toBeInTheDocument();

    // セクション側に <form> が残っていないこと（ネスト form の構造的回帰検査）
    expect(container.querySelectorAll("form")).toHaveLength(1);

    await user.click(screen.getByRole("button", { name: "追加" }));

    await waitFor(() => expect(posts).toHaveLength(1));
    expect(posts[0]).toMatchObject({
      available_type: "weekly",
      day_of_week: 1,
      start_time: "09:45",
      is_active: true,
    });
    expect(outerAction).not.toHaveBeenCalled();
  });
});
