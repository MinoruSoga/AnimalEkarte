import React, { useState } from "react";
import { ImageGalleryFilter } from "./ImageGalleryFilter";
import { ImageGalleryGroup } from "./ImageGalleryGroup";

export function MedicalRecordImage({ isNewRecord = false }: { isNewRecord?: boolean }) {
  const [searchTerm, setSearchTerm] = useState("");
  const [dateStart, setDateStart] = useState("");
  const [dateEnd, setDateEnd] = useState("");
  const [sortOrder, setSortOrder] = useState("desc");

  // Mock data for image groups
  const imageGroups = isNewRecord ? [] : [
    {
      id: 1,
      date: "2025/10/10 10:10:10",
      images: [
        { id: 1, name: "画像名", src: null, label: "画像1" },
        { id: 2, name: "画像名", src: null, label: "画像2" },
        { id: 3, name: "画像名", src: null, label: "画像3" },
      ],
    },
    {
      id: 2,
      date: "2025/11/11 10:10:10",
      images: [
        { id: 4, name: "画像名", src: null, label: "画像4" },
        { id: 5, name: "画像名", src: null, label: "画像5" },
      ],
    },
  ];

  return (
    <div className="h-[calc(100vh-220px)] min-h-[500px] flex flex-col gap-3 overflow-y-auto pb-20 pr-1">
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
      />

      {/* Results Title */}
      <div>
         <h2 className="text-sm font-bold text-[#37352F] pl-1">検査結果</h2>
      </div>

      {/* Image Groups */}
      <div className="flex flex-col gap-6 pl-1">
        {imageGroups.map((group) => (
          <ImageGalleryGroup key={group.id} group={group} />
        ))}
      </div>
    </div>
  );
}
