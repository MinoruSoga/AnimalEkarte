// React/Framework
import { useRef, memo } from "react";

// External
import { Upload } from "lucide-react";
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

  const handleUploadClick = () => {
    fileInputRef.current?.click();
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
        <div className="flex justify-end">
          <input
            ref={fileInputRef}
            type="file"
            accept="image/jpeg,image/png,image/gif,application/pdf"
            multiple
            className="hidden"
            onChange={handleFileChange}
          />
          <Button
            type="button"
            size="sm"
            className={`${C.bgBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} ${C.textOnBrand} gap-2 h-10 text-sm shadow-none rounded-full border-transparent px-4`}
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
            className={`${C.bgWhite} ${C.borderMedium} h-10 text-sm`}
          />
        </div>

        <div className="flex flex-col gap-1.5 w-[400px]">
          <Label className={`text-sm font-medium ${C.text60}`}>期間</Label>
          <div className="flex items-center gap-2">
            <DatePicker
              value={dateStart}
              onChange={onDateStartChange}
              placeholder="開始日"
              className="flex-1"
            />
            <span className={`${C.text} font-medium text-sm`}>〜</span>
            <DatePicker
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
            className={`h-10 ${C.bgWhite} ${C.text} ${C.borderMedium} ${C.hoverBgPage} text-sm px-3`}
          >
            クリア
          </Button>
          <Button
            type="button"
            className={`h-10 ${C.bgBrand} ${C.textOnBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} border-transparent text-sm shadow-none rounded-full px-3`}
          >
            検索
          </Button>
          <Select value={sortOrder} onValueChange={onSortOrderChange}>
            <SelectTrigger className={`w-[80px] h-10 ${C.bgWhite} ${C.borderMedium} text-sm`}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{SORT_ORDER_SELECT_ITEMS}</SelectContent>
          </Select>
        </div>
      </div>
    </div>
  );
});
