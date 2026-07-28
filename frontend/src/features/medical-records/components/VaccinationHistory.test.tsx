import { describe, it, expect } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { VaccinationHistory } from "./VaccinationHistory";

const BLANK = { vaccineId: 1, lot1: "", lot2: "", lot3: "", lot4: "", nextDate: "", remarks: "" };

const ITEMS = [
  { ...BLANK, id: 1, name: "狂犬病ワクチン", date: "2026-01-01", next: "2027-01-01" },
  { ...BLANK, id: 2, name: "混合ワクチン", date: "2026-01-02", next: "2027-01-02" },
];

describe("VaccinationHistory — カナ混同検索", () => {
  it("mobileでは全幅1列、lg以上で既存の右6列に戻る", () => {
    const { container } = render(<VaccinationHistory historyItems={ITEMS} />);

    expect(container.firstElementChild).toHaveClass("col-span-1", "lg:col-span-6");
    expect(container.firstElementChild).not.toHaveClass("col-span-6");
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
