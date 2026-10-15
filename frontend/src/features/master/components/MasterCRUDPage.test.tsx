import type { ReactNode, TransitionStartFunction } from "react";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { UseMasterCRUDReturn } from "../hooks/use-master-crud";
import { MasterCRUDPage } from "./MasterCRUDPage";
import { ResourceMasterPermission } from "@/types/generated/models";

const mocks = vi.hoisted(() => ({
  canCreate: true,
  canEdit: false,
  canDelete: false,
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => mocks,
}));

const masterListPagePropsSpy = vi.fn();
vi.mock("./MasterListPage", () => ({
  MasterListPage: (props: { sidePanel: ReactNode; maxWidth?: string }) => {
    masterListPagePropsSpy(props);
    return <>{props.sidePanel}</>;
  },
}));

interface TestEntity {
  id: string;
  name: string;
}

function buildCrud(editTarget: TestEntity | "new"): UseMasterCRUDReturn<TestEntity> {
  const panelItem = editTarget === "new" ? null : editTarget;
  const startSaveTransition: TransitionStartFunction = (callback) => {
    callback();
  };

  return {
    editTarget,
    setEditTarget: vi.fn(),
    searchTerm: "",
    setSearchTerm: vi.fn(),
    pendingDelete: null,
    setPendingDelete: vi.fn(),
    isSavePending: false,
    activeFilters: [],
    setActiveFilters: vi.fn(),
    activeSorts: [],
    setActiveSorts: vi.fn(),
    filteredItems: panelItem === null ? [] : [panelItem],
    panelItem,
    isEditing: true,
    handleClose: vi.fn(),
    handleNew: vi.fn(),
    handleEdit: vi.fn(),
    handleDeleteRequest: vi.fn(),
    handleDeleteConfirm: vi.fn(),
    handleDeleteCancel: vi.fn(),
    handleSortChange: vi.fn(),
    startSaveTransition,
  };
}

function renderPage(editTarget: TestEntity | "new") {
  render(
    <MasterCRUDPage
      title="権限グループマスタ"
      resource={ResourceMasterPermission}
      entityLabel="グループ"
      searchPlaceholder="検索"
      emptyMessage="空"
      crud={buildCrud(editTarget)}
      handleSave={vi.fn()}
      columns={[]}
      renderRow={() => null}
      renderSidePanel={({ readOnly }) => <p>{readOnly ? "閲覧のみ" : "編集可能"}</p>}
    >
      <div>一覧</div>
    </MasterCRUDPage>,
  );
}

describe("MasterCRUDPage permission-specific panel mode", () => {
  it.each([
    {
      label: "新規作成",
      editTarget: "new" as const,
      expected: "編集可能",
    },
    {
      label: "既存編集",
      editTarget: { id: "1", name: "管理者" },
      expected: "閲覧のみ",
    },
  ])(
    "create-only persona: $labelパネルの操作可否をaction-specificに判定する",
    ({ editTarget, expected }) => {
      renderPage(editTarget);

      expect(screen.getByText(expected)).toBeInTheDocument();
    },
  );
});

describe("MasterCRUDPage maxWidth auto-assignment (EMR-333)", () => {
  it("列数が5列以下のときは tableNarrow (max-w-5xl) を MasterListPage に渡す", () => {
    masterListPagePropsSpy.mockClear();
    render(
      <MasterCRUDPage
        title="職種マスタ"
        entityLabel="職種"
        searchPlaceholder="検索"
        emptyMessage="空"
        crud={buildCrud("new")}
        handleSave={vi.fn()}
        columns={[
          { header: "名称" },
          { header: "説明" },
          { header: "ステータス" },
          { header: "操作" },
        ]}
        renderRow={() => null}
        renderSidePanel={() => null}
      />,
    );

    expect(masterListPagePropsSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        maxWidth: "max-w-5xl",
      }),
    );
  });

  it("列数が6列以上のときは full (max-w-full) を MasterListPage に渡す", () => {
    masterListPagePropsSpy.mockClear();
    render(
      <MasterCRUDPage
        title="多列マスタ"
        entityLabel="マスタ"
        searchPlaceholder="検索"
        emptyMessage="空"
        crud={buildCrud("new")}
        handleSave={vi.fn()}
        columns={[
          { header: "1" },
          { header: "2" },
          { header: "3" },
          { header: "4" },
          { header: "5" },
          { header: "6" },
        ]}
        renderRow={() => null}
        renderSidePanel={() => null}
      />,
    );

    expect(masterListPagePropsSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        maxWidth: "max-w-full",
      }),
    );
  });

  it("maxWidth が明示指定されたときは列数に関わらずその値が優先される", () => {
    masterListPagePropsSpy.mockClear();
    render(
      <MasterCRUDPage
        title="カスタム幅マスタ"
        entityLabel="マスタ"
        searchPlaceholder="検索"
        emptyMessage="空"
        crud={buildCrud("new")}
        handleSave={vi.fn()}
        maxWidth="max-w-3xl"
        columns={[{ header: "1" }]}
        renderRow={() => null}
        renderSidePanel={() => null}
      />,
    );

    expect(masterListPagePropsSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        maxWidth: "max-w-3xl",
      }),
    );
  });
});
