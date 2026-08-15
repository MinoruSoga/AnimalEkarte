// React/Framework
import { useCallback, memo } from "react";

// External
import { Trash2, FileText } from "lucide-react";

// Internal
import { C, ICON } from "@/lib/design-tokens";

interface ImageItem {
  id: number;
  name: string;
  src: string | null;
  label: string;
  mimeType?: string;
}

interface ImageGalleryGroupProps {
  group: {
    id: number;
    date: string;
    images: ImageItem[];
  };
  onDeleteImage?: (imageId: number) => void;
  isDeletingId?: number | null;
}

export const ImageGalleryGroup = memo(function ImageGalleryGroup({
  group,
  onDeleteImage,
  isDeletingId,
}: ImageGalleryGroupProps) {
  const handleDeleteClick = useCallback(
    (e: React.MouseEvent, imageId: number) => {
      e.stopPropagation();
      onDeleteImage?.(imageId);
    },
    [onDeleteImage],
  );

  const handleImageClick = useCallback((src: string | null) => {
    if (src) {
      window.open(src, "_blank", "noopener,noreferrer");
    }
  }, []);

  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-col gap-0 w-full">
        <h3 className={`text-sm font-bold font-mono mb-1 ${C.text}`}>
          {group.date}
        </h3>
        <div className={`h-[1px] w-full ${C.bgLight}`} />
      </div>

      <div className="flex flex-wrap gap-4 mt-2">
        {group.images.map((img) => {
          const isPdf =
            img.mimeType === "application/pdf" ||
            img.name.toLowerCase().endsWith(".pdf");
          const isDeleting = isDeletingId === img.id;

          return (
            <div
              key={img.id}
              role="button"
              tabIndex={0}
              aria-label={`${img.name}を開く`}
              className="flex flex-col gap-1 w-[160px] group relative cursor-pointer"
              onClick={() => handleImageClick(img.src)}
              onKeyDown={(e) => {
                if (e.target !== e.currentTarget) return;
                if (e.key === "Enter" || e.key === " ") {
                  e.preventDefault();
                  handleImageClick(img.src);
                }
              }}
            >
              <div
                className={`h-[160px] w-full ${C.bgPage} border ${C.borderMedium} flex items-center justify-center rounded-sm ${C.hoverBorderMedium40} transition-colors overflow-hidden relative`}
              >
                {isPdf ? (
                  <div className="flex flex-col items-center gap-2">
                    <FileText className={`${ICON.xl} ${C.text50}`} />
                    <span className={`text-xs ${C.text50}`}>PDF</span>
                  </div>
                ) : img.src ? (
                  <img
                    src={img.src}
                    alt={img.name}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <span className={`text-sm font-medium ${C.text50}`}>
                    {img.label}
                  </span>
                )}

                {/* Delete button — visible on hover */}
                {onDeleteImage ? (
                  <button
                    type="button"
                    aria-label={`${img.name}を削除`}
                    onClick={(e) => handleDeleteClick(e, img.id)}
                    disabled={isDeleting}
                    className={`absolute top-1 right-1 p-1 rounded opacity-0 group-hover:opacity-100 transition-opacity ${C.bgWhite} border ${C.borderDanger20} ${C.hoverBgDanger5} disabled:opacity-50`}
                  >
                    <Trash2 className={`${ICON.smXs} ${C.danger}`} />
                  </button>
                ) : null}

                {isDeleting ? (
                  <div className={`absolute inset-0 flex items-center justify-center ${C.bgWhite}/70`}>
                    <span className={`text-xs ${C.text50}`}>削除中...</span>
                  </div>
                ) : null}
              </div>
              <p
                className={`text-sm font-medium truncate transition-colors ${C.text} ${C.hoverTextBrand}`}
              >
                {img.name}
              </p>
            </div>
          );
        })}
      </div>
    </div>
  );
});
