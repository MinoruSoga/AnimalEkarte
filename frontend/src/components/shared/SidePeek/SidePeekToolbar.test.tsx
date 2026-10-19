import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { C } from "@/lib/design-tokens";
import { SidePeekToolbar } from "./SidePeekToolbar";

describe("SidePeekToolbar", () => {
  it("閉じる操作の色を読みやすくし、削除操作には赤を維持する", () => {
    render(<SidePeekToolbar isNew={false} onClose={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByRole("button", { name: "閉じる" })).toHaveClass(C.text60);
    expect(screen.getByRole("button", { name: "削除" })).toHaveClass(C.danger);
  });
  it.each([
    { isNew: true, readOnly: false, label: "新規登録" },
    { isNew: false, readOnly: false, label: "編集" },
    { isNew: false, readOnly: true, label: "詳細" },
  ])(
    "shows $label with readable caption text and accessible controls",
    ({ isNew, readOnly, label }) => {
      render(
        <SidePeekToolbar isNew={isNew} readOnly={readOnly} onClose={vi.fn()} onDelete={vi.fn()} />,
      );

      expect(screen.getByText(label)).toHaveClass("text-xs", C.text65);
      expect(screen.getByRole("button", { name: "削除" })).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "閉じる" })).toBeInTheDocument();
    },
  );
});
