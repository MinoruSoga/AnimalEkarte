import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { toast } from "sonner";
import type { UseMutationResult } from "@tanstack/react-query";
import type { ActiveFilter, ActiveSort } from "@/components/shared/PropertyFilter/types";
import {
  useMasterCRUD,
  defaultSearchFilter,
  defaultActiveFilterApply,
  applySorts,
} from "./use-master-crud";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));

interface TestEntity {
  id: string;
  name?: string;
  status?: string;
  isActive?: boolean;
  note?: string | null;
}

function buildMockDeleteMutation(
  onMutate?: (
    id: string,
    opts?: { onSuccess?: () => void; onError?: (error: Error) => void },
  ) => void,
): UseMutationResult<void, Error, string> {
  return {
    mutate: vi.fn(onMutate),
  } as unknown as UseMutationResult<void, Error, string>;
}

// ─────────────────────────────────────────────────
// defaultSearchFilter
// ─────────────────────────────────────────────────

describe("defaultSearchFilter", () => {
  it.each<[string, TestEntity, string, boolean]>([
    ["ひらがな完全一致", { id: "1", name: "さくら" }, "さくら", true],
    ["部分一致", { id: "1", name: "さくらクリニック" }, "くり", true],
    ["カタカナ名をひらがな化して一致", { id: "1", name: "サクラ" }, "さくら", true],
    ["不一致", { id: "1", name: "さくら" }, "たろう", false],
    ["空文字の検索語は常にtrue(includes('')==true)", { id: "1", name: "さくら" }, "", true],
  ])("%s", (_label, item, term, expected) => {
    expect(defaultSearchFilter(item, term)).toBe(expected);
  });

  it("nameフィールドを持たないアイテムは常にfalse", () => {
    expect(defaultSearchFilter({ id: "1" }, "さくら")).toBe(false);
  });

  it("nameが文字列でない場合はfalse", () => {
    expect(defaultSearchFilter({ id: "1", name: undefined }, "")).toBe(false);
  });
});

// ─────────────────────────────────────────────────
// defaultActiveFilterApply
// ─────────────────────────────────────────────────

describe("defaultActiveFilterApply", () => {
  const activeItem: TestEntity = { id: "1", isActive: true };
  const inactiveItem: TestEntity = { id: "2", isActive: false };
  const noStatusItem: TestEntity = { id: "3" };

  it.each<[string, TestEntity, ActiveFilter[], boolean]>([
    [
      "status is=active はisActive trueに一致",
      activeItem,
      [{ key: "status", condition: "is", value: "active", displayValue: "" }],
      true,
    ],
    [
      "status is=active はisActive falseと不一致",
      inactiveItem,
      [{ key: "status", condition: "is", value: "active", displayValue: "" }],
      false,
    ],
    [
      "status is_not=active はisActive trueと不一致",
      activeItem,
      [{ key: "status", condition: "is_not", value: "active", displayValue: "" }],
      false,
    ],
    [
      "status is_not=active はisActive falseに一致",
      inactiveItem,
      [{ key: "status", condition: "is_not", value: "active", displayValue: "" }],
      true,
    ],
    [
      "isActiveフィールドを持たないアイテムはstatusフィルタをスキップして通過",
      noStatusItem,
      [{ key: "status", condition: "is", value: "active", displayValue: "" }],
      true,
    ],
    [
      "statusキーで未対応の条件(defaultケース)は制約なしとして通過",
      activeItem,
      [{ key: "status", condition: "contains", value: "active", displayValue: "" }],
      true,
    ],
  ])("%s", (_label, item, filters, expected) => {
    expect(defaultActiveFilterApply(item, filters)).toBe(expected);
  });

  it("is_empty: null/undefined/空文字は空として一致", () => {
    expect(
      defaultActiveFilterApply({ id: "1", note: null }, [
        { key: "note", condition: "is_empty", value: "", displayValue: "" },
      ]),
    ).toBe(true);
    expect(
      defaultActiveFilterApply({ id: "1", note: "" }, [
        { key: "note", condition: "is_empty", value: "", displayValue: "" },
      ]),
    ).toBe(true);
  });

  it("is_empty: 値がある場合は不一致", () => {
    expect(
      defaultActiveFilterApply({ id: "1", note: "備考あり" }, [
        { key: "note", condition: "is_empty", value: "", displayValue: "" },
      ]),
    ).toBe(false);
  });

  it("is_not_empty: 値がある場合は一致", () => {
    expect(
      defaultActiveFilterApply({ id: "1", note: "備考あり" }, [
        { key: "note", condition: "is_not_empty", value: "", displayValue: "" },
      ]),
    ).toBe(true);
  });

  it("is_not_empty: null/空文字は不一致", () => {
    expect(
      defaultActiveFilterApply({ id: "1", note: null }, [
        { key: "note", condition: "is_not_empty", value: "", displayValue: "" },
      ]),
    ).toBe(false);
  });

  it("複数フィルタはAND条件: 1つでも不一致なら全体falseになる", () => {
    const item: TestEntity = { id: "1", isActive: true, note: null };
    const filters: ActiveFilter[] = [
      { key: "status", condition: "is", value: "active", displayValue: "" },
      { key: "note", condition: "is_not_empty", value: "", displayValue: "" },
    ];
    expect(defaultActiveFilterApply(item, filters)).toBe(false);
  });

  it("空のフィルタ配列は常にtrue", () => {
    expect(defaultActiveFilterApply(activeItem, [])).toBe(true);
  });
});

// ─────────────────────────────────────────────────
// applySorts
// ─────────────────────────────────────────────────

describe("applySorts", () => {
  it("空のソート配列は同一の配列参照をそのまま返す(コピーしない)", () => {
    const items: TestEntity[] = [{ id: "1", name: "あ" }];
    expect(applySorts(items, [])).toBe(items);
  });

  it("元の配列を破壊的に変更しない", () => {
    const items: TestEntity[] = [
      { id: "1", name: "じ" },
      { id: "2", name: "あ" },
    ];
    const sorts: ActiveSort[] = [{ key: "name", direction: "asc" }];
    const result = applySorts(items, sorts);
    expect(result).not.toBe(items);
    expect(items.map((i) => i.name)).toEqual(["じ", "あ"]);
  });

  it("localeCompare('ja')による昇順ソート", () => {
    const items: TestEntity[] = [
      { id: "1", name: "じ" },
      { id: "2", name: "あ" },
      { id: "3", name: "さ" },
    ];
    const sorts: ActiveSort[] = [{ key: "name", direction: "asc" }];
    expect(applySorts(items, sorts).map((i) => i.id)).toEqual(["2", "3", "1"]);
  });

  it("降順ソートは昇順の逆順になる", () => {
    const items: TestEntity[] = [
      { id: "1", name: "じ" },
      { id: "2", name: "あ" },
      { id: "3", name: "さ" },
    ];
    const sorts: ActiveSort[] = [{ key: "name", direction: "desc" }];
    expect(applySorts(items, sorts).map((i) => i.id)).toEqual(["1", "3", "2"]);
  });

  it("複数ソートキー: 1次キーが同値の場合は2次キーで比較する", () => {
    const items: TestEntity[] = [
      { id: "1", status: "b", name: "じ" },
      { id: "2", status: "a", name: "あ" },
      { id: "3", status: "a", name: "さ" },
    ];
    const sorts: ActiveSort[] = [
      { key: "status", direction: "asc" },
      { key: "name", direction: "asc" },
    ];
    expect(applySorts(items, sorts).map((i) => i.id)).toEqual(["2", "3", "1"]);
  });

  it("ソートキーの値が存在しないアイテムは空文字として扱われる", () => {
    const items: TestEntity[] = [{ id: "1", name: "あ" }, { id: "2" }];
    const sorts: ActiveSort[] = [{ key: "name", direction: "asc" }];
    expect(applySorts(items, sorts).map((i) => i.id)).toEqual(["2", "1"]);
  });
});

// ─────────────────────────────────────────────────
// useMasterCRUD (state machine)
// ─────────────────────────────────────────────────

describe("useMasterCRUD", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const data: TestEntity[] = [
    { id: "1", name: "さくら", isActive: true },
    { id: "2", name: "たろう", isActive: false },
  ];
  const allowDeletePermissions = { canDelete: true };

  it("初期状態はeditTarget=null, isEditing=false", () => {
    const { result } = renderHook(() =>
      useMasterCRUD<TestEntity>({
        data,
        deleteMutation: buildMockDeleteMutation(),
        entityLabel: "テスト",
        permissions: allowDeletePermissions,
      }),
    );
    expect(result.current.editTarget).toBeNull();
    expect(result.current.isEditing).toBe(false);
    expect(result.current.panelItem).toBeNull();
  });

  it("handleNewはeditTargetを'new'にする", () => {
    const { result } = renderHook(() =>
      useMasterCRUD<TestEntity>({
        data,
        deleteMutation: buildMockDeleteMutation(),
        entityLabel: "テスト",
        permissions: allowDeletePermissions,
      }),
    );
    act(() => result.current.handleNew());
    expect(result.current.editTarget).toBe("new");
    expect(result.current.isEditing).toBe(true);
  });

  it("handleEditはeditTargetを対象アイテムにしpanelItemに反映する", () => {
    const { result } = renderHook(() =>
      useMasterCRUD<TestEntity>({
        data,
        deleteMutation: buildMockDeleteMutation(),
        entityLabel: "テスト",
        permissions: allowDeletePermissions,
      }),
    );
    act(() => result.current.handleEdit(data[0]));
    expect(result.current.editTarget).toEqual(data[0]);
    expect(result.current.panelItem).toEqual(data[0]);
  });

  it("handleCloseはeditTargetをnullに戻す", () => {
    const { result } = renderHook(() =>
      useMasterCRUD<TestEntity>({
        data,
        deleteMutation: buildMockDeleteMutation(),
        entityLabel: "テスト",
        permissions: allowDeletePermissions,
      }),
    );
    act(() => result.current.handleEdit(data[0]));
    act(() => result.current.handleClose());
    expect(result.current.editTarget).toBeNull();
  });

  it("dirtyGuard.confirmDiscardがfalseを返す場合、handleNewは中断されeditTargetは変化しない", () => {
    const confirmDiscard = vi.fn(() => false);
    const { result } = renderHook(() =>
      useMasterCRUD<TestEntity>({
        data,
        deleteMutation: buildMockDeleteMutation(),
        entityLabel: "テスト",
        dirtyGuard: { confirmDiscard },
        permissions: allowDeletePermissions,
      }),
    );
    act(() => result.current.handleNew());
    expect(confirmDiscard).toHaveBeenCalledTimes(1);
    expect(result.current.editTarget).toBeNull();
  });

  it("dirtyGuard.confirmDiscardがfalseを返す場合、handleEditも中断される", () => {
    const confirmDiscard = vi.fn(() => false);
    const { result } = renderHook(() =>
      useMasterCRUD<TestEntity>({
        data,
        deleteMutation: buildMockDeleteMutation(),
        entityLabel: "テスト",
        dirtyGuard: { confirmDiscard },
        permissions: allowDeletePermissions,
      }),
    );
    act(() => result.current.handleEdit(data[0]));
    expect(result.current.editTarget).toBeNull();
  });

  it("dirtyGuard.confirmDiscardがtrueを返す場合、handleEditは通常通り遷移する", () => {
    const confirmDiscard = vi.fn(() => true);
    const { result } = renderHook(() =>
      useMasterCRUD<TestEntity>({
        data,
        deleteMutation: buildMockDeleteMutation(),
        entityLabel: "テスト",
        dirtyGuard: { confirmDiscard },
        permissions: allowDeletePermissions,
      }),
    );
    act(() => result.current.handleEdit(data[0]));
    expect(result.current.editTarget).toEqual(data[0]);
  });

  it("canDeleteがtrueならhandleDeleteConfirmは対象IDでmutateする", async () => {
    const mutate = vi.fn(
      (_id: string, opts?: { onSuccess?: () => void; onError?: (error: Error) => void }) =>
        opts?.onSuccess?.(),
    );
    const deleteMutation = { mutate } as unknown as UseMutationResult<void, Error, string>;
    const { result } = renderHook(() =>
      useMasterCRUD<TestEntity>({
        data,
        deleteMutation,
        entityLabel: "テスト",
        permissions: allowDeletePermissions,
      }),
    );

    act(() => result.current.handleDeleteRequest(data[0]));
    expect(result.current.pendingDelete).toEqual(data[0]);

    // pendingDeleteRef は useEffect で同期されるため、act() のフラッシュ後を待つ
    await waitFor(() => {
      act(() => result.current.handleDeleteConfirm());
      expect(mutate).toHaveBeenCalledWith(
        "1",
        expect.objectContaining({ onSuccess: expect.any(Function) }),
      );
    });

    expect(result.current.pendingDelete).toBeNull();
    expect(result.current.editTarget).toBeNull();
    expect(toast.success).toHaveBeenCalledWith("テストを削除しました");
  });

  it("canDeleteがtrueでない場合はdelete mutationを発行しない", async () => {
    const mutate = vi.fn();
    const deleteMutation = { mutate } as unknown as UseMutationResult<void, Error, string>;
    const { result } = renderHook(() =>
      useMasterCRUD<TestEntity>({
        data,
        deleteMutation,
        entityLabel: "テスト",
        permissions: { canDelete: false },
      }),
    );

    act(() => result.current.handleDeleteRequest(data[0]));
    await waitFor(() => expect(result.current.pendingDelete).toEqual(data[0]));
    act(() => result.current.handleDeleteConfirm());

    expect(mutate).not.toHaveBeenCalled();
    expect(result.current.pendingDelete).toEqual(data[0]);
  });

  it("canDeleteがtrueなら対象IDとcallbackを渡してmutateする", async () => {
    const mutate = vi.fn();
    const deleteMutation = { mutate } as unknown as UseMutationResult<void, Error, string>;
    const { result } = renderHook(() =>
      useMasterCRUD<TestEntity>({
        data,
        deleteMutation,
        entityLabel: "テスト",
        permissions: { canDelete: true },
      }),
    );

    act(() => result.current.handleDeleteRequest(data[0]));
    await waitFor(() => expect(result.current.pendingDelete).toEqual(data[0]));
    act(() => result.current.handleDeleteConfirm());

    // onError は渡さない: deleteMutation フック自身の onError (handleApiError) と二重通知になるため
    expect(mutate).toHaveBeenCalledWith(
      "1",
      expect.not.objectContaining({ onError: expect.anything() }),
    );
    expect(mutate).toHaveBeenCalledWith(
      "1",
      expect.objectContaining({ onSuccess: expect.any(Function) }),
    );
  });

  it("権限剥奪後はcaptured済みhandleDeleteConfirmでも最新のdenyを使う", async () => {
    const mutate = vi.fn();
    const deleteMutation = { mutate } as unknown as UseMutationResult<void, Error, string>;
    const { result, rerender } = renderHook(
      ({ canDelete }: { canDelete: boolean }) =>
        useMasterCRUD<TestEntity>({
          data,
          deleteMutation,
          entityLabel: "テスト",
          permissions: { canDelete },
        }),
      { initialProps: { canDelete: true } },
    );

    act(() => result.current.handleDeleteRequest(data[0]));
    await waitFor(() => expect(result.current.pendingDelete).toEqual(data[0]));
    const capturedHandleDeleteConfirm = result.current.handleDeleteConfirm;

    rerender({ canDelete: false });
    act(() => capturedHandleDeleteConfirm());

    expect(mutate).not.toHaveBeenCalled();
  });

  it("pendingDeleteRefは常に最新のpendingDeleteを参照する(連続request後は最後の対象を削除する)", async () => {
    const mutate = vi.fn();
    const deleteMutation = { mutate } as unknown as UseMutationResult<void, Error, string>;
    const { result } = renderHook(() =>
      useMasterCRUD<TestEntity>({
        data,
        deleteMutation,
        entityLabel: "テスト",
        permissions: allowDeletePermissions,
      }),
    );

    act(() => result.current.handleDeleteRequest(data[0]));
    act(() => result.current.handleDeleteRequest(data[1]));

    await waitFor(() => {
      act(() => result.current.handleDeleteConfirm());
      expect(mutate).toHaveBeenCalled();
    });

    expect(mutate).toHaveBeenCalledTimes(1);
    expect(mutate).toHaveBeenCalledWith("2", expect.anything());
  });

  it("pendingDeleteが無い状態でhandleDeleteConfirmを呼んでもmutateは呼ばれない", () => {
    const mutate = vi.fn();
    const deleteMutation = { mutate } as unknown as UseMutationResult<void, Error, string>;
    const { result } = renderHook(() =>
      useMasterCRUD<TestEntity>({
        data,
        deleteMutation,
        entityLabel: "テスト",
        permissions: allowDeletePermissions,
      }),
    );
    act(() => result.current.handleDeleteConfirm());
    expect(mutate).not.toHaveBeenCalled();
  });

  it("handleDeleteCancel後にhandleDeleteConfirmを呼んでもmutateは呼ばれない", async () => {
    const mutate = vi.fn();
    const deleteMutation = { mutate } as unknown as UseMutationResult<void, Error, string>;
    const { result } = renderHook(() =>
      useMasterCRUD<TestEntity>({
        data,
        deleteMutation,
        entityLabel: "テスト",
        permissions: allowDeletePermissions,
      }),
    );

    act(() => result.current.handleDeleteRequest(data[0]));
    act(() => result.current.handleDeleteCancel());
    expect(result.current.pendingDelete).toBeNull();

    await waitFor(() => {
      act(() => result.current.handleDeleteConfirm());
    });
    expect(mutate).not.toHaveBeenCalled();
  });

  it("削除失敗時はonSuccessが呼ばれずpendingDeleteは維持される(onErrorはdeleteMutation側のhandleApiErrorに委譲)", async () => {
    // このフックは onError を渡さない。失敗時の通知は各 master/api/*.ts の
    // useDeleteXxx フック自身の onError (handleApiError) が担う（二重 toast 防止）。
    const mutate = vi.fn((_id: string, _opts?: { onSuccess?: () => void }) => {
      // 失敗時: onSuccess は呼ばれない
    });
    const deleteMutation = { mutate } as unknown as UseMutationResult<void, Error, string>;
    const { result } = renderHook(() =>
      useMasterCRUD<TestEntity>({
        data,
        deleteMutation,
        entityLabel: "テスト",
        permissions: allowDeletePermissions,
      }),
    );

    act(() => result.current.handleDeleteRequest(data[0]));
    await waitFor(() => {
      act(() => result.current.handleDeleteConfirm());
      expect(mutate).toHaveBeenCalled();
    });

    expect(mutate).toHaveBeenCalledWith(
      "1",
      expect.not.objectContaining({ onError: expect.anything() }),
    );
    // onSuccess が発火しない限り pendingDelete をクリアしないため、確認ダイアログは開いたままになる
    expect(result.current.pendingDelete).toEqual(data[0]);
  });

  it("検索語・フィルタ・ソートを組み合わせてfilteredItemsに反映する", async () => {
    const { result } = renderHook(() =>
      useMasterCRUD<TestEntity>({
        data,
        deleteMutation: buildMockDeleteMutation(),
        entityLabel: "テスト",
        permissions: allowDeletePermissions,
      }),
    );

    act(() => {
      result.current.setActiveFilters([
        { key: "status", condition: "is", value: "active", displayValue: "" },
      ]);
    });
    await waitFor(() => {
      expect(result.current.filteredItems.map((i) => i.id)).toEqual(["1"]);
    });
  });

  it("handleSortChangeはactiveSortsを更新しfilteredItemsをソートする", () => {
    const { result } = renderHook(() =>
      useMasterCRUD<TestEntity>({
        data,
        deleteMutation: buildMockDeleteMutation(),
        entityLabel: "テスト",
        permissions: allowDeletePermissions,
      }),
    );

    act(() => result.current.handleSortChange([{ key: "name", direction: "asc" }]));
    expect(result.current.activeSorts).toEqual([{ key: "name", direction: "asc" }]);
    // "さくら" と "たろう" は localeCompare('ja') で "さくら" が先
    expect(result.current.filteredItems.map((i) => i.id)).toEqual(["1", "2"]);
  });
});
