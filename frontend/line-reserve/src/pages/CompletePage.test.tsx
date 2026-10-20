import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { CompletePage } from "./CompletePage";
import type { ReservationFlow } from "../types/models";

const baseFlow: ReservationFlow = {
  customerInfo: {
    name: "",
    phone: "",
    ownerName: "",
    pets: [],
  },
  courseId: null,
  courseName: "",
  courseCategory: "general",
  staffId: 0,
  staffName: "",
  date: "2026-08-01",
  startTime: "1000",
  endTime: "1030",
  requestText: "",
  trimmingCourseId: null,
  trimmingCourseName: "",
  trimmingOptionIds: [],
};

function renderCompletePage(flow: ReservationFlow) {
  render(
    <CompletePage
      reservationId={0}
      notes=""
      flow={flow}
      onMyReservations={vi.fn()}
      onNewReservation={vi.fn()}
    />,
  );
}

describe("CompletePage", () => {
  it("予約日時を日本語の日付と HH:MM の時間帯で表示する", () => {
    renderCompletePage(baseFlow);

    expect(screen.getByText("2026年08月01日(土) 10:00〜10:30")).toBeInTheDocument();
  });

  it("4文字を超える時刻は先頭4文字に切り詰めて表示する", () => {
    renderCompletePage({
      ...baseFlow,
      startTime: "12345",
      endTime: "67890",
    });

    expect(screen.getByText("2026年08月01日(土) 12:34〜67:89")).toBeInTheDocument();
  });

  it("EMR-474: 再診問診ハンドラがあるときだけ任意の CTA を表示し、押すと問診へ進む", async () => {
    const user = userEvent.setup();
    const onRevisitIntake = vi.fn();
    render(
      <CompletePage
        reservationId={12001}
        notes="R-20260801-0001"
        flow={baseFlow}
        onMyReservations={vi.fn()}
        onNewReservation={vi.fn()}
        onRevisitIntake={onRevisitIntake}
      />,
    );

    expect(screen.getByText(/回答しなくても予約は確定しています/)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "事前問診に回答する（任意）" }));
    expect(onRevisitIntake).toHaveBeenCalledTimes(1);
  });

  it("EMR-474: ハンドラが無ければ問診 CTA を表示しない", () => {
    renderCompletePage(baseFlow);

    expect(
      screen.queryByRole("button", { name: "事前問診に回答する（任意）" }),
    ).not.toBeInTheDocument();
  });

  it("EMR-474: 複数頭を選んだ予約では、ほかのペットは予約一覧の各予約から回答する旨を示す", () => {
    render(
      <CompletePage
        reservationId={12001}
        notes=""
        flow={{
          ...baseFlow,
          customerInfo: {
            ...baseFlow.customerInfo,
            pets: [
              { name: "合成ペットA", type: "犬" },
              { name: "合成ペットB", type: "猫" },
            ],
          },
        }}
        onMyReservations={vi.fn()}
        onNewReservation={vi.fn()}
        onRevisitIntake={vi.fn()}
      />,
    );

    expect(screen.getByText(/問診は予約ごとに1頭分です/)).toBeInTheDocument();
  });
});
