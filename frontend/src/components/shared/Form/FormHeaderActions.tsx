import type { ReactNode } from "react";

import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { Button } from "@/components/ui/button";

interface FormHeaderActionsProps {
  onCancel: () => void;
  cancelLabel?: string;
  submitLabel?: string;
  submitDisabled?: boolean;
  submitFormId?: string;
  extra?: ReactNode;
}

export function FormHeaderActions({
  onCancel,
  cancelLabel = "キャンセル",
  submitLabel,
  submitDisabled,
  submitFormId,
  extra,
}: FormHeaderActionsProps) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      {extra}
      <Button type="button" variant="outline" size="sm" onClick={onCancel} className="h-10">
        {cancelLabel}
      </Button>
      {submitLabel ? (
        <SubmitButton form={submitFormId} disabled={submitDisabled} className="h-10 px-4">
          {submitLabel}
        </SubmitButton>
      ) : null}
    </div>
  );
}
