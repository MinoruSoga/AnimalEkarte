import { render, screen } from "@testing-library/react";
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
});
