// React/Framework
import { memo } from "react";

// Internal
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { C } from "@/lib/design-tokens";

interface EstimateFormProps {
  subject: string;
  onSubjectChange: (value: string) => void;
  canEdit: boolean;
}

export const EstimateForm = memo(function EstimateForm({
  subject,
  onSubjectChange,
  canEdit,
}: EstimateFormProps) {
  return (
    <div className="flex flex-col gap-1.5 w-[300px]">
      <Label className={`text-sm font-medium ${C.text60}`}>見積書件名</Label>
      <Input
        value={subject}
        onChange={(e) => onSubjectChange(e.target.value)}
        className={`${C.bgWhite} ${C.borderMedium} h-10 text-sm`}
        disabled={!canEdit}
      />
    </div>
  );
});
