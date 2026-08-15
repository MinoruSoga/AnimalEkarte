import { useEffect, useState, useTransition } from "react";
import { Navigate } from "react-router";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { C } from "@/lib/design-tokens";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { usePermission } from "@/hooks/use-permission";
import { ResourceIdentityLinks } from "@/types/generated/models";
import type { OwnerSearchItem, PetSearchItem } from "@/types/generated/identitylink-responses";

import {
  createOwnerIdentityGroup,
  createPetIdentityGroup,
  getLinkedTreatmentHistory,
  searchOwnersForLink,
  searchPetsForLink,
  unlinkOwnerIdentityMember,
  unlinkPetIdentityMember,
} from "../api/identity-links-api";

const SEARCH_DEBOUNCE_MS = 300;

type SelectedOwner = OwnerSearchItem;
type SelectedPet = PetSearchItem;

/**
 * Phase 1 manual identity-link UI: search → select → link/unlink → minimal history.
 * view gates page; edit gates mutation controls (DEC-15 manage → edit action).
 */
export function IdentityLinksPage() {
  const { canView, canEdit } = usePermission(ResourceIdentityLinks);
  if (!canView) {
    return <Navigate to="/" replace />;
  }

  return <IdentityLinksWorkbench canEdit={canEdit} />;
}

function IdentityLinksWorkbench({ canEdit }: { canEdit: boolean }) {
  const [ownerQuery, setOwnerQuery] = useState("");
  const [petQuery, setPetQuery] = useState("");
  const debouncedOwnerQuery = useDebouncedValue(ownerQuery, SEARCH_DEBOUNCE_MS);
  const debouncedPetQuery = useDebouncedValue(petQuery, SEARCH_DEBOUNCE_MS);

  const [ownerHits, setOwnerHits] = useState<OwnerSearchItem[]>([]);
  const [petHits, setPetHits] = useState<PetSearchItem[]>([]);
  const [selectedOwners, setSelectedOwners] = useState<SelectedOwner[]>([]);
  const [selectedPets, setSelectedPets] = useState<SelectedPet[]>([]);
  const [ownerGroupId, setOwnerGroupId] = useState<number | null>(null);
  const [petGroupId, setPetGroupId] = useState<number | null>(null);
  const [historyText, setHistoryText] = useState<string>("");
  const [error, setError] = useState<string | null>(null);
  const [pending, startTransition] = useTransition();

  // Clear hits when the debounced query becomes empty via render-time state
  // adjustment (React-recommended pattern) so we do not call setState inside
  // the effect body. Timing still follows debouncedOwnerQuery / debouncedPetQuery,
  // matching the previous early-return setHits([]) path.
  const trimmedOwnerQuery = debouncedOwnerQuery.trim();
  const trimmedPetQuery = debouncedPetQuery.trim();
  const [prevTrimmedOwnerQuery, setPrevTrimmedOwnerQuery] = useState(trimmedOwnerQuery);
  const [prevTrimmedPetQuery, setPrevTrimmedPetQuery] = useState(trimmedPetQuery);
  if (trimmedOwnerQuery !== prevTrimmedOwnerQuery) {
    setPrevTrimmedOwnerQuery(trimmedOwnerQuery);
    if (trimmedOwnerQuery.length < 1) {
      setOwnerHits([]);
    }
  }
  if (trimmedPetQuery !== prevTrimmedPetQuery) {
    setPrevTrimmedPetQuery(trimmedPetQuery);
    if (trimmedPetQuery.length < 1) {
      setPetHits([]);
    }
  }

  useEffect(() => {
    if (trimmedOwnerQuery.length < 1) {
      return;
    }
    let cancelled = false;
    void searchOwnersForLink(trimmedOwnerQuery)
      .then((items) => {
        if (!cancelled) setOwnerHits(items);
      })
      .catch((e: unknown) => {
        if (!cancelled) setError(e instanceof Error ? e.message : "飼主検索に失敗しました");
      });
    return () => {
      cancelled = true;
    };
  }, [trimmedOwnerQuery]);

  useEffect(() => {
    if (trimmedPetQuery.length < 1) {
      return;
    }
    let cancelled = false;
    void searchPetsForLink(trimmedPetQuery)
      .then((items) => {
        if (!cancelled) setPetHits(items);
      })
      .catch((e: unknown) => {
        if (!cancelled) setError(e instanceof Error ? e.message : "ペット検索に失敗しました");
      });
    return () => {
      cancelled = true;
    };
  }, [trimmedPetQuery]);

  const toggleOwner = (item: OwnerSearchItem) => {
    setSelectedOwners((prev) => {
      const exists = prev.some((p) => p.clinic_id === item.clinic_id && p.owner_id === item.owner_id);
      if (exists) {
        return prev.filter((p) => !(p.clinic_id === item.clinic_id && p.owner_id === item.owner_id));
      }
      return [...prev, item];
    });
  };

  const togglePet = (item: PetSearchItem) => {
    setSelectedPets((prev) => {
      const exists = prev.some((p) => p.clinic_id === item.clinic_id && p.pet_id === item.pet_id);
      if (exists) {
        return prev.filter((p) => !(p.clinic_id === item.clinic_id && p.pet_id === item.pet_id));
      }
      return [...prev, item];
    });
  };

  const onLinkOwners = () => {
    if (!canEdit || selectedOwners.length < 2) return;
    setError(null);
    startTransition(async () => {
      try {
        const group = await createOwnerIdentityGroup(
          selectedOwners.map((o) => ({ clinic_id: o.clinic_id, owner_id: o.owner_id })),
        );
        setOwnerGroupId(group.id);
      } catch (e: unknown) {
        setError(e instanceof Error ? e.message : "飼主リンクに失敗しました");
      }
    });
  };

  const onUnlinkOwner = (item: SelectedOwner) => {
    if (!canEdit || ownerGroupId == null) return;
    setError(null);
    startTransition(async () => {
      try {
        await unlinkOwnerIdentityMember(ownerGroupId, {
          clinic_id: item.clinic_id,
          owner_id: item.owner_id,
        });
        setSelectedOwners((prev) =>
          prev.filter((p) => !(p.clinic_id === item.clinic_id && p.owner_id === item.owner_id)),
        );
      } catch (e: unknown) {
        setError(e instanceof Error ? e.message : "飼主unlinkに失敗しました");
      }
    });
  };

  const onLinkPets = () => {
    if (!canEdit || ownerGroupId == null || selectedPets.length < 2) return;
    setError(null);
    startTransition(async () => {
      try {
        const group = await createPetIdentityGroup(
          ownerGroupId,
          selectedPets.map((p) => ({ clinic_id: p.clinic_id, pet_id: p.pet_id })),
        );
        setPetGroupId(group.id);
      } catch (e: unknown) {
        setError(e instanceof Error ? e.message : "ペットリンクに失敗しました");
      }
    });
  };

  const onUnlinkPet = (item: SelectedPet) => {
    if (!canEdit || petGroupId == null) return;
    setError(null);
    startTransition(async () => {
      try {
        await unlinkPetIdentityMember(petGroupId, {
          clinic_id: item.clinic_id,
          pet_id: item.pet_id,
        });
        setSelectedPets((prev) =>
          prev.filter((p) => !(p.clinic_id === item.clinic_id && p.pet_id === item.pet_id)),
        );
      } catch (e: unknown) {
        setError(e instanceof Error ? e.message : "ペットunlinkに失敗しました");
      }
    });
  };

  const onLoadHistory = (item: SelectedPet) => {
    setError(null);
    startTransition(async () => {
      try {
        const hist = await getLinkedTreatmentHistory(item.clinic_id, item.pet_id, true, 1, 20);
        setHistoryText(
          hist.items
            .map(
              (h) =>
                `[clinic ${h.clinic_id}/pet ${h.pet_id}] ${h.record_date} ${h.record_no}: ${h.content}`,
            )
            .join("\n") || "（履歴なし）",
        );
      } catch (e: unknown) {
        setError(e instanceof Error ? e.message : "履歴取得に失敗しました");
      }
    });
  };

  return (
    <PageLayout
      title="同一飼主・ペット連携"
      description="所属医院内の手動リンクのみ。権限のない医院の ID はサーバ側で拒否されます。"
      resource={ResourceIdentityLinks}
    >
      <div className="space-y-6">
        {!canEdit && (
          <p className={`text-sm ${C.textWarning}`} role="status">
            閲覧のみ（link/unlink 権限なし）
          </p>
        )}

        {error ? (
          <div
            className={`rounded border p-3 text-sm ${C.borderDanger} ${C.bgDanger10} ${C.textWarning}`}
            role="alert"
          >
            {error}
          </div>
        ) : null}

        <section className={`rounded border p-4 space-y-3 ${C.borderLight} ${C.bgWhite}`} aria-label="飼主リンク">
          <h2 className={`font-semibold ${C.textInk}`}>飼主リンク</h2>
          <label className="block text-sm">
            <span className={C.textInkMuted}>検索（300ms debounce）</span>
            <input
              className={`mt-1 w-full rounded border px-3 py-2 ${C.borderLight} ${C.bgWhite} ${C.textInk}`}
              value={ownerQuery}
              onChange={(e) => setOwnerQuery(e.target.value)}
              placeholder="氏名・カナ・電話"
            />
          </label>
          <ul className="space-y-1 max-h-40 overflow-auto text-sm">
            {ownerHits.map((o) => (
              <li key={`${o.clinic_id}-${o.owner_id}`}>
                <button
                  type="button"
                  className={`w-full text-left px-2 py-1 rounded ${C.bgHover}`}
                  onClick={() => toggleOwner(o)}
                >
                  [clinic {o.clinic_id}] {o.name} ({o.phone})
                </button>
              </li>
            ))}
          </ul>
          <div className={`text-sm ${C.textInkSecondary}`}>
            選択: {selectedOwners.map((o) => `${o.clinic_id}/${o.owner_id}`).join(", ") || "なし"}
            {ownerGroupId != null && ` / group #${ownerGroupId}`}
          </div>
          {canEdit ? (
            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                className={`px-3 py-1.5 rounded text-sm ${C.bgBrand} ${C.textOnBrand}`}
                disabled={pending || selectedOwners.length < 2}
                onClick={onLinkOwners}
              >
                飼主をリンク
              </button>
              {selectedOwners.map((o) => (
                <button
                  key={`unlink-o-${o.clinic_id}-${o.owner_id}`}
                  type="button"
                  className={`px-3 py-1.5 rounded text-sm border ${C.borderLight}`}
                  disabled={pending || ownerGroupId == null}
                  onClick={() => onUnlinkOwner(o)}
                >
                  unlink {o.clinic_id}/{o.owner_id}
                </button>
              ))}
            </div>
          ) : null}
        </section>

        <section className={`rounded border p-4 space-y-3 ${C.borderLight} ${C.bgWhite}`} aria-label="ペットリンク">
          <h2 className={`font-semibold ${C.textInk}`}>ペットリンク</h2>
          <p className={`text-xs ${C.textInkMuted}`}>親の飼主 identity group が必要です。</p>
          <label className="block text-sm">
            <span className={C.textInkMuted}>検索（300ms debounce）</span>
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
                  [clinic {p.clinic_id}] {p.name}
                </button>
              </li>
            ))}
          </ul>
          <div className={`text-sm ${C.textInkSecondary}`}>
            選択: {selectedPets.map((p) => `${p.clinic_id}/${p.pet_id}`).join(", ") || "なし"}
            {petGroupId != null && ` / group #${petGroupId}`}
          </div>
          {canEdit ? (
            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                className={`px-3 py-1.5 rounded text-sm ${C.bgBrand} ${C.textOnBrand}`}
                disabled={pending || ownerGroupId == null || selectedPets.length < 2}
                onClick={onLinkPets}
              >
                ペットをリンク
              </button>
              {selectedPets.map((p) => (
                <button
                  key={`unlink-p-${p.clinic_id}-${p.pet_id}`}
                  type="button"
                  className={`px-3 py-1.5 rounded text-sm border ${C.borderLight}`}
                  disabled={pending || petGroupId == null}
                  onClick={() => onUnlinkPet(p)}
                >
                  unlink {p.clinic_id}/{p.pet_id}
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
      </div>
    </PageLayout>
  );
}
