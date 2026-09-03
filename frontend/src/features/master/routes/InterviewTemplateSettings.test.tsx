import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue, AuthUser } from "@/types/auth";
import { InterviewTemplateSettings } from "./InterviewTemplateSettings";

const mocks = vi.hoisted(() => ({
  queryResult: {
    data: [] as Array<{
      id: string;
      category: string;
      title: string;
      content?: string;
      isActive: boolean;
    }>,
    isPending: false,
    isError: false,
    error: null as Error | null,
  },
  handleNew: vi.fn(),
  crudData: [] as unknown[],
  crudCanDelete: true,
  saveCanCreate: true,
  saveCanEdit: true,
  resourcePermissions: {
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  },
  usePermissionResource: null as string | null,
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: (resource: string) => {
    mocks.usePermissionResource = resource;
    return { ...mocks.resourcePermissions };
  },
}));

vi.mock("@/hooks/use-side-peek-dirty", () => ({
  useSidePeekDirty: () => ({
    markDirty: vi.fn(),
    markClean: vi.fn(),
    confirmDiscard: vi.fn(() => true),
  }),
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
  useMasterSave: ({ permissions }: { permissions: { canCreate: boolean; canEdit: boolean } }) => {
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
        問診テンプレートを追加
      </button>
      {children}
    </>
  ),
}));

vi.mock("../components/InterviewTemplateSidePanel", () => ({
  InterviewTemplateSidePanel: () => null,
}));

function mutationStub() {
  return { mutate: vi.fn(), mutateAsync: vi.fn() };
}

vi.mock("../api/inquiry-templates", () => ({
  useGetInquiryTemplates: () => mocks.queryResult,
  useCreateInquiryTemplate: mutationStub,
  useUpdateInquiryTemplate: mutationStub,
  useDeleteInquiryTemplate: mutationStub,
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

function setPermissions({
  canView = true,
  canCreate = true,
  canEdit = true,
  canDelete = true,
}: Partial<typeof mocks.resourcePermissions> = {}) {
  mocks.resourcePermissions = { canView, canCreate, canEdit, canDelete };
}

function authValue(): AuthContextValue {
  const user = {
    id: "1",
    email: "test@example.com",
    displayName: "Test",
    isSystemAdmin: false,
    mainClinicId: "1",
    clinic: null,
    clinics: [],
    permissions: {},
  } satisfies AuthUser;

  return {
    user,
    currentClinicId: "1",
    isAuthenticated: true,
    isLoading: false,
    login: vi.fn(),
    logout: vi.fn(),
    switchClinic: vi.fn(),
    hasPermission: vi.fn(() => true),
    refreshPermissions: vi.fn(),
  };
}

function renderPage() {
  return render(
    <AuthContext.Provider value={authValue()}>
      <InterviewTemplateSettings />
    </AuthContext.Provider>,
  );
}

describe("InterviewTemplateSettings", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setQueryResult({});
    setPermissions();
  });

  it("取得成功時に一覧データを CRUD hook へ渡し、作成/編集権限を保存 hook へ伝える", () => {
    setQueryResult({
      data: [
        {
          id: "1",
          category: "chief_complaint",
          title: "初診",
          content: "内容",
          isActive: true,
        },
      ],
    });

    renderPage();

    expect(mocks.usePermissionResource).toBe("master-medical");
    expect(mocks.crudData).toEqual([
      {
        id: "1",
        category: "chief_complaint",
        title: "初診",
        content: "内容",
        isActive: true,
      },
    ]);
    expect(mocks.crudCanDelete).toBe(true);
    expect(mocks.saveCanCreate).toBe(true);
    expect(mocks.saveCanEdit).toBe(true);
  });

  it("view-only 権限では create/edit/delete を閉じる", () => {
    setPermissions({
      canView: true,
      canCreate: false,
      canEdit: false,
      canDelete: false,
    });
    setQueryResult({
      data: [{ id: "1", category: "history", title: "既往", isActive: true }],
    });

    renderPage();

    expect(mocks.crudCanDelete).toBe(false);
    expect(mocks.saveCanCreate).toBe(false);
    expect(mocks.saveCanEdit).toBe(false);
  });

  it("新規追加操作が handleNew を呼ぶ", async () => {
    const user = userEvent.setup();
    renderPage();

    await user.click(screen.getByRole("button", { name: "問診テンプレートを追加" }));
    expect(mocks.handleNew).toHaveBeenCalledOnce();
  });
});
