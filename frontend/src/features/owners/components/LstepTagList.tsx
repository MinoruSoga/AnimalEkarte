import { useState } from "react";
import { X } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { C, ICON } from "@/lib/design-tokens";
import { isAutoManagedTag } from "@/constants/lstep-auto-tag-prefixes";
const COLLAPSE_THRESHOLD = 10;

interface LstepTagListProps {
  tags: string[];
  onRemove: (tagName: string) => void;
  disabled?: boolean;
  canEdit?: boolean;
}

export function LstepTagList({
  tags,
  onRemove,
  disabled = false,
  canEdit = false,
}: LstepTagListProps) {
  const [expanded, setExpanded] = useState(false);

  const visibleTags =
    expanded || tags.length <= COLLAPSE_THRESHOLD ? tags : tags.slice(0, COLLAPSE_THRESHOLD);
  const hiddenCount = tags.length - COLLAPSE_THRESHOLD;

  if (tags.length === 0) {
    return <p className={`text-sm ${C.text40} italic`}>タグはありません</p>;
  }

  return (
    <div className="flex flex-col gap-2">
      <div className={`flex flex-wrap gap-1.5 ${disabled ? "opacity-50" : ""}`}>
        {visibleTags.map((tag) => {
          const isAuto = isAutoManagedTag(tag);
          return (
            <span key={tag} className="inline-flex items-center gap-0.5">
              {isAuto ? (
                <Badge
                  variant="outline"
                  className={`text-xs px-2 py-0.5 h-auto ${C.bgStatusGray} ${C.textStatusGray} border-0`}
                >
                  {tag}
                </Badge>
              ) : (
                <Badge
                  variant="secondary"
                  className="text-xs px-2 py-0.5 h-auto flex items-center gap-0.5"
                >
                  {tag}
                  {canEdit && !disabled ? (
                    <button
                      type="button"
                      aria-label={`${tag}を削除`}
                      onClick={() => onRemove(tag)}
                      className={`ml-0.5 rounded-full ${C.hoverBgMedium} transition-colors`}
                    >
                      <X className={ICON.xxs} />
                    </button>
                  ) : null}
                </Badge>
              )}
            </span>
          );
        })}
      </div>

      {tags.length > COLLAPSE_THRESHOLD ? (
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className={`self-start h-auto py-0.5 px-1 text-xs ${C.text50} ${C.hoverText}`}
          onClick={() => setExpanded((prev) => !prev)}
        >
          {expanded ? "折りたたむ" : `もっと見る（残り ${hiddenCount} 件）`}
        </Button>
      ) : null}
    </div>
  );
}
