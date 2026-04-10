// React/Framework
import { useRef, memo } from "react";

// External
import { Upload } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { NotionDatePicker } from "@/components/shared/NotionDatePicker/NotionDatePicker";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { C, ICON } from "@/lib/design-tokens";

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
    const files = Array.from(e.target.files ?? []);
    if (files.length > 0) {
      onFilesSelected(files);
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
            className={`${C.bgAccent} ${C.bgAccentHover} ${C.textWhite} gap-2 h-10 text-sm shadow-sm border-transparent px-4`}
            onClick={handleUploadClick}
            disabled={isUploading}
          >
            <Upload className={ICON.action} />
            {isUploading ? "アップロード中..." : "画像アップロード"}
          </Button>
        </div>
      ) : null}

      {/* Filters */}
      <div className={`flex items-end gap-4 flex-wrap bg-white p-4 rounded-lg border ${C.borderMedium} shadow-sm`}>
        <div className="flex flex-col gap-1.5 w-[300px]">
          <Label className={`text-sm font-medium ${C.text60}`}>
            検索単語
          </Label>
          <Input
            value={searchTerm}
            onChange={(e) => onSearchChange(e.target.value)}
            className={`bg-white ${C.borderMedium} h-10 text-sm`}
          />
        </div>

        <div className="flex flex-col gap-1.5 w-[400px]">
          <Label className={`text-sm font-medium ${C.text60}`}>
            期間
          </Label>
          <div className="flex items-center gap-2">
            <NotionDatePicker
              value={dateStart}
              onChange={onDateStartChange}
              placeholder="開始日"
              className="flex-1"
            />
            <span className={`${C.text} font-medium text-sm`}>〜</span>
            <NotionDatePicker
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
            className={`h-10 bg-white ${C.text} ${C.borderMedium} ${C.hoverBgPage} text-sm shadow-sm px-3`}
          >
            クリア
          </Button>
          <Button
            type="button"
            variant="outline"
            className={`h-10 ${C.bgAccent} ${C.textWhite} ${C.bgAccentHover} hover:text-white border-transparent text-sm shadow-sm px-3`}
          >
            検索
          </Button>
          <Select value={sortOrder} onValueChange={onSortOrderChange}>
            <SelectTrigger className={`w-[80px] h-10 bg-white ${C.borderMedium} text-sm`}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{SORT_ORDER_SELECT_ITEMS}</SelectContent>
          </Select>
        </div>
      </div>
    </div>
  );
});
