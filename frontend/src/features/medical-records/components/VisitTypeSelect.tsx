import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { C } from "@/lib/design-tokens";
import { VISIT_TYPE_OPTIONS } from "../routes/medical-record-form-model";

interface VisitTypeSelectProps {
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
}

export function VisitTypeSelect({ value, onChange, disabled = false }: VisitTypeSelectProps) {
  return (
    <div className="flex flex-col gap-0 min-w-[72px] shrink-0">
      <span className={`text-xs ${C.text50}`}>来院種別</span>
      <Select value={value} onValueChange={onChange} disabled={disabled}>
        <SelectTrigger
          aria-label="来院種別"
          className={`h-8 text-sm border-none bg-transparent p-0 focus:ring-0 gap-1 ${C.text} font-medium`}
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
