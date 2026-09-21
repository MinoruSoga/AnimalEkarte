import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { TreatmentTable } from "./TreatmentTable";

describe("TreatmentTable", () => {
  it("prefers onOpenSearch over onAddRow for the visible add button", async () => {
    const user = userEvent.setup();
    const onOpenSearch = vi.fn();
    const onAddRow = vi.fn();

    render(
      <TreatmentTable
        items={[]}
        onUpdate={vi.fn()}
        onRemove={vi.fn()}
        onOpenSearch={onOpenSearch}
        onAddRow={onAddRow}
      />,
    );

    await user.click(screen.getByRole("button", { name: "行を追加（検索）" }));

    expect(onOpenSearch).toHaveBeenCalledTimes(1);
    expect(onAddRow).not.toHaveBeenCalled();
  });

  it("falls back to onAddRow when onOpenSearch is absent", async () => {
    const user = userEvent.setup();
    const onAddRow = vi.fn();

    render(<TreatmentTable items={[]} onUpdate={vi.fn()} onRemove={vi.fn()} onAddRow={onAddRow} />);

    await user.click(screen.getByRole("button", { name: "行を追加" }));

    expect(onAddRow).toHaveBeenCalledTimes(1);
  });
});
