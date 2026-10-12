import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { addDays, format, startOfWeek } from "date-fns";
import { ja } from "date-fns/locale";
import type { Reservation } from "@/types";
import { WeekView } from "./WeekView";

const BASE_DATE = new Date("2026-05-20T00:00:00"); // Wednesday

// startOfWeek with ja locale (weekStartsOn: 0) → Sunday 2026-05-17
// Week: 17(日) 18(月) 19(火) 20(水) 21(木) 22(金) 23(土)
const WEEK_START = startOfWeek(BASE_DATE, { locale: ja });
const EXPECTED_DATE_LABELS = [0, 1, 2, 3, 4, 5, 6].map((i) => format(addDays(WEEK_START, i), "d")); // ["17","18","19","20","21","22","23"]

describe("WeekView — days prop", () => {
  it("days 未指定（デフォルト5）→ 7列全曜日が表示される（週全体スクロール）", () => {
    render(<WeekView currentDate={BASE_DATE} appointments={[]} onAppointmentClick={() => {}} />);
    const headers = screen.getAllByTestId("week-header-day");
    expect(headers).toHaveLength(7);
    EXPECTED_DATE_LABELS.forEach((label, i) => {
      expect(headers[i]).toHaveTextContent(label);
    });
  });

  it("days=5 → 7列全曜日が表示される（週全体スクロール — 以前のデフォルト動作）", () => {
    render(
      <WeekView currentDate={BASE_DATE} appointments={[]} onAppointmentClick={() => {}} days={5} />,
    );
    const headers = screen.getAllByTestId("week-header-day");
    expect(headers).toHaveLength(7);
    // 日曜(17)〜土曜(23)の全7曜日が含まれることを確認
    EXPECTED_DATE_LABELS.forEach((label, i) => {
      expect(headers[i]).toHaveTextContent(label);
    });
  });

  it("days=7 → 7列全曜日が表示される（全曜日を一画面表示）", () => {
    render(
      <WeekView currentDate={BASE_DATE} appointments={[]} onAppointmentClick={() => {}} days={7} />,
    );
    const headers = screen.getAllByTestId("week-header-day");
    expect(headers).toHaveLength(7);
    EXPECTED_DATE_LABELS.forEach((label, i) => {
      expect(headers[i]).toHaveTextContent(label);
    });
  });
});

describe("WeekView — tentative 常時表示帯（Codex PR606 #9）", () => {
  /** 意図日の JST 00:00 sentinel start を持つ tentative 行の最小フィクスチャ。 */
  function tentativeOf(id: string, dayIndex: number, petName: string): Reservation {
    const day = addDays(WEEK_START, dayIndex);
    return {
      id,
      status: "tentative",
      isTentative: true,
      start: new Date(day.getFullYear(), day.getMonth(), day.getDate(), 0, 0),
      end: new Date(day.getFullYear(), day.getMonth(), day.getDate(), 0, 15),
      petName,
      ownerName: "山田",
    } as unknown as Reservation;
  }

  it("tentative はタイムグリッドではなく常時表示帯に出て、クリックで詳細を開く", async () => {
    const onClick = vi.fn();
    render(
      <WeekView
        currentDate={BASE_DATE}
        appointments={[
          tentativeOf("t-1", 1, "ポチ"),
          {
            id: "r-1",
            status: "confirmed",
            isTentative: false,
            start: addDays(WEEK_START, 2),
            end: addDays(WEEK_START, 2),
            petName: "タマ",
            ownerName: "佐藤",
          } as unknown as Reservation,
        ]}
        onAppointmentClick={onClick}
      />,
    );

    const strip = screen.getByTestId("week-tentative-row");
    expect(strip).toHaveTextContent("ポチ");
    // 実予約は帯に出ない
    expect(strip).not.toHaveTextContent("タマ");

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /ポチ/ }));
    expect(onClick).toHaveBeenCalledWith(expect.objectContaining({ id: "t-1" }));
  });

  it("tentative が無ければ帯自体を描画しない", () => {
    render(<WeekView currentDate={BASE_DATE} appointments={[]} onAppointmentClick={() => {}} />);
    expect(screen.queryByTestId("week-tentative-row")).not.toBeInTheDocument();
  });
});
