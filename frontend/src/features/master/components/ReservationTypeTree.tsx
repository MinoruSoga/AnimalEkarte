import { useState, useMemo } from "react";
import { ChevronDown, ChevronRight, Circle } from "lucide-react";
import { C, ICON } from "@/lib/design-tokens";
import type { ReservationType } from "../api/reservation-types";

interface Props {
  types: ReservationType[];
  selectedId: string | null;
  onSelect: (id: string) => void;
}

export function ReservationTypeTree({ types, selectedId, onSelect }: Props) {
  const parentNodes = types.filter((t) => !t.isLeaf);
  const childLeaves = types.filter((t) => t.isLeaf && t.depth === 1);
  const rootLeaves = types.filter((t) => t.isLeaf && t.depth === 0);

  // ユーザーが手動で開いた親ID のみを state として持つ
  const [openedIds, setOpenedIds] = useState<Set<string>>(new Set<string>());

  // 選択中 leaf の parentId を render 時に導出（effect 不要）
  const selectedParentId = useMemo(() => {
    if (!selectedId) return undefined;
    return types.find((t) => t.id === selectedId)?.parentId;
  }, [selectedId, types]);

  // 実効展開 Set = ユーザー開閉 + 選択中親（常に展開）
  const effectiveOpenedIds = useMemo(() => {
    if (selectedParentId) return new Set([...openedIds, selectedParentId]);
    return openedIds;
  }, [openedIds, selectedParentId]);

  function toggleOpen(id: string) {
    setOpenedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }

  return (
    <div className="py-2">
      {parentNodes.map((parent) => {
        const isOpen = effectiveOpenedIds.has(parent.id);
        const leaves = childLeaves.filter((c) => c.parentId === parent.id);

        return (
          <div key={parent.id}>
            {/* 親ノードヘッダー */}
            <button
              type="button"
              onClick={() => toggleOpen(parent.id)}
              className={`w-full flex items-center gap-1.5 px-3 py-2 text-sm font-medium text-left transition-colors ${C.hoverBgLight} ${
                parent.isActive ? C.text : C.text40
              }`}
            >
              {isOpen ? (
                <ChevronDown className={`${ICON.smXs} shrink-0`} />
              ) : (
                <ChevronRight className={`${ICON.smXs} shrink-0`} />
              )}
              <span className="truncate">
                {parent.name}
                {parent.isActive ? null : "（無効）"}
              </span>
            </button>

            {/* 子 leaf */}
            {isOpen ? (
              <div>
                {leaves.map((leaf) => {
                  const isSelected = leaf.id === selectedId;
                  return (
                    <button
                      key={leaf.id}
                      type="button"
                      onClick={() => onSelect(leaf.id)}
                      className={`w-full flex items-center gap-1.5 pl-6 pr-3 py-2 text-sm text-left transition-colors ${C.hoverBgLight} ${
                        isSelected ? `${C.bgAccent8} ${C.accent}` : leaf.isActive ? C.text : C.text40
                      }`}
                    >
                      <Circle
                        className={`${ICON.smXs} shrink-0 ${isSelected ? "fill-current" : ""}`}
                      />
                      <span className="truncate">
                        {leaf.name}
                        {leaf.isActive ? null : "（無効）"}
                      </span>
                    </button>
                  );
                })}
              </div>
            ) : null}
          </div>
        );
      })}

      {/* root-only leaf（親なし leaf） */}
      {rootLeaves.map((leaf) => {
        const isSelected = leaf.id === selectedId;
        return (
          <button
            key={leaf.id}
            type="button"
            onClick={() => onSelect(leaf.id)}
            className={`w-full flex items-center gap-1.5 px-3 py-2 text-sm text-left transition-colors ${C.hoverBgLight} ${
              isSelected ? `${C.bgAccent8} ${C.accent}` : leaf.isActive ? C.text : C.text40
            }`}
          >
            <Circle
              className={`${ICON.smXs} shrink-0 ${isSelected ? "fill-current" : ""}`}
            />
            <span className="truncate">
              {leaf.name}
              {leaf.isActive ? null : "（無効）"}
            </span>
          </button>
        );
      })}
    </div>
  );
}
