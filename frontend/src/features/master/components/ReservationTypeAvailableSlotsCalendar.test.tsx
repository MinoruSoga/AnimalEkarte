import { describe, expect, it, afterEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { ReservationTypeAvailableSlotsCalendar } from "./ReservationTypeAvailableSlotsCalendar";

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

afterEach(() => {
  server.resetHandlers();
});

const SLOTS = [
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
    specific_date: "2026-06-15",
    start_time: "14:00",
    is_active: true,
    created_at: "2026-05-29T00:00:00Z",
    updated_at: "2026-05-29T00:00:00Z",
  },
];

function renderCalendar() {
  return render(
    <ReservationTypeAvailableSlotsCalendar
      clinicId="1"
      reservationTypeId="5"
      initialMonth={new Date(2026, 5, 1)}
    />,
    { wrapper: createWrapper() },
  );
}

describe("ReservationTypeAvailableSlotsCalendar", () => {
  it("月グリッドに特定日枠と毎週枠を表示する", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types/5/available-slots", () =>
        HttpResponse.json(SLOTS),
      ),
    );

    renderCalendar();

    expect(await screen.findByText("2026年 6月")).toBeInTheDocument();

    // 特定日枠は該当日セルにのみ表示
    const cell = await screen.findByRole("button", { name: "2026年6月15日" });
    expect(await within(cell).findByText("14:00")).toBeInTheDocument();

    // 毎週(月曜)枠は 6月の全月曜セルに表示される(6/1,8,15,22,29)
    await waitFor(() => {
      expect(screen.getAllByText("09:45").length).toBeGreaterThanOrEqual(5);
    });
  });

  it("日付クリックで編集パネルが開き、特定日枠を削除できる", async () => {
    let deletedId: string | null = null;
    server.use(
      http.get("/api/v1/masters/reservation-types/5/available-slots", () =>
        HttpResponse.json(SLOTS),
      ),
      http.delete(
        "/api/v1/masters/reservation-types/5/available-slots/:id",
        ({ params }) => {
          deletedId = String(params.id);
          return new HttpResponse(null, { status: 204 });
        },
      ),
    );

    const user = userEvent.setup();
    renderCalendar();

    const cell = await screen.findByRole("button", { name: "2026年6月15日" });
    await user.click(cell);

    expect(await screen.findByText("6月15日（月）")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "14:00の枠を削除" }));

    await waitFor(() => {
      expect(deletedId).toBe("2");
    });
  });

  it("日付を選択して枠を追加すると specific_date 付きで POST する", async () => {
    let posted: unknown = null;
    server.use(
      http.get("/api/v1/masters/reservation-types/5/available-slots", () =>
        HttpResponse.json(SLOTS),
      ),
      http.post(
        "/api/v1/masters/reservation-types/5/available-slots",
        async ({ request }) => {
          posted = await request.json();
          return HttpResponse.json({}, { status: 201 });
        },
      ),
    );

    const user = userEvent.setup();
    renderCalendar();

    const cell = await screen.findByRole("button", { name: "2026年6月20日" });
    await user.click(cell);

    await user.click(await screen.findByRole("button", { name: /追加/ }));

    await waitFor(() => {
      expect(posted).toEqual({
        available_type: "specific",
        start_time: "09:00",
        is_active: true,
        specific_date: "2026-06-20",
      });
    });
  });
});
