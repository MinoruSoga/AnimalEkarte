import { memo } from "react";
import { ICON, C, STYLE } from "@/lib/design-tokens";
import { Button } from "@/components/ui/button";
import { ChevronLeft } from "lucide-react";

interface FormHeaderProps {
  title: string;
  description?: string;
  icon?: React.ReactNode;
  onBack?: () => void;
  action?: React.ReactNode;
}

export const FormHeader = memo(function FormHeader({
  title,
  description,
  icon,
  onBack,
  action,
}: FormHeaderProps) {
  return (
    // flex-wrap + min-w-0: narrow viewports stack title/action instead of clipping (BUG-458).
    <div className={`${STYLE.formHeader} flex-wrap min-w-0`}>
      <div className="flex min-w-0 flex-1 items-center gap-2">
        {onBack ? (
          <Button
            variant="ghost"
            type="button"
            onClick={onBack}
            className={`${STYLE.btnGhost} pl-0 size-11 shrink-0`}
          >
            <ChevronLeft className={ICON.page} />
            <span className="sr-only">戻る</span>
          </Button>
        ) : null}
        <div className="flex min-w-0 items-center gap-2">
          {icon ? <div className="shrink-0">{icon}</div> : null}
          <div className="flex min-w-0 flex-col">
            <h1 className={`text-xl font-semibold ${C.text} break-words`}>{title}</h1>
            {description ? (
              <p className={`text-sm ${C.text50} mt-0.5 break-words`}>{description}</p>
            ) : null}
          </div>
        </div>
      </div>
      {action ? (
        <div className="flex min-w-0 flex-wrap items-center gap-2 shrink-0">{action}</div>
      ) : null}
    </div>
  );
});
