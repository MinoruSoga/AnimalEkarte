import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { NextVisitDateField } from "./NextVisitDateField";

const { todayJSTISO } = vi.hoisted(() => ({
  todayJSTISO: vi.fn(() => "2026-09-03"),
}));

vi.mock("@/lib/jst-date", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/jst-date")>();
  return {
    ...actual,
    todayJSTISO,
    isPastJSTDate: (date: string) =>
      /^\d{4}-\d{2}-\d{2}$/.test(date) && date < todayJSTISO(),
  };
});

const PAST_ERROR = "今日より前の日付は設定できません";
const TWO_YEAR_ERROR = "今日から2年以内の日付を設定してください";

function renderField(value: string) {
  const onChange = vi.fn();
  const onValidationChange = vi.fn();
  const view = render(
    <NextVisitDateField
      value={value}
      onChange={onChange}
      onValidationChange={onValidationChange}
    />,
  );
  return { ...view, onChange, onValidationChange };
}

describe("NextVisitDateField", () => {
  beforeEach(() => {
    todayJSTISO.mockReturnValue("2026-09-03");
  });

  it("空欄ではエラーを出さない", () => {
    renderField("");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
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

  it("日付入力の変更で過去日なら onValidationChange(false) を呼ぶ", () => {
    const { onValidationChange } = renderField("2026-09-03");
    fireEvent.change(screen.getByLabelText("次回来院推奨日"), {
      target: { value: "2026-09-02" },
    });
    expect(onValidationChange).toHaveBeenCalledWith(false);
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
