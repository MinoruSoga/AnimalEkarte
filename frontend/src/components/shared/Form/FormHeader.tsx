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

export const FormHeader = memo(function FormHeader({ title, description, icon, onBack, action }: FormHeaderProps) {
  return (
    <div className={STYLE.formHeader}>
      <div className="flex items-center gap-2">
        {onBack ? (
          <Button
            variant="ghost"
            type="button"
            onClick={onBack}
            className={`${C.text60} ${C.hoverText} hover:bg-transparent pl-0 size-11`}
          >
            <ChevronLeft className={ICON.page} />
            <span className="sr-only">戻る</span>
          </Button>
        ) : null}
        <div className="flex items-center gap-2">
          {icon ? <div className="shrink-0">{icon}</div> : null}
          <div className="flex flex-col">
            <h1 className={`text-base ${C.text} leading-tight tracking-[var(--tracking-notion)]`}>{title}</h1>
            {description ? <p className={`text-sm ${C.text50} mt-0.5 tracking-[var(--tracking-notion-sm)]`}>{description}</p> : null}
          </div>
        </div>
      </div>
      {action ? <div>{action}</div> : null}
    </div>
  );
});
