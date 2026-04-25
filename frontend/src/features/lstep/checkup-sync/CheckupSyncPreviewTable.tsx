import { C, STYLE, PALETTE } from "@/lib/design-tokens";
import type { CheckupSyncPreviewOwner } from "../api/get-checkup-sync-preview";

interface CheckupSyncPreviewTableProps {
  owners: CheckupSyncPreviewOwner[];
  selectedIds: Set<string>;
  onSelectionChange: (ids: Set<string>) => void;
  lineLinkedCount: number;
  totalCount: number;
}

export function CheckupSyncPreviewTable({
  owners,
  selectedIds,
  onSelectionChange,
  lineLinkedCount,
  totalCount,
}: CheckupSyncPreviewTableProps) {
  const lineLinkedOwners = owners.filter((o) => o.has_line);
  const allLineLinkedSelected =
    lineLinkedOwners.length > 0 &&
    lineLinkedOwners.every((o) => selectedIds.has(o.owner_id));

  function handleSelectAll(checked: boolean) {
    if (checked) {
      onSelectionChange(new Set(lineLinkedOwners.map((o) => o.owner_id)));
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

  return (
    <div className="space-y-3">
      {/* サマリー行 */}
      <p className={`text-sm ${C.text70}`}>
        {totalCount}件中{" "}
        <span className={`font-medium ${C.textStatusGreen}`}>
          {lineLinkedCount}件
        </span>
        がLINE連携済み（送信可能）
      </p>

      <div className={STYLE.tableContainer}>
        <table className="w-full table-fixed">
          <thead>
            <tr className={STYLE.tableHeaderRow}>
              <th className={`${STYLE.tableHeaderCell} w-12 px-4`}>
                <input
                  type="checkbox"
                  checked={allLineLinkedSelected}
                  onChange={(e) => handleSelectAll(e.target.checked)}
                  disabled={lineLinkedOwners.length === 0}
                  className="size-4 cursor-pointer accent-[#038B94]"
                  aria-label="LINE連携済みをすべて選択"
                />
              </th>
              <th className={`${STYLE.tableHeaderCell} text-left px-4`}>飼い主名</th>
              <th className={`${STYLE.tableHeaderCell} text-left px-4`}>ペット名</th>
              <th className={`${STYLE.tableHeaderCell} text-left px-4`}>最終来院日</th>
              <th className={`${STYLE.tableHeaderCell} text-left px-4 w-32`}>LINE連携</th>
            </tr>
          </thead>
          <tbody>
            {owners.length === 0 ? (
              <tr>
                <td colSpan={5} className={STYLE.tableEmpty}>
                  対象者が見つかりませんでした
                </td>
              </tr>
            ) : null}
            {owners.map((owner) => (
              <tr
                key={owner.owner_id}
                className={`${STYLE.tableRow} ${owner.has_line ? "" : "opacity-40"}`}
              >
                <td className="px-4">
                  <input
                    type="checkbox"
                    checked={selectedIds.has(owner.owner_id)}
                    onChange={(e) =>
                      handleRowToggle(owner.owner_id, e.target.checked)
                    }
                    disabled={!owner.has_line}
                    className="size-4 cursor-pointer accent-[#038B94] disabled:cursor-not-allowed"
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
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
