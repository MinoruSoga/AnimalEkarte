import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AnimalSpeciesSettings } from "./AnimalSpeciesSettings";

const mocks = vi.hoisted(() => ({
  queryResult: {
    data: [] as Array<{ id: string; name: string; isActive: boolean }>,
    isPending: false,
    isError: false,
    error: null as Error | null,
  },
  handleNew: vi.fn(),
  crudData: [] as unknown[],
  crudCanDelete: true,
  saveCanCreate: true,
  saveCanEdit: true,
  reorderCallback: (_ids: string[]) => {},
  reorderMutation: vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canCreate: true, canEdit: true, canDelete: true }),
}));

vi.mock("@/hooks/use-side-peek-dirty", () => ({
  useSidePeekDirty: () => ({
    markDirty: vi.fn(),
    markClean: vi.fn(),
    confirmDiscard: vi.fn(() => true),
  }),
}));

vi.mock("@/hooks/use-sortable-list", () => ({
  useSortableList: ({
    items,
    onReorder,
  }: {
    items: unknown[];
    onReorder: (ids: string[]) => void;
  }) => {
    mocks.reorderCallback = onReorder;
    return {
      orderedItems: items,
      sensors: [],
      handleDragEnd: vi.fn(),
    };
  },
}));

vi.mock("../hooks/use-master-crud", () => ({
  useMasterCRUD: ({
    data,
    permissions,
  }: {
    data?: unknown[];
    permissions: { canDelete: boolean };
  }) => {
    mocks.crudData = data ?? [];
    mocks.crudCanDelete = permissions.canDelete;
    return {
      filteredItems: data ?? [],
      pendingDelete: null,
      handleEdit: vi.fn(),
      handleNew: mocks.handleNew,
    };
  },
}));

vi.mock("../hooks/use-master-save", () => ({
  useMasterSave: ({
    permissions,
  }: {
    permissions: { canCreate: boolean; canEdit: boolean };
  }) => {
    mocks.saveCanCreate = permissions.canCreate;
    mocks.saveCanEdit = permissions.canEdit;
    return { handleSave: vi.fn() };
  },
}));

vi.mock("../components/MasterCRUDPage", () => ({
  MasterCRUDPage: ({
    children,
    crud,
  }: {
    children?: ReactNode;
    crud: { handleNew: () => void };
  }) => (
    <>
      <button type="button" onClick={crud.handleNew}>
        動物種類を追加
      </button>
      {children}
    </>
  ),
}));

vi.mock("../components/AnimalSpeciesSidePanel", () => ({
  AnimalSpeciesSidePanel: () => null,
}));

vi.mock("../components/AnimalSpeciesSortableTable", () => ({
  ANIMAL_SPECIES_COLUMNS: [],
  AnimalSpeciesSortableTable: () => (
    <div role="table" aria-label="動物種類一覧" />
  ),
}));

function mutationStub() {
  return { mutate: vi.fn(), mutateAsync: vi.fn() };
}

vi.mock("../api/animal-species", () => ({
  useGetAnimalSpecies: () => mocks.queryResult,
  useCreateAnimalSpecies: mutationStub,
  useUpdateAnimalSpecies: mutationStub,
  useDeleteAnimalSpecies: mutationStub,
  useReorderAnimalSpecies: () => ({
    mutate: mocks.reorderMutation,
    mutateAsync: vi.fn(),
  }),
}));

function setQueryResult({
  data = [],
  isPending = false,
  isError = false,
  error = null,
}: Partial<typeof mocks.queryResult>) {
  mocks.queryResult.data = data;
  mocks.queryResult.isPending = isPending;
  mocks.queryResult.isError = isError;
  mocks.queryResult.error = error;
}

describe("AnimalSpeciesSettings", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setQueryResult({});
  });

  it("取得失敗を読み込み中やデータより優先して表示し、生のエラー詳細を隠す", () => {
    setQueryResult({
      data: [{ id: "1", name: "犬", isActive: true }],
      isPending: true,
      isError: true,
      error: new Error("GET /v1/masters/animal-species 500"),
    });

    render(<AnimalSpeciesSettings />);

    const alert = screen.getByRole("alert");
    expect(alert).toHaveTextContent("動物種の取得に失敗しました。");
    expect(alert).toHaveAttribute("aria-atomic", "true");
    expect(screen.queryByText("動物種を読み込み中です。")).not.toBeInTheDocument();
    expect(
      screen.queryByText("動物種マスタが登録されていません。"),
    ).not.toBeInTheDocument();
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    expect(
      screen.queryByText("GET /v1/masters/animal-species 500"),
    ).not.toBeInTheDocument();
    expect(mocks.crudData).toEqual([]);
    expect(mocks.crudCanDelete).toBe(false);
    expect(mocks.saveCanCreate).toBe(true);
    expect(mocks.saveCanEdit).toBe(false);
    mocks.reorderCallback(["1"]);
    expect(mocks.reorderMutation).not.toHaveBeenCalled();
  });

  it("読み込み中をデータより優先してaccessibleなstatusで表示する", () => {
    setQueryResult({
      data: [{ id: "1", name: "犬", isActive: true }],
      isPending: true,
    });

    render(<AnimalSpeciesSettings />);

    const status = screen.getByRole("status");
    expect(status).toHaveTextContent("動物種を読み込み中です。");
    expect(status).toHaveAttribute("aria-live", "polite");
    expect(status).toHaveAttribute("aria-atomic", "true");
    expect(
      screen.queryByText("動物種マスタが登録されていません。"),
    ).not.toBeInTheDocument();
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    expect(mocks.crudData).toEqual([]);
    expect(mocks.crudCanDelete).toBe(false);
    expect(mocks.saveCanCreate).toBe(true);
    expect(mocks.saveCanEdit).toBe(false);
    mocks.reorderCallback(["1"]);
    expect(mocks.reorderMutation).not.toHaveBeenCalled();
  });

  it("取得成功かつ0件をdistinctなaccessible statusで表示する", () => {
    render(<AnimalSpeciesSettings />);

    const status = screen.getByRole("status");
    expect(status).toHaveTextContent(
      "動物種マスタが登録されていません。",
    );
    expect(status).toHaveAttribute("aria-live", "polite");
    expect(status).toHaveAttribute("aria-atomic", "true");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
  });

  it("取得成功かつデータありでは状態表示を消して一覧を表示する", () => {
    setQueryResult({
      data: [{ id: "1", name: "犬", isActive: true }],
    });

    render(<AnimalSpeciesSettings />);

    expect(
      screen.getByRole("table", { name: "動物種類一覧" }),
    ).toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(mocks.crudData).toEqual([
      { id: "1", name: "犬", isActive: true },
    ]);
    expect(mocks.crudCanDelete).toBe(true);
    expect(mocks.saveCanCreate).toBe(true);
    expect(mocks.saveCanEdit).toBe(true);
  });

  it.each([
    ["取得失敗", { isError: true, error: new Error("raw error") }],
    ["読み込み中", { isPending: true }],
    ["取得成功かつ0件", {}],
  ])("%sでも動物種類の追加操作を使える", async (_state, queryResult) => {
    const user = userEvent.setup();
    setQueryResult(queryResult);

    render(<AnimalSpeciesSettings />);
    await user.click(screen.getByRole("button", { name: "動物種類を追加" }));

    expect(mocks.handleNew).toHaveBeenCalledOnce();
  });
});
