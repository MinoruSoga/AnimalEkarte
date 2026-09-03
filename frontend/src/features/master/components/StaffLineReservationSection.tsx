import type { Dispatch, SetStateAction } from "react";
import { MessageCircle } from "lucide-react";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { PropertyRow } from "@/components/shared/SidePeek";
import { C, ICON, PALETTE, STYLE } from "@/lib/design-tokens";

import { MASTER_INPUT_CLASS } from "../constants/styles";
import type { StaffFormData } from "../lib/staff-side-panel-model";

const STAFF_TYPE_SELECT_ITEMS = (
  <>
    <SelectItem value="doctor">医師</SelectItem>
    <SelectItem value="nurse">看護師</SelectItem>
    <SelectItem value="resource">設備</SelectItem>
  </>
);

interface StaffLineReservationSectionProps {
  formData: StaffFormData;
  setFormDataDirty: Dispatch<SetStateAction<StaffFormData>>;
}

export function StaffLineReservationSection({
  formData,
  setFormDataDirty,
}: StaffLineReservationSectionProps) {
  return (
    <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
      <div className="flex items-center gap-1.5 mb-3">
        <MessageCircle className={ICON.smXs} style={{ color: PALETTE.lineGreen }} />
        <p className={`text-xs font-medium ${C.text50}`}>LINE予約設定</p>
      </div>

      <PropertyRow label="LINE表示名">
        <input
          type="text"
          aria-label="LINE表示名"
          className={MASTER_INPUT_CLASS}
          value={formData.reservationDisplayName}
          onChange={(event) =>
            setFormDataDirty((prev) => ({ ...prev, reservationDisplayName: event.target.value }))
          }
          placeholder={formData.name || "空欄なら氏名を使用"}
        />
      </PropertyRow>

      <PropertyRow label="予約ページに表示">
        <Switch
          checked={formData.reservationVisible}
          onCheckedChange={(value) =>
            setFormDataDirty((prev) => ({ ...prev, reservationVisible: value }))
          }
        />
      </PropertyRow>

      <PropertyRow label="スタッフ種別">
        <Select
          value={formData.staffType}
          onValueChange={(value) => setFormDataDirty((prev) => ({ ...prev, staffType: value }))}
        >
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>{STAFF_TYPE_SELECT_ITEMS}</SelectContent>
        </Select>
      </PropertyRow>

      <PropertyRow label="LINE説明文">
        <input
          type="text"
          aria-label="LINE説明文"
          className={MASTER_INPUT_CLASS}
          value={formData.reservationComment}
          onChange={(event) =>
            setFormDataDirty((prev) => ({ ...prev, reservationComment: event.target.value }))
          }
          placeholder="LINE予約画面に表示する説明"
        />
      </PropertyRow>

      <PropertyRow label="画像URL">
        <input
          type="text"
          aria-label="画像URL"
          className={MASTER_INPUT_CLASS}
          value={formData.reservationImageUrl}
          onChange={(event) =>
            setFormDataDirty((prev) => ({ ...prev, reservationImageUrl: event.target.value }))
          }
          placeholder="https://..."
        />
      </PropertyRow>
    </div>
  );
}
