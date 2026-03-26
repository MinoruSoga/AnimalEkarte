// React/Framework
import { memo, useDeferredValue, useMemo, useState } from "react";

// Relative
import { useGetRecordImages } from "../api/get-record-images";
import { ImageGalleryFilter } from "./ImageGalleryFilter";
import { ImageGalleryGroup } from "./ImageGalleryGroup";

interface MedicalRecordImageProps {
  isNewRecord?: boolean;
  medicalRecordId?: string;
}

export const MedicalRecordImage = memo(function MedicalRecordImage({
  isNewRecord = false,
  medicalRecordId,
}: MedicalRecordImageProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearch = useDeferredValue(searchTerm);
  const [dateStart, setDateStart] = useState("");
  const [dateEnd, setDateEnd] = useState("");
  const [sortOrder, setSortOrder] = useState("desc");

  const { data: apiImageGroups = [], isLoading } = useGetRecordImages(
    isNewRecord ? undefined : medicalRecordId,
  );

  const imageGroups = useMemo(
    () =>
      apiImageGroups.filter((g) =>
        deferredSearch
          ? g.images.some((img) => img.name.includes(deferredSearch))
          : true,
      ),
    [apiImageGroups, deferredSearch],
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
      />

      {/* Results Title */}
      <div>
        <h2 className="text-sm font-bold text-[#37352F] pl-1">検査結果</h2>
      </div>

      {/* Image Groups */}
      {isLoading ? (
        <div className="flex items-center justify-center h-24 text-sm text-[#37352F]/40 pl-1">
          読み込み中...
        </div>
      ) : imageGroups.length === 0 ? (
        <div className="flex items-center justify-center h-24 text-sm text-[#37352F]/40 pl-1">
          画像がありません
        </div>
      ) : null}
      <div className="flex flex-col gap-6 pl-1">
        {!isLoading &&
          imageGroups.map((group) => (
            <ImageGalleryGroup key={group.id} group={group} />
          ))}
      </div>
    </div>
  );
});
