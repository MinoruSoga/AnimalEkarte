import { render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { C } from "@/lib/design-tokens";
import type { VaccinationRecord } from "../api/transforms";
import { VaccinationCard } from "./VaccinationCard";

const vaccination = {
  id: "vaccination-1",
  petId: "pet-1",
  ownerName: "合成監査飼主",
  petName: "合成監査ペット",
  vaccineId: "vaccine-1",
  vaccineName: "合成監査ワクチン",
  doctor: "合成監査スタッフ",
  date: "2026-07-23",
  nextDate: "2026-07-24",
  nextScheduleType: "1year",
  lot1: "SYN-LOT-1",
  lot2: undefined,
  lot3: undefined,
  lot4: undefined,
  supplemental: "合成補足",
  remarks: "合成備考",
} satisfies VaccinationRecord;

function renderCard(nextDate: string) {
  return render(<VaccinationCard vaccination={{ ...vaccination, nextDate }} />);
}

describe("VaccinationCard 期限超過表示", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-07-24T06:00:00Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("JST当日は期限超過にならない", () => {
    renderCard("2026-07-24");

    expect(screen.getByText("次回: 2026-07-24")).toBeInTheDocument();
    expect(screen.queryByText("（期限超過）")).not.toBeInTheDocument();
  });

  it("過去日は既存のdanger表現で期限超過になる", () => {
    renderCard("2026-07-23");

    const overdueLabel = screen.getByText("（期限超過）");
    expect(overdueLabel).toBeInTheDocument();
    expect(overdueLabel.closest("div")).toHaveClass(C.danger);
  });

  it("未来日は期限超過にならない", () => {
    renderCard("2026-07-25");

    expect(screen.getByText("次回: 2026-07-25")).toBeInTheDocument();
    expect(screen.queryByText("（期限超過）")).not.toBeInTheDocument();
  });

  it("次回日が空欄なら次回表示を出さない", () => {
    renderCard("");

    expect(screen.queryByText(/次回:/)).not.toBeInTheDocument();
    expect(screen.queryByText("（期限超過）")).not.toBeInTheDocument();
  });
});
