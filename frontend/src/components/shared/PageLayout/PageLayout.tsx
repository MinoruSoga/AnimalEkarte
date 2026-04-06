import { ReactNode, type RefObject } from "react";
import { FormHeader } from "@/components/shared/Form/FormHeader";
import { PermissionBadges } from "@/components/shared/PermissionBadges/PermissionBadges";
import { C } from "@/lib/design-tokens";
import type { Resource } from "@/types/generated/models";

interface PageLayoutProps {
  children: ReactNode;
  title: string;
  description?: string;
  onBack?: () => void;
  icon?: ReactNode;
  headerAction?: ReactNode;
  /** 権限バッジ表示用リソース。指定するとヘッダー右側に現在の権限をバッジ表示する。 */
  resource?: Resource;
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
  resource,
  maxWidth = "max-w-[1440px]",
  className,
  align = "center",
  scrollContainerRef,
}: PageLayoutProps) {
  const actionContent = (resource != null || headerAction != null) ? (
    <div className="flex items-center gap-3">
      {resource != null ? <PermissionBadges resource={resource} /> : null}
      {headerAction != null ? headerAction : null}
    </div>
  ) : null;

  return (
    <div
      className={`flex flex-col h-full ${C.bgPage} overflow-hidden ${className || ""}`}
    >
      <FormHeader
        title={title}
        onBack={onBack}
        icon={icon}
        action={actionContent}
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
