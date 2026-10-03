import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { FilterAddPopover } from "./FilterAddPopover";
import type { FilterProperty } from "./types";

const dateProp: FilterProperty = {
  key: "diagnosis_date",
  label: "診療日",
  type: "date-range",
};

describe("FilterAddPopover date-range", () => {
  it("日付プロパティ選択で開始/終了切替エディタを表示し、プリセット適用で onAdd する", async () => {
    const onAdd = vi.fn();
    render(<FilterAddPopover properties={[dateProp]} activeFilters={[]} onAdd={onAdd} />);
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "フィルタを追加" }));
    await user.click(screen.getByRole("option", { name: "診療日" }));

    // DateValueEditor: 開始日/終了日の切替ボタンと年付きタイトルが出る
    expect(screen.getByRole("button", { name: "開始日" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "終了日" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /\d{4}年 \d{1,2}月/ })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "今日" }));

    expect(onAdd).toHaveBeenCalledTimes(1);
    const filter = onAdd.mock.calls[0][0];
    expect(filter).toMatchObject({
      key: "diagnosis_date",
      condition: "is_between",
      displayValue: "今日",
    });
    expect(filter.value.from).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    expect(filter.value.from).toBe(filter.value.to);
  });

  it("カレンダーで同日を開始→終了と選ぶと単日レンジで onAdd する", async () => {
    const onAdd = vi.fn();
    render(<FilterAddPopover properties={[dateProp]} activeFilters={[]} onAdd={onAdd} />);
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "フィルタを追加" }));
    await user.click(screen.getByRole("option", { name: "診療日" }));

    const today = new Date();
    const dayName = new RegExp(
      `${today.getFullYear()}年${today.getMonth() + 1}月${today.getDate()}日`,
    );
    await user.click(screen.getByRole("button", { name: dayName }));
    await user.click(screen.getByRole("button", { name: dayName }));

    expect(onAdd).toHaveBeenCalledTimes(1);
    const filter = onAdd.mock.calls[0][0];
    const pad = (n: number) => String(n).padStart(2, "0");
    const iso = `${today.getFullYear()}-${pad(today.getMonth() + 1)}-${pad(today.getDate())}`;
    expect(filter).toMatchObject({
      key: "diagnosis_date",
      condition: "is_between",
      value: { from: iso, to: iso },
    });
  });
});
