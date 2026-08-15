import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { Reservation } from "@/types";
import { AppointmentCard } from "./WeekViewAppointmentCard";

const appointment: Reservation = {
  id: "reservation-1",
  start: new Date("2026-07-21T00:23:00"),
  end: new Date("2026-07-21T00:38:00"),
  ownerName: "宇野 いづみ",
  petName: "マメラ",
  visitType: "return",
  type: "general_revisit",
  doctor: "",
  isDesignated: false,
  status: "confirmed",
  notes: "",
  petId: "pet-1",
  source: "manual",
};

describe("WeekView AppointmentCard touch target", () => {
  it("連続する15分予約の実buttonを44pxにし、hit areaを重ねない", () => {
    const nextAppointment: Reservation = {
      ...appointment,
      id: "reservation-2",
      start: new Date("2026-07-21T00:38:00"),
      end: new Date("2026-07-21T00:53:00"),
      petName: "ココ",
    };

    render(
      <>
        <AppointmentCard
          appointment={appointment}
          layoutStyle={{ left: "0%", width: "100%" }}
          onClick={vi.fn()}
        />
        <AppointmentCard
          appointment={nextAppointment}
          layoutStyle={{ left: "0%", width: "100%" }}
          onClick={vi.fn()}
        />
      </>,
    );

    const firstCard = screen.getByRole("button", { name: /00:23〜00:38/ });
    const secondCard = screen.getByRole("button", { name: /00:38〜00:53/ });
    const firstTop = Number.parseFloat(firstCard.style.top);
    const secondTop = Number.parseFloat(secondCard.style.top);
    const firstHeight = Number.parseFloat(firstCard.style.height);

    expect(firstCard).toHaveClass("min-h-11");
    expect(secondCard).toHaveClass("min-h-11");
    expect(firstCard).not.toHaveClass("overflow-visible", "after:h-11", "after:w-full");
    expect(firstCard).toHaveStyle({ height: "44px" });
    expect(secondTop).toBeGreaterThanOrEqual(firstTop + firstHeight);
  });
});
