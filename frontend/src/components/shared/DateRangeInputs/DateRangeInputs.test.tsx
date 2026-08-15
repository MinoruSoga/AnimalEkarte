import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DateRangeInputs } from "./DateRangeInputs";

describe("DateRangeInputs", () => {
  it("from/to の値がそれぞれの入力欄に反映される", () => {
    render(
      <DateRangeInputs
        fromValue="2026-01-01"
        toValue="2026-01-31"
        onFromChange={() => {}}
        onToChange={() => {}}
        fromTestId="from"
        toTestId="to"
      />,
    );
    expect(screen.getByTestId("from")).toHaveValue("2026-01-01");
    expect(screen.getByTestId("to")).toHaveValue("2026-01-31");
  });

  it("区切り文字のデフォルトは 〜", () => {
    render(
      <DateRangeInputs fromValue="" toValue="" onFromChange={() => {}} onToChange={() => {}} />,
    );
    expect(screen.getByText("〜")).toBeInTheDocument();
  });

  it("2つの日付入力に個別の accessible name を付ける", () => {
    render(
      <DateRangeInputs fromValue="" toValue="" onFromChange={() => {}} onToChange={() => {}} />,
    );

    expect(screen.getByLabelText("開始日")).toBeInTheDocument();
    expect(screen.getByLabelText("終了日")).toBeInTheDocument();
  });

  it("separator を差し替えられる", () => {
    render(
      <DateRangeInputs
        fromValue=""
        toValue=""
        onFromChange={() => {}}
        onToChange={() => {}}
        separator="to"
      />,
    );
    expect(screen.getByText("to")).toBeInTheDocument();
  });

  it("from 欄の変更で onFromChange が呼ばれる", async () => {
    const user = userEvent.setup();
    const onFromChange = vi.fn();
    render(
      <DateRangeInputs
        fromValue=""
        toValue=""
        onFromChange={onFromChange}
        onToChange={() => {}}
        fromTestId="from"
      />,
    );

    await user.type(screen.getByTestId("from"), "2026-02-01");

    expect(onFromChange).toHaveBeenCalled();
  });

  it("to 欄の変更で onToChange が呼ばれる", async () => {
    const user = userEvent.setup();
    const onToChange = vi.fn();
    render(
      <DateRangeInputs
        fromValue=""
        toValue=""
        onFromChange={() => {}}
        onToChange={onToChange}
        toTestId="to"
      />,
    );

    await user.type(screen.getByTestId("to"), "2026-02-28");

    expect(onToChange).toHaveBeenCalled();
  });
});
