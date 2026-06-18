import { useRef, type KeyboardEvent } from "react";

import { C } from "@/lib/design-tokens";
import type { Pet } from "@/types";

/** ペットタブと対応するレポート本文（tabpanel）を関連付ける id 群。OwnerReport と共有する。 */
export const OWNER_REPORT_TABPANEL_ID = "owner-report-tabpanel";
export const ownerReportPetTabId = (petId: string) => `owner-report-pet-tab-${petId}`;

interface PetSwitcherProps {
  pets: Pet[];
  selectedPetId: string | undefined;
  onSelect: (petId: string) => void;
}

/**
 * #158 R5: ペット切替タブ。選択でページ遷移せず（state 更新 + URL ?petId= 同期は親が担当）。
 * tablist/tab パターンを完成させるため、各タブは aria-controls で本文 tabpanel を指す。
 *
 * キーボード操作は WAI-ARIA APG「Tabs（手動アクティベーション）」準拠:
 * - roving tabindex: Tab で到達できるタブは常に 1 つ（選択中タブ）。他は tabIndex=-1。
 * - ←/→/↑/↓ は端で巻き戻りつつフォーカスのみ移動する（ページ遷移・選択はしない）。
 * - Home/End で先頭/末尾へフォーカス。
 * - Enter/Space で activation（ネイティブ button の活性化 → onClick）。
 *   ペット切替は履歴データ再取得を伴うため、矢印移動では選択を発火させない手動アクティベーションを採る。
 */
export function PetSwitcher({ pets, selectedPetId, onSelect }: PetSwitcherProps) {
  const tabRefs = useRef<(HTMLButtonElement | null)[]>([]);

  // 選択が一致しない場合でも tablist 全体が Tab 到達不能にならないよう先頭へフォールバックする。
  const activeIndex = pets.findIndex((pet) => pet.id === selectedPetId);
  const tabbableIndex = activeIndex >= 0 ? activeIndex : 0;

  const focusTabAt = (index: number) => {
    const count = pets.length;
    if (count === 0) return;
    const wrapped = ((index % count) + count) % count; // 両端で巻き戻る
    tabRefs.current[wrapped]?.focus();
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLButtonElement>, index: number) => {
    switch (event.key) {
      case "ArrowRight":
      case "ArrowDown":
        event.preventDefault();
        focusTabAt(index + 1);
        break;
      case "ArrowLeft":
      case "ArrowUp":
        event.preventDefault();
        focusTabAt(index - 1);
        break;
      case "Home":
        event.preventDefault();
        focusTabAt(0);
        break;
      case "End":
        event.preventDefault();
        focusTabAt(pets.length - 1);
        break;
      // Enter / Space はネイティブ button の活性化（onClick）に委ねる。
      default:
        break;
    }
  };

  return (
    <div className="flex flex-wrap gap-1.5" role="tablist" aria-label="ペット切替">
      {pets.map((pet, index) => {
        const isActive = pet.id === selectedPetId;
        return (
          <button
            key={pet.id}
            ref={(el) => {
              tabRefs.current[index] = el;
            }}
            type="button"
            role="tab"
            id={ownerReportPetTabId(pet.id)}
            aria-selected={isActive}
            aria-controls={OWNER_REPORT_TABPANEL_ID}
            tabIndex={index === tabbableIndex ? 0 : -1}
            onClick={() => onSelect(pet.id)}
            onKeyDown={(event) => handleKeyDown(event, index)}
            className={`rounded-full border px-2.5 py-0.5 text-sm whitespace-nowrap transition-colors focus-visible:outline-none focus-visible:ring-2 ${C.focusRingAccent} focus-visible:ring-offset-1 ${
              isActive
                ? `${C.bgBrand} border-transparent text-white`
                : `${C.bgWhite} ${C.borderLight} ${C.text} ${C.hoverBgPage}`
            }`}
          >
            {pet.name}
            {pet.species ? <span className="ml-1 text-xs opacity-70">{pet.species}</span> : null}
          </button>
        );
      })}
    </div>
  );
}
