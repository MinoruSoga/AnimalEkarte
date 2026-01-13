import React from "react";
import { Input } from "../../../components/ui/input";
import { Label } from "../../../components/ui/label";

interface EstimateFormProps {
  subject: string;
  onSubjectChange: (value: string) => void;
}

export const EstimateForm = React.memo(function EstimateForm({
  subject,
  onSubjectChange,
}: EstimateFormProps) {
  return (
    <div className="flex flex-col gap-1.5 w-[300px]">
      <Label className="text-sm font-medium text-[#37352F]/60">
        見積書件名
      </Label>
      <Input
        value={subject}
        onChange={(e) => onSubjectChange(e.target.value)}
        className="bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm"
      />
    </div>
  );
});
