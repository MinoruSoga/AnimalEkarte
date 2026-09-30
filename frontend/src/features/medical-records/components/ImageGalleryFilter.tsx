// React/Framework
import { useId, useRef, memo } from "react";

// External
import { Camera, Upload } from "lucide-react";
import { toast } from "sonner";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { DatePicker } from "@/components/shared/DatePicker/DatePicker";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { C, ICON } from "@/lib/design-tokens";

/** 1ファイルあたりの上限（MB） */
const MAX_FILE_SIZE_MB = 10;
export const MAX_FILE_SIZE_BYTES = MAX_FILE_SIZE_MB * 1024 * 1024;
/** SEC-CS-F08: 1回の選択で受け付ける最大ファイル数 */
export const MAX_UPLOAD_FILES = 10;
/** SEC-CS-F08: 1回の選択で受け付ける合計バイト上限（50MiB） */
export const MAX_UPLOAD_BATCH_BYTES = 50 * 1024 * 1024;

// rendering-hoist-jsx: 静的 SelectItem JSX をモジュール定数に巻き上げ
const SORT_ORDER_SELECT_ITEMS = (
  <>
    <SelectItem value="desc">降順</SelectItem>
    <SelectItem value="asc">昇順</SelectItem>
  </>
);

interface ImageGalleryFilterProps {
  searchTerm: string;
  onSearchChange: (value: string) => void;
  dateStart: string;
  onDateStartChange: (value: string) => void;
  dateEnd: string;
  onDateEndChange: (value: string) => void;
  sortOrder: string;
  onSortOrderChange: (value: string) => void;
  isUploading?: boolean;
  onFilesSelected: (files: File[]) => void;
  canUpload?: boolean;
}

export const ImageGalleryFilter = memo(function ImageGalleryFilter({
  searchTerm,
  onSearchChange,
  dateStart,
  onDateStartChange,
  dateEnd,
  onDateEndChange,
  sortOrder,
  onSortOrderChange,
  isUploading = false,
  onFilesSelected,
  canUpload = true,
}: ImageGalleryFilterProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const captureInputRef = useRef<HTMLInputElement>(null);
  const dateRangeId = useId();
  const startDateId = `${dateRangeId}-start-date`;
  const endDateId = `${dateRangeId}-end-date`;

  const handleClearFilters = () => {
    onSearchChange("");
    onDateStartChange("");
    onDateEndChange("");
    onSortOrderChange("desc");
  };

  const handleUploadClick = () => {
    fileInputRef.current?.click();
  };

  const handleCaptureClick = () => {
    captureInputRef.current?.click();
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const allFiles = Array.from(e.target.files ?? []);
    // SEC-CS-F08: 件数・合計バイトを onFilesSelected 前に fail-closed で拒否
    if (allFiles.length > MAX_UPLOAD_FILES) {
      toast.error(`一度にアップロードできるファイルは${MAX_UPLOAD_FILES}件までです`);
      e.target.value = "";
      return;
    }
    const totalBytes = allFiles.reduce((sum, f) => sum + f.size, 0);
    if (totalBytes > MAX_UPLOAD_BATCH_BYTES) {
      toast.error("合計ファイルサイズが上限（50MB）を超えています");
      e.target.value = "";
      return;
    }
    // SEC-CS-F08-R1: any oversized file rejects the whole batch (no partial upload).
    const oversized = allFiles.filter((f) => f.size > MAX_FILE_SIZE_BYTES);
    if (oversized.length > 0) {
      toast.error(
        `ファイルサイズが上限（${MAX_FILE_SIZE_MB}MB）を超えています: ${oversized.map((f) => f.name).join(", ")}`,
      );
      e.target.value = "";
      return;
    }
    if (allFiles.length > 0) {
      onFilesSelected(allFiles);
    }
    // Reset input so the same file can be re-selected
    e.target.value = "";
  };

  return (
    <div className="flex flex-col gap-3">
      {canUpload ? (
        <div className="flex justify-end gap-2">
          <input
            ref={fileInputRef}
            type="file"
            accept="image/jpeg,image/png,image/gif,application/pdf"
            multiple
            className="hidden"
            onChange={handleFileChange}
          />
          <input
            ref={captureInputRef}
            type="file"
            accept="image/jpeg,image/png,image/gif"
            capture="environment"
            className="hidden"
            onChange={handleFileChange}
          />
          <Button
            type="button"
            size="sm"
            variant="outline"
            className={`gap-2 h-11 text-sm shadow-none rounded-md px-4 ${C.borderMedium} ${C.text}`}
            onClick={handleCaptureClick}
            disabled={isUploading}
          >
            <Camera className={ICON.action} />
            撮影
          </Button>
          <Button
            type="button"
            size="sm"
            className={`${C.bgActionPrimarySolid} ${C.textOnActionPrimary} ${C.hoverBgActionPrimarySolid} ${C.hoverTextOnActionPrimary} ${C.activeBgActionPrimarySolid} ${C.activeTextOnActionPrimary} gap-2 h-11 text-sm shadow-none rounded-full border-transparent px-4`}
            onClick={handleUploadClick}
            disabled={isUploading}
          >
            <Upload className={ICON.action} />
            {isUploading ? "アップロード中..." : "画像アップロード"}
          </Button>
        </div>
      ) : null}

      {/* Filters */}
      <div
        className={`flex items-end gap-4 flex-wrap ${C.bgWhite} p-4 rounded-lg border ${C.borderMedium}`}
      >
        <div className="flex flex-col gap-1.5 w-[300px]">
          <Label htmlFor="image-gallery-search" className={`text-sm font-medium ${C.text60}`}>
            検索単語
          </Label>
          <Input
            id="image-gallery-search"
            value={searchTerm}
            onChange={(e) => onSearchChange(e.target.value)}
            className={`${C.bgWhite} ${C.borderMedium} text-sm`}
          />
        </div>

        <div className="flex flex-col gap-1.5 w-[400px]">
          {/* 期間は2入力のグループ名のため span + 各ピッカーへ sr-only Label */}
          <span className={`text-sm font-medium ${C.text60}`}>期間</span>
          <div className="flex items-center gap-2">
            <Label htmlFor={startDateId} className="sr-only">
              開始日
            </Label>
            <DatePicker
              id={startDateId}
              value={dateStart}
              onChange={onDateStartChange}
              placeholder="開始日"
              className="flex-1"
            />
            <span className={`${C.text} font-medium text-sm`}>〜</span>
            <Label htmlFor={endDateId} className="sr-only">
              終了日
            </Label>
            <DatePicker
              id={endDateId}
              value={dateEnd}
              onChange={onDateEndChange}
              placeholder="終了日"
              className="flex-1"
            />
          </div>
        </div>

        <div className="flex items-end gap-2 pb-[1px]">
          <Button
            type="button"
            variant="outline"
            onClick={handleClearFilters}
            className={`h-11 ${C.bgWhite} ${C.text} ${C.borderMedium} ${C.hoverBgPage} text-sm px-3`}
          >
            クリア
          </Button>
          <Select value={sortOrder} onValueChange={onSortOrderChange}>
            <SelectTrigger
              aria-label="並び順"
              className={`w-[80px] h-11 ${C.bgWhite} ${C.borderMedium} text-sm focus-visible:ring-2 ${C.focusVisibleRingActionPrimary} focus-visible:ring-offset-1`}
            >
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{SORT_ORDER_SELECT_ITEMS}</SelectContent>
          </Select>
        </div>
      </div>
    </div>
  );
});
