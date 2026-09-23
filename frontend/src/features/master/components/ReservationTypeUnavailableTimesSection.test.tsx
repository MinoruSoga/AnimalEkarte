import { describe, expect, it, afterEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { ReservationTypeUnavailableTimesSection } from "./ReservationTypeUnavailableTimesSection";

function createWrapper() {
  return createTestWrapper({ router: true });
}

afterEach(() => {
  server.resetHandlers();
});

describe("ReservationTypeUnavailableTimesSection", () => {
  it("予約不可時間を表示する", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types/5/unavailable-times", () =>
        HttpResponse.json([
          {
            id: 1,
            clinic_id: 1,
            reservation_type_id: 5,
            unavailable_type: "weekly",
            day_of_week: 1,
            start_time: "09:00",
            end_time: "12:00",
            created_at: "2026-05-29T00:00:00Z",
            updated_at: "2026-05-29T00:00:00Z",
          },
          {
            id: 2,
            clinic_id: 1,
            reservation_type_id: 5,
            unavailable_type: "specific",
            specific_date: "2026-06-01",
            start_time: "14:00",
            end_time: "18:00",
            created_at: "2026-05-29T00:00:00Z",
            updated_at: "2026-05-29T00:00:00Z",
          },
        ]),
      ),
    );

    render(<ReservationTypeUnavailableTimesSection clinicId="1" reservationTypeId="5" />, {
      wrapper: createWrapper(),
    });

    expect(await screen.findByText("予約不可時間")).toBeInTheDocument();
    expect(await screen.findByText("毎週月曜日")).toBeInTheDocument();
    expect(screen.getByText("2026-06-01")).toBeInTheDocument();
    expect(screen.getAllByText("09:00〜12:00").length).toBeGreaterThan(0);
    expect(screen.getAllByText("14:00〜18:00").length).toBeGreaterThan(0);
  });

  // EMR-212 (EMR-208 同型): MasterSidePanel の外側 <form> 内にネストした <form> は
  // HTML では無効でブラウザが内側を破棄し、「追加」が外側 action へ submit される
  // 無音失敗だった。回帰防止: セクション内に <form> を持たず、SubmitButton の
  // formAction で POST を発行する。
  it("親フォーム内で「追加」を押すと POST /unavailable-times が発行され、外側 action は呼ばれない", async () => {
    const posts: unknown[] = [];
    const outerAction = vi.fn();
    server.use(
      http.get("/api/v1/masters/reservation-types/5/unavailable-times", () =>
        HttpResponse.json([]),
      ),
      http.post("/api/v1/masters/reservation-types/5/unavailable-times", async ({ request }) => {
        posts.push(await request.json());
        return HttpResponse.json({ id: 3 }, { status: 201 });
      }),
    );

    const user = userEvent.setup();
    const { container } = render(
      <form action={outerAction}>
        <ReservationTypeUnavailableTimesSection clinicId="1" reservationTypeId="5" />
      </form>,
      { wrapper: createWrapper() },
    );

    expect(await screen.findByText("予約不可時間")).toBeInTheDocument();

    // セクション側に <form> が残っていないこと（ネスト form の構造的回帰検査）
    expect(container.querySelectorAll("form")).toHaveLength(1);

    await user.click(screen.getByRole("button", { name: "追加" }));

    await waitFor(() => expect(posts).toHaveLength(1));
    expect(posts[0]).toMatchObject({
      unavailable_type: "weekly",
      day_of_week: 1,
      start_time: "09:00",
      end_time: "18:00",
    });
    expect(outerAction).not.toHaveBeenCalled();
  });
});
