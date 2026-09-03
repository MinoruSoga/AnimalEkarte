import { C } from "@/lib/design-tokens";
import type { PetSearchItem } from "@/types/generated/identitylink-responses";

interface PetLinkSectionProps {
  canEdit: boolean;
  pending: boolean;
  petQuery: string;
  setPetQuery: (value: string) => void;
  petHits: PetSearchItem[];
  selectedPets: PetSearchItem[];
  petGroupId: number | null;
  canLinkPets: boolean;
  historyText: string;
  togglePet: (item: PetSearchItem) => void;
  onLinkPets: () => void;
  onUnlinkPet: (item: PetSearchItem) => void;
  onLoadHistory: (item: PetSearchItem) => void;
  resolvePetGroupId: (item: PetSearchItem) => number | null;
}

export function PetLinkSection({
  canEdit,
  pending,
  petQuery,
  setPetQuery,
  petHits,
  selectedPets,
  petGroupId,
  canLinkPets,
  historyText,
  togglePet,
  onLinkPets,
  onUnlinkPet,
  onLoadHistory,
  resolvePetGroupId,
}: PetLinkSectionProps) {
  return (
    <section className={`rounded border p-4 space-y-3 ${C.borderLight} ${C.bgWhite}`} aria-label="ペットリンク">
      <h2 className={`font-semibold ${C.textInk}`}>ペットリンク</h2>
      <p className={`text-xs ${C.textInkMuted}`}>親となる飼主の連携グループが必要です。</p>
      <label className="block text-sm">
        <span className={C.textInkMuted}>検索</span>
        <input
          className={`mt-1 w-full rounded border px-3 py-2 ${C.borderLight} ${C.bgWhite} ${C.textInk}`}
          value={petQuery}
          onChange={(e) => setPetQuery(e.target.value)}
          placeholder="ペット名・番号"
        />
      </label>
      <ul className="space-y-1 max-h-40 overflow-auto text-sm">
        {petHits.map((p) => (
          <li key={`${p.clinic_id}-${p.pet_id}`}>
            <button
              type="button"
              className={`w-full text-left px-2 py-1 rounded ${C.bgHover}`}
              onClick={() => togglePet(p)}
            >
              [医院 {p.clinic_id}] {p.name}
            </button>
          </li>
        ))}
      </ul>
      <div className={`text-sm ${C.textInkSecondary}`}>
        選択: {selectedPets.map((p) => `${p.clinic_id}/${p.pet_id}`).join(", ") || "なし"}
        {petGroupId != null ? ` / 連携グループ #${petGroupId}` : null}
      </div>
      {canEdit ? (
        <div className="flex flex-wrap gap-2">
          <button
            type="button"
            className={`px-3 py-1.5 rounded text-sm ${C.bgBrand} ${C.textOnBrand}`}
            disabled={pending || !canLinkPets || selectedPets.length < 2}
            onClick={onLinkPets}
          >
            ペットをリンク
          </button>
          {selectedPets.map((p) => (
            <button
              key={`unlink-p-${p.clinic_id}-${p.pet_id}`}
              type="button"
              className={`px-3 py-1.5 rounded text-sm border ${C.borderLight}`}
              disabled={pending || resolvePetGroupId(p) == null}
              onClick={() => onUnlinkPet(p)}
            >
              連携解除 {p.clinic_id}/{p.pet_id}
            </button>
          ))}
        </div>
      ) : null}
      <div className="flex flex-wrap gap-2">
        {selectedPets.map((p) => (
          <button
            key={`hist-${p.clinic_id}-${p.pet_id}`}
            type="button"
            className={`px-3 py-1.5 rounded text-sm border ${C.borderLight}`}
            disabled={pending}
            onClick={() => onLoadHistory(p)}
          >
            連携履歴 {p.clinic_id}/{p.pet_id}
          </button>
        ))}
      </div>
      {historyText ? (
        <pre className={`text-xs whitespace-pre-wrap rounded p-2 ${C.bgMuted} ${C.textInk}`}>{historyText}</pre>
      ) : null}
    </section>
  );
}
