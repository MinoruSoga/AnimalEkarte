import { useState } from "react";
import { describe, it, expect, vi } from "vitest";
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

describe("ReservationDateTimeFields — カレンダーナビゲーション", () => {
  function renderFields(
    overrides: Partial<React.ComponentProps<typeof ReservationDateTimeFields>> = {},
  ) {
    return render(
      <ReservationDateTimeFields
        formData={{ start: new Date(2026, 5, 1, 9, 0), end: new Date(2026, 5, 1, 9, 30) }}
        onChange={() => undefined}
        isCalendarDateDisabled={() => false}
        handleMonthChange={() => undefined}
        startTimeOptions={TIME_OPTIONS}
        availableTimeSlotMap={undefined}
        {...overrides}
      />,
    );
  }

  it("タイトル→年ナビ→月グリッドで遠い月へ移動し handleMonthChange に通知する", async () => {
    const user = userEvent.setup({ delay: null });
    const handleMonthChange = vi.fn();
    renderFields({ handleMonthChange });

    await user.click(screen.getByRole("button", { name: /2026\/06\/01/ }));
    expect(screen.getByRole("button", { name: "2026年 6月" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "2026年 6月" }));
    await user.click(screen.getByRole("button", { name: "次の年" }));
    await user.click(screen.getByRole("button", { name: "12月" }));

    expect(screen.getByRole("button", { name: "2027年 12月" })).toBeInTheDocument();
    expect(handleMonthChange).toHaveBeenLastCalledWith(new Date(2027, 11, 1));
  });

  it("日付を選ぶと既存の時刻を保持した start/end を onChange する", async () => {
    const user = userEvent.setup({ delay: null });
    const onChange = vi.fn();
    renderFields({
      formData: {
        start: new Date(2026, 5, 1, 9, 30),
        end: new Date(2026, 5, 1, 10, 0),
      },
      onChange,
    });

    await user.click(screen.getByRole("button", { name: /2026\/06\/01/ }));
    await user.click(screen.getByRole("button", { name: /2026年6月15日/ }));

    const last = onChange.mock.calls.at(-1)?.[0] as Partial<Reservation>;
    expect(last.start).toEqual(new Date(2026, 5, 15, 9, 30));
    expect(last.end).toEqual(new Date(2026, 5, 15, 10, 0));
  });
});
