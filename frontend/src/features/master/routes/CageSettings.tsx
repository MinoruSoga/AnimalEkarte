import { useCallback, useLayoutEffect, useMemo, useRef, useState } from "react";
import { toast } from "sonner";
import { useSortableList } from "@/hooks/use-sortable-list";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { useMergeCages, useUnmergeCages } from "@/hooks/use-cage-merge";
import { useCageOccupancy } from "@/hooks/use-cage-occupancy";
import { Building2 } from "lucide-react";
import { ErrorFallback, LoadingFallback } from "@/components/shared/DataStates";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { C, ICON } from "@/lib/design-tokens";
import { resolveCageNameById } from "@/lib/cage-name";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import {
  useGetAllCages,
  useCreateCage,
  useUpdateCage,
  useDeleteCage,
  useReorderCages,
} from "../api/cages";
import type { Cage, CreateCageRequest, UpdateCageRequest } from "../api/cages";
import { CageSidePanel } from "../components/CageSidePanel";
import { CageSortableTable } from "../components/CageSortableTable";
import type { CageFormData } from "../lib/cage-side-panel-model";
import { buildCageCreateRequest, buildCageUpdateRequest } from "./cage-settings-model";
import { ResourceMasterHospitalization } from "@/types/generated/models";
import { usePermission } from "@/hooks/use-permission";

// ─── EMR-306: 連結代表ケージ選択ダイアログ ───
// ConfirmDialog は description に string しか受けられずフォーム部品を載せられない
// ため、代表ケージ選択を持つこの確認ダイアログは ui/alert-dialog + ui/radio-group
// から feature ローカルに組み立てる（shared コンポーネントは新設しない）。
interface CageMergeDialogProps {
  open: boolean;
  /** 連結対象として選択中のケージ（代表候補のラジオ選択肢になる） */
  cages: Cage[];
  representativeId: string | null;
  onRepresentativeChange: (id: string) => void;
  isPending: boolean;
  onClose: () => void;
  onConfirm: () => void;
}

function CageMergeDialog({
  open,
  cages,
  representativeId,
  onRepresentativeChange,
  isPending,
  onClose,
  onConfirm,
}: CageMergeDialogProps) {
  return (
    <AlertDialog
      open={open}
      onOpenChange={(nextOpen) => {
        if (!nextOpen) onClose();
      }}
    >
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>ケージを連結しますか？</AlertDialogTitle>
          <AlertDialogDescription>
            代表ケージを選択してください。選択した他のケージは代表ケージに連結されます。
          </AlertDialogDescription>
        </AlertDialogHeader>
        <RadioGroup
          value={representativeId ?? ""}
          onValueChange={onRepresentativeChange}
          aria-label="代表ケージ"
        >
          {cages.map((cage) => (
            <div key={cage.id} className="flex items-center gap-2">
              <RadioGroupItem id={`cage-merge-rep-${cage.id}`} value={cage.id} />
              <Label htmlFor={`cage-merge-rep-${cage.id}`} className="mb-0 font-normal">
                {cage.name}
              </Label>
            </div>
          ))}
        </RadioGroup>
        <AlertDialogFooter>
          <AlertDialogCancel>キャンセル</AlertDialogCancel>
          <AlertDialogAction onClick={onConfirm} disabled={!representativeId || isPending}>
            連結する
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

// ─── Page ───
export function CageSettings() {
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterHospitalization);
  const { data, isPending, isError } = useGetAllCages();
  const createMutation = useCreateCage();
  const updateMutation = useUpdateCage();
  const deleteMutation = useDeleteCage();
  const reorderMutation = useReorderCages();
  const mergeMutation = useMergeCages();
  const unmergeMutation = useUnmergeCages();

  // EMR-306 (Codex review P2): 連結ピッカーの占有無効化。spec「選択できるのは
  // 未連結・有効・占有中でないケージのみ」に合わせる。canEdit でない閲覧者には
  // チェックボックス自体が disabled なので占有クエリは発行しない。
  // undefined（読込中/エラー/打ち切り）は table 側で fail-closed に選択不可扱い。
  const { occupiedCageIds } = useCageOccupancy({ enabled: canEdit === true });

  // EMR-306: 連結選択状態はページローカル。use-editable-id-selection は
  // フォーム dirty 連動のためここでは素の Set state を使う。
  const [selectedCageIds, setSelectedCageIds] = useState<ReadonlySet<string>>(new Set());
  const [mergeDialogOpen, setMergeDialogOpen] = useState(false);
  const [representativeId, setRepresentativeId] = useState<string | null>(null);
  const [unmergeRepresentativeId, setUnmergeRepresentativeId] = useState<string | null>(null);

  // 臨床安全境界: 権限は UI 表示だけでなく mutation 直前にも再検査する
  // （use-master-crud の permissionsRef と同型）。
  const canEditRef = useRef(canEdit === true);
  useLayoutEffect(() => {
    canEditRef.current = canEdit === true;
  }, [canEdit]);

  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<Cage>({
    data,
    deleteMutation,
    entityLabel: "ケージ",
    dirtyGuard: dirty,
    permissions: { canDelete },
  });
  const handleDirtyChange = useCallback(
    (d: boolean) => {
      if (d) dirty.markDirty();
      else dirty.markClean();
    },
    [dirty],
  );

  const {
    orderedItems: sortedCages,
    sensors,
    activeId,
    handleDragStart,
    handleDragCancel,
    handleDragEnd,
    resetOrder,
  } = useSortableList({
    items: crud.filteredItems,
    onReorder: (newIds) => {
      if (!canEdit) return;
      reorderMutation.mutate({ ids: newIds.map(Number) }, { onSuccess: resetOrder });
    },
  });

  const { handleSave } = useMasterSave<Cage, CageFormData, CreateCageRequest, UpdateCageRequest>({
    crud,
    createMutation,
    updateMutation,
    validate: (d) => (!d.name.trim() ? "ケージ名は必須です" : null),
    toCreateRequest: buildCageCreateRequest,
    toUpdateRequest: buildCageUpdateRequest,
    permissions: { canCreate, canEdit },
  });

  // 選択後に占有と判明したケージ（入院登録との競合）は表示・送信から除外する —
  // チェックボックスの disabled と合わせて二重に防ぐ。effect による state 書き戻しは
  // react-hooks/set-state-in-effect 禁止のため、レンダー中の派生集合で表現する。
  const effectiveSelectedCageIds = useMemo(() => {
    if (occupiedCageIds === undefined) return selectedCageIds;
    if ([...selectedCageIds].every((id) => !occupiedCageIds.has(id))) {
      return selectedCageIds;
    }
    return new Set([...selectedCageIds].filter((id) => !occupiedCageIds.has(id)));
  }, [selectedCageIds, occupiedCageIds]);

  const handleToggleCageSelect = useCallback((id: string, checked: boolean) => {
    setSelectedCageIds((prev) => {
      const next = new Set(prev);
      if (checked) {
        next.add(id);
      } else {
        next.delete(id);
      }
      return next;
    });
  }, []);

  const selectedCages = useMemo(
    () => (data ?? []).filter((cage) => effectiveSelectedCageIds.has(cage.id)),
    [data, effectiveSelectedCageIds],
  );

  const handleOpenMergeDialog = useCallback(() => {
    setRepresentativeId(selectedCages[0]?.id ?? null);
    setMergeDialogOpen(true);
  }, [selectedCages]);

  const handleMergeConfirm = useCallback(() => {
    if (canEditRef.current !== true) return;
    if (representativeId === null) return;
    const memberCageIds = selectedCages
      .filter((cage) => cage.id !== representativeId)
      .map((cage) => Number(cage.id));
    if (memberCageIds.length === 0) return;
    mergeMutation.mutate(
      { representativeId, memberCageIds },
      {
        onSuccess: () => {
          setMergeDialogOpen(false);
          setSelectedCageIds(new Set());
          toast.success("ケージを連結しました");
        },
      },
    );
  }, [representativeId, selectedCages, mergeMutation]);

  const handleRequestUnmerge = useCallback((repId: string) => {
    setUnmergeRepresentativeId(repId);
  }, []);

  const unmergeTargetName = useMemo(() => {
    if (unmergeRepresentativeId === null) return "";
    return resolveCageNameById(data ?? [], unmergeRepresentativeId);
  }, [data, unmergeRepresentativeId]);

  const handleUnmergeConfirm = useCallback(() => {
    if (canEditRef.current !== true) return;
    if (unmergeRepresentativeId === null) return;
    unmergeMutation.mutate(unmergeRepresentativeId, {
      onSuccess: () => {
        setUnmergeRepresentativeId(null);
        toast.success("ケージの連結を解除しました");
      },
    });
  }, [unmergeRepresentativeId, unmergeMutation]);

  return (
    <>
      <MasterCRUDPage
        title="ケージマスタ"
        icon={<Building2 className={`${ICON.page} ${C.text}`} />}
        resource={ResourceMasterHospitalization}
        entityLabel="ケージ"
        searchPlaceholder="ケージ名で検索..."
        emptyMessage="ケージが登録されていません"
        crud={crud}
        handleSave={handleSave}
        filterProperties={[MASTER_STATUS_FILTER]}
        columns={[]}
        renderRow={() => null}
        renderSidePanel={({ readOnly, ...props }) => (
          <CageSidePanel
            key={props.item?.id ?? "new"}
            {...props}
            cages={data ?? []}
            readOnly={readOnly}
            onDirtyChange={handleDirtyChange}
          />
        )}
      >
        {isError ? (
          <ErrorFallback message="ケージの取得に失敗しました" />
        ) : isPending ? (
          <LoadingFallback />
        ) : (
          <>
            {canEdit ? (
              <div className="flex items-center justify-end gap-3">
                <span className={`text-sm ${C.text65}`}>
                  {effectiveSelectedCageIds.size > 0
                    ? `${effectiveSelectedCageIds.size}件選択中`
                    : "連結するケージを2件以上選択"}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={effectiveSelectedCageIds.size < 2}
                  onClick={handleOpenMergeDialog}
                >
                  選択したケージを連結
                </Button>
              </div>
            ) : null}
            <CageSortableTable
              items={sortedCages}
              sensors={sensors}
              activeId={activeId}
              onDragStart={handleDragStart}
              onDragEnd={handleDragEnd}
              onDragCancel={handleDragCancel}
              canEdit={canEdit}
              onEdit={crud.handleEdit}
              selectedIds={effectiveSelectedCageIds}
              onToggleSelect={handleToggleCageSelect}
              occupiedCageIds={occupiedCageIds}
              onRequestUnmerge={handleRequestUnmerge}
            />
          </>
        )}
      </MasterCRUDPage>
      <CageMergeDialog
        open={mergeDialogOpen}
        cages={selectedCages}
        representativeId={representativeId}
        onRepresentativeChange={setRepresentativeId}
        isPending={mergeMutation.isPending}
        onClose={() => setMergeDialogOpen(false)}
        onConfirm={handleMergeConfirm}
      />
      <ConfirmDialog
        open={unmergeRepresentativeId !== null}
        onClose={() => setUnmergeRepresentativeId(null)}
        title="ケージの連結を解除しますか？"
        description={`「${unmergeTargetName}」の連結を解除します。構成ケージはそれぞれ独立したケージに戻ります。`}
        confirmLabel="解除"
        variant="destructive"
        isPending={unmergeMutation.isPending}
        onConfirm={handleUnmergeConfirm}
      />
      {dirty.discardDialog}
    </>
  );
}
