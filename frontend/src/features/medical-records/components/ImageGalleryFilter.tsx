// React/Framework
import React, { useRef } from "react";

// External
import { Upload } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { NotionDatePicker } from "@/components/shared/NotionDatePicker/NotionDatePicker";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ICON } from "@/lib/design-tokens";

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
}

export const ImageGalleryFilter = React.memo(function ImageGalleryFilter({
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
          size="sm"
          className="bg-[#2383E2] hover:bg-[#1B6EC2] text-white gap-2 h-10 text-sm shadow-sm border-transparent px-4"
          onClick={handleUploadClick}
          disabled={isUploading}
        >
          <Upload className={ICON.action} />
          {isUploading ? "アップロード中..." : "画像アップロード"}
        </Button>
      </div>

      {/* Filters */}
      <div className="flex items-end gap-4 flex-wrap bg-white p-4 rounded-lg border border-[rgba(55,53,47,0.16)] shadow-sm">
        <div className="flex flex-col gap-1.5 w-[300px]">
          <Label className="text-sm font-medium text-[#37352F]/60">
            検索単語
          </Label>
          <Input
            value={searchTerm}
            onChange={(e) => onSearchChange(e.target.value)}
            className="bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm"
          />
        </div>

        <div className="flex flex-col gap-1.5 w-[400px]">
          <Label className="text-sm font-medium text-[#37352F]/60">
            期間
          </Label>
          <div className="flex items-center gap-2">
            <NotionDatePicker
              value={dateStart}
              onChange={onDateStartChange}
              placeholder="開始日"
              className="flex-1"
            />
            <span className="text-[#37352F] font-medium text-sm">〜</span>
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
            variant="outline"
            className="h-10 bg-[#37352F] text-white hover:bg-[#37352F]/90 hover:text-white border-transparent text-sm shadow-sm px-3"
          >
            クリア
          </Button>
          <Button
            variant="outline"
            className="h-10 bg-[#37352F] text-white hover:bg-[#37352F]/90 hover:text-white border-transparent text-sm shadow-sm px-3"
          >
            検索
          </Button>
          <Select value={sortOrder} onValueChange={onSortOrderChange}>
            <SelectTrigger className="w-[80px] h-10 bg-white border-[rgba(55,53,47,0.16)] text-sm">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="desc">降順</SelectItem>
              <SelectItem value="asc">昇順</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>
    </div>
  );
});
