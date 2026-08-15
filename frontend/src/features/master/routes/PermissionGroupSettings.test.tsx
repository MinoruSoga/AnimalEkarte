import type { ReactNode } from "react";
import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { PermissionGroup } from "../api/permission-groups";
import type { PermissionGroupFormData } from "../components/permission-group-side-panel-model";
import { PermissionGroupSettings } from "./PermissionGroupSettings";

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
  return { mutate, mutateAsync: mutate };
}

function echoPermissionGroupFromRequest(
  req: {
    name: string;
    description?: string;
    color?: string;
    is_active?: boolean;
    rules?: Array<{
      resource: string;
      can_view: boolean;
      can_create: boolean;
      can_edit: boolean;
      can_delete: boolean;
    }>;
  },
  id = "1",
): PermissionGroup {
  return {
    id,
    clinicId: "1",
    name: req.name,
    description: req.description ?? "",
    color: req.color ?? "#000000",
    isActive: req.is_active ?? true,
    sortOrder: 1,
    rules:
      req.rules?.map((rule, index) => ({
        id: String(index + 1),
        groupId: id,
        resource: rule.resource,
        canView: rule.can_view,
        canCreate: rule.can_create,
        canEdit: rule.can_edit,
        canDelete: rule.can_delete,
        createdAt: "",
        updatedAt: "",
      })) ?? [],
    createdAt: "",
    updatedAt: "",
  };
}

vi.mock("../api/permission-groups", () => ({
  useGetPermissionGroups: () => ({ data: [] }),
  useCreatePermissionGroup: () => mutationStub(mocks.createMutate),
  useUpdatePermissionGroup: () => mutationStub(mocks.updateMutate),
  useDeletePermissionGroup: () => mutationStub(mocks.deleteMutate),
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
        resource: "owners",
        canView: true,
        canCreate: true,
        canEdit: false,
        canDelete: false,
      },
    ],
  };
}

describe("PermissionGroupSettings permission mutation boundaries", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.reorderCallbacks.length = 0;
    mocks.editTarget = "new";
    mocks.permissions.canCreate = true;
    mocks.permissions.canEdit = true;
    mocks.permissions.canDelete = true;
    mocks.createMutate.mockImplementation(async (req) =>
      echoPermissionGroupFromRequest(req, "9"),
    );
    mocks.updateMutate.mockImplementation(async ({ id, req }) =>
      echoPermissionGroupFromRequest(req, id),
    );
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

  it("create payloadにrulesを含め、権限グループAPI mutationを1回だけ発行する", async () => {
    const user = userEvent.setup();
    render(<PermissionGroupSettings />);

    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(mocks.createMutate).toHaveBeenCalledTimes(1);
    const createPayload = mocks.createMutate.mock.calls[0]?.[0] as {
      name: string;
      rules: Array<{ resource: string; can_view: boolean; can_create: boolean }>;
    };
    expect(createPayload.name).toBe("受付");
    expect(createPayload.rules.length).toBeGreaterThan(1);
    expect(createPayload.rules.find((rule) => rule.resource === "owners")).toEqual(
      expect.objectContaining({
        resource: "owners",
        can_view: true,
        can_create: true,
        can_edit: false,
        can_delete: false,
      }),
    );
    expect(createPayload.rules.find((rule) => rule.resource === "reception")).toEqual(
      expect.objectContaining({
        resource: "reception",
        can_view: false,
        can_create: false,
      }),
    );

    expect(mocks.updateMutate).not.toHaveBeenCalled();
  });

  it("update payloadにrulesを含め、権限グループAPI mutationを1回だけ発行する", async () => {
    mocks.editTarget = { id: "1" };
    const user = userEvent.setup();
    render(<PermissionGroupSettings />);

    await user.click(screen.getByRole("button", { name: "保存" }));
    expect(mocks.updateMutate).toHaveBeenCalledTimes(1);
    const updateArg = mocks.updateMutate.mock.calls[0]?.[0] as {
      id: string;
      req: {
        name: string;
        rules: Array<{ resource: string; can_view: boolean }>;
      };
    };
    expect(updateArg.id).toBe("1");
    expect(updateArg.req.name).toBe("受付");
    expect(updateArg.req.rules.length).toBeGreaterThan(1);
    expect(updateArg.req.rules.find((rule) => rule.resource === "owners")).toEqual(
      expect.objectContaining({
        resource: "owners",
        can_view: true,
        can_create: true,
        can_edit: false,
        can_delete: false,
      }),
    );

    expect(mocks.createMutate).not.toHaveBeenCalled();
  });
});
