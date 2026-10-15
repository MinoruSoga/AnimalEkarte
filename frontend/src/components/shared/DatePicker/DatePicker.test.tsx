import { describe, expect, it, vi } from "vitest";
import { useState } from "react";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { DatePicker } from "./DatePicker";
import { CalendarNav, ClearButton, MonthGrid, YearNav } from "./DatePickerParts";
import { C } from "@/lib/design-tokens";

// react-day-picker v9 の日ボタン名は ja locale で "2026年10月20日火曜日" 形式
function dayButton(isoDay: string) {
  const [year, month, day] = isoDay.split("-").map(Number);
  return screen.getByRole("button", {
    name: new RegExp(`${year}年${month}月${day}日`),
  });
}

describe("DatePicker accessible naming", () => {
  it("業務ラベルをテキスト入力とカレンダー操作へ伝える", () => {
    render(<DatePicker value="" onChange={vi.fn()} ariaLabel="開始日" />);

    expect(screen.getByRole("textbox", { name: "開始日" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "開始日を選択" })).toBeInTheDocument();
  });
});

describe("DatePicker single mode form serialization", () => {
  it("名前付き日付を表示形式ではなく ISO 形式で 1 つだけ送信する", () => {
    const { container } = render(
      <form>
        <DatePicker id="visit-date" name="visit_date" value="2026-10-14" onChange={vi.fn()} />
      </form>,
    );
    const form = container.querySelector("form");

    expect(form).not.toBeNull();
    expect(new FormData(form!).getAll("visit_date")).toEqual(["2026-10-14"]);
    expect(screen.getByRole("textbox")).not.toHaveAttribute("name");
    expect(container.querySelectorAll('[name="visit_date"]')).toHaveLength(1);
  });

  it("空の日付も空文字で送信する", () => {
    const { container } = render(
      <form>
        <DatePicker name="visit_date" value="" onChange={vi.fn()} />
      </form>,
    );
    const form = container.querySelector("form");

    expect(form).not.toBeNull();
    expect(new FormData(form!).getAll("visit_date")).toEqual([""]);
  });

  it("disabled では表示欄と clear 操作を無効にし FormData から除外する", () => {
    const { container } = render(
      <form>
        <DatePicker name="visit_date" value="2026-10-14" onChange={vi.fn()} disabled />
      </form>,
    );
    const form = container.querySelector("form");

    expect(form).not.toBeNull();
    expect(screen.getByRole("textbox")).toBeDisabled();
    expect(screen.getByRole("button", { name: "カレンダーを開く" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "日付をクリア" })).toBeDisabled();
    expect(new FormData(form!).has("visit_date")).toBe(false);
  });

  it("disabled では clear/Today と開いたカレンダー操作が値を変更しない", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const { rerender } = render(<DatePicker value="2026-10-14" onChange={onChange} />);
    await user.click(screen.getByRole("button", { name: "カレンダーを開く" }));
    expect(screen.getByRole("button", { name: "Today" })).toBeInTheDocument();

    rerender(<DatePicker value="2026-10-14" onChange={onChange} disabled />);
    expect(screen.queryByRole("button", { name: "Today" })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "日付をクリア" })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "日付をクリア" }));
    expect(onChange).not.toHaveBeenCalled();
  });
});

describe("DatePicker disabledDays single mode behavior", () => {
  it.each([
    ["before", "2026-10-13"],
    ["after", "2026-10-17"],
  ])("手入力の %s 日付 %s は値を通知しつつ native validity で無効化する", async (_label, value) => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <DatePicker
        value=""
        onChange={onChange}
        disabledDays={{ before: new Date(2026, 9, 14), after: new Date(2026, 9, 16) }}
      />,
    );
    const input = screen.getByRole<HTMLInputElement>("textbox");

    await user.click(input);
    await user.type(input, value);
    await user.keyboard("{Enter}");

    expect(onChange).toHaveBeenCalledWith(value);
    expect(input.validity.valid).toBe(false);
    expect(input.validationMessage).toBeTruthy();
  });

  it("disabledDays に一致する今日を Today 操作で選択できない", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    render(<DatePicker value="" onChange={onChange} disabledDays={today} />);

    await user.click(screen.getByRole("button", { name: "カレンダーを開く" }));

    const todayButton = screen.getByRole("button", { name: "Today" });
    expect(todayButton).toBeDisabled();
    expect(onChange).not.toHaveBeenCalled();
  });

  it("範囲外の手入力は ISO hidden 値を保持しつつ native validation で送信を拒否する", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn((event: React.FormEvent<HTMLFormElement>) => event.preventDefault());
    function Form() {
      const [value, setValue] = useState("");
      return (
        <form onSubmit={onSubmit}>
          <DatePicker
            name="visit_date"
            ariaLabel="来院日"
            value={value}
            onChange={setValue}
            required
            disabledDays={{ before: new Date(2026, 9, 14) }}
          />
          <button type="submit">送信</button>
        </form>
      );
    }
    render(<Form />);
    const input = screen.getByRole<HTMLInputElement>("textbox", { name: "来院日" });
    await user.type(input, "2026/10/13");
    await user.tab();

    const form = input.closest("form");
    expect(form).not.toBeNull();
    expect(new FormData(form!).get("visit_date")).toBe("2026-10-13");
    expect(input.validity.valid).toBe(false);
    await user.click(screen.getByRole("button", { name: "送信" }));
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it.each([
    ["min", "2026-10-14", "2026/10/13"],
    ["max", "2026-10-14", "2026/10/15"],
  ])("%s 境界外の手入力は値を通知し、フォーム送信を拒否する", async (bound, limit, typedValue) => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const onSubmit = vi.fn((event: React.FormEvent<HTMLFormElement>) => event.preventDefault());
    function Form() {
      const [value, setValue] = useState("");
      return (
        <form onSubmit={onSubmit}>
          <DatePicker
            name="visit_date"
            value={value}
            onChange={(nextValue) => {
              onChange(nextValue);
              setValue(nextValue);
            }}
            min={bound === "min" ? limit : undefined}
            max={bound === "max" ? limit : undefined}
          />
          <button type="submit">送信</button>
        </form>
      );
    }
    render(<Form />);
    const input = screen.getByRole<HTMLInputElement>("textbox");

    await user.type(input, typedValue);
    await user.tab();

    const form = input.closest("form");
    expect(form).not.toBeNull();
    expect(onChange).toHaveBeenCalledWith(typedValue.replaceAll("/", "-"));
    expect(new FormData(form!).get("visit_date")).toBe(typedValue.replaceAll("/", "-"));
    expect(input.validity.valid).toBe(false);
    await user.click(screen.getByRole("button", { name: "送信" }));
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("popup 表示中に disabled へ変わったら閉じ、既存 callback を実行できない", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const { rerender } = render(<DatePicker value="2026-10-14" onChange={onChange} />);

    await user.click(screen.getByRole("button", { name: "カレンダーを開く" }));
    expect(screen.getByRole("button", { name: "Today" })).toBeInTheDocument();

    rerender(<DatePicker value="2026-10-14" onChange={onChange} disabled />);

    expect(screen.queryByRole("button", { name: "Today" })).not.toBeInTheDocument();
    expect(onChange).not.toHaveBeenCalled();
  });
});

describe("DatePicker — 44px touch targets", () => {
  it("CalendarNavの前月・タイトル・次月を44x44px以上に保つ", () => {
    render(
      <CalendarNav
        displayMonth={new Date(2026, 3, 1)}
        onPrev={() => {}}
        onNext={() => {}}
        onTitleClick={() => {}}
      />,
    );

    const prev = screen.getByRole("button", { name: "前の月" });
    const title = screen.getByRole("button", { name: "2026年 4月" });
    const next = screen.getByRole("button", { name: "次の月" });

    for (const button of [prev, title, next]) {
      expect(button).toHaveClass("min-h-11", "min-w-11");
    }
    expect(prev.querySelector("svg")).toHaveClass("size-5");
    expect(title).toHaveClass("text-sm");
    expect(next.querySelector("svg")).toHaveClass("size-5");
  });

  it("YearNavの前年・次年を44x44px以上に保つ", () => {
    render(<YearNav year={2026} onPrevYear={() => {}} onNextYear={() => {}} />);

    const prev = screen.getByRole("button", { name: "前の年" });
    const next = screen.getByRole("button", { name: "次の年" });

    for (const button of [prev, next]) {
      expect(button).toHaveClass("min-h-11", "min-w-11");
      expect(button.querySelector("svg")).toHaveClass("size-5");
    }
  });

  it("MonthGridの各月を44x44px以上にし、文字サイズを維持する", () => {
    render(<MonthGrid currentMonth={3} onSelect={() => {}} />);

    const monthButtons = screen.getAllByRole("button");
    expect(monthButtons).toHaveLength(12);
    for (const button of monthButtons) {
      expect(button).toHaveClass("min-h-11", "min-w-11", "text-sm");
    }
  });

  it("ClearButtonを44x44px以上にし、glyphサイズを維持する", () => {
    render(<ClearButton onClick={() => {}} />);

    const clear = screen.getByRole("button", { name: "日付をクリア" });
    expect(clear).toHaveClass("min-h-11", "min-w-11", "-my-px");
    expect(clear.querySelector("svg")).toHaveClass("size-5");
  });

  it("single modeのcalendar triggerを44x44px以上に保つ", () => {
    render(<DatePicker value="" onChange={() => {}} />);

    expect(screen.getByRole("button", { name: "カレンダーを開く" })).toHaveClass(
      "min-h-11",
      "min-w-11",
      "-my-px",
    );
    const input = screen.getByPlaceholderText("日付を選択…");
    expect(input).toHaveClass("min-h-11");
    expect(input.parentElement).toHaveClass("focus-within:ring-1", "focus-within:ring-ring");
  });

  it("single modeのToday buttonを44x44px以上に保つ", async () => {
    const user = userEvent.setup();
    render(<DatePicker value="" onChange={() => {}} />);

    await user.click(screen.getByRole("button", { name: "カレンダーを開く" }));

    expect(screen.getByRole("button", { name: "Today" })).toHaveClass("min-h-11", "min-w-11");
  });

  it("flex内で複数配置しても親幅を押し広げない", () => {
    render(<DatePicker value="" onChange={() => {}} placeholder="開始日" />);

    expect(screen.getByPlaceholderText("開始日").parentElement).toHaveClass("min-w-0");
  });

  it("range modeのpopover triggerとclear操作を別buttonにしてinteractive要素をネストしない", () => {
    const { container } = render(
      <DatePicker mode="range" value="2026-04-01~2026-04-30" onChange={() => {}} />,
    );

    expect(screen.getByRole("button", { name: /2026\/4\/1.*2026\/4\/30/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "日付をクリア" })).toBeInTheDocument();
    expect(container.querySelector("button button")).toBeNull();
  });

  // EMR-319: 必須未入力時に親フォームが ariaInvalid を渡すと aria-invalid+赤枠になる
  it("single modeでariaInvalidを渡すと入力にaria-invalidと赤枠が付く", () => {
    render(<DatePicker value="" onChange={() => {}} ariaInvalid />);

    const input = screen.getByPlaceholderText("日付を選択…");
    expect(input).toHaveAttribute("aria-invalid", "true");
    expect(input.parentElement).toHaveClass("ring-2");
  });

  it("single modeでariaInvalid未指定ならaria-invalidは付かない", () => {
    render(<DatePicker value="" onChange={() => {}} />);

    expect(screen.getByPlaceholderText("日付を選択…")).not.toHaveAttribute("aria-invalid");
  });

  // EMR-319: エラー要素 id を内部 input の aria-describedby へ転送する
  // （FormFieldError とセットで使う NumberInput と同じ契約）。
  it("single modeでariaDescribedByを渡すと内部 input に aria-describedby が付く", () => {
    render(<DatePicker value="" onChange={() => {}} ariaDescribedBy="some-date-error" />);

    expect(screen.getByPlaceholderText("日付を選択…")).toHaveAttribute(
      "aria-describedby",
      "some-date-error",
    );
  });

  it("single modeでariaDescribedBy未指定ならaria-describedbyは付かない", () => {
    render(<DatePicker value="" onChange={() => {}} />);

    expect(screen.getByPlaceholderText("日付を選択…")).not.toHaveAttribute("aria-describedby");
  });
});

describe("DatePicker native validity と ARIA エラーの分離", () => {
  it("ariaInvalid=true だけではネイティブ validity を無効化しない", () => {
    render(<DatePicker value="2026-10-13" onChange={vi.fn()} ariaInvalid />);

    const input = screen.getByRole<HTMLInputElement>("textbox");
    expect(input).toHaveAttribute("aria-invalid", "true");
    expect(input.validity.valid).toBe(true);
    expect(input.validationMessage).toBe("");
  });

  it("ariaInvalid=true で空値でもネイティブ validity は有効のまま", () => {
    render(<DatePicker value="" onChange={vi.fn()} ariaInvalid />);

    expect(screen.getByRole<HTMLInputElement>("textbox").validity.valid).toBe(true);
  });

  it("非空の validationMessage は値が正当でもネイティブ validity を無効化する", () => {
    render(
      <DatePicker
        value="2026-10-13"
        onChange={vi.fn()}
        validationMessage="今日より前の日付は設定できません"
      />,
    );

    const input = screen.getByRole<HTMLInputElement>("textbox");
    expect(input.validity.valid).toBe(false);
    expect(input.validationMessage).toBe("今日より前の日付は設定できません");
  });

  it("validationMessage が消えると custom validity が解除される", () => {
    const { rerender } = render(
      <DatePicker value="2026-10-13" onChange={vi.fn()} validationMessage="エラー" />,
    );
    const input = screen.getByRole<HTMLInputElement>("textbox");
    expect(input.validity.valid).toBe(false);

    rerender(<DatePicker value="2026-10-13" onChange={vi.fn()} />);
    expect(input.validity.valid).toBe(true);
    expect(input.validationMessage).toBe("");
  });

  it("disabled 化すると残っていた custom validity を解除する", () => {
    const { rerender } = render(
      <DatePicker
        value="2026-10-13"
        onChange={vi.fn()}
        disabledDays={{ after: new Date(2026, 9, 1) }}
      />,
    );
    const input = screen.getByRole<HTMLInputElement>("textbox");
    expect(input.validationMessage).toBe("入力した日付は選択できません");

    rerender(
      <DatePicker
        value="2026-10-13"
        onChange={vi.fn()}
        disabled
        disabledDays={{ after: new Date(2026, 9, 1) }}
      />,
    );
    expect(input.validationMessage).toBe("");
  });

  it("ariaRequired を内部 input の aria-required に転送する", () => {
    const { rerender } = render(<DatePicker value="" onChange={vi.fn()} ariaRequired />);

    expect(screen.getByRole("textbox")).toHaveAttribute("aria-required", "true");

    rerender(<DatePicker value="" onChange={vi.fn()} />);
    expect(screen.getByRole("textbox")).not.toHaveAttribute("aria-required");
  });

  it("showClearButton={false} では値があっても内蔵クリアボタンを出さない", () => {
    const { rerender } = render(
      <DatePicker value="2026-10-13" onChange={vi.fn()} showClearButton={false} />,
    );

    expect(screen.queryByRole("button", { name: "日付をクリア" })).not.toBeInTheDocument();

    rerender(<DatePicker value="2026-10-13" onChange={vi.fn()} />);
    expect(screen.getByRole("button", { name: "日付をクリア" })).toBeInTheDocument();
  });
});

describe("DatePicker single keyboard commit session dedup", () => {
  it("変更した値を Enter で確定してから blur しても onChange は1回だけ", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<DatePicker value="2026-10-13" onChange={onChange} ariaLabel="日付" />);

    const input = screen.getByRole("textbox", { name: "日付" });
    await user.click(input);
    await user.clear(input);
    await user.type(input, "20261014");
    await user.keyboard("{Enter}");
    await user.tab();
    expect(onChange).toHaveBeenCalledTimes(1);
    expect(onChange).toHaveBeenCalledWith("2026-10-14");
  });

  it("同じ編集で Enter を繰り返しても重複 emit しない", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<DatePicker value="2026-10-13" onChange={onChange} ariaLabel="日付" />);

    const input = screen.getByRole("textbox", { name: "日付" });
    await user.click(input);
    await user.clear(input);
    await user.type(input, "20261014");
    await user.keyboard("{Enter}");
    await user.keyboard("{Enter}");
    await user.tab();
    expect(onChange).toHaveBeenCalledTimes(1);
    expect(onChange).toHaveBeenCalledWith("2026-10-14");
  });

  it("Enter 後にさらに別の日へ編集すると2回目を emit する", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<DatePicker value="2026-10-13" onChange={onChange} ariaLabel="日付" />);

    const input = screen.getByRole("textbox", { name: "日付" });
    await user.click(input);
    await user.clear(input);
    await user.type(input, "20261014");
    await user.keyboard("{Enter}");
    await user.clear(input);
    await user.type(input, "20261015");
    await user.tab();
    expect(onChange).toHaveBeenCalledTimes(2);
    expect(onChange).toHaveBeenLastCalledWith("2026-10-15");
  });

  it("prop が更新されなくても再フォーカス・再編集で同じ値を emit できる", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<DatePicker value="2026-10-13" onChange={onChange} ariaLabel="日付" />);

    const input = screen.getByRole("textbox", { name: "日付" });
    await user.click(input);
    await user.clear(input);
    await user.type(input, "20261014");
    await user.keyboard("{Enter}");
    await user.tab();
    expect(onChange).toHaveBeenCalledTimes(1);

    await user.click(input);
    await user.clear(input);
    await user.type(input, "20261014");
    await user.keyboard("{Enter}");
    await user.tab();
    expect(onChange).toHaveBeenCalledTimes(2);
    expect(onChange).toHaveBeenLastCalledWith("2026-10-14");
  });
});

/**
 * Issue #48 回帰テスト
 *
 * 飼主・ペット登録フォームで使用される日付選択 UI は、月によって
 * グリッドの行数（4〜6 週）が変わると Popover 全体の高さが変動し、
 * 上部の前月/翌月ナビゲーション (◀ ▶) の位置が見かけ上ずれてしまう。
 *
 * `react-day-picker` の `fixedWeeks` を有効化することで、
 * 常に 6 週固定（42 セル）でレンダリングされることを保証する。
 *
 * NOTE: `role="gridcell"` のセル数を検証することで、
 * fixedWeeks プロパティが実際に反映されているかを確認する。
 */
describe("DatePicker — fixedWeeks layout (Issue #48)", () => {
  it.each([
    ["2026-02-15", "うるう年でない 2 月（28 日 / 通常 4 週）"],
    ["2026-04-15", "30 日（通常 5 週）"],
    ["2026-08-15", "31 日 + 月初土曜（通常 6 週）"],
    ["2027-01-15", "31 日 + 月初金曜（通常 5 週）"],
  ])(
    "value=%s (%s) を開いたとき、常に 42 個の日付セル（6 週 × 7 曜日）を表示する",
    async (value) => {
      const user = userEvent.setup();
      render(<DatePicker value={value} onChange={() => {}} />);

      await user.click(screen.getByLabelText("カレンダーを開く"));

      const cells = screen.getAllByRole("gridcell");
      expect(cells).toHaveLength(42);
    },
  );

  it("前月/翌月ナビゲーションで月を切り替えても 42 セルを維持する", async () => {
    const user = userEvent.setup();
    render(<DatePicker value="2026-04-15" onChange={() => {}} />);

    await user.click(screen.getByLabelText("カレンダーを開く"));
    expect(screen.getAllByRole("gridcell")).toHaveLength(42);

    await user.click(screen.getByLabelText("前の月"));
    expect(screen.getAllByRole("gridcell")).toHaveLength(42);

    await user.click(screen.getByLabelText("次の月"));
    await user.click(screen.getByLabelText("次の月"));
    expect(screen.getAllByRole("gridcell")).toHaveLength(42);
  });

  it("range mode でも 2 ヶ月分とも fixedWeeks が適用される（合計 84 セル）", async () => {
    const user = userEvent.setup();
    render(<DatePicker mode="range" value="" onChange={() => {}} />);

    await user.click(screen.getByRole("button", { name: /期間を選択/ }));

    const cells = screen.getAllByRole("gridcell");
    expect(cells).toHaveLength(84);
  });
});

describe("DatePicker — range mode の開始/終了切替", () => {
  it("開くと終了日が編集対象になり、終了日だけを再選択して閉じる", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<DatePicker mode="range" value="2026-10-01~2026-10-05" onChange={onChange} />);

    await user.click(screen.getByRole("button", { name: /2026\/10\/1/ }));

    const toButton = screen.getByRole("button", { name: "終了日 2026/10/5" });
    expect(toButton).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "開始日 2026/10/1" })).toHaveAttribute(
      "aria-pressed",
      "false",
    );

    await user.click(dayButton("2026-10-10"));

    expect(onChange).toHaveBeenLastCalledWith("2026-10-01~2026-10-10");
    await waitFor(() => expect(screen.queryByRole("grid")).toBeNull());
  });

  it("開始日ボタンで開始日だけを更新し、逆転しない限り終了日を保持する", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<DatePicker mode="range" value="2026-10-01~2026-10-05" onChange={onChange} />);

    await user.click(screen.getByRole("button", { name: /2026\/10\/1/ }));
    await user.click(screen.getByRole("button", { name: "開始日 2026/10/1" }));
    await user.click(dayButton("2026-10-03"));

    expect(onChange).toHaveBeenLastCalledWith("2026-10-03~2026-10-05");
    await waitFor(() => expect(screen.queryByRole("grid")).toBeNull());
  });

  it("終了日に開始日より前を選ぶと端点をスワップして確定する", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<DatePicker mode="range" value="2026-10-01~2026-10-05" onChange={onChange} />);

    await user.click(screen.getByRole("button", { name: /2026\/10\/1/ }));
    await user.click(screen.getByRole("button", { name: "終了日 2026/10/5" }));
    await user.click(screen.getByRole("button", { name: "前の月" }));
    // 2ヶ月表示のため隣月のオーバーフロー日は重複表示される。9/15 は9月グリッドのみに現れる
    await user.click(dayButton("2026-09-15"));

    expect(onChange).toHaveBeenLastCalledWith("2026-09-15~2026-10-01");
    await waitFor(() => expect(screen.queryByRole("grid")).toBeNull());
  });

  it("同日を終了日に選ぶと単日レンジとして確定する", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<DatePicker mode="range" value="2026-10-20~" onChange={onChange} />);

    await user.click(screen.getByRole("button", { name: /2026\/10\/20 〜/ }));
    await user.click(dayButton("2026-10-20"));

    expect(onChange).toHaveBeenLastCalledWith("2026-10-20~2026-10-20");
    await waitFor(() => expect(screen.queryByRole("grid")).toBeNull());
  });

  it("終了日未設定の部分値で開始日を選び直すと部分値をemitして開いたまま", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<DatePicker mode="range" value="2026-10-20~" onChange={onChange} />);

    await user.click(screen.getByRole("button", { name: /2026\/10\/20 〜/ }));
    expect(screen.getByRole("button", { name: "終了日" })).toHaveAttribute("aria-pressed", "true");

    await user.click(screen.getByRole("button", { name: "開始日 2026/10/20" }));
    await user.click(dayButton("2026-10-18"));

    expect(onChange).toHaveBeenLastCalledWith("2026-10-18~");
    // 2ヶ月表示のため grid は2個
    expect(screen.getAllByRole("grid")).toHaveLength(2);
  });
});

// EMR-411 C-1c: カレンダーグリフのコントラストを C.text60 へ引き上げる。
// （range picker は対象外。JSDOM では描画色を計測できないため DOM クラスで検証）
describe("DatePickerSingle calendar glyph contrast (EMR-411 C-1c)", () => {
  it("カレンダーアイコンが C.text60 を持つ", () => {
    render(<DatePicker value="" onChange={vi.fn()} />);

    const trigger = screen.getByRole("button", { name: "カレンダーを開く" });
    const glyph = trigger.querySelector("svg");
    expect(glyph).not.toBeNull();
    expect(glyph).toHaveClass(C.text60);
  });
});
