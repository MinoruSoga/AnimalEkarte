import { describe, expect, it, afterEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { toJSTWallDate } from "@/lib/jst-date";
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

  it("前の月・次の月ナビで表示月が切り替わり、今日ボタンで当月に戻る", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types/5/available-slots", () =>
        HttpResponse.json([]),
      ),
    );

    const user = userEvent.setup();
    renderCalendar();

    expect(await screen.findByText("2026年 6月")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "次の月" }));
    expect(screen.getByText("2026年 7月")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "前の月" }));
    await user.click(screen.getByRole("button", { name: "前の月" }));
    expect(screen.getByText("2026年 5月")).toBeInTheDocument();

    // 「今日」は実行時の当月へ戻る（テスト実行日に依存しないよう動的に算出）
    await user.click(screen.getByRole("button", { name: "今日" }));
    const currentMonthLabel = format(toJSTWallDate(new Date()), "yyyy年 M月", { locale: ja });
    expect(screen.getByText(currentMonthLabel)).toBeInTheDocument();
  });

  it("枠が5件以上の日は4件まで表示し残りを「他 N 件」に省略する", async () => {
    const weeklySlots = ["09:00", "10:00", "11:00", "12:00", "13:00"].map((startTime, i) => ({
      id: i + 1,
      clinic_id: 1,
      reservation_type_id: 5,
      available_type: "weekly",
      day_of_week: 1,
      start_time: startTime,
      is_active: true,
      created_at: "2026-05-29T00:00:00Z",
      updated_at: "2026-05-29T00:00:00Z",
    }));
    server.use(
      http.get("/api/v1/masters/reservation-types/5/available-slots", () =>
        HttpResponse.json(weeklySlots),
      ),
    );

    renderCalendar();

    const cell = await screen.findByRole("button", { name: "2026年6月15日" });
    expect(await within(cell).findByText("12:00")).toBeInTheDocument();
    expect(within(cell).queryByText("13:00")).not.toBeInTheDocument();
    expect(within(cell).getByText("他 1 件")).toBeInTheDocument();
  });

  it("編集パネルの毎週枠は読み取り専用で、削除ボタンは特定日枠のみに表示する", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types/5/available-slots", () =>
        HttpResponse.json(SLOTS),
      ),
    );

    const user = userEvent.setup();
    renderCalendar();

    const cell = await screen.findByRole("button", { name: "2026年6月15日" });
    await user.click(cell);

    expect(await screen.findByText("毎週の枠（リストで管理）:")).toBeInTheDocument();
    expect(screen.getByText("この日の枠:")).toBeInTheDocument();
    // 特定日枠(14:00)のみ削除可能。毎週枠(09:45)に削除ボタンはない
    expect(screen.getByRole("button", { name: "14:00の枠を削除" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "09:45の枠を削除" })).not.toBeInTheDocument();
  });
});
