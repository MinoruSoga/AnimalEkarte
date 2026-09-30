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
    (imageId: number) => {
      onDeleteImage?.(imageId);
    },
    [onDeleteImage],
  );

  const renderThumb = (img: ImageItem) => {
    const isPdf = img.mimeType === "application/pdf" || img.name.toLowerCase().endsWith(".pdf");
    return (
      <div
        className={`h-[160px] w-full ${C.bgPage} border ${C.borderMedium} flex items-center justify-center rounded-lg ${C.hoverBorderMedium40} transition-colors overflow-hidden`}
      >
        {isPdf ? (
          <div className="flex flex-col items-center gap-2">
            <FileText className={`${ICON.xl} ${C.text50}`} />
            <span className={`text-xs ${C.text50}`}>PDF</span>
          </div>
        ) : img.src ? (
          <img src={img.src} alt={img.name} className="w-full h-full object-cover" />
        ) : (
          <span className={`text-sm font-medium ${C.text50}`}>{img.label}</span>
        )}
      </div>
    );
  };

  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-col gap-0 w-full">
        <h3 className={`text-sm font-bold font-mono mb-1 ${C.text}`}>{group.date}</h3>
        <div className={`h-[1px] w-full ${C.bgLight}`} />
      </div>

      <div className="flex flex-wrap gap-4 mt-2">
        {group.images.map((img) => {
          const isDeleting = isDeletingId === img.id;

          return (
            <div key={img.id} className="flex flex-col gap-1 w-[160px] group relative">
              {/* EMR-227: カードのクリックはネイティブリンク化。削除ボタンはリンク内にネストせず兄弟配置 */}
              {img.src ? (
                <a
                  href={img.src}
                  target="_blank"
                  rel="noopener noreferrer"
                  aria-label={`${img.name}を開く`}
                  className={`flex flex-col gap-1 rounded-lg focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-1 ${C.focusVisibleRingActionPrimary}`}
                >
                  {renderThumb(img)}
                  <p
                    className={`text-sm font-medium truncate transition-colors ${C.textActionPrimary}`}
                  >
                    {img.name}
                  </p>
                </a>
              ) : (
                <>
                  {renderThumb(img)}
                  <p className={`text-sm font-medium truncate ${C.text}`}>{img.name}</p>
                </>
              )}

              {/* Delete button — visible on hover/focus */}
              {onDeleteImage ? (
                <button
                  type="button"
                  aria-label={`${img.name}を削除`}
                  onClick={() => handleDeleteClick(img.id)}
                  disabled={isDeleting}
                  className={`absolute top-1 right-1 min-h-11 min-w-11 flex items-center justify-center rounded opacity-0 group-hover:opacity-100 focus-visible:opacity-100 group-focus-within:opacity-100 transition-opacity ${C.bgWhite} border ${C.borderDanger20} ${C.hoverBgDanger5} disabled:opacity-50`}
                >
                  <Trash2 className={`${ICON.smXs} ${C.danger}`} />
                </button>
              ) : null}

              {isDeleting ? (
                <div
                  className={`absolute inset-x-0 top-0 h-[160px] flex items-center justify-center rounded-lg ${C.bgWhite60}`}
                >
                  <span className={`text-xs ${C.text50}`}>削除中...</span>
                </div>
              ) : null}
            </div>
          );
        })}
      </div>
    </div>
  );
});
