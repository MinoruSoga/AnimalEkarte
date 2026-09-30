// React/Framework
import { memo, useCallback, useDeferredValue, useMemo, useState } from "react";

// Internal
import { normalizedIncludes } from "@/lib/normalize-kana";

// Relative
import { useGetMedicalRecordImages } from "../api/get-medical-record-images";
import { useCreateMedicalRecordImages, useDeleteImage } from "../api/medical-record-images";
import { ImageGalleryFilter } from "./ImageGalleryFilter";
import { ImageGalleryGroup } from "./ImageGalleryGroup";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { usePermission } from "@/hooks/use-permission";
import { C } from "@/lib/design-tokens";

interface MedicalRecordImageProps {
  isNewRecord?: boolean;
  medicalRecordId?: string;
  /** P2-15: 拠点横断で開いたカルテの子リソース操作用。レコード自身の clinicId */
  recordClinicId?: string;
  /** SEC-CS-F14: 死亡ペットでは画像アップロードを UI から無効化する */
  isPetDeceased?: boolean;
}

export const MedicalRecordImage = memo(function MedicalRecordImage({
  isNewRecord = false,
  medicalRecordId,
  recordClinicId,
  isPetDeceased = false,
}: MedicalRecordImageProps) {
  const { canCreate, canDelete } = usePermission("medical-records");
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearch = useDeferredValue(searchTerm);
  const [dateStart, setDateStart] = useState("");
  const [dateEnd, setDateEnd] = useState("");
  const [sortOrder, setSortOrder] = useState("desc");
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<{ id: number; name: string } | null>(null);

  const resolvedId = isNewRecord ? undefined : medicalRecordId;

  const { data: apiImageGroups = [], isLoading } = useGetMedicalRecordImages(
    resolvedId,
    recordClinicId,
  );

  const uploadMutation = useCreateMedicalRecordImages(resolvedId ?? "", recordClinicId);
  const deleteMutation = useDeleteImage(resolvedId ?? "", recordClinicId);

  // EMR-227: 期間・並び順フィルタを接続（以前は state に保持するだけで未適用だった）。
  // group.date は "YYYY/MM/DD HH:MM:SS"（不明時 "-"）なので ISO へ正規化して比較する。
  const imageGroups = useMemo(() => {
    const filtered = apiImageGroups.filter((g) => {
      if (deferredSearch && !g.images.some((img) => normalizedIncludes(img.name, deferredSearch))) {
        return false;
      }
      if (dateStart || dateEnd) {
        const dateKey = g.date.slice(0, 10).replace(/\//g, "-");
        if (!/^\d{4}-\d{2}-\d{2}$/.test(dateKey)) return false;
        if (dateStart && dateKey < dateStart) return false;
        if (dateEnd && dateKey > dateEnd) return false;
      }
      return true;
    });
    const direction = sortOrder === "asc" ? 1 : -1;
    return [...filtered].sort(
      (a, b) =>
        direction *
        a.date
          .slice(0, 10)
          .replace(/\//g, "-")
          .localeCompare(b.date.slice(0, 10).replace(/\//g, "-")),
    );
  }, [apiImageGroups, deferredSearch, dateStart, dateEnd, sortOrder]);

  const canUpload = canCreate && !isPetDeceased;
  // FE12 二重防壁: 死亡ペットでは削除ボタン自体を出さず、callback 側でも拒否する
  const canDeleteImage = canDelete && !isPetDeceased;

  const { mutate: uploadImagesFn } = uploadMutation;
  const handleFilesSelected = useCallback(
    (files: File[]) => {
      if (!canUpload || !resolvedId) return;
      uploadImagesFn(files);
    },
    [canUpload, resolvedId, uploadImagesFn],
  );

  const { mutate: deleteImageFn } = deleteMutation;
  const handleDeleteRequest = useCallback(
    (imageId: number) => {
      if (!canDeleteImage || !resolvedId) return;
      const target = apiImageGroups.flatMap((g) => g.images).find((img) => img.id === imageId);
      setDeleteTarget({ id: imageId, name: target?.name ?? `画像${imageId}` });
    },
    [canDeleteImage, resolvedId, apiImageGroups],
  );
  const handleDeleteConfirm = useCallback(() => {
    if (!canDeleteImage || !resolvedId || !deleteTarget) return;
    const targetId = deleteTarget.id;
    setDeleteTarget(null);
    setDeletingId(targetId);
    deleteImageFn(targetId, {
      onSettled: () => setDeletingId(null),
    });
  }, [canDeleteImage, resolvedId, deleteTarget, deleteImageFn]);
  const handleDeleteCancel = useCallback(() => setDeleteTarget(null), []);

  return (
    <div className="flex flex-col gap-3 flex-1 min-h-0 overflow-y-auto relative pb-20 pr-1">
      {/* Search & Upload Header */}
      <ImageGalleryFilter
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        dateStart={dateStart}
        onDateStartChange={setDateStart}
        dateEnd={dateEnd}
        onDateEndChange={setDateEnd}
        sortOrder={sortOrder}
        onSortOrderChange={setSortOrder}
        isUploading={uploadMutation.isPending}
        onFilesSelected={handleFilesSelected}
        canUpload={canUpload}
      />

      {/* Results Title */}
      <div>
        <h2 className={`text-sm font-bold ${C.text} pl-1`}>画像</h2>
      </div>

      {/* Image Groups */}
      {isLoading ? (
        <div className={`flex items-center justify-center h-24 text-sm ${C.text60} pl-1`}>
          読み込み中...
        </div>
      ) : imageGroups.length === 0 ? (
        <div className={`flex items-center justify-center h-24 text-sm ${C.text60} pl-1`}>
          画像がありません
        </div>
      ) : null}
      <div className="flex flex-col gap-6 pl-1">
        {!isLoading
          ? imageGroups.map((group) => (
              <ImageGalleryGroup
                key={group.id}
                group={group}
                onDeleteImage={resolvedId && canDeleteImage ? handleDeleteRequest : undefined}
                isDeletingId={deletingId}
              />
            ))
          : null}
      </div>

      {/* EMR-227: 画像削除は即時実行せず ConfirmDialog を挟む */}
      <ConfirmDialog
        open={deleteTarget !== null}
        onClose={handleDeleteCancel}
        onConfirm={handleDeleteConfirm}
        title="画像を削除しますか？"
        description={`「${deleteTarget?.name ?? ""}」を削除します。この操作は元に戻せません。`}
        confirmLabel="削除"
        variant="destructive"
      />
    </div>
  );
});
