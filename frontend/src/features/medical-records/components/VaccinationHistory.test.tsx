import { describe, it, expect } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { VaccinationHistory } from "./VaccinationHistory";

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
    expect(screen.getByText("2026/01/01")).toHaveClass("w-[110px]");
    expect(screen.getByText("2027/01/01")).toHaveClass("w-[110px]");
  });
});
