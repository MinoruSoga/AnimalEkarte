// React/Framework
import { useState, useCallback, useLayoutEffect, useMemo, useRef, memo } from "react";

// External
import { Loader2 } from "lucide-react";

// Internal
import { C, ICON } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";

// Relative
import { useGetCarePlanItems, useCreateCarePlanItem, useUpdateCarePlanItem, useDeleteCarePlanItem } from "../../api/care-plan-items";
import { HOSPITALIZATION_DECEASED_BLOCK_MESSAGE } from "../../constants";
import { EditRow } from "./EditRow";
import { ItemRow } from "./ItemRow";
import { AddForm } from "./AddForm";

// Types
import type { CreateCarePlanItemInput, UpdateCarePlanItemInput } from "../../api/care-plan-items";

interface CarePlanTabProps {
  hospitalizationId: string;
  petIsDeceased: boolean;
}

type CarePlanMutation = "create" | "edit" | "delete";

const PERMISSION_BY_MUTATION = {
  create: "canCreate",
  edit: "canEdit",
  delete: "canDelete",
} as const;

export const CarePlanTab = memo(function CarePlanTab({ hospitalizationId, petIsDeceased }: CarePlanTabProps) {
  const { canCreate, canEdit, canDelete } = usePermission("hospitalization");
  const permissionsRef = useRef({ canCreate, canEdit, canDelete });
  const petIsDeceasedRef = useRef(petIsDeceased);
  useLayoutEffect(() => {
    permissionsRef.current = { canCreate, canEdit, canDelete };
    petIsDeceasedRef.current = petIsDeceased;
  }, [canCreate, canDelete, canEdit, petIsDeceased]);
  const isMutationAllowed = useCallback(
    (action: CarePlanMutation) =>
      permissionsRef.current[PERMISSION_BY_MUTATION[action]] === true && petIsDeceasedRef.current !== true,
    [],
  );
  const { data: items, isLoading } = useGetCarePlanItems(hospitalizationId);
  // rerender-dependencies: useMutation の戻り値オブジェクト全体でなく、安定参照の関数のみを deps に置く。
  const { mutateAsync: createItemAsync } = useCreateCarePlanItem(hospitalizationId);
  const { mutateAsync: updateItemAsync } = useUpdateCarePlanItem(hospitalizationId);
  const { mutate: deleteItemMutate } = useDeleteCarePlanItem(hospitalizationId);

  const [editingId, setEditingId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);

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

  const handleDelete = useCallback(
    (itemId: string) => {
      if (!isMutationAllowed("delete")) return;
      setDeletingId(itemId);
      deleteItemMutate(itemId, {
        onSettled: () => {
          setDeletingId(null);
        },
      });
    },
    [deleteItemMutate, isMutationAllowed],
  );

  const handleAdd = useCallback(
    async (input: CreateCarePlanItemInput) => {
      if (!isMutationAllowed("create")) return;
      await createItemAsync(input);
    },
    [createItemAsync, isMutationAllowed],
  );

  // 臨床安全境界1: 死亡ペットは render 側でも操作要素を出さない（callback 側は isMutationAllowed で維持）。
  const canCreateNow = canCreate && !petIsDeceased;
  const canEditNow = canEdit && !petIsDeceased;
  const canDeleteNow = canDelete && !petIsDeceased;
  const showDeceasedNotice = petIsDeceased && (canCreate || canEdit || canDelete);

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
  }, [items, editingId, deletingId, handleEdit, handleDelete, handleSaveEdit, handleCancelEdit, canEditNow, canDeleteNow]);

  if (isLoading) {
    return (
      <div className={`flex items-center justify-center py-10 ${C.text40}`}>
        <Loader2 className={`${ICON.page} animate-spin mr-2`} />
        <span className="text-sm">読み込み中...</span>
      </div>
    );
  }

  return (
    <div className="flex flex-col">
      {items && items.length === 0 ? (
        <p className={`text-sm ${C.text40} py-4 text-center`}>ケアプラン項目がありません</p>
      ) : (
        <div className="flex flex-col gap-1">{itemRows}</div>
      )}
      {canCreateNow ? (
        <AddForm onSubmit={handleAdd} />
      ) : showDeceasedNotice ? (
        <p role="status" className={`text-xs ${C.text50} pt-3 mt-2 border-t ${C.borderLight}`}>
          {HOSPITALIZATION_DECEASED_BLOCK_MESSAGE.CARE_PLAN}
        </p>
      ) : null}
    </div>
  );
});
