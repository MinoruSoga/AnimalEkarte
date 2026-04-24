// React/Framework
import { memo, useCallback, useDeferredValue, useMemo, useState } from "react";

// Relative
import { useGetMedicalRecordImages } from "../api/get-medical-record-images";
import { useCreateMedicalRecordImages, useDeleteImage } from "../api/medical-record-images";
import { ImageGalleryFilter } from "./ImageGalleryFilter";
import { ImageGalleryGroup } from "./ImageGalleryGroup";
import { usePermission } from "@/hooks/use-permission";
import { C } from "@/lib/design-tokens";

interface MedicalRecordImageProps {
  isNewRecord?: boolean;
  medicalRecordId?: string;
}

export const MedicalRecordImage = memo(function MedicalRecordImage({
  isNewRecord = false,
  medicalRecordId,
}: MedicalRecordImageProps) {
  const { canCreate, canDelete } = usePermission("medical-records");
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearch = useDeferredValue(searchTerm);
  const [dateStart, setDateStart] = useState("");
  const [dateEnd, setDateEnd] = useState("");
  const [sortOrder, setSortOrder] = useState("desc");
  const [deletingId, setDeletingId] = useState<number | null>(null);

  const resolvedId = isNewRecord ? undefined : medicalRecordId;

  const { data: apiImageGroups = [], isLoading } = useGetMedicalRecordImages(resolvedId);

  const uploadMutation = useCreateMedicalRecordImages(resolvedId ?? "");
  const deleteMutation = useDeleteImage(resolvedId ?? "");

  const imageGroups = useMemo(
    () =>
      apiImageGroups.filter((g) =>
        deferredSearch
          ? g.images.some((img) => img.name.includes(deferredSearch))
          : true,
      ),
    [apiImageGroups, deferredSearch],
  );

  const handleFilesSelected = useCallback(
    (files: File[]) => {
      if (!canCreate || !resolvedId) return;
      uploadMutation.mutate(files);
    },
    [canCreate, resolvedId, uploadMutation],
  );

  const { mutate: deleteImageFn } = deleteMutation;
  const handleDeleteImage = useCallback(
    (imageId: number) => {
      if (!canDelete || !resolvedId) return;
      setDeletingId(imageId);
      deleteImageFn(imageId, {
        onSettled: () => setDeletingId(null),
      });
    },
    [canDelete, resolvedId, deleteImageFn],
  );

  return (
    <div className="h-[calc(100vh-220px)] min-h-[500px] flex flex-col gap-3 overflow-y-auto pb-20 pr-1">
      {/* Search & Upload Header */}
      <ImageGalleryFilter
        searchTerm={deferredSearch}
        onSearchChange={setSearchTerm}
        dateStart={dateStart}
        onDateStartChange={setDateStart}
        dateEnd={dateEnd}
        onDateEndChange={setDateEnd}
        sortOrder={sortOrder}
        onSortOrderChange={setSortOrder}
        isUploading={uploadMutation.isPending}
        onFilesSelected={handleFilesSelected}
        canUpload={canCreate}
      />

      {/* Results Title */}
      <div>
        <h2 className={`text-sm font-bold ${C.text} pl-1`}>検査結果</h2>
      </div>

      {/* Image Groups */}
      {isLoading ? (
        <div className={`flex items-center justify-center h-24 text-sm ${C.text40} pl-1`}>
          読み込み中...
        </div>
      ) : imageGroups.length === 0 ? (
        <div className={`flex items-center justify-center h-24 text-sm ${C.text40} pl-1`}>
          画像がありません
        </div>
      ) : null}
      <div className="flex flex-col gap-6 pl-1">
        {!isLoading
          ? imageGroups.map((group) => (
              <ImageGalleryGroup
                key={group.id}
                group={group}
                onDeleteImage={resolvedId && canDelete ? handleDeleteImage : undefined}
                isDeletingId={deletingId}
              />
            ))
          : null}
      </div>
    </div>
  );
});
