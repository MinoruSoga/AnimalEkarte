import React from "react";

interface ImageItem {
  id: number;
  name: string;
  src: string | null;
  label: string;
}

interface ImageGalleryGroupProps {
  group: {
    id: number;
    date: string;
    images: ImageItem[];
  };
}

export const ImageGalleryGroup = React.memo(function ImageGalleryGroup({
  group,
}: ImageGalleryGroupProps) {
  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-col gap-0 w-full">
        <h3 className="text-sm font-bold text-[#37352F] mb-1 font-mono">
          {group.date}
        </h3>
        <div className="h-[1px] w-full bg-[#37352F]/10" />
      </div>

      <div className="flex flex-wrap gap-4 mt-2">
        {group.images.map((img) => (
          <div
            key={img.id}
            className="flex flex-col gap-1 w-[160px] group cursor-pointer"
          >
            <div className="h-[160px] w-full bg-[#F7F6F3] border border-[rgba(55,53,47,0.16)] flex items-center justify-center rounded-sm hover:border-[#37352F]/40 transition-colors">
              <span className="text-sm font-medium text-[#37352F]/60">
                {img.label}
              </span>
            </div>
            <p className="text-sm font-medium text-[#37352F] truncate group-hover:text-blue-600 transition-colors">
              {img.name}
            </p>
          </div>
        ))}
      </div>
    </div>
  );
});
