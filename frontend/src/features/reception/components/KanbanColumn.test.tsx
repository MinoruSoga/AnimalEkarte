import { describe, expect, it, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";
import { C } from "@/lib/design-tokens";
import type { ColumnData } from "@/types";
import { KanbanColumn } from "./KanbanColumn";

const columnTitles = ["受付予約", "受付済", "診察中", "会計待ち", "会計済"];

describe("KanbanColumn visual text contracts", () => {
  it("all five empty columns keep their semantic labels and use readable heading, count, and add text", () => {
    const onAddClick = vi.fn();
    const columns: ColumnData[] = columnTitles.map((title) => ({ title, appointments: [] }));
    const onCardClick = vi.fn();

    render(
      <div>
        {columns.map((data) => (
          <KanbanColumn key={data.title} data={data} onAddClick={onAddClick} onCardClick={onCardClick} />
        ))}
      </div>,
    );

    for (const title of columnTitles) {
      const region = screen.getByRole("region", { name: `${title} — 0件` });
      const heading = within(region).getByRole("heading", { name: title });
      expect(heading).toHaveClass(C.text);
      expect(within(heading.parentElement!).getByText("0")).toHaveClass(C.text65);

      const addButton = within(region).getByRole("button", { name: "新規追加" });
      expect(addButton).toHaveClass(C.text65, C.hoverText, "min-h-11", "min-w-11");
    }
  });
});
