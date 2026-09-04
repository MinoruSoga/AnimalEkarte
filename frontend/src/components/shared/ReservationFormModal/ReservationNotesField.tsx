import { C } from "@/lib/design-tokens";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { Reservation } from "@/types";
import { FieldLabel } from "./ReservationDateTimeFields";

const TRIGGER_CLASS = `h-9 text-sm bg-white ${C.borderMediumLight} ${C.text} ${C.hoverBgSubtle} transition-colors`;

interface ReservationNotesFieldProps {
  formData: Partial<Reservation>;
  onChange: (data: Partial<Reservation>) => void;
}

export function ReservationNotesField({ formData, onChange }: ReservationNotesFieldProps) {
  return (
    <>
      <div className="space-y-1.5">
        <FieldLabel>メモ</FieldLabel>
        <Textarea
          value={formData.notes || ""}
          onChange={(e) => onChange({ ...formData, notes: e.target.value })}
          placeholder="詳細や備考を入力..."
          className={`min-h-[80px] text-sm resize-none bg-white ${C.borderMedium} ${C.text} ${C.textPlaceholder} ${C.focusRingMedium}`}
        />
      </div>

      <div className="space-y-1.5">
        <FieldLabel>予約ソース</FieldLabel>
        <Select
          value={formData.source || "manual"}
          onValueChange={(v: string) => onChange({ ...formData, source: v as "manual" | "line" })}
        >
          <SelectTrigger className={TRIGGER_CLASS}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="manual">手動予約</SelectItem>
            <SelectItem value="line">LINE予約</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </>
  );
}
