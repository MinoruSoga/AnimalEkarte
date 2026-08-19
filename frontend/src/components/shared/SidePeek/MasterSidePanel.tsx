import { memo, useEffect, type ReactNode } from "react";
import { STYLE } from "@/lib/design-tokens";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker/NavigationBlocker";
import { SidePeekPanel } from "@/components/shared/SidePeek/SidePeekPanel";
import { SidePeekToolbar } from "@/components/shared/SidePeek/SidePeekToolbar";
import { SidePeekBody } from "@/components/shared/SidePeek/SidePeekBody";
import { SidePeekTitleInput } from "@/components/shared/SidePeek/SidePeekTitleInput";
import { SidePeekFooter } from "@/components/shared/SidePeek/SidePeekFooter";

interface MasterSidePanelProps {
  isNew: boolean;
  title: string;
  onTitleChange: (v: string) => void;
  onClose: () => void;
  onSave?: () => void;
  action?: (formData: FormData) => void;
  onDelete?: () => void;
  icon: ReactNode;
  isPending?: boolean;
  titlePlaceholder?: string;
  /** BUG-083: inline error message shown below the title field */
  titleError?: string;
  /** Maximum number of characters allowed in the title field */
  titleMaxLength?: number;
  /** When true, shows a navigation blocker dialog if the user tries to navigate away */
  isDirty?: boolean;
  /** BUG-158: true の場合、保存・削除ボタンを非表示にし、閲覧モードで表示 */
  readOnly?: boolean;
  className?: string;
  children: ReactNode;
}

export const MasterSidePanel = memo(function MasterSidePanel({
  isNew,
  title,
  onTitleChange,
  onClose,
  onSave,
  action,
  onDelete,
  icon,
  isPending,
  titlePlaceholder,
  titleError,
  titleMaxLength,
  isDirty = false,
  readOnly = false,
  className,
  children,
}: MasterSidePanelProps) {
  // --- Focus Management (Accessibility) ---
  // BUG-084: パネルが開いたとき（マウント時）にタイトル入力へ自動フォーカス
  useEffect(() => {
    const element = document.getElementById("master-title");
    if (element) {
      element.focus();
    }
  }, []);

  // バリデーションエラー発生時もタイトルフィールドへフォーカス
  useEffect(() => {
    if (titleError) {
      const element = document.getElementById("master-title");
      if (element) {
        element.focus();
      }
    }
  }, [titleError]);

  // BUG-048: Save on Enter in any text input (not in textarea or button)
  const handleKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
    const tag = (e.target as HTMLElement).tagName;
    if (e.key === "Enter" && tag !== "TEXTAREA" && tag !== "BUTTON") {
      if (onSave) {
        e.preventDefault();
        onSave();
      }
      // If action is used, standard form submission will handle Enter
    }
  };

  const content = (
    <>
      <SidePeekToolbar isNew={isNew} onClose={onClose} onDelete={readOnly ? undefined : onDelete} readOnly={readOnly} />
      <SidePeekBody>
        <div className="pt-4 pb-2">
          <div className={STYLE.pageIcon}>{icon}</div>
        </div>
        <SidePeekTitleInput
          id="master-title"
          value={title}
          onChange={onTitleChange}
          placeholder={titlePlaceholder}
          onSave={onSave}
          error={titleError}
          maxLength={titleMaxLength}
        />
        <div className={`${STYLE.sectionDivider} mb-1`} />
        <div className="py-1">{children}</div>
      </SidePeekBody>
      <SidePeekFooter onCancel={onClose} onSave={readOnly ? undefined : onSave} isPending={isPending} readOnly={readOnly} />
    </>
  );

  return (
    <SidePeekPanel className={className} onKeyDown={handleKeyDown}>
      <NavigationBlocker
        when={isDirty}
        title="変更が保存されていません"
        description="変更が保存されていません。ページを離れますか？"
      />
      {action ? (
        <form action={action} noValidate className="flex-1 flex flex-col min-h-0">
          {content}
        </form>
      ) : (
        content
      )}
    </SidePeekPanel>
  );
});
