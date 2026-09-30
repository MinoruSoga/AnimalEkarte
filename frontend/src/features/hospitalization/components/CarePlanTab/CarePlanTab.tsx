// React/Framework
import { useState, useCallback, useLayoutEffect, useMemo, useRef, memo } from "react";

// External
import { Loader2 } from "lucide-react";

// Internal
import { C, ICON } from "@/lib/design-tokens";
import { EmptyState } from "@/components/shared/DataStates";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { usePermission } from "@/hooks/use-permission";

// Relative
import {
  useGetCarePlanItems,
  useCreateCarePlanItem,
  useUpdateCarePlanItem,
  useDeleteCarePlanItem,
} from "../../api/care-plan-items";
import {
  HOSPITALIZATION_DECEASED_BLOCK_MESSAGE,
  HOSPITALIZATION_DISCHARGED_BLOCK_MESSAGE,
} from "../../constants";
import { EditRow } from "./EditRow";
import { ItemRow } from "./ItemRow";
import { AddForm } from "./AddForm";

// Types
import type { CreateCarePlanItemInput, UpdateCarePlanItemInput } from "../../api/care-plan-items";

interface CarePlanTabProps {
  hospitalizationId: string;
  petIsDeceased: boolean;
  isDischarged: boolean;
}

type CarePlanMutation = "create" | "edit" | "delete";

const PERMISSION_BY_MUTATION = {
  create: "canCreate",
  edit: "canEdit",
  delete: "canDelete",
} as const;

export const CarePlanTab = memo(function CarePlanTab({
  hospitalizationId,
  petIsDeceased,
  isDischarged,
}: CarePlanTabProps) {
  const { canCreate, canEdit, canDelete } = usePermission("hospitalization");
  const permissionsRef = useRef({ canCreate, canEdit, canDelete });
  const petIsDeceasedRef = useRef(petIsDeceased);
  const isDischargedRef = useRef(isDischarged);
  useLayoutEffect(() => {
    permissionsRef.current = { canCreate, canEdit, canDelete };
    petIsDeceasedRef.current = petIsDeceased;
    isDischargedRef.current = isDischarged;
  }, [canCreate, canDelete, canEdit, petIsDeceased, isDischarged]);
  const isMutationAllowed = useCallback(
    (action: CarePlanMutation) =>
      permissionsRef.current[PERMISSION_BY_MUTATION[action]] === true &&
      petIsDeceasedRef.current !== true &&
      isDischargedRef.current !== true,
    [],
  );
  const { data: items, isLoading } = useGetCarePlanItems(hospitalizationId);
  // rerender-dependencies: useMutation の戻り値オブジェクト全体でなく、安定参照の関数のみを deps に置く。
  const { mutateAsync: createItemAsync } = useCreateCarePlanItem(hospitalizationId);
  const { mutateAsync: updateItemAsync } = useUpdateCarePlanItem(hospitalizationId);
  const { mutate: deleteItemMutate } = useDeleteCarePlanItem(hospitalizationId);

  const [editingId, setEditingId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [pendingDeleteId, setPendingDeleteId] = useState<string | null>(null);

  const handleEdit = useCallback((id: string) => {
    setEditingId(id);
  }, []);

  const handleCancelEdit = useCallback(() => {
    setEditingId(null);
  }, []);

  // 臨床安全境界1&2: mutation 直前に permission と petIsDeceased を再検査する。
  const handleSaveEdit = useCallback(
    async (itemId: string, input: UpdateCarePlanItemInput) => {
      if (!isMutationAllowed("edit")) return;
      try {
        await updateItemAsync({ itemId, input });
        setEditingId(null);
      } catch {
        // useUpdateCarePlanItem.onError → handleApiError 済み
      }
    },
    [isMutationAllowed, updateItemAsync],
  );

  // 破壊操作は ConfirmDialog 経由: ここでは確認ダイアログを開くだけで mutation は実行しない。
  const handleDelete = useCallback(
    (itemId: string) => {
      if (!isMutationAllowed("delete")) return;
      setPendingDeleteId(itemId);
    },
    [isMutationAllowed],
  );

  // 臨床安全境界1&2: 確認直前に permission / petIsDeceased / isDischarged を再検査する。
  const handleConfirmDelete = useCallback(() => {
    if (!pendingDeleteId) return;
    if (!isMutationAllowed("delete")) {
      setPendingDeleteId(null);
      return;
    }
    const itemId = pendingDeleteId;
    setPendingDeleteId(null);
    setDeletingId(itemId);
    deleteItemMutate(itemId, {
      onSettled: () => {
        setDeletingId(null);
      },
    });
  }, [pendingDeleteId, deleteItemMutate, isMutationAllowed]);

  const handleAdd = useCallback(
    async (input: CreateCarePlanItemInput) => {
      if (!isMutationAllowed("create")) return;
      await createItemAsync(input);
    },
    [createItemAsync, isMutationAllowed],
  );

  // 臨床安全境界1: 死亡ペット/退院済み入院は render 側でも操作要素を出さない（callback 側は isMutationAllowed で維持）。
  const canCreateNow = canCreate && !petIsDeceased && !isDischarged;
  const canEditNow = canEdit && !petIsDeceased && !isDischarged;
  const canDeleteNow = canDelete && !petIsDeceased && !isDischarged;
  const showDeceasedNotice = petIsDeceased && (canCreate || canEdit || canDelete);
  // 死亡と退院済みが両立する場合は死亡センチネルを優先する。
  const showDischargedNotice =
    !petIsDeceased && isDischarged && (canCreate || canEdit || canDelete);

  const itemRows = useMemo(() => {
    if (!items) return null;
    return items.map((item) =>
      editingId === item.id ? (
        <EditRow
          key={item.id}
          item={item}
          onSave={(input) => handleSaveEdit(item.id, input)}
          onCancel={handleCancelEdit}
        />
      ) : (
        <ItemRow
          key={item.id}
          item={item}
          onEdit={canEditNow ? handleEdit : undefined}
          onDelete={canDeleteNow ? handleDelete : undefined}
          isDeleting={deletingId === item.id}
        />
      ),
    );
  }, [
    items,
    editingId,
    deletingId,
    handleEdit,
    handleDelete,
    handleSaveEdit,
    handleCancelEdit,
    canEditNow,
    canDeleteNow,
  ]);

  if (isLoading) {
    return (
      <div className={`flex items-center justify-center py-10 ${C.text60}`}>
        <Loader2 className={`${ICON.page} animate-spin mr-2`} />
        <span className="text-sm">読み込み中...</span>
      </div>
    );
  }

  return (
    <div className="flex flex-col">
      {items && items.length === 0 ? (
        <EmptyState message="ケアプラン項目がありません" />
      ) : (
        <div className="flex flex-col gap-1">{itemRows}</div>
      )}
      {canCreateNow ? (
        <AddForm onSubmit={handleAdd} />
      ) : showDeceasedNotice ? (
        <p role="status" className={`text-xs ${C.text50} pt-3 mt-2 border-t ${C.borderLight}`}>
          {HOSPITALIZATION_DECEASED_BLOCK_MESSAGE.CARE_PLAN}
        </p>
      ) : showDischargedNotice ? (
        <p role="status" className={`text-xs ${C.text50} pt-3 mt-2 border-t ${C.borderLight}`}>
          {HOSPITALIZATION_DISCHARGED_BLOCK_MESSAGE.CARE_PLAN}
        </p>
      ) : null}
      <ConfirmDialog
        open={pendingDeleteId !== null}
        onClose={() => setPendingDeleteId(null)}
        title="ケアプラン項目の削除"
        description="このケアプラン項目を削除しますか？この操作は取り消せません。"
        confirmLabel="削除"
        cancelLabel="キャンセル"
        variant="destructive"
        onConfirm={handleConfirmDelete}
      />
    </div>
  );
});
