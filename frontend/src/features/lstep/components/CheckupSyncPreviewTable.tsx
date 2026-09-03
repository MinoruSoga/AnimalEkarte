import { memo, useCallback, useLayoutEffect, useRef } from "react";

import { CHECKUP_SYNC_OWNER_LIMIT } from "@/constants/lstep-checkup-sync";
import { C, STYLE, PALETTE } from "@/lib/design-tokens";
import type { CheckupSyncPreviewOwner } from "../api/get-checkup-sync-preview";

interface CheckupSyncPreviewRowProps {
  owner: CheckupSyncPreviewOwner;
  eligible: boolean;
  selected: boolean;
  selectionLimitReached: boolean;
  onToggle: (ownerId: string, checked: boolean) => void;
}

const CheckupSyncPreviewRow = memo(function CheckupSyncPreviewRow({
  owner,
  eligible,
  selected,
  selectionLimitReached,
  onToggle,
}: CheckupSyncPreviewRowProps) {
  return (
    <tr className={`${STYLE.tableRow} ${eligible ? "" : "opacity-50"}`}>
      <td className={STYLE.tableCell}>
        <input
          type="checkbox"
          checked={selected}
          onChange={(e) => onToggle(owner.owner_id, e.target.checked)}
          disabled={!eligible || (!selected && selectionLimitReached)}
          className={`size-4 cursor-pointer ${C.accentBrand} disabled:cursor-not-allowed`}
          aria-label={`${owner.owner_name}を選択`}
        />
      </td>
      <td className={`${STYLE.tableCell} px-4`}>{owner.owner_name}</td>
      <td className={`${STYLE.tableCell} px-4 ${C.text70}`}>{owner.pet_names.join(", ") || "—"}</td>
      <td className={`${STYLE.tableCell} px-4 ${C.text70}`}>
        {owner.last_visit_date ? owner.last_visit_date : "—"}
      </td>
      <td className={STYLE.tableCell}>
        {owner.has_line ? (
          <span
            className="inline-flex items-center gap-1 text-sm font-medium rounded-full px-2 py-0.5"
            style={{
              color: PALETTE.lineGreen,
              backgroundColor: `${PALETTE.lineGreen}1A`,
            }}
          >
            LINE連携済
          </span>
        ) : (
          <span className={`text-sm ${C.textStatusGray}`}>未連携</span>
        )}
      </td>
      <td className={`${STYLE.tableCell} ${C.danger}`}>{owner.exclusion_reason ?? null}</td>
    </tr>
  );
});

interface CheckupSyncPreviewTableProps {
  owners: CheckupSyncPreviewOwner[];
  selectedIds: Set<string>;
  onSelectionChange: (ids: Set<string>) => void;
  eligibleCount: number;
  lineLinkedCount: number;
  optOutCount: number;
  noLivingPetCount: number;
  totalCount: number;
}

function isEligible(owner: CheckupSyncPreviewOwner): boolean {
  return owner.has_line && !owner.is_opt_out && owner.has_living_pet;
}

export function CheckupSyncPreviewTable({
  owners,
  selectedIds,
  onSelectionChange,
  eligibleCount,
  lineLinkedCount,
  optOutCount,
  noLivingPetCount,
  totalCount,
}: CheckupSyncPreviewTableProps) {
  const eligibleOwners = owners.filter(isEligible);
  const selectableOwners = eligibleOwners.slice(0, CHECKUP_SYNC_OWNER_LIMIT);
  const allEligibleSelected =
    selectableOwners.length > 0 && selectableOwners.every((o) => selectedIds.has(o.owner_id));
  const selectionLimitReached = selectedIds.size >= CHECKUP_SYNC_OWNER_LIMIT;

  const handleSelectAll = useCallback(
    (checked: boolean) => {
      if (checked) {
        onSelectionChange(new Set(selectableOwners.map((o) => o.owner_id)));
      } else {
        onSelectionChange(new Set());
      }
    },
    [selectableOwners, onSelectionChange],
  );

  const selectedIdsRef = useRef(selectedIds);
  useLayoutEffect(() => {
    selectedIdsRef.current = selectedIds;
  }, [selectedIds]);

  const handleRowToggle = useCallback(
    (ownerId: string, checked: boolean) => {
      const next = new Set(selectedIdsRef.current);
      if (checked) {
        if (next.size >= CHECKUP_SYNC_OWNER_LIMIT) {
          return;
        }
        next.add(ownerId);
      } else {
        next.delete(ownerId);
      }
      onSelectionChange(next);
    },
    [onSelectionChange],
  );

  const lineUnlinkedCount = totalCount - lineLinkedCount;

  return (
    <div className="space-y-3">
      {/* サマリー行 */}
      <div
        className={`rounded-md border ${C.borderLight} px-4 py-2.5 flex flex-wrap gap-x-4 gap-y-1 text-sm`}
      >
        <span>
          合計 <span className={`font-semibold ${C.text}`}>{totalCount}件</span>
        </span>
        <span>
          送信可能 <span className={`font-semibold ${C.textStatusGreen}`}>{eligibleCount}件</span>
        </span>
        {lineUnlinkedCount > 0 ? (
          <span className={C.text50}>LINE未連携 {lineUnlinkedCount}件</span>
        ) : null}
        {optOutCount > 0 ? <span className={C.text50}>配信停止中 {optOutCount}件</span> : null}
        {noLivingPetCount > 0 ? (
          <span className={C.text50}>生存ペットなし {noLivingPetCount}件</span>
        ) : null}
      </div>

      {eligibleOwners.length > CHECKUP_SYNC_OWNER_LIMIT || selectionLimitReached ? (
        <p className={`text-sm ${C.text50}`}>
          一度に選択できるのは最大{CHECKUP_SYNC_OWNER_LIMIT}名です
        </p>
      ) : null}

      <div className={STYLE.tableContainer}>
        <table className="w-full table-fixed">
          <thead>
            <tr className={STYLE.tableHeaderRow}>
              <th className={`${STYLE.tableHeaderCell} w-12 px-4`}>
                <input
                  type="checkbox"
                  checked={allEligibleSelected}
                  onChange={(e) => handleSelectAll(e.target.checked)}
                  disabled={eligibleOwners.length === 0}
                  className={`size-4 cursor-pointer ${C.accentBrand}`}
                  aria-label="送信可能対象をすべて選択"
                />
              </th>
              <th className={`${STYLE.tableHeaderCell} text-left px-4`}>飼主名</th>
              <th className={`${STYLE.tableHeaderCell} text-left px-4`}>ペット名</th>
              <th className={`${STYLE.tableHeaderCell} text-left px-4`}>最終来院日</th>
              <th className={`${STYLE.tableHeaderCell} text-left px-4 w-32`}>LINE連携</th>
              <th className={`${STYLE.tableHeaderCell} text-left px-4 w-36`}>対象外理由</th>
            </tr>
          </thead>
          <tbody>
            {owners.length === 0 ? (
              <tr>
                <td colSpan={6} className={STYLE.tableEmpty}>
                  対象者が見つかりませんでした
                </td>
              </tr>
            ) : null}
            {owners.map((owner) => (
              <CheckupSyncPreviewRow
                key={owner.owner_id}
                owner={owner}
                eligible={isEligible(owner)}
                selected={selectedIds.has(owner.owner_id)}
                selectionLimitReached={selectionLimitReached}
                onToggle={handleRowToggle}
              />
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
