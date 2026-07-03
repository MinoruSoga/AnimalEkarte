import { C, STYLE, PALETTE } from "@/lib/design-tokens";
import type { CheckupSyncPreviewOwner } from "../api/get-checkup-sync-preview";

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
  const allEligibleSelected =
    eligibleOwners.length > 0 &&
    eligibleOwners.every((o) => selectedIds.has(o.owner_id));

  function handleSelectAll(checked: boolean) {
    if (checked) {
      onSelectionChange(new Set(eligibleOwners.map((o) => o.owner_id)));
    } else {
      onSelectionChange(new Set());
    }
  }

  function handleRowToggle(ownerId: string, checked: boolean) {
    const next = new Set(selectedIds);
    if (checked) {
      next.add(ownerId);
    } else {
      next.delete(ownerId);
    }
    onSelectionChange(next);
  }

  const lineUnlinkedCount = totalCount - lineLinkedCount;

  return (
    <div className="space-y-3">
      {/* サマリー行 */}
      <div className={`rounded-md border ${C.borderLight} px-4 py-2.5 flex flex-wrap gap-x-4 gap-y-1 text-sm`}>
        <span>
          合計{" "}
          <span className={`font-semibold ${C.text}`}>{totalCount}件</span>
        </span>
        <span>
          送信可能{" "}
          <span className={`font-semibold ${C.textStatusGreen}`}>{eligibleCount}件</span>
        </span>
        {lineUnlinkedCount > 0 ? (
          <span className={C.text50}>
            LINE未連携 {lineUnlinkedCount}件
          </span>
        ) : null}
        {optOutCount > 0 ? (
          <span className={C.text50}>
            配信停止中 {optOutCount}件
          </span>
        ) : null}
        {noLivingPetCount > 0 ? (
          <span className={C.text50}>
            生存ペットなし {noLivingPetCount}件
          </span>
        ) : null}
      </div>

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
            {owners.map((owner) => {
              const eligible = isEligible(owner);
              return (
                <tr
                  key={owner.owner_id}
                  className={`${STYLE.tableRow} ${eligible ? "" : "opacity-50"}`}
                >
                  <td className="px-4">
                    <input
                      type="checkbox"
                      checked={selectedIds.has(owner.owner_id)}
                      onChange={(e) =>
                        handleRowToggle(owner.owner_id, e.target.checked)
                      }
                      disabled={!eligible}
                      className={`size-4 cursor-pointer ${C.accentBrand} disabled:cursor-not-allowed`}
                      aria-label={`${owner.owner_name}を選択`}
                    />
                  </td>
                  <td className={`${STYLE.tableCell} px-4`}>{owner.owner_name}</td>
                  <td className={`${STYLE.tableCell} px-4 ${C.text70}`}>
                    {owner.pet_names.join(", ") || "—"}
                  </td>
                  <td className={`${STYLE.tableCell} px-4 ${C.text70}`}>
                    {owner.last_visit_date ? owner.last_visit_date : "—"}
                  </td>
                  <td className="px-4 py-2.5">
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
                  <td className={`px-4 py-2.5 text-sm ${C.danger}`}>
                    {owner.exclusion_reason ?? null}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
