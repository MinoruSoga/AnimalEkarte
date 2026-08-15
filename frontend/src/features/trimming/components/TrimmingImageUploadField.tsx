import type { ChangeEvent } from "react";
import { Upload, X } from "lucide-react";

import { Label } from "@/components/ui/label";
import { C, ICON } from "@/lib/design-tokens";

interface TrimmingImageUploadFieldProps {
  label: string;
  preview: string | null;
  previewAlt: string;
  onImageChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onRemoveImage: () => void;
}

export function TrimmingImageUploadField({
  label,
  preview,
  previewAlt,
  onImageChange,
  onRemoveImage,
}: TrimmingImageUploadFieldProps) {
  return (
    <div>
      <Label className={`text-sm ${C.text60} mb-2 block`}>{label}</Label>
      {preview ? (
        <div className="relative">
          <img
            src={preview}
            alt={previewAlt}
            className={`w-full h-32 object-cover rounded-md border ${C.borderPrimary20}`}
          />
          <button
            type="button"
            onClick={onRemoveImage}
            className={`absolute top-1 right-1 p-1 ${C.bgWhite} rounded-full shadow-level1 ${C.hoverBgPage}`}
          >
            <X className={`${ICON.action} ${C.text}`} />
          </button>
        </div>
      ) : (
        <label
          className={`flex items-center justify-center w-full h-32 border-2 border-dashed ${C.borderMedium} rounded-md cursor-pointer ${C.hoverBgPage}`}
        >
          <div className="flex flex-col items-center">
            <Upload className={`${ICON.lg} ${C.text40} mb-1`} />
            <span className={`text-sm ${C.text60}`}>画像をアップロード</span>
          </div>
          <input type="file" accept="image/*" onChange={onImageChange} className="hidden" />
        </label>
      )}
    </div>
  );
}
