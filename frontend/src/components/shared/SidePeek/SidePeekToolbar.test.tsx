import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { C } from "@/lib/design-tokens";
import { SidePeekToolbar } from "./SidePeekToolbar";

describe("SidePeekToolbar", () => {
  it.each([
    { isNew: true, readOnly: false, label: "新規作成" },
    { isNew: false, readOnly: false, label: "編集" },
    { isNew: false, readOnly: true, label: "詳細" },
  ])("shows $label with readable caption text and accessible controls", ({ isNew, readOnly, label }) => {
    render(
      <SidePeekToolbar
        isNew={isNew}
        readOnly={readOnly}
        onClose={vi.fn()}
        onDelete={vi.fn()}
      />,
    );

    expect(screen.getByText(label)).toHaveClass("text-xs", C.text65);
    expect(screen.getByRole("button", { name: "削除" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "閉じる" })).toBeInTheDocument();
  });
});
