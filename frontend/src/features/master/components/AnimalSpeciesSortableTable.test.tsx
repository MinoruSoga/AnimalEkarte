import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { AnimalSpeciesSortableTable } from "./AnimalSpeciesSortableTable";
import type { AnimalSpecies } from "../api/animal-species";

function speciesFixture(id: number, name: string): AnimalSpecies {
  return {
    id: String(id),
    name,
    isActive: true,
    sortOrder: 0,
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
  };
}

const ITEMS = [speciesFixture(1, "イヌ"), speciesFixture(2, "ネコ")];

function renderTable({ canEdit }: { canEdit: boolean }) {
  const onEdit = vi.fn();
  render(
    <AnimalSpeciesSortableTable
      items={ITEMS}
      sensors={[]}
      onDragEnd={vi.fn()}
      canEdit={canEdit}
      onEdit={onEdit}
    />,
  );
  return { onEdit };
}

describe("AnimalSpeciesSortableTable", () => {
  it("編集可能な場合は「操作」列ヘッダと編集ボタンを出し、編集で onEdit を呼ぶ", async () => {
    const user = userEvent.setup();
    const { onEdit } = renderTable({ canEdit: true });

    expect(screen.getByRole("columnheader", { name: "操作" })).toBeInTheDocument();
    const editButton = screen.getByRole("button", { name: "編集: 動物種類 イヌ (ID 1)" });
    await user.click(editButton);
    expect(onEdit).toHaveBeenCalledWith(ITEMS[0]);
  });

  it("編集不可の場合は「操作」列ヘッダも編集ボタンも出さない", () => {
    renderTable({ canEdit: false });

    expect(screen.queryByRole("columnheader", { name: "操作" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /^編集: / })).not.toBeInTheDocument();
  });

  it("編集不可の場合も名称ボタンによる詳細アクセスとドラッグハンドル表示は保たれる", async () => {
    const user = userEvent.setup();
    const { onEdit } = renderTable({ canEdit: false });

    const detailButton = screen.getByRole("button", { name: "詳細: 動物種類 イヌ (ID 1)" });
    await user.click(detailButton);
    expect(onEdit).toHaveBeenCalledWith(ITEMS[0]);

    expect(screen.getByRole("button", { name: "並べ替え: 動物種類 イヌ (ID 1)" })).toBeDisabled();
  });
});
