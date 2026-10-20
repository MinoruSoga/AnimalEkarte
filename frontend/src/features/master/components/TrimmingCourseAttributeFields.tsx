import { useId, useTransition, useState, type ChangeEvent } from "react";
import { Upload, X } from "lucide-react";

import { FormFieldError } from "@/components/shared/FormFieldError";
import { PropertyRow } from "@/components/shared/SidePeek";
import { Button } from "@/components/ui/button";
import { C, ICON } from "@/lib/design-tokens";

import { uploadTrimmingCourseImage } from "../api/upload-trimming-course-image";

// BE の受理形式（jpeg/png/gif）と一致させる
const COURSE_IMAGE_ACCEPT = "image/jpeg,image/png,image/gif";

/** 0=日〜6=土（BE allowed_weekdays と同じ番号） */
const WEEKDAY_LABELS = ["日", "月", "火", "水", "木", "金", "土"] as const;

interface WeekdayFieldProps {
  value: number[];
  onChange: (days: number[]) => void;
  readOnly?: boolean;
}

export function TrimmingCourseWeekdayField({ value, onChange, readOnly }: WeekdayFieldProps) {
  const toggle = (day: number) => {
    onChange(value.includes(day) ? value.filter((d) => d !== day) : [...value, day]);
  };
  return (
    <PropertyRow
      as="div"
      label="予約可能な曜日"
      description="未選択なら全曜日で選べます。曜日を選んでも休診・勤務・空き枠の条件はそのまま適用されます。"
    >
      <div className="flex flex-wrap items-center gap-1" role="group" aria-label="予約可能な曜日">
        {WEEKDAY_LABELS.map((label, day) => {
          const selected = value.includes(day);
          return (
            <Button
              key={label}
              type="button"
              variant={selected ? "default" : "outline"}
              aria-pressed={selected}
              disabled={readOnly}
              onClick={() => toggle(day)}
              className="min-h-11 min-w-11 px-0 text-sm"
            >
              {label}
            </Button>
          );
        })}
        {value.length === 0 ? <span className={`text-sm ${C.text60}`}>全曜日</span> : null}
      </div>
    </PropertyRow>
  );
}

interface ImageFieldProps {
  currentKey: string;
  previewUrl: string;
  hasImage: boolean;
  onUploaded: (key: string, previewUrl: string) => void;
  onRemove: () => void;
  onUploadingChange: (uploading: boolean) => void;
  readOnly?: boolean;
}

export function TrimmingCourseImageField({
  previewUrl,
  hasImage,
  onUploaded,
  onRemove,
  onUploadingChange,
  readOnly,
}: ImageFieldProps) {
  const inputId = useId();
  const [isUploading, startUpload] = useTransition();
  const [error, setError] = useState("");
  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file || readOnly || isUploading) return;
    setError("");
    onUploadingChange(true);
    startUpload(async () => {
      try {
        const result = await uploadTrimmingCourseImage(file);
        onUploaded(result.key, result.signedUrl);
      } catch {
        // 失敗時は直前の画像を保持する
        setError("画像をアップロードできませんでした。もう一度お試しください。");
      } finally {
        onUploadingChange(false);
      }
    });
  };

  return (
    <PropertyRow as="div" label="画像" description="任意です。未登録でもコースは使えます。">
      <div className="flex flex-col gap-1 w-full">
        {hasImage ? (
          <div className="relative w-32">
            {previewUrl ? (
              <img
                src={previewUrl}
                alt="コース画像"
                className={`w-32 h-24 object-cover rounded-md border ${C.borderPrimary20}`}
              />
            ) : (
              <span className={`text-sm ${C.text60}`}>登録済み（プレビューを表示できません）</span>
            )}
            {readOnly ? null : (
              <button
                type="button"
                onClick={onRemove}
                disabled={isUploading || readOnly}
                aria-label="画像を削除"
                className={`absolute top-1 right-1 min-h-11 min-w-11 flex items-center justify-center ${C.bgWhite} rounded-full shadow-level1 ${C.hoverBgPage}`}
              >
                <X className={`${ICON.action} ${C.text}`} />
              </button>
            )}
          </div>
        ) : null}
        {readOnly ? null : (
          <label
            htmlFor={inputId}
            className={`inline-flex items-center gap-2 min-h-11 px-2 rounded-xxs cursor-pointer text-sm ${C.text60} ${C.hoverBgLight} has-[:focus-visible]:ring-2 ${C.focusRingAccent40}`}
          >
            <Upload className={ICON.action} aria-hidden="true" />
            {isUploading ? "アップロード中…" : hasImage ? "画像を差し替える" : "画像をアップロード"}
            <input
              id={inputId}
              type="file"
              accept={COURSE_IMAGE_ACCEPT}
              onChange={handleChange}
              disabled={isUploading || readOnly}
              className="sr-only"
            />
          </label>
        )}
        <FormFieldError message={error} />
      </div>
    </PropertyRow>
  );
}
