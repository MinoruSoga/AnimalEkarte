import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { addDays, format, startOfWeek } from "date-fns";
import { ja } from "date-fns/locale";
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
