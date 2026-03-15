import type { ReactNode } from "react";
import { STYLE } from "@/lib/design-tokens";
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
  children,
}: MasterSidePanelProps) {
  return (
    <SidePeekPanel>
      <SidePeekToolbar isNew={isNew} onClose={onClose} onDelete={onDelete} />
      <SidePeekBody>
        <div className="pt-4 pb-2">
          <div className={STYLE.pageIcon}>{icon}</div>
        </div>
        <SidePeekTitleInput
          value={title}
          onChange={onTitleChange}
          placeholder={titlePlaceholder}
        />
        <div className={`${STYLE.sectionDivider} mb-1`} />
        <div className="py-1">{children}</div>
      </SidePeekBody>
      <SidePeekFooter onCancel={onClose} onSave={onSave} isPending={isPending} />
    </SidePeekPanel>
  );
}
