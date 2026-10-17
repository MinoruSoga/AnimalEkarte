import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { VaccinationHistory } from "./VaccinationHistory";
import { MEDICAL_RECORD_FIXED_ACTION_LAYOUT } from "../constants/fixed-action-layout";

const BLANK = {
  vaccineId: 1,
  lot1: "",
  lot2: "",
  lot3: "",
  lot4: "",
  dateISO: "",
  nextDate: "",
  remarks: "",
};

const ITEMS = [
  {
    ...BLANK,
    id: 1,
    name: "狂犬病ワクチン",
    date: "26/1/1",
    dateISO: "2026-01-01",
    next: "27/1/1",
    nextDate: "2027-01-01",
  },
  {
    ...BLANK,
    id: 2,
    name: "混合ワクチン",
    date: "26/1/2",
    dateISO: "2026-01-02",
    next: "27/1/2",
    nextDate: "2027-01-02",
  },
];

describe("VaccinationHistory — カナ混同検索", () => {
  it("mobileでは全幅1列、lg以上で既存の右6列に戻る", () => {
    const { container } = render(<VaccinationHistory historyItems={ITEMS} />);

    expect(container.firstElementChild).toHaveClass("col-span-1", "lg:col-span-2");
    expect(container.firstElementChild).not.toHaveClass("col-span-2");

    // EMR-432: 右下固定ボタン群が過去の接種履歴の表を覆わない下余白
    const scroller = container.querySelector(".overflow-y-auto") as HTMLElement;
    expect(scroller).not.toBeNull();
    expect(scroller.className).toContain(MEDICAL_RECORD_FIXED_ACTION_LAYOUT.scrollBottomClearance);
  });

  it("空の検索語は全件表示する", () => {
    render(<VaccinationHistory historyItems={ITEMS} />);
    expect(screen.getByText("狂犬病ワクチン")).toBeInTheDocument();
    expect(screen.getByText("混合ワクチン")).toBeInTheDocument();
  });

  it("ひらがなで検索するとカタカナのワクチン名にヒットする", async () => {
    const katakanaItems = [
      { ...BLANK, id: 1, name: "キョウケンビョウワクチン", date: "2026-01-01", next: "2027-01-01" },
      { ...BLANK, id: 2, name: "コンゴウワクチン", date: "2026-01-02", next: "2027-01-02" },
    ];
    const user = userEvent.setup();
    render(<VaccinationHistory historyItems={katakanaItems} />);

    await user.type(screen.getByPlaceholderText("検索..."), "きょうけんびょう");

    await waitFor(() => {
      expect(screen.getByText("キョウケンビョウワクチン")).toBeInTheDocument();
      expect(screen.queryByText("コンゴウワクチン")).not.toBeInTheDocument();
    });
  });

  it("カタカナで検索するとひらがなのワクチン名にヒットする", async () => {
    const hiraganaItems = [
      { ...BLANK, id: 1, name: "きょうけんびょうわくちん", date: "2026-01-01", next: "2027-01-01" },
      { ...BLANK, id: 2, name: "こんごうわくちん", date: "2026-01-02", next: "2027-01-02" },
    ];
    const user = userEvent.setup();
    render(<VaccinationHistory historyItems={hiraganaItems} />);

    await user.type(screen.getByPlaceholderText("検索..."), "キョウケンビョウ");

    await waitFor(() => {
      expect(screen.getByText("きょうけんびょうわくちん")).toBeInTheDocument();
      expect(screen.queryByText("こんごうわくちん")).not.toBeInTheDocument();
    });
  });

  it("クリアボタンで全件に戻る", async () => {
    const user = userEvent.setup();
    render(<VaccinationHistory historyItems={ITEMS} />);

    await user.type(screen.getByPlaceholderText("検索..."), "狂犬病");
    await user.click(screen.getByRole("button", { name: "クリア" }));

    await waitFor(() => {
      expect(screen.getByText("狂犬病ワクチン")).toBeInTheDocument();
      expect(screen.getByText("混合ワクチン")).toBeInTheDocument();
    });
  });
});

describe("VaccinationHistory — 日付表示 (EMR-429)", () => {
  it("実施日・次予定は dateISO / nextDate を YYYY/MM/DD で表示する", () => {
    render(<VaccinationHistory historyItems={ITEMS} />);

    expect(screen.getAllByText("2026/01/01")).not.toHaveLength(0);
    expect(screen.getAllByText("2027/01/01")).not.toHaveLength(0);
    expect(screen.queryByText("26/1/1")).toBeNull();
  });

  it("日付未設定の項目は '-' を表示する", () => {
    const items = [
      { ...BLANK, id: 1, name: "未設定ワクチン", date: "-", dateISO: "", next: "-", nextDate: "" },
    ];
    render(<VaccinationHistory historyItems={items} />);

    const row = screen.getByText("未設定ワクチン").parentElement;
    expect(row?.querySelectorAll("div")[1]).toHaveTextContent("-");
    expect(row?.querySelectorAll("div")[2]).toHaveTextContent("-");
  });

  it("ヘッダと行の日付列は折り返さない幅 w-[110px] を持つ", () => {
    render(<VaccinationHistory historyItems={ITEMS} />);

    const headerDate = screen
      .getAllByText("実施日")
      .find((el) => el.classList.contains("w-[110px]"));
    const headerNext = screen
      .getAllByText("次予定")
      .find((el) => el.classList.contains("w-[110px]"));
    expect(headerDate).toBeDefined();
    expect(headerNext).toBeDefined();
    // EMR-433: 次予定セルは日付（nowrap span）とバッジの縦積み。日付の span は折り返さず、セルは w-[110px] のまま
    for (const dateText of ["2026/01/01", "2027/01/01"]) {
      const text = screen.getByText(dateText);
      expect(text).toHaveClass("whitespace-nowrap");
      expect(text.closest('[class*="w-[110px]"]')).not.toBeNull();
    }
  });
});

describe("VaccinationHistory 次予定の期限バッジ (EMR-433)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-07-24T06:00:00Z")); // JST 2026-07-24
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  function makeItem(overrides: Record<string, unknown>) {
    return {
      ...BLANK,
      id: 0,
      name: "ワクチン",
      date: "",
      dateISO: "",
      next: "",
      nextDate: "",
      ...overrides,
    } as (typeof ITEMS)[number];
  }

  function nextCellOf(dateText: string) {
    const dateEl = screen.getByText(dateText);
    const cell = dateEl.closest('[class*="w-[110px]"]');
    expect(cell).not.toBeNull();
    return { cell: cell as HTMLElement, row: (cell as HTMLElement).parentElement as HTMLElement };
  }

  it("(a) 次回予定日が過去の記録（最新）に期限切れバッジを日付の下に出す", () => {
    render(
      <VaccinationHistory
        historyItems={[makeItem({ id: 1, dateISO: "2026-01-01", nextDate: "2026-07-23" })]}
      />,
    );

    const { cell } = nextCellOf("2026/07/23");
    expect(cell).toHaveClass("w-[110px]");
    expect(cell.textContent).toContain("期限切れ");
    // 日付とバッジは縦に並ぶ
    expect(cell.querySelector(".flex-col")).not.toBeNull();
  });

  it("(b) 次回予定日が30日以内の記録に期限間近バッジを出す", () => {
    render(
      <VaccinationHistory
        historyItems={[makeItem({ id: 1, dateISO: "2026-01-01", nextDate: "2026-07-25" })]}
      />,
    );

    const { cell } = nextCellOf("2026/07/25");
    expect(cell.textContent).toContain("期限間近");
    expect(cell.textContent).not.toContain("期限切れ");
  });

  it("(c) 置き換わった記録にはバッジを出さず、最新の記録には出す", () => {
    render(
      <VaccinationHistory
        historyItems={[
          makeItem({
            id: 1,
            name: "混合ワクチン",
            vaccineId: 10,
            dateISO: "2026-01-01",
            nextDate: "2026-07-23",
          }),
          makeItem({
            id: 2,
            name: "混合ワクチン",
            vaccineId: 10,
            dateISO: "2026-02-01",
            nextDate: "2026-07-20",
          }),
        ]}
      />,
    );

    const oldRow = nextCellOf("2026/07/23").row;
    const newRow = nextCellOf("2026/07/20").row;
    expect(oldRow.textContent).not.toContain("期限切れ");
    expect(oldRow.textContent).not.toContain("期限間近");
    expect(newRow.textContent).toContain("期限切れ");
  });

  it("(d) 名前が「他院」で始まる記録は置き換わらず記録ごとにバッジを出す", () => {
    render(
      <VaccinationHistory
        historyItems={[
          makeItem({
            id: 1,
            name: "他院で接種したワクチン",
            vaccineId: 10,
            dateISO: "2026-01-01",
            nextDate: "2026-07-23",
          }),
          makeItem({
            id: 2,
            name: "混合ワクチン",
            vaccineId: 10,
            dateISO: "2026-02-01",
            nextDate: "2026-07-20",
          }),
        ]}
      />,
    );

    const externalRow = nextCellOf("2026/07/23").row;
    const newRow = nextCellOf("2026/07/20").row;
    expect(externalRow.textContent).toContain("期限切れ");
    expect(newRow.textContent).toContain("期限切れ");
  });

  it("(e) 絞り込みで新しい記録が隠れても、古い記録にバッジが戻らない", () => {
    render(
      <VaccinationHistory
        historyItems={[
          makeItem({
            id: 1,
            name: "混合ワクチン",
            vaccineId: 10,
            dateISO: "2026-01-01",
            nextDate: "2026-07-23",
          }),
          makeItem({
            id: 2,
            name: "フィラリア",
            vaccineId: 10,
            dateISO: "2026-02-01",
            nextDate: "2026-07-20",
          }),
        ]}
      />,
    );

    fireEvent.change(screen.getByLabelText("検索単語"), { target: { value: "混合" } });
    act(() => {
      vi.advanceTimersByTime(500);
    });

    // 新しい記録（フィラリア）は検索で隠れる
    expect(screen.queryByText("2026/07/20")).not.toBeInTheDocument();
    // それでも古い記録にバッジは戻らない
    const oldRow = nextCellOf("2026/07/23").row;
    expect(oldRow.textContent).not.toContain("期限切れ");
    expect(oldRow.textContent).not.toContain("期限間近");
  });
});
