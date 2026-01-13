import { Button } from "../ui/button";
import { ChevronLeft } from "lucide-react";

interface FormHeaderProps {
  title: string;
  description?: string;
  icon?: React.ReactNode;
  onBack?: () => void;
  action?: React.ReactNode;
}

export default function FormHeader({ title, description, icon, onBack, action }: FormHeaderProps) {
  return (
    <div className="sticky top-0 z-10 bg-[#F7F6F3] border-b border-[rgba(55,53,47,0.16)] px-4 py-2 flex items-center justify-between h-12">
      <div className="flex items-center gap-2">
        {onBack && (
          <Button
            variant="ghost"
            size="sm"
            onClick={onBack}
            className="text-[#37352F]/60 hover:text-[#37352F] hover:bg-transparent pl-0 h-10"
          >
            <ChevronLeft className="size-4" />
            <span className="sr-only">戻る</span>
          </Button>
        )}
        <div className="flex items-center gap-2">
          {icon && <div className="shrink-0">{icon}</div>}
          <div className="flex flex-col">
            <h1 className="text-sm font-bold text-[#37352F] leading-tight">{title}</h1>
            {description && <p className="text-sm text-[#37352F]/60 mt-0.5">{description}</p>}
          </div>
        </div>
      </div>
      {action && <div>{action}</div>}
    </div>
  );
}
