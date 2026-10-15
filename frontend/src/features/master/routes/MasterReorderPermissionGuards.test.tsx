import type { ReactNode } from "react";
import { render } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { AnimalSpeciesSettings } from "./AnimalSpeciesSettings";
import { CageSettings } from "./CageSettings";
import { MerchandiseItemSettings } from "./MerchandiseItemSettings";
import { PermissionGroupSettings } from "./PermissionGroupSettings";
import type { FrontendMerchandiseItem } from "../api/merchandise-items";

const mocks = vi.hoisted(() => ({
  reorderCallbacks: [] as Array<(ids: string[]) => void>,
  animalSpeciesReorder: vi.fn(),
  cageReorder: vi.fn(),
  merchandiseReorder: vi.fn(),
  permissionGroupReorder: vi.fn(),
  canEdit: false,
  merchandisePending: false,
  merchandiseFiltered: false,
  merchandiseItems: [] as FrontendMerchandiseItem[],
  merchandiseMove: undefined as
    ((item: FrontendMerchandiseItem, direction: "up" | "down") => void) | undefined,
  merchandiseIsReordering: undefined as boolean | undefined,
  merchandiseCanEdit: undefined as boolean | undefined,
  merchandiseCanReorder: undefined as boolean | undefined,
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canCreate: false, canEdit: mocks.canEdit, canDelete: false }),
}));

vi.mock("@/hooks/use-side-peek-dirty", () => ({
  useSidePeekDirty: () => ({
    markDirty: vi.fn(),
    markClean: vi.fn(),
    confirmDiscard: vi.fn(() => true),
  }),
}));

// EMR-306: CageSettings が新規に使う連結 mutation hook。
// 本 suite は QueryClientProvider なしで render するため stub 化する。
vi.mock("@/hooks/use-cage-merge", () => ({
  useMergeCages: () => mutationStub(vi.fn()),
  useUnmergeCages: () => mutationStub(vi.fn()),
}));

// EMR-306: CageSettings の占有クエリ hook も同じ理由で stub 化する。
vi.mock("@/hooks/use-cage-occupancy", () => ({
  useCageOccupancy: () => ({ occupiedCageIds: undefined }),
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
      activeId: null,
      handleDragStart: vi.fn(),
      handleDragEnd: vi.fn(),
      handleDragCancel: vi.fn(),
      resetOrder: vi.fn(),
    };
  },
}));

vi.mock("../hooks/use-master-crud", () => ({
  useMasterCRUD: ({ data = [] }: { data?: Array<{ id: string }> }) => ({
    filteredItems: mocks.merchandiseFiltered ? data.filter((item) => item.id !== "2") : data,
    pendingDelete: null,
    handleEdit: vi.fn(),
    handleNew: vi.fn(),
  }),
}));

vi.mock("../hooks/use-master-save", () => ({
  useMasterSave: () => ({ handleSave: vi.fn() }),
}));

vi.mock("../components/MasterCRUDPage", () => ({
  MasterCRUDPage: ({ children }: { children?: ReactNode }) => <>{children}</>,
}));

vi.mock("../components/AnimalSpeciesSidePanel", () => ({
  AnimalSpeciesSidePanel: () => null,
}));

vi.mock("../components/AnimalSpeciesSortableTable", () => ({
  ANIMAL_SPECIES_COLUMNS: [],
  AnimalSpeciesSortableTable: () => null,
}));

vi.mock("../components/CageSidePanel", () => ({ CageSidePanel: () => null }));
vi.mock("../components/CageSortableTable", () => ({ CageSortableTable: () => null }));
vi.mock("../components/MerchandiseSidePanel", () => ({ MerchandiseSidePanel: () => null }));
vi.mock("../components/MerchandiseSortableTable", () => ({
  MerchandiseSortableTable: ({
    onMoveItem,
    isReordering,
    canEdit,
    canReorder,
  }: {
    onMoveItem?: (item: FrontendMerchandiseItem, direction: "up" | "down") => void;
    isReordering?: boolean;
    canEdit?: boolean;
    canReorder?: boolean;
  }) => {
    mocks.merchandiseMove = onMoveItem;
    mocks.merchandiseIsReordering = isReordering;
    mocks.merchandiseCanEdit = canEdit;
    mocks.merchandiseCanReorder = canReorder;
    return null;
  },
}));
vi.mock("../components/PermissionGroupSidePanel", () => ({ PermissionGroupSidePanel: () => null }));
vi.mock("../components/PermissionGroupSortableTable", () => ({
  PERMISSION_GROUP_COLUMNS: [],
  PermissionGroupSortableTable: () => null,
}));

function mutationStub(mutate: ReturnType<typeof vi.fn>) {
  return { mutate, mutateAsync: vi.fn() };
}

vi.mock("../api/animal-species", () => ({
  useGetAnimalSpecies: () => ({ data: [] }),
  useCreateAnimalSpecies: () => mutationStub(vi.fn()),
  useUpdateAnimalSpecies: () => mutationStub(vi.fn()),
  useDeleteAnimalSpecies: () => mutationStub(vi.fn()),
  useReorderAnimalSpecies: () => mutationStub(mocks.animalSpeciesReorder),
}));

vi.mock("../api/cages", () => ({
  useGetAllCages: () => ({ data: [] }),
  useCreateCage: () => mutationStub(vi.fn()),
  useUpdateCage: () => mutationStub(vi.fn()),
  useDeleteCage: () => mutationStub(vi.fn()),
  useReorderCages: () => mutationStub(mocks.cageReorder),
}));

vi.mock("../api/merchandise-items", () => ({
  useGetAllMerchandiseItems: () => ({ data: mocks.merchandiseItems }),
  useCreateMerchandiseItem: () => mutationStub(vi.fn()),
  useUpdateMerchandiseItem: () => mutationStub(vi.fn()),
  useDeleteMerchandiseItem: () => mutationStub(vi.fn()),
  useReorderMerchandiseItems: () => ({
    ...mutationStub(mocks.merchandiseReorder),
    isPending: mocks.merchandisePending,
  }),
}));

vi.mock("../api/permission-groups", () => ({
  useGetPermissionGroups: () => ({ data: [] }),
  useCreatePermissionGroup: () => mutationStub(vi.fn()),
  useUpdatePermissionGroup: () => mutationStub(vi.fn()),
  useDeletePermissionGroup: () => mutationStub(vi.fn()),
  useReorderPermissionGroups: () => mutationStub(mocks.permissionGroupReorder),
}));

function callLatestReorder() {
  const callback = mocks.reorderCallbacks.at(-1);
  expect(callback).toBeDefined();
  callback?.(["2", "1"]);
}

describe("master route reorder permission guards", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.reorderCallbacks.length = 0;
    mocks.canEdit = false;
    mocks.merchandisePending = false;
    mocks.merchandiseFiltered = false;
    mocks.merchandiseMove = undefined;
    mocks.merchandiseIsReordering = undefined;
    mocks.merchandiseCanEdit = undefined;
    mocks.merchandiseCanReorder = undefined;
    mocks.merchandiseItems = Array.from({ length: 101 }, (_, i) => ({
      id: String(i + 1),
      name: `Item ${i + 1}`,
      category: "goods",
      unitPrice: 1100,
      taxType: "included",
      taxRate: 0.1,
      isActive: true,
      sortOrder: i,
    }));
  });

  it("blocks animal species reorder mutation when edit permission is absent", () => {
    render(<AnimalSpeciesSettings />);
    callLatestReorder();
    expect(mocks.animalSpeciesReorder).not.toHaveBeenCalled();
  });

  it("blocks permission group reorder mutation when edit permission is absent", () => {
    render(<PermissionGroupSettings />);
    callLatestReorder();
    expect(mocks.permissionGroupReorder).not.toHaveBeenCalled();
  });

  it("blocks cage reorder mutation when edit permission is absent", () => {
    render(<CageSettings />);
    callLatestReorder();
    expect(mocks.cageReorder).not.toHaveBeenCalled();
  });

  it("blocks merchandise reorder mutation when edit permission is absent", () => {
    render(<MerchandiseItemSettings />);
    callLatestReorder();
    expect(mocks.merchandiseReorder).not.toHaveBeenCalled();
  });

  it("絞り込みで非表示の行がある場合は移動・ドラッグの部分順序を送信しない", () => {
    mocks.canEdit = true;
    mocks.merchandiseFiltered = true;
    render(<MerchandiseItemSettings />);
    expect(mocks.merchandiseCanEdit).toBe(true);
    expect(mocks.merchandiseCanReorder).toBe(false);
    mocks.merchandiseMove?.(mocks.merchandiseItems[0]!, "down");
    callLatestReorder();
    expect(mocks.merchandiseReorder).not.toHaveBeenCalled();
  });

  it.each([
    [false, false],
    [true, true],
  ])("blocks merchandise move with canEdit=%s and pending=%s", (canEdit, pending) => {
    mocks.canEdit = canEdit;
    mocks.merchandisePending = pending;
    render(<MerchandiseItemSettings />);
    expect(mocks.merchandiseMove).toBeDefined();
    mocks.merchandiseMove?.(mocks.merchandiseItems[49]!, "down");
    expect(mocks.merchandiseReorder).not.toHaveBeenCalled();
  });

  it.each([
    [49, "down"],
    [50, "up"],
  ] as const)("sends full cross-page order for row index %s moved %s", (index, direction) => {
    mocks.canEdit = true;
    render(<MerchandiseItemSettings />);
    mocks.merchandiseMove?.(mocks.merchandiseItems[index]!, direction);
    const ids = Array.from({ length: 101 }, (_, i) => i + 1);
    [ids[49], ids[50]] = [ids[50]!, ids[49]!];
    expect(mocks.merchandiseReorder).toHaveBeenCalledWith(
      { ids },
      expect.objectContaining({ onSuccess: expect.any(Function) }),
    );
  });

  it("PR665: onReorder callback ignores invocation while reorder mutation is pending", () => {
    mocks.canEdit = true;
    mocks.merchandisePending = true;
    render(<MerchandiseItemSettings />);
    callLatestReorder();
    expect(mocks.merchandiseReorder).not.toHaveBeenCalled();
  });

  it("PR665: route propagates isReordering while reorder mutation is pending (dragDisabled includes pending)", () => {
    mocks.canEdit = true;
    mocks.merchandisePending = true;
    render(<MerchandiseItemSettings />);
    expect(mocks.merchandiseIsReordering).toBe(true);
    expect(mocks.merchandiseCanEdit).toBe(true);
  });

  it("PR665: route propagates isReordering=false with edit permission and no pending reorder", () => {
    mocks.canEdit = true;
    mocks.merchandisePending = false;
    render(<MerchandiseItemSettings />);
    expect(mocks.merchandiseIsReordering).toBe(false);
  });
});
