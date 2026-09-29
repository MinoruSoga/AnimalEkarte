import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { DiagnosisNameTab, DiagnosisTypeTab } from "./DiagnosisTabs";

const mocks = vi.hoisted(() => ({
  reorderCallbacks: [] as Array<(ids: string[]) => void>,
  reorderNames: vi.fn(),
  reorderTypes: vi.fn(),
  typeFixture: {
    id: "1",
    clinicId: 1,
    name: "テストカテゴリ",
    isActive: true,
    description: "",
    sortOrder: 1,
    createdAt: "2026-09-28T00:00:00Z",
    updatedAt: "2026-09-28T00:00:00Z",
  },
  nameFixture: {
    id: "7",
    clinicId: 1,
    name: "テスト病名",
    isActive: true,
    description: "",
    diagnosisTypeId: "1",
    sortOrder: 1,
    createdAt: "2026-09-28T00:00:00Z",
    updatedAt: "2026-09-28T00:00:00Z",
  },
}));

vi.mock("@dnd-kit/sortable", () => ({
  useSortable: () => ({
    attributes: {},
    listeners: {},
    setNodeRef: vi.fn(),
    setActivatorNodeRef: vi.fn(),
    transform: null,
    transition: undefined,
    isDragging: false,
  }),
}));

vi.mock("@/hooks/use-sortable-list", () => ({
  useSortableList: ({
    items,
    onReorder,
  }: {
    items: Array<{ id: string }>;
    onReorder: (ids: string[]) => void;
  }) => {
    mocks.reorderCallbacks.push(onReorder);
    return {
      orderedItems: items,
      sensors: [],
      handleDragEnd: vi.fn(),
    };
  },
}));

vi.mock("@/components/shared/PropertyFilter/PropertyFilter", () => ({
  PropertyFilter: () => null,
}));

vi.mock("./DiagnosisSortableTable", () => ({
  DiagnosisSortableTable: ({
    items,
    renderRow,
  }: {
    items: Array<{ id: string }>;
    renderRow: (item: { id: string }) => ReactNode;
  }) => (
    <table>
      <tbody>{items.map((item) => renderRow(item))}</tbody>
    </table>
  ),
}));

vi.mock("../api/diagnosis", () => ({
  useGetDiagnosisNames: () => ({ data: [mocks.nameFixture] }),
  useGetDiagnosisTypes: () => ({ data: [mocks.typeFixture] }),
  useReorderDiagnosisNames: () => ({ mutate: mocks.reorderNames }),
  useReorderDiagnosisTypes: () => ({ mutate: mocks.reorderTypes }),
}));

function callLatestReorder() {
  const callback = mocks.reorderCallbacks.at(-1);
  expect(callback).toBeDefined();
  callback?.(["2", "1"]);
}

describe("diagnosis reorder permission guards", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.reorderCallbacks.length = 0;
  });

  it("blocks diagnosis type reorder mutation when edit permission is absent", () => {
    render(<DiagnosisTypeTab canEdit={false} onEditTargetChange={vi.fn()} />);
    callLatestReorder();
    expect(mocks.reorderTypes).not.toHaveBeenCalled();
  });

  it("blocks diagnosis name reorder mutation when edit permission is absent", () => {
    render(<DiagnosisNameTab canEdit={false} onEditTargetChange={vi.fn()} />);
    callLatestReorder();
    expect(mocks.reorderNames).not.toHaveBeenCalled();
  });
});

describe("detail affordance by permission", () => {
  it("view のみでも診断カテゴリ行の詳細ボタンを表示し、クリックで詳細を開く", async () => {
    const onEditTargetChange = vi.fn();
    render(<DiagnosisTypeTab canEdit={false} onEditTargetChange={onEditTargetChange} />);

    await userEvent.click(
      screen.getByRole("button", { name: "詳細: 診断カテゴリ テストカテゴリ (ID 1)" }),
    );
    expect(onEditTargetChange).toHaveBeenCalledWith(
      expect.objectContaining({ id: "1", name: "テストカテゴリ" }),
    );
    expect(
      screen.queryByRole("button", { name: "編集: 診断カテゴリ テストカテゴリ (ID 1)" }),
    ).toBeNull();
    expect(
      screen.getByRole("button", { name: "並べ替え: 診断カテゴリ テストカテゴリ (ID 1)" }),
    ).toBeDisabled();
  });

  it("view のみでも診断病名行の詳細ボタンを表示し、クリックで詳細を開く", async () => {
    const onEditTargetChange = vi.fn();
    render(<DiagnosisNameTab canEdit={false} onEditTargetChange={onEditTargetChange} />);

    await userEvent.click(screen.getByRole("button", { name: "詳細: 診断病名 テスト病名 (ID 7)" }));
    expect(onEditTargetChange).toHaveBeenCalledWith(
      expect.objectContaining({ id: "7", name: "テスト病名" }),
    );
    expect(screen.queryByRole("button", { name: "編集: 診断病名 テスト病名 (ID 7)" })).toBeNull();
    expect(
      screen.getByRole("button", { name: "並べ替え: 診断病名 テスト病名 (ID 7)" }),
    ).toBeDisabled();
  });

  it("編集権限ありでは詳細・編集ボタンと有効な並べ替えハンドルを従来どおり描画する", () => {
    render(<DiagnosisTypeTab canEdit onEditTargetChange={vi.fn()} />);

    expect(
      screen.getByRole("button", { name: "詳細: 診断カテゴリ テストカテゴリ (ID 1)" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "編集: 診断カテゴリ テストカテゴリ (ID 1)" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "並べ替え: 診断カテゴリ テストカテゴリ (ID 1)" }),
    ).toBeEnabled();
  });
});
