import { useId } from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Label } from "@/components/ui/label";
import { C } from "@/lib/design-tokens";
import { VISIT_TYPE_OPTIONS } from "../routes/medical-record-form-model";

interface VisitTypeSelectProps {
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
}

export function VisitTypeSelect({ value, onChange, disabled = false }: VisitTypeSelectProps) {
  const triggerId = useId();
  return (
    <div className="flex flex-col gap-0 min-w-[72px] shrink-0">
      <Label htmlFor={triggerId} className={`text-xs ${C.text50}`}>
        来院種別
      </Label>
      <Select value={value} onValueChange={onChange} disabled={disabled}>
        <SelectTrigger
          id={triggerId}
          className={`h-8 text-sm border-none bg-transparent p-0 gap-1 ${C.text} font-medium focus-visible:ring-2 ${C.focusVisibleRingActionPrimary} focus-visible:ring-offset-1`}
        >
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {VISIT_TYPE_OPTIONS.map((opt) => (
            <SelectItem key={opt} value={opt}>
              {opt}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}
