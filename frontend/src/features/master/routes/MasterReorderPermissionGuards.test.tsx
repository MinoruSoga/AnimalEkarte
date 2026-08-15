import type { ReactNode } from "react";
import { render } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { AnimalSpeciesSettings } from "./AnimalSpeciesSettings";
import { CageSettings } from "./CageSettings";
import { MerchandiseItemSettings } from "./MerchandiseItemSettings";
import { PermissionGroupSettings } from "./PermissionGroupSettings";

const mocks = vi.hoisted(() => ({
  reorderCallbacks: [] as Array<(ids: string[]) => void>,
  animalSpeciesReorder: vi.fn(),
  cageReorder: vi.fn(),
  merchandiseReorder: vi.fn(),
  permissionGroupReorder: vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canCreate: false, canEdit: false, canDelete: false }),
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
  useMasterCRUD: () => ({
    filteredItems: [],
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
vi.mock("../components/MerchandiseSortableTable", () => ({ MerchandiseSortableTable: () => null }));
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
  useGetAllMerchandiseItems: () => ({ data: [] }),
  useCreateMerchandiseItem: () => mutationStub(vi.fn()),
  useUpdateMerchandiseItem: () => mutationStub(vi.fn()),
  useDeleteMerchandiseItem: () => mutationStub(vi.fn()),
  useReorderMerchandiseItems: () => mutationStub(mocks.merchandiseReorder),
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
});
