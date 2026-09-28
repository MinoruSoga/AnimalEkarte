import { useState } from "react";
import { describe, it, expect } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ReservationDateTimeFields } from "./ReservationDateTimeFields";
import { TIME_OPTIONS } from "./reservation-time-utils";
import type { Reservation } from "@/types";

// 終了 Select が非15分刻みの値を空白表示しないことの確認（EMR-191 追補）。
// ReservationDateTimeFields は表示専用コンポーネントなので、
// onChange を受けて formData を更新する最小ハーネスで durationMinutes 経路を直接検証する。
function Harness({ durationMinutes }: { durationMinutes: number }) {
  const [formData, setFormData] = useState<Partial<Reservation>>({
    start: new Date(2026, 5, 1, 9, 0, 0),
    end: new Date(2026, 5, 1, 9, 30, 0),
  });
  return (
    <ReservationDateTimeFields
      formData={formData}
      onChange={setFormData}
      isCalendarDateDisabled={() => false}
      handleMonthChange={() => undefined}
      startTimeOptions={TIME_OPTIONS}
      availableTimeSlotMap={undefined}
      durationMinutes={durationMinutes}
    />
  );
}

describe("ReservationDateTimeFields — 終了時刻 Select の現在値注入 (EMR-191)", () => {
  it("20分の区分で開始10:00を選ぶと終了Selectに10:20が表示される", async () => {
    const user = userEvent.setup({ delay: null });
    render(<Harness durationMinutes={20} />);

    await user.click(screen.getByTestId("res-start-time-trigger"));
    fireEvent.click(await screen.findByRole("option", { name: "10:00" }));

    // 修正前は options に無い値のため trigger が空白になった
    expect(screen.getByTestId("res-end-time-trigger")).toHaveTextContent("10:20");

    // 終了 Select を開いても 10:20 の option が 1 件だけ存在する
    await user.click(screen.getByTestId("res-end-time-trigger"));
    expect(await screen.findAllByRole("option", { name: "10:20" })).toHaveLength(1);
  });

  it("15分刻みに揃う終了時刻では options が増えない", async () => {
    const user = userEvent.setup({ delay: null });
    render(<Harness durationMinutes={15} />);

    await user.click(screen.getByTestId("res-start-time-trigger"));
    fireEvent.click(await screen.findByRole("option", { name: "10:00" }));

    expect(screen.getByTestId("res-end-time-trigger")).toHaveTextContent("10:15");

    await user.click(screen.getByTestId("res-end-time-trigger"));
    expect(await screen.findAllByRole("option")).toHaveLength(TIME_OPTIONS.length);
  });
});
