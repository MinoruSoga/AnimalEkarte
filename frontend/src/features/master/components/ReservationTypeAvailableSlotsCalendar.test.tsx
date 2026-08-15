import { describe, expect, it, afterEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/utils";
import { toJSTWallDate } from "@/lib/jst-date";
import { ReservationTypeAvailableSlotsCalendar } from "./ReservationTypeAvailableSlotsCalendar";

function createWrapper() {
  return createTestWrapper({ router: true });
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
  it("狭幅でも7日の日付buttonを44px以上に保ちcalendar内だけ横scrollできる", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types/5/available-slots", () =>
        HttpResponse.json([]),
      ),
    );

    const { container } = renderCalendar();

    await screen.findByText("2026/06/01 - 2026/06/07");
    const dayButtons = screen.getAllByRole("button", { name: /^2026\/06\/0[1-7]$/ });
    expect(dayButtons).toHaveLength(7);
    for (const button of dayButtons) {
      expect(button).toHaveClass("min-w-11");
    }

    const calendarCanvas = dayButtons[0]?.parentElement?.parentElement;
    const scrollViewport = calendarCanvas?.parentElement;
    expect(calendarCanvas).toHaveClass("min-w-[308px]", "w-full");
    expect(scrollViewport).toHaveClass("min-w-0", "max-w-full", "overflow-x-auto");
    expect(container.firstElementChild).not.toHaveClass("overflow-x-auto");
  });

  it("週グリッドに特定日枠と毎週枠を表示する", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types/5/available-slots", () =>
        HttpResponse.json(SLOTS),
      ),
    );

    renderCalendar();

    expect(await screen.findByText("2026/06/01 - 2026/06/07")).toBeInTheDocument();

    const cell = await screen.findByRole("button", { name: "2026/06/01" });
    expect(await within(cell).findByText("09:45")).toBeInTheDocument();
    expect(screen.queryByText("14:00")).not.toBeInTheDocument();
  });

  it("次週に切り替えると該当週の特定日枠を表示する", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types/5/available-slots", () =>
        HttpResponse.json(SLOTS),
      ),
    );

    const user = userEvent.setup();
    renderCalendar();

    await user.click(await screen.findByRole("button", { name: "次の週" }));
    await user.click(screen.getByRole("button", { name: "次の週" }));

    expect(screen.getByText("2026/06/15 - 2026/06/21")).toBeInTheDocument();
    const cell = await screen.findByRole("button", { name: "2026/06/15" });
    expect(await within(cell).findByText("14:00")).toBeInTheDocument();
    expect(within(cell).getByText("09:45")).toBeInTheDocument();
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

    await user.click(await screen.findByRole("button", { name: "次の週" }));
    await user.click(screen.getByRole("button", { name: "次の週" }));

    const cell = await screen.findByRole("button", { name: "2026/06/15" });
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

    await user.click(await screen.findByRole("button", { name: "次の週" }));
    await user.click(screen.getByRole("button", { name: "次の週" }));

    const cell = await screen.findByRole("button", { name: "2026/06/20" });
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

  it("前の週・次の週ナビで表示週が切り替わり、今日ボタンで当週に戻る", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types/5/available-slots", () =>
        HttpResponse.json([]),
      ),
    );

    const user = userEvent.setup();
    renderCalendar();

    expect(await screen.findByText("2026/06/01 - 2026/06/07")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "次の週" }));
    expect(screen.getByText("2026/06/08 - 2026/06/14")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "前の週" }));
    await user.click(screen.getByRole("button", { name: "前の週" }));
    expect(screen.getByText("2026/05/25 - 2026/05/31")).toBeInTheDocument();

    // 「今日」は実行時の当週へ戻る（月曜始まりで動的に算出）
    await user.click(screen.getByRole("button", { name: "今日" }));
    const today = toJSTWallDate(new Date());
    const dow = today.getDay(); // 0=Sun, 1=Mon, ..., 6=Sat
    const daysFromMonday = dow === 0 ? 6 : dow - 1;
    const currentWeekStart = new Date(today);
    currentWeekStart.setDate(today.getDate() - daysFromMonday);
    const currentWeekEnd = new Date(currentWeekStart);
    currentWeekEnd.setDate(currentWeekStart.getDate() + 6);
    // FE5-28/M3: 週見出しフォーマットは開始・終了とも "yyyy/MM/dd" (DISPLAY_DATE_FORMAT) に統一
    const currentWeekLabel = `${format(currentWeekStart, "yyyy/MM/dd", { locale: ja })} - ${format(currentWeekEnd, "yyyy/MM/dd", { locale: ja })}`;
    expect(screen.getByText(currentWeekLabel)).toBeInTheDocument();
  });

  it("週表示では日別の枠を省略せず表示する", async () => {
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

    const user = userEvent.setup();
    renderCalendar();

    await user.click(await screen.findByRole("button", { name: "次の週" }));
    await user.click(screen.getByRole("button", { name: "次の週" }));

    const cell = await screen.findByRole("button", { name: "2026/06/15" });
    expect(await within(cell).findByText("12:00")).toBeInTheDocument();
    expect(within(cell).getByText("13:00")).toBeInTheDocument();
  });

  it("同一日・同一 start_time で追加ボタンが disabled になる", async () => {
    const specificSlots = [
      {
        id: 10,
        clinic_id: 1,
        reservation_type_id: 5,
        available_type: "specific",
        specific_date: "2026-06-15",
        start_time: "09:00",
        is_active: true,
        created_at: "2026-05-29T00:00:00Z",
        updated_at: "2026-05-29T00:00:00Z",
      },
    ];
    server.use(
      http.get("/api/v1/masters/reservation-types/5/available-slots", () =>
        HttpResponse.json(specificSlots),
      ),
    );

    const user = userEvent.setup();
    renderCalendar();

    await user.click(await screen.findByRole("button", { name: "次の週" }));
    await user.click(screen.getByRole("button", { name: "次の週" }));

    const cell = await screen.findByRole("button", { name: "2026/06/15" });
    await user.click(cell);

    // デフォルト start_time は 09:00 → 既存と同じなので disabled
    await waitFor(() => {
      expect(screen.getByText("この時刻は既に登録済みです")).toBeInTheDocument();
    });
    const addButton = screen.getByRole("button", { name: /追加/ });
    expect(addButton).toBeDisabled();
  });

  it("同一日・異なる start_time では disabled にならない", async () => {
    const specificSlots = [
      {
        id: 11,
        clinic_id: 1,
        reservation_type_id: 5,
        available_type: "specific",
        specific_date: "2026-06-15",
        start_time: "10:00",
        is_active: true,
        created_at: "2026-05-29T00:00:00Z",
        updated_at: "2026-05-29T00:00:00Z",
      },
    ];
    server.use(
      http.get("/api/v1/masters/reservation-types/5/available-slots", () =>
        HttpResponse.json(specificSlots),
      ),
    );

    const user = userEvent.setup();
    renderCalendar();

    await user.click(await screen.findByRole("button", { name: "次の週" }));
    await user.click(screen.getByRole("button", { name: "次の週" }));

    const cell = await screen.findByRole("button", { name: "2026/06/15" });
    await user.click(cell);

    // デフォルト start_time は 09:00 → 既存 (10:00) と異なるので disabled にならない
    await waitFor(() => {
      expect(screen.queryByText("この時刻は既に登録済みです")).not.toBeInTheDocument();
    });
    const addButton = screen.getByRole("button", { name: /追加/ });
    expect(addButton).not.toBeDisabled();
  });

  it("weekly slot と同じ start_time でも specific 追加は許可される（disabled にならない）", async () => {
    const weeklySlots = [
      {
        id: 20,
        clinic_id: 1,
        reservation_type_id: 5,
        available_type: "weekly",
        day_of_week: 1, // 月曜
        start_time: "09:00",
        is_active: true,
        created_at: "2026-05-29T00:00:00Z",
        updated_at: "2026-05-29T00:00:00Z",
      },
    ];
    server.use(
      http.get("/api/v1/masters/reservation-types/5/available-slots", () =>
        HttpResponse.json(weeklySlots),
      ),
    );

    const user = userEvent.setup();
    renderCalendar();

    await user.click(await screen.findByRole("button", { name: "次の週" }));
    await user.click(screen.getByRole("button", { name: "次の週" }));

    // 2026-06-15 は月曜
    const cell = await screen.findByRole("button", { name: "2026/06/15" });
    await user.click(cell);

    // weekly の 09:00 があるが specific は別管理 → disabled にならない
    await waitFor(() => {
      expect(screen.queryByText("この時刻は既に登録済みです")).not.toBeInTheDocument();
    });
    const addButton = screen.getByRole("button", { name: /追加/ });
    expect(addButton).not.toBeDisabled();
  });

  it("編集パネルの毎週枠は読み取り専用で、削除ボタンは特定日枠のみに表示する", async () => {
    server.use(
      http.get("/api/v1/masters/reservation-types/5/available-slots", () =>
        HttpResponse.json(SLOTS),
      ),
    );

    const user = userEvent.setup();
    renderCalendar();

    await user.click(await screen.findByRole("button", { name: "次の週" }));
    await user.click(screen.getByRole("button", { name: "次の週" }));

    const cell = await screen.findByRole("button", { name: "2026/06/15" });
    await user.click(cell);

    expect(await screen.findByText("毎週の枠（リストで管理）:")).toBeInTheDocument();
    expect(screen.getByText("この日の枠:")).toBeInTheDocument();
    // 特定日枠(14:00)のみ削除可能。毎週枠(09:45)に削除ボタンはない
    expect(screen.getByRole("button", { name: "14:00の枠を削除" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "09:45の枠を削除" })).not.toBeInTheDocument();
  });
});
