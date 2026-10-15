import { useId, useRef, type ChangeEvent } from "react";
import { Camera, Upload, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { C, ICON } from "@/lib/design-tokens";

// BE の受理形式（jpeg/png/gif）と一致させ、WebP/HEIC 等は picker 段階で選べなくする。
// capture は picker と別 input。同じ input だとファイル選択のヒントを上書きし得る。
const TRIMMING_IMAGE_ACCEPT = "image/jpeg,image/png,image/gif";

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
  const inputId = useId();
  const captureInputRef = useRef<HTMLInputElement>(null);

  const handleCaptureClick = () => {
    captureInputRef.current?.click();
  };

  return (
    <div>
      <Label htmlFor={inputId} className={`${C.text60} mb-2 block`}>
        {label}
      </Label>
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
            aria-label="画像を削除"
            className={`absolute top-1 right-1 min-h-11 min-w-11 flex items-center justify-center ${C.bgWhite} rounded-full shadow-level1 ${C.hoverBgPage}`}
          >
            <X className={`${ICON.action} ${C.text}`} />
          </button>
        </div>
      ) : (
        <label
          htmlFor={inputId}
          className={`flex items-center justify-center w-full h-32 min-h-11 min-w-11 border-2 border-dashed ${C.borderMedium} rounded-md cursor-pointer ${C.hoverBgPage} has-[:focus-visible]:ring-2 ${C.focusRingAccent40}`}
        >
          <div className="flex flex-col items-center">
            <Upload className={`${ICON.lg} ${C.text40} mb-1`} aria-hidden="true" />
            <span className={`text-sm ${C.text60}`}>画像をアップロード</span>
          </div>
          <input
            id={inputId}
            type="file"
            accept={TRIMMING_IMAGE_ACCEPT}
            onChange={onImageChange}
            className="sr-only"
          />
        </label>
      )}
      <input
        ref={captureInputRef}
        type="file"
        accept={TRIMMING_IMAGE_ACCEPT}
        capture="environment"
        onChange={onImageChange}
        // hidden 属性は Tailwind 未適用時もタブ順と a11y tree から外す
        hidden
        className="hidden"
      />
      <Button
        type="button"
        variant="outline"
        className={`mt-2 gap-2 h-11 min-h-11 min-w-11 text-sm shadow-none rounded-md px-4 ${C.borderMedium} ${C.text}`}
        onClick={handleCaptureClick}
      >
        <Camera className={ICON.action} aria-hidden="true" />
        撮影
      </Button>
    </div>
  );
}
