import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { Reservation } from "@/types";
import type { ReservationTypeColor } from "./week-view-grid-constants";
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

function colorEntry(overrides: Partial<ReservationTypeColor> = {}): ReservationTypeColor {
  return {
    style: { backgroundColor: "rgba(255, 102, 0, 0.1)", color: "#ff6600", borderColor: "rgba(255, 102, 0, 0.3)" },
    dotStyle: { backgroundColor: "#ff6600" },
    hex: "#ff6600",
    ...overrides,
  };
}

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

describe("WeekView AppointmentCard inactive type label (BUG-016)", () => {
  it("appends （無効） to accessible name and title when dynamicColorMap marks type inactive", () => {
    const dynamicColorMap = new Map<string, ReservationTypeColor>([
      [appointment.type, colorEntry({ isInactive: true })],
    ]);

    render(
      <AppointmentCard
        appointment={appointment}
        layoutStyle={{ left: "0%", width: "100%" }}
        onClick={vi.fn()}
        dynamicColorMap={dynamicColorMap}
      />,
    );

    const card = screen.getByRole("button", { name: /general_revisit（無効）/ });
    expect(card).toHaveAccessibleName(expect.stringContaining("（無効）"));
    expect(card).toHaveAttribute("title", expect.stringContaining("general_revisit（無効）"));
  });

  it("does not append （無効） when the mapped type is active", () => {
    const dynamicColorMap = new Map<string, ReservationTypeColor>([
      [appointment.type, colorEntry({ isInactive: false })],
    ]);

    render(
      <AppointmentCard
        appointment={appointment}
        layoutStyle={{ left: "0%", width: "100%" }}
        onClick={vi.fn()}
        dynamicColorMap={dynamicColorMap}
      />,
    );

    const card = screen.getByRole("button", { name: /00:23〜00:38/ });
    expect(card).toHaveAccessibleName(expect.not.stringContaining("（無効）"));
    expect(card.getAttribute("title") ?? "").not.toContain("（無効）");
  });

  it("keeps cancelled dimming classes when the type is also inactive", () => {
    const dynamicColorMap = new Map<string, ReservationTypeColor>([
      [appointment.type, colorEntry({ isInactive: true })],
    ]);

    render(
      <AppointmentCard
        appointment={{ ...appointment, status: "cancelled" }}
        layoutStyle={{ left: "0%", width: "100%" }}
        onClick={vi.fn()}
        dynamicColorMap={dynamicColorMap}
      />,
    );

    const card = screen.getByRole("button", { name: /general_revisit（無効）/ });
    expect(card).toHaveClass("opacity-60");
    expect(card).toHaveClass("line-through");
  });
});
