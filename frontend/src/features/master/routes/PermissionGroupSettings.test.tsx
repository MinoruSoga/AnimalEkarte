import type { ReactNode } from "react";
import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type {
  CreatePermissionGroupRequest,
  PermissionGroup,
  UpdatePermissionGroupRequest,
} from "../api/permission-groups";
import type { PermissionGroupFormData } from "../components/permission-group-side-panel-model";
import { PermissionGroupSettings } from "./PermissionGroupSettings";

interface MutationCallbacks {
  onSuccess?: (saved: PermissionGroup) => Promise<void> | void;
}

const mocks = vi.hoisted(() => ({
  permissions: {
    canCreate: true,
    canEdit: true,
    canDelete: true,
  },
  editTarget: "new" as "new" | { id: string } | null,
  reorderCallbacks: [] as Array<(ids: string[]) => void>,
  createMutate: vi.fn(),
  updateMutate: vi.fn(),
  deleteMutate: vi.fn(),
  updateRulesMutateAsync: vi.fn(),
  reorderMutate: vi.fn(),
  setEditTarget: vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => mocks.permissions,
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
    items: PermissionGroup[];
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
    editTarget: mocks.editTarget,
    setEditTarget: mocks.setEditTarget,
    startSaveTransition: (callback: () => void) => callback(),
    filteredItems: [],
    pendingDelete: null,
    handleEdit: vi.fn(),
    handleNew: vi.fn(),
  }),
}));

vi.mock("../components/MasterCRUDPage", () => ({
  MasterCRUDPage: ({
    children,
    handleSave,
  }: {
    children?: ReactNode;
    handleSave: (data: PermissionGroupFormData) => void;
  }) => (
    <>
      <button type="button" onClick={() => handleSave(buildFormData())}>
        保存
      </button>
      {children}
    </>
  ),
}));

vi.mock("../components/PermissionGroupSidePanel", () => ({
  PermissionGroupSidePanel: () => null,
}));

vi.mock("../components/PermissionGroupSortableTable", () => ({
  PERMISSION_GROUP_COLUMNS: [],
  PermissionGroupSortableTable: () => null,
}));

function mutationStub(mutate: ReturnType<typeof vi.fn>) {
  return { mutate, mutateAsync: vi.fn() };
}

vi.mock("../api/permission-groups", () => ({
  useGetPermissionGroups: () => ({ data: [] }),
  useCreatePermissionGroup: () => mutationStub(mocks.createMutate),
  useUpdatePermissionGroup: () => mutationStub(mocks.updateMutate),
  useDeletePermissionGroup: () => mutationStub(mocks.deleteMutate),
  useUpdatePermissionGroupRules: () => ({
    mutate: vi.fn(),
    mutateAsync: mocks.updateRulesMutateAsync,
  }),
  useReorderPermissionGroups: () => mutationStub(mocks.reorderMutate),
}));

function buildFormData(): PermissionGroupFormData {
  return {
    name: "受付",
    description: "受付担当",
    color: "#000000",
    isActive: true,
    rules: [
      {
        resource: "owner",
        canView: true,
        canCreate: true,
        canEdit: false,
        canDelete: false,
      },
    ],
  };
}

function buildSavedGroup(): PermissionGroup {
  return {
    id: "1",
    clinicId: "1",
    name: "受付",
    description: "受付担当",
    color: "#000000",
    isActive: true,
    sortOrder: 1,
    rules: [],
    createdAt: "2026-07-26T00:00:00Z",
    updatedAt: "2026-07-26T00:00:00Z",
  };
}

function latestMutationSuccess<TRequest>(
  mutate: ReturnType<typeof vi.fn>,
): MutationCallbacks["onSuccess"] {
  const calls = mutate.mock.calls as unknown as Array<
    [TRequest, MutationCallbacks]
  >;
  return calls.at(-1)?.[1].onSuccess;
}

describe("PermissionGroupSettings permission mutation boundaries", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.reorderCallbacks.length = 0;
    mocks.editTarget = "new";
    mocks.permissions.canCreate = true;
    mocks.permissions.canEdit = true;
    mocks.permissions.canDelete = true;
  });

  it("captured reorder callback blocks mutation after edit permission is revoked", () => {
    const { rerender } = render(<PermissionGroupSettings />);
    const capturedReorder = mocks.reorderCallbacks.at(-1);
    expect(capturedReorder).toBeDefined();

    mocks.permissions.canEdit = false;
    rerender(<PermissionGroupSettings />);
    act(() => capturedReorder?.(["2", "1"]));

    expect(mocks.reorderMutate).not.toHaveBeenCalled();
  });

  it("create-only persona may update rules after create while edit permission is absent", async () => {
    mocks.permissions.canEdit = false;
    const user = userEvent.setup();
    render(<PermissionGroupSettings />);

    await user.click(screen.getByRole("button", { name: "保存" }));
    const onSuccess =
      latestMutationSuccess<CreatePermissionGroupRequest>(mocks.createMutate);
    expect(onSuccess).toBeDefined();

    await act(async () => {
      await onSuccess?.(buildSavedGroup());
    });

    expect(mocks.updateRulesMutateAsync).toHaveBeenCalledTimes(1);
  });

  it("create後のrules mutationは最新create permissionが剥奪済みなら発行しない", async () => {
    const user = userEvent.setup();
    const { rerender } = render(<PermissionGroupSettings />);

    await user.click(screen.getByRole("button", { name: "保存" }));
    const onSuccess =
      latestMutationSuccess<CreatePermissionGroupRequest>(mocks.createMutate);
    expect(onSuccess).toBeDefined();

    mocks.permissions.canCreate = false;
    rerender(<PermissionGroupSettings />);
    await act(async () => {
      await onSuccess?.(buildSavedGroup());
    });

    expect(mocks.updateRulesMutateAsync).not.toHaveBeenCalled();
  });

  it("update後のrules mutationは最新edit permissionが剥奪済みなら発行しない", async () => {
    mocks.editTarget = { id: "1" };
    const user = userEvent.setup();
    const { rerender } = render(<PermissionGroupSettings />);

    await user.click(screen.getByRole("button", { name: "保存" }));
    const onSuccess =
      latestMutationSuccess<{
        id: string;
        req: UpdatePermissionGroupRequest;
      }>(mocks.updateMutate);
    expect(onSuccess).toBeDefined();

    mocks.permissions.canEdit = false;
    rerender(<PermissionGroupSettings />);
    await act(async () => {
      await onSuccess?.(buildSavedGroup());
    });

    expect(mocks.updateRulesMutateAsync).not.toHaveBeenCalled();
  });
});
