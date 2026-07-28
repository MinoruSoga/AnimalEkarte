// React/Framework
import { ICON, C } from "@/lib/design-tokens";
import { useState, useCallback, useLayoutEffect, useMemo, useRef, memo } from "react";

// External
import { Loader2 } from "lucide-react";

// Relative
import { useGetCarePlanItems, useCreateCarePlanItem, useUpdateCarePlanItem, useDeleteCarePlanItem } from "../../api/care-plan-items";
import { usePermission } from "@/hooks/use-permission";
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
            permissionsRef.current[PERMISSION_BY_MUTATION[action]] === true &&
            petIsDeceasedRef.current !== true,
        []
    );
    const { data: items, isLoading } = useGetCarePlanItems(hospitalizationId);
    const createItem = useCreateCarePlanItem(hospitalizationId);
    const updateItem = useUpdateCarePlanItem(hospitalizationId);
    const deleteItem = useDeleteCarePlanItem(hospitalizationId);

    const [editingId, setEditingId] = useState<string | null>(null);
    const [deletingId, setDeletingId] = useState<string | null>(null);

    const handleEdit = useCallback((id: string) => {
        setEditingId(id);
    }, []);

    const handleCancelEdit = useCallback(() => {
        setEditingId(null);
    }, []);

    const handleSaveEdit = useCallback(
        (itemId: string, input: UpdateCarePlanItemInput) => {
            if (!isMutationAllowed("edit")) return;
            updateItem.mutate(
                { itemId, input },
                {
                    onSuccess: () => {
                        setEditingId(null);
                    },
                }
            );
        },
        [isMutationAllowed, updateItem]
    );

    const handleDelete = useCallback(
        (itemId: string) => {
            if (!isMutationAllowed("delete")) return;
            setDeletingId(itemId);
            deleteItem.mutate(itemId, {
                onSettled: () => {
                    setDeletingId(null);
                },
            });
        },
        [deleteItem, isMutationAllowed]
    );

    const handleAdd = useCallback(
        (input: CreateCarePlanItemInput) => {
            if (!isMutationAllowed("create")) return;
            createItem.mutate(input);
        },
        [createItem, isMutationAllowed]
    );

    const itemRows = useMemo(() => {
        if (!items) return null;
        return items.map((item) =>
            editingId === item.id ? (
                <EditRow
                    key={item.id}
                    item={item}
                    onSave={(input) => handleSaveEdit(item.id, input)}
                    onCancel={handleCancelEdit}
                    isSaving={updateItem.isPending}
                />
            ) : (
                <ItemRow
                    key={item.id}
                    item={item}
                    onEdit={canEdit ? handleEdit : undefined}
                    onDelete={canDelete ? handleDelete : undefined}
                    isDeleting={deletingId === item.id}
                />
            )
        );
    }, [items, editingId, deletingId, updateItem.isPending, handleEdit, handleDelete, handleSaveEdit, handleCancelEdit, canEdit, canDelete]);

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
                <p className={`text-sm ${C.text40} py-4 text-center`}>
                    ケアプラン項目がありません
                </p>
            ) : (
                <div className="flex flex-col gap-1">
                    {itemRows}
                </div>
            )}
            {canCreate ? (
                <AddForm onSubmit={handleAdd} isSubmitting={createItem.isPending} />
            ) : null}
        </div>
    );
});
