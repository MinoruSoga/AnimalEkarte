import { ReactNode, type RefObject } from "react";
import { FormHeader } from "@/components/shared/Form/FormHeader";

interface PageLayoutProps {
  children: ReactNode;
  title: string;
  description?: string;
  onBack?: () => void;
  icon?: ReactNode;
  headerAction?: ReactNode;
  maxWidth?: string;
  className?: string;
  align?: "center" | "left";
  /** BUG-MEDI-005: スクロールコンテナへの ref（タブ切替時に scrollTop = 0 に使用） */
  scrollContainerRef?: RefObject<HTMLDivElement | null>;
}

export function PageLayout({
  children,
  title,
  description: _description,
  onBack,
  icon,
  headerAction,
  maxWidth = "max-w-[1440px]",
  className,
  align = "center",
  scrollContainerRef,
}: PageLayoutProps) {
  return (
    <div
      className={`flex flex-col h-full bg-[#F7F6F3] overflow-hidden ${className || ""}`}
    >
      <FormHeader
        title={title}
        onBack={onBack}
        icon={icon}
        action={headerAction}
      />
      <div ref={scrollContainerRef} className="flex-1 overflow-y-auto w-full flex flex-col">
        <div
          className={`${maxWidth} ${align === "center" ? "mx-auto" : ""} w-full px-3 py-5 flex-1 flex flex-col`}
        >
          {children}
        </div>
      </div>
    </div>
  );
}
