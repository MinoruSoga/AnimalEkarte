import { render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { Reservation } from "@/types";
import { MonthView } from "./MonthView";

function reservationOf(overrides: Partial<Reservation> = {}): Reservation {
  return {
    id: "reservation-1",
    start: new Date("2026-10-15T10:00:00"),
    end: new Date("2026-10-15T10:15:00"),
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
    ...overrides,
  };
}

function renderMonthView(appointments: Reservation[]) {
  render(
    <MonthView
      currentDate={new Date("2026-10-01T00:00:00")}
      appointments={appointments}
      onAppointmentClick={vi.fn()}
    />,
  );
}

describe("MonthView 誕生月バッジ (EMR-301)", () => {
  beforeEach(() => {
    // 現在JST月を 2026-10 に固定して決定的にする
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-10T00:00:00+09:00"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("今月(JST)誕生月のペットの予約チップに Cake 小アイコンを表示する", () => {
    renderMonthView([reservationOf({ petBirthDate: "2018-10-05" })]);

    // チップ自体が button のため static 変種（role=img + title=生年月日）
    const badge = screen.getByRole("img", { name: "今月誕生日" });
    expect(badge).toHaveAttribute("title", "マメラの生年月日: 2018-10-05");
    // チップ内に入れ子 button を生成しない
    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it.each([
    ["誕生月でない", "2018-09-05"],
    ["birth_date 未登録", undefined],
    ["暦上無効", "2018-02-30"],
  ])("%s のペットには Cake アイコンを表示しない", (_label, petBirthDate) => {
    renderMonthView([reservationOf({ petBirthDate })]);

    expect(screen.queryByRole("img", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("死亡ペット（petStatus=deceased）には誕生月でも Cake アイコンを表示しない", () => {
    renderMonthView([reservationOf({ petBirthDate: "2018-10-05", petStatus: "deceased" })]);

    expect(screen.queryByRole("img", { name: "今月誕生日" })).not.toBeInTheDocument();
  });
});
