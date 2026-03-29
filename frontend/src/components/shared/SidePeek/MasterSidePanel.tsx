import type { ReactNode } from "react";
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
  onSave: () => void;
  onDelete?: () => void;
  icon: ReactNode;
  isPending?: boolean;
  titlePlaceholder?: string;
  /** When true, shows a navigation blocker dialog if the user tries to navigate away */
  isDirty?: boolean;
  children: ReactNode;
}

export function MasterSidePanel({
  isNew,
  title,
  onTitleChange,
  onClose,
  onSave,
  onDelete,
  icon,
  isPending,
  titlePlaceholder,
  isDirty = false,
  children,
}: MasterSidePanelProps) {
  // BUG-048: Save on Enter in any text input (not in textarea or button)
  const handleKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
    const tag = (e.target as HTMLElement).tagName;
    if (e.key === "Enter" && tag !== "TEXTAREA" && tag !== "BUTTON") {
      e.preventDefault();
      onSave();
    }
  };

  return (
    <SidePeekPanel onKeyDown={handleKeyDown}>
      <NavigationBlocker
        when={isDirty}
        title="変更が保存されていません"
        description="変更が保存されていません。ページを離れますか？"
      />
      <SidePeekToolbar isNew={isNew} onClose={onClose} onDelete={onDelete} />
      <SidePeekBody>
        <div className="pt-4 pb-2">
          <div className={STYLE.pageIcon}>{icon}</div>
        </div>
        <SidePeekTitleInput
          value={title}
          onChange={onTitleChange}
          placeholder={titlePlaceholder}
          onSave={onSave}
        />
        <div className={`${STYLE.sectionDivider} mb-1`} />
        <div className="py-1">{children}</div>
      </SidePeekBody>
      <SidePeekFooter onCancel={onClose} onSave={onSave} isPending={isPending} />
    </SidePeekPanel>
  );
}
