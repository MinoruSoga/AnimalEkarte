import { memo } from "react";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { C } from "@/lib/design-tokens";

import type { TrimmingMiddleColumnProps } from "./trimming-form-column-types";
import { TrimmingImageUploadField } from "./TrimmingImageUploadField";

const BW_UNIT_ITEMS = (
  <>
    <SelectItem value="Kg">Kg</SelectItem>
    <SelectItem value="g">g</SelectItem>
  </>
);

export const TrimmingMiddleColumn = memo(function TrimmingMiddleColumn({
  formData,
  completedImagePreview,
  onFormChange,
  onCompletedImageChange,
  onRemoveCompletedImage,
}: TrimmingMiddleColumnProps) {
  return (
    <div className={`${C.bgWhite} rounded-lg border ${C.borderMedium} p-3 space-y-4`}>
      <div>
        <Label htmlFor="trimming-bw" className={`text-sm ${C.text60} mb-2 block`}>体重 (BW)</Label>
        <div className="flex gap-2">
          <NumberInput
            id="trimming-bw"
            value={formData.bw}
            onChange={(value) => onFormChange({ bw: value })}
            placeholder="体重"
            className="flex-1 text-sm"
          />
          <Select
            value={formData.bwUnit}
            onValueChange={(value) => onFormChange({ bwUnit: value as "Kg" | "g" })}
          >
            <SelectTrigger className="w-[80px]" aria-label="体重単位">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {BW_UNIT_ITEMS}
            </SelectContent>
          </Select>
        </div>
      </div>

      <div>
        <Label htmlFor="trimming-bt" className={`text-sm ${C.text60} mb-2 block`}>体温 (BT)</Label>
        <NumberInput
          id="trimming-bt"
          step={0.1}
          value={formData.bt}
          onChange={(value) => onFormChange({ bt: value })}
          placeholder="体温"
          suffix="℃"
          className="text-sm"
        />
      </div>

      <div>
        <Label htmlFor="trimming-used-shampoo" className={`text-sm ${C.text60} mb-2 block`}>
          使用シャンプー
        </Label>
        <Input
          id="trimming-used-shampoo"
          value={formData.usedShampoo}
          onChange={(event) => onFormChange({ usedShampoo: event.target.value })}
          placeholder="シャンプー名"
          className="text-sm"
        />
      </div>

      <div>
        <Label htmlFor="trimming-used-ribbon" className={`text-sm ${C.text60} mb-2 block`}>
          使用リボン
        </Label>
        <Input
          id="trimming-used-ribbon"
          value={formData.usedRibbon}
          onChange={(event) => onFormChange({ usedRibbon: event.target.value })}
          placeholder="リボン"
          className="text-sm"
        />
      </div>

      <div>
        <Label htmlFor="trimming-remarks" className={`text-sm ${C.text60} mb-2 block`}>備考</Label>
        <Textarea
          id="trimming-remarks"
          value={formData.remarks}
          onChange={(event) => onFormChange({ remarks: event.target.value })}
          placeholder="備考を入力..."
          className="min-h-[60px] text-sm"
        />
      </div>

      <TrimmingImageUploadField
        label="完成画像"
        preview={completedImagePreview}
        previewAlt="Completed preview"
        onImageChange={onCompletedImageChange}
        onRemoveImage={onRemoveCompletedImage}
      />
    </div>
  );
});
