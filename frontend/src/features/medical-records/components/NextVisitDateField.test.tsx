import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { formatJSTWallDate } from "@/lib/jst-date";
import { NextVisitDateField } from "./NextVisitDateField";

const { todayJSTISO } = vi.hoisted(() => ({
  todayJSTISO: vi.fn(() => "2026-09-03"),
}));

vi.mock("@/lib/jst-date", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/jst-date")>();
  return {
    ...actual,
    todayJSTISO,
    isPastJSTDate: (date: string) => /^\d{4}-\d{2}-\d{2}$/.test(date) && date < todayJSTISO(),
  };
});

const PAST_ERROR = "今日より前の日付は設定できません";
const TWO_YEAR_ERROR = "今日から2年以内の日付を設定してください";

function renderField(value: string) {
  const onChange = vi.fn();
  const view = render(<NextVisitDateField value={value} onChange={onChange} />);
  return { ...view, onChange };
}

describe("NextVisitDateField", () => {
  beforeEach(() => {
    todayJSTISO.mockReturnValue("2026-09-03");
  });

  it("空欄ではエラーを出さない", () => {
    renderField("");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("日付設定不可・解除可では入力と候補を無効化し解除だけを許可する", () => {
    const onChange = vi.fn();
    render(
      <NextVisitDateField
        value="2026-09-04"
        onChange={onChange}
        setDisabled
        clearDisabled={false}
      />,
    );
    expect(screen.getByLabelText("次回来院推奨日")).toBeDisabled();
    expect(screen.getByRole("button", { name: "1ヶ月後" })).toBeDisabled();
    fireEvent.click(screen.getByRole("button", { name: "クリア" }));
    expect(onChange).toHaveBeenCalledWith("");
  });

  it("設定不可では同じ DatePicker を無効化し、値とエラー関連付けを保つ", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<NextVisitDateField value="2026-09-02" onChange={onChange} setDisabled clearDisabled />);

    const input = screen.getByRole<HTMLInputElement>("textbox", { name: "次回来院推奨日" });
    const error = screen.getByRole("alert");
    expect(input).toBeDisabled();
    expect(input).toHaveAttribute("aria-invalid", "true");
    expect(input).toHaveAttribute("aria-describedby", error.id);
    expect(error).toHaveTextContent(PAST_ERROR);
    expect(input.value).toMatch(/2026年9月2日/);
    expect(screen.getByRole("button", { name: "次回来院推奨日を選択" })).toBeDisabled();
    expect(screen.queryByRole("button", { name: "日付をクリア" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Today" })).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "次回来院推奨日を選択" }));
    expect(screen.queryByRole("button", { name: "Today" })).not.toBeInTheDocument();
    expect(onChange).not.toHaveBeenCalled();
  });

  it("日付設定可・解除不可では解除ボタンを無効化する", () => {
    const onChange = vi.fn();
    render(
      <NextVisitDateField
        value="2026-09-04"
        onChange={onChange}
        setDisabled={false}
        clearDisabled
      />,
    );
    expect(screen.getByLabelText("次回来院推奨日")).toBeEnabled();
    expect(screen.getByRole("button", { name: "1ヶ月後" })).toBeEnabled();
    fireEvent.click(screen.getByRole("button", { name: "クリア" }));
    expect(onChange).not.toHaveBeenCalled();
  });

  it("フォーカス→編集なしで blur しても onChange を呼ばない（重複 PUT 抑止）", async () => {
    const user = userEvent.setup();
    const { onChange } = renderField("2026-09-04");
    const dateInput = screen.getByRole("textbox", { name: "次回来院推奨日" });
    await user.click(dateInput);
    await user.tab();
    expect(onChange).not.toHaveBeenCalled();
  });

  it("有効な日付へ編集して blur すると onChange を1回だけ ISO 値で呼ぶ", async () => {
    const user = userEvent.setup();
    const { onChange } = renderField("2026-09-04");
    const dateInput = screen.getByRole<HTMLInputElement>("textbox", { name: "次回来院推奨日" });
    await user.clear(dateInput);
    await user.type(dateInput, "2026/09/10");
    await user.tab();
    expect(onChange).toHaveBeenCalledTimes(1);
    expect(onChange).toHaveBeenCalledWith("2026-09-10");
  });

  it("未編集で Enter→blur しても重複した onChange を呼ばない", async () => {
    const user = userEvent.setup();
    const { onChange } = renderField("2026-09-04");
    const dateInput = screen.getByRole("textbox", { name: "次回来院推奨日" });
    await user.click(dateInput);
    await user.keyboard("{Enter}");
    await user.tab();
    expect(onChange).not.toHaveBeenCalled();
  });

  it("DatePicker 内蔵のクリアボタンは表示しない（専用クリアチップで権限制御する）", () => {
    renderField("2026-09-04");
    expect(screen.queryByRole("button", { name: "日付をクリア" })).not.toBeInTheDocument();
  });

  it("クリア可では専用クリアボタンが有効で onChange('') を1回だけ呼ぶ", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<NextVisitDateField value="2026-09-04" onChange={onChange} clearDisabled={false} />);
    const clearChip = screen.getByRole("button", { name: "クリア" });
    expect(clearChip).toBeEnabled();
    await user.click(clearChip);
    expect(onChange).toHaveBeenCalledTimes(1);
    expect(onChange).toHaveBeenCalledWith("");
  });

  it("クイック選択が現在値と同じ日付でも重複した onChange を呼ばない", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const oneYearLater = new Date();
    oneYearLater.setFullYear(oneYearLater.getFullYear() + 1);
    render(<NextVisitDateField value={formatJSTWallDate(oneYearLater)} onChange={onChange} />);
    await user.click(screen.getByRole("button", { name: "1年後" }));
    expect(onChange).not.toHaveBeenCalled();
  });

  it("過去日では過去日エラーを表示する", () => {
    renderField("2026-09-02");
    expect(screen.getByRole("alert")).toHaveTextContent(PAST_ERROR);
  });

  it("当日はエラーにしない", () => {
    renderField("2026-09-03");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("2年以内の未来日はエラーにしない", () => {
    renderField("2026-09-04");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("ちょうど2年後はエラーにしない", () => {
    renderField("2028-09-03");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("2年を超える日付では上限エラーを表示する", () => {
    renderField("2028-09-04");
    expect(screen.getByRole("alert")).toHaveTextContent(TWO_YEAR_ERROR);
  });

  it("日付入力の変更で過去日でも onChange(newValue) を呼ぶ（バリデーションは表示側で行う）", async () => {
    const user = userEvent.setup();
    const { onChange } = renderField("2026-09-03");
    const dateInput = screen.getByRole<HTMLInputElement>("textbox", { name: "次回来院推奨日" });
    await user.clear(dateInput);
    await user.type(dateInput, "2026/09/02");
    await user.tab();
    expect(onChange).toHaveBeenCalledWith("2026-09-02");
  });

  it("親の過去日エラーを表示し native validity でフォーム送信を止める", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn((event: React.FormEvent<HTMLFormElement>) => event.preventDefault());
    const onChange = vi.fn();
    const { rerender } = render(
      <form onSubmit={onSubmit}>
        <NextVisitDateField value="2026-09-03" onChange={onChange} />
        <button type="submit">送信</button>
      </form>,
    );
    const dateInput = screen.getByRole<HTMLInputElement>("textbox", { name: "次回来院推奨日" });
    await user.clear(dateInput);
    await user.type(dateInput, "2026/09/02");
    await user.tab();

    expect(onChange).toHaveBeenCalledWith("2026-09-02");
    rerender(
      <form onSubmit={onSubmit}>
        <NextVisitDateField value="2026-09-02" onChange={onChange} />
        <button type="submit">送信</button>
      </form>,
    );
    expect(screen.getByRole("alert")).toHaveTextContent(PAST_ERROR);
    const updatedInput = screen.getByRole<HTMLInputElement>("textbox", { name: "次回来院推奨日" });
    expect(updatedInput.validity.valid).toBe(false);
    await user.click(screen.getByRole("button", { name: "送信" }));
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("2月29日起点の2年上限は非閏年2月28日まで（3月1日は上限エラー）", () => {
    todayJSTISO.mockReturnValue("2024-02-29");
    renderField("2026-03-01");
    expect(screen.getByRole("alert")).toHaveTextContent(TWO_YEAR_ERROR);
  });

  it("2月29日起点のちょうど2年後（非閏年2月28日）はエラーにしない", () => {
    todayJSTISO.mockReturnValue("2024-02-29");
    renderField("2026-02-28");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});
