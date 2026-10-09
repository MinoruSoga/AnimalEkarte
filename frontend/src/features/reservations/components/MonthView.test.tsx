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
    visitType: "revisit",
    type: "general_revisit",
    category: "general",
    doctor: "",
    doctorId: undefined,
    isDesignated: false,
    isTentative: false,
    intendedDate: undefined,
    hadIntendedDate: false,
    sourceMedicalRecordId: undefined,
    status: "confirmed",
    notes: "",
    petId: "pet-1",
    petType: undefined,
    ownerId: undefined,
    reservationTypeId: undefined,
    reservationRoute: null,
    clinicId: undefined,
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

  // PR#614 指摘: static バッジは pointer-events-none 配下で title に届かないため、
  // 誕生日情報はチップ（親 button）の title / accessible name に含める。
  it("誕生月ペットの予約チップは title と aria-label に誕生日・生年月日を含む", () => {
    renderMonthView([reservationOf({ petBirthDate: "2018-10-05" })]);

    const chip = screen.getByRole("button", { name: /再診 マメラ/ });
    expect(chip).toHaveAttribute("title", expect.stringContaining("今月誕生日"));
    expect(chip).toHaveAttribute("title", expect.stringContaining("2018-10-05"));
    expect(chip.getAttribute("aria-label")).toContain("今月誕生日");
    expect(chip.getAttribute("aria-label")).toContain("2018-10-05");
  });

  it("誕生月でないペットの予約チップは title・aria-label に誕生日情報を含まない", () => {
    renderMonthView([reservationOf({ petBirthDate: "2018-09-05" })]);

    const chip = screen.getByRole("button", { name: /再診 マメラ/ });
    expect(chip.getAttribute("title") ?? "").not.toContain("今月誕生日");
    expect(chip.getAttribute("aria-label") ?? "").not.toContain("今月誕生日");
  });
});

describe("MonthView 予約チップの accessible name (PR#614 再レビュー指摘1)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-10T00:00:00+09:00"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  // 子テキストで表示される意味情報（初診/再診・ペット名・飼主名・担当医・誕生月）は
  // 全て親 button の accessible name に合成する（WeekViewAppointmentCard と同方針
  // + MonthView では担当医も含める）。
  it("accessible name は 再診・ペット名・飼主名・担当医・誕生月情報を全て含む", () => {
    renderMonthView([reservationOf({ doctor: "田中 一郎", petBirthDate: "2018-10-05" })]);

    expect(
      screen.getByRole("button", {
        name: "再診 マメラ 宇野 いづみ 田中 一郎 今月誕生日（生年月日: 2018-10-05）",
      }),
    ).toBeInTheDocument();
  });

  it("担当医が空のとき accessible name に担当医を含めない", () => {
    renderMonthView([reservationOf({ doctor: "" })]);

    expect(screen.getByRole("button", { name: "再診 マメラ 宇野 いづみ" })).toBeInTheDocument();
  });

  it("tentative 予約の accessible name は 次回予定・未予約 ラベルを含む", () => {
    renderMonthView([reservationOf({ isTentative: true })]);

    const chip = screen.getByRole("button", {
      name: "次回予定・未予約 再診 マメラ 宇野 いづみ",
    });
    expect(chip).toBeInTheDocument();
    expect(chip).toHaveTextContent("次回予定・未予約");
  });

  it.each([
    ["petStatus=deceased", { petStatus: "deceased" as const }],
    ["petDeceasedAt 非null", { petDeceasedAt: "2026-09-01" }],
  ])(
    "死亡ペット（%s）はバッジ非描画かつ accessible name に誕生月情報を含まない",
    (_label, overrides) => {
      renderMonthView([reservationOf({ petBirthDate: "2018-10-05", ...overrides })]);

      expect(screen.queryByRole("img", { name: "今月誕生日" })).not.toBeInTheDocument();
      const chip = screen.getByRole("button", { name: /再診 マメラ/ });
      expect(chip).toHaveAccessibleName(expect.not.stringContaining("今月誕生日"));
      expect(chip).toHaveAccessibleName(expect.not.stringContaining("2018-10-05"));
    },
  );
});
