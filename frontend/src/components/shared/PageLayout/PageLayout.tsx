import { memo, type ReactNode, type RefObject } from "react";
import { FormHeader } from "@/components/shared/Form/FormHeader";
import { PermissionBadges } from "@/components/shared/PermissionBadges/PermissionBadges";
import { LAYOUT, STYLE } from "@/lib/design-tokens";
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

export const PageLayout = memo(function PageLayout({
  children,
  title,
  description,
  onBack,
  icon,
  headerAction,
  resource,
  maxWidth = LAYOUT.pageContentMaxWidth.default,
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
      className={`flex min-w-0 flex-col h-full ${STYLE.page} ${className || ""}`}
    >
      <FormHeader
        title={title}
        description={description}
        onBack={onBack}
        icon={icon}
        action={actionContent}
      />
      <div ref={scrollContainerRef} className="flex-1 min-w-0 overflow-y-auto w-full flex flex-col">
        <div
          className={`${maxWidth} ${align === "center" ? "mx-auto" : ""} min-w-0 w-full px-3 py-6 flex-1 flex flex-col`}
        >
          {children}
        </div>
      </div>
    </div>
  );
});
