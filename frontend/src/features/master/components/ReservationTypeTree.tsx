import { useState, useMemo } from "react";
import { ChevronDown, ChevronRight, Circle } from "lucide-react";
import { C, ICON } from "@/lib/design-tokens";
import type { ReservationType } from "../api/reservation-types";

interface Props {
  types: ReservationType[];
  selectedId: string | null;
  onSelect: (id: string) => void;
}

function leafClassName(isSelected: boolean, isActive: boolean): string {
  const base = `w-full min-h-11 flex items-center gap-1.5 px-3 py-2 text-sm text-left transition-colors border-l-2 ${C.hoverBgLight}`;
  if (isSelected) return `${base} ${C.bgBrand8} ${C.textBrand} ${C.borderBrand}`;
  return `${base} border-transparent ${isActive ? C.text : C.text40}`;
}

export function ReservationTypeTree({ types, selectedId, onSelect }: Props) {
  const parentNodes = types.filter((t) => !t.isLeaf);
  const childLeaves = types.filter((t) => t.isLeaf && t.depth === 1);
  const rootLeaves = types.filter((t) => t.isLeaf && t.depth === 0);

  const [openedIds, setOpenedIds] = useState<Set<string>>(new Set<string>());

  const selectedParentId = useMemo(() => {
    if (!selectedId) return undefined;
    return types.find((t) => t.id === selectedId)?.parentId;
  }, [selectedId, types]);

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
        const parentAriaLabel = parent.isActive
          ? `${parent.name} グループ`
          : `${parent.name}（無効） グループ`;

        return (
          <div key={parent.id}>
            <button
              type="button"
              onClick={() => toggleOpen(parent.id)}
              aria-expanded={isOpen}
              aria-label={parentAriaLabel}
              className={`w-full min-h-11 flex items-center gap-1.5 px-3 py-2 text-sm font-medium text-left transition-colors ${C.bgPage} ${C.hoverBgLight} ${
                parent.isActive ? C.text : C.text40
              }`}
            >
              {isOpen ? (
                <ChevronDown className={`${ICON.smXs} shrink-0`} />
              ) : (
                <ChevronRight className={`${ICON.smXs} shrink-0`} />
              )}
              <span className="truncate flex-1">{parent.name}</span>
              {!parent.isActive ? (
                <span className={`shrink-0 text-2xs ${C.text40}`}>（無効）</span>
              ) : null}
              <span
                className={`shrink-0 text-2xs ${C.bgLight} ${C.text40} rounded-sm px-1.5 py-0.5`}
              >
                グループ
              </span>
            </button>

            {isOpen ? (
              <div
                className={`ml-3 border-l-2 ${C.borderMedium} pl-3`}
                role="group"
                aria-label={parent.name}
              >
                {leaves.map((leaf) => {
                  const isSelected = leaf.id === selectedId;
                  const leafAriaLabel = leaf.isActive ? undefined : `${leaf.name}（無効）`;
                  return (
                    <button
                      key={leaf.id}
                      type="button"
                      onClick={() => onSelect(leaf.id)}
                      aria-current={isSelected ? "true" : undefined}
                      aria-label={leafAriaLabel}
                      className={leafClassName(isSelected, leaf.isActive)}
                    >
                      <Circle
                        className={`${ICON.smXs} shrink-0 ${isSelected ? "fill-current" : ""}`}
                      />
                      <span className="truncate flex-1">{leaf.name}</span>
                      {!leaf.isActive ? (
                        <span className={`shrink-0 text-2xs ${C.text40}`}>（無効）</span>
                      ) : null}
                    </button>
                  );
                })}
              </div>
            ) : null}
          </div>
        );
      })}

      {rootLeaves.map((leaf) => {
        const isSelected = leaf.id === selectedId;
        const leafAriaLabel = leaf.isActive ? undefined : `${leaf.name}（無効）`;
        return (
          <button
            key={leaf.id}
            type="button"
            onClick={() => onSelect(leaf.id)}
            aria-current={isSelected ? "true" : undefined}
            aria-label={leafAriaLabel}
            className={leafClassName(isSelected, leaf.isActive)}
          >
            <Circle className={`${ICON.smXs} shrink-0 ${isSelected ? "fill-current" : ""}`} />
            <span className="truncate flex-1">{leaf.name}</span>
            {!leaf.isActive ? (
              <span className={`shrink-0 text-2xs ${C.text40}`}>（無効）</span>
            ) : null}
          </button>
        );
      })}
    </div>
  );
}
