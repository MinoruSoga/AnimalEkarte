import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { Reservation } from "@/types";
import type { ReservationTypeColor } from "../lib/week-view-grid-constants";
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
    style: {
      backgroundColor: "rgba(255, 102, 0, 0.1)",
      color: "#ff6600",
      borderColor: "rgba(255, 102, 0, 0.3)",
    },
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

describe("WeekView AppointmentCard tentative (EMR-308)", () => {
  // tentative（次回予定・未予約）は意図日 00:00 JST の sentinel 時刻を持つ。
  // 時刻表示ではなく明示テキストで識別し、ドラッグも無効化する。
  const tentativeAppointment: Reservation = {
    ...appointment,
    id: "tentative-1",
    start: new Date("2026-07-21T00:00:00"),
    end: new Date("2026-07-21T00:30:00"),
    isTentative: true,
    intendedDate: "2026-07-21",
  };

  it("次回予定・未予約 の明示テキストを表示し、aria-label にも含める", () => {
    render(
      <AppointmentCard
        appointment={tentativeAppointment}
        layoutStyle={{ left: "0%", width: "100%" }}
        onClick={vi.fn()}
      />,
    );

    const card = screen.getByRole("button", { name: /次回予定・未予約/ });
    expect(card).toBeInTheDocument();
    // 色だけに頼らない識別子: 破線スタイル + title でも識別できる
    expect(card).toHaveClass("border-dashed");
    expect(card).toHaveAttribute("title", expect.stringContaining("次回予定・未予約"));
  });

  it("aria-label は sentinel 時刻を読み上げない（次回予定・未予約 + 患者情報のみ）", () => {
    // EMR-308 review fix: tentative の 00:00〜00:30 は sentinel であり実時刻ではない。
    // スクリーンリーダーに実在しない時刻を読ませない（spec: UI・通知・帳票は時刻なし）。
    render(
      <AppointmentCard
        appointment={tentativeAppointment}
        layoutStyle={{ left: "0%", width: "100%" }}
        onClick={vi.fn()}
      />,
    );

    const card = screen.getByRole("button", { name: /次回予定・未予約/ });
    expect(card).toHaveAccessibleName(expect.not.stringMatching(/\d{2}:\d{2}/));
    expect(card).toHaveAccessibleName(expect.stringContaining("マメラ"));
  });

  it("tentative はドラッグ用カーソルではなく通常カーソルにする", () => {
    render(
      <AppointmentCard
        appointment={tentativeAppointment}
        layoutStyle={{ left: "0%", width: "100%" }}
        onClick={vi.fn()}
        onUpdate={vi.fn()}
      />,
    );

    const card = screen.getByRole("button", { name: /次回予定・未予約/ });
    expect(card).toHaveClass("cursor-pointer");
    expect(card).not.toHaveClass("cursor-grab");
  });

  it("通常予約は従来どおり時刻表示で tentative バッジを出さない", () => {
    render(
      <AppointmentCard
        appointment={appointment}
        layoutStyle={{ left: "0%", width: "100%" }}
        onClick={vi.fn()}
      />,
    );

    const card = screen.getByRole("button", { name: /00:23〜00:38/ });
    expect(card).not.toHaveClass("border-dashed");
    expect(card).toHaveClass("cursor-grab");
    expect(screen.queryByText("次回予定・未予約")).not.toBeInTheDocument();
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
