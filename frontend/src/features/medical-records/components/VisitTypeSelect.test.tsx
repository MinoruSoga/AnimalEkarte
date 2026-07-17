import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { VisitTypeSelect } from "./VisitTypeSelect";

describe("VisitTypeSelect", () => {
  it("SelectTrigger に現在の value が表示される", () => {
    render(
      <VisitTypeSelect value="初診" onChange={vi.fn()} />,
    );
    expect(screen.getByRole("combobox")).toBeInTheDocument();
    expect(screen.getByText("初診")).toBeInTheDocument();
  });

  it("初診/再診/緊急/往診 の選択肢が含まれる", async () => {
    const user = userEvent.setup();
    render(
      <VisitTypeSelect value="再診" onChange={vi.fn()} />,
    );
    await user.click(screen.getByRole("combobox"));
    // Radix Select がオプションをレンダリングするのを確認
    // "再診" は SelectTrigger の表示値とドロップダウンの両方に存在するため getAllByText で確認
    expect(screen.getByText("初診")).toBeInTheDocument();
    expect(screen.getAllByText("再診").length).toBeGreaterThan(0);
    expect(screen.getByText("緊急")).toBeInTheDocument();
    expect(screen.getByText("往診")).toBeInTheDocument();
  });

  it("選択時 onChange が呼ばれる", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();
    render(<VisitTypeSelect value="再診" onChange={onChange} />);
    await user.click(screen.getByRole("combobox"));
    await user.click(screen.getByText("初診"));
    expect(onChange).toHaveBeenCalledWith("初診");
  });

  it("disabled 時に SelectTrigger が disabled になる", () => {
    render(
      <VisitTypeSelect value="再診" onChange={vi.fn()} disabled />,
    );
    expect(screen.getByRole("combobox")).toBeDisabled();
  });
});
