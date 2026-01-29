// React/Framework
import React from "react";

// External
import { Upload } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface ImageGalleryFilterProps {
  searchTerm: string;
  onSearchChange: (value: string) => void;
  dateStart: string;
  onDateStartChange: (value: string) => void;
  dateEnd: string;
  onDateEndChange: (value: string) => void;
  sortOrder: string;
  onSortOrderChange: (value: string) => void;
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
}: ImageGalleryFilterProps) {
  return (
    <div className="flex flex-col gap-3">
      <div className="flex justify-end">
        <Button
          size="sm"
          className="bg-[#37352F] hover:bg-[#37352F]/90 text-white gap-2 h-10 text-sm shadow-sm border-transparent px-4"
        >
          <Upload className="size-4" />
          画像アップロード
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
            <Input
              type="date"
              value={dateStart}
              onChange={(e) => onDateStartChange(e.target.value)}
              className="bg-white border-[rgba(55,53,47,0.16)] h-10 flex-1 text-sm"
            />
            <span className="text-[#37352F] font-medium text-sm">〜</span>
            <Input
              type="date"
              value={dateEnd}
              onChange={(e) => onDateEndChange(e.target.value)}
              className="bg-white border-[rgba(55,53,47,0.16)] h-10 flex-1 text-sm"
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
