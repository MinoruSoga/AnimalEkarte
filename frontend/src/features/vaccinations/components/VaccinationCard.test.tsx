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
  version: 1,
} satisfies VaccinationRecord;

function renderCard(nextDate: string) {
  return render(<VaccinationCard vaccination={{ ...vaccination, nextDate }} />);
}

describe("VaccinationCard 次回予定の期限バッジと日付表記 (EMR-415 Q2/X4)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-07-24T06:00:00Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("X4: 接種日と次回予定日を YYYY/MM/DD で表示する", () => {
    renderCard("2026-07-24");

    expect(screen.getByText("接種: 2026/07/23")).toBeInTheDocument();
    expect(screen.getByText(/次回: 2026\/07\/24/)).toBeInTheDocument();
    expect(screen.queryByText(/次回: 2026-07-24/)).not.toBeInTheDocument();
  });

  it("JST当日は期限切れにならず、30日以内なら期限間近バッジ (Q2)", () => {
    renderCard("2026-07-24");

    expect(screen.getByText("期限間近")).toBeInTheDocument();
    expect(screen.queryByText("期限切れ")).not.toBeInTheDocument();
    expect(screen.queryByText("（期限超過）")).not.toBeInTheDocument();
  });

  it("過去日は期限切れバッジで、「（期限超過）」の文言は出さない (Q2)", () => {
    renderCard("2026-07-23");

    expect(screen.getByText("期限切れ")).toBeInTheDocument();
    expect(screen.queryByText("（期限超過）")).not.toBeInTheDocument();
  });

  it("30日超の未来日はバッジなし (Q2)", () => {
    renderCard("2026-12-01");

    expect(screen.queryByText("期限切れ")).not.toBeInTheDocument();
    expect(screen.queryByText("期限間近")).not.toBeInTheDocument();
  });

  it("次回日が空欄なら次回表示を出さない", () => {
    renderCard("");

    expect(screen.queryByText(/次回:/)).not.toBeInTheDocument();
    expect(screen.queryByText("期限切れ")).not.toBeInTheDocument();
  });
});
