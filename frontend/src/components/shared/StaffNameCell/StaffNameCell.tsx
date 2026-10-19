import { C, STYLE } from "@/lib/design-tokens";
import { TableCell } from "@/components/ui/table";

interface StaffNameCellProps {
  /** 解決済みの担当者名。doctor_id など名前解決に失敗した生値が入ることもある。 */
  staff: string;
  /** アクティブなスタッフ名集合による検証（useStaffValidation）。 */
  isValidStaff: (name: string) => boolean;
  /** 追加の TableCell クラス（列の表示切替 hidden lg:table-cell など）。 */
  className?: string;
}

/**
 * 名前解決に失敗したことを表す共有マーカー（lib/transforms の doctor 変換が返す）。
 * 生の doctor_id を氏名として見せることはしない。
 */
// EMR-436: 外部からの参照はないため非export（同一ファイル内の判定でのみ使用）。
const UNRESOLVED_STAFF_LABEL = "担当者不明";

/** 担当者の名前が数字のみ（= doctor_id 等の生 ID）かどうか。氏名は数値だけにならない前提。 */
const isUnresolvedStaffValue = (value: string): boolean =>
  value === UNRESOLVED_STAFF_LABEL || /^\d+$/.test(value);

/**
 * EMR-412 X3: カルテ一覧・トリミング一覧で共通の担当者セル。
 * - 空欄は「-」（X15 の記号）
 * - 名前解決できない生 ID のみは「担当者不明」
 * - 退職等で無効は「（無効）」を文字で併記（色警告・アイコンに依存しない）
 * - 名前は C.text。whitespace-nowrap・min-w-[160px]・STYLE.tableCell。
 */
export function StaffNameCell({ staff, isValidStaff, className = "" }: StaffNameCellProps) {
  if (!staff) {
    return (
      <TableCell className={`${STYLE.tableCell} min-w-[160px] whitespace-nowrap ${className}`}>
        <span className={C.text60}>-</span>
      </TableCell>
    );
  }
  if (!isValidStaff(staff)) {
    if (isUnresolvedStaffValue(staff)) {
      return (
        <TableCell className={`${STYLE.tableCell} min-w-[160px] whitespace-nowrap ${className}`}>
          <span className={C.text60}>担当者不明</span>
        </TableCell>
      );
    }
    return (
      <TableCell className={`${STYLE.tableCell} min-w-[160px] whitespace-nowrap ${className}`}>
        <span className={C.text}>{staff}（無効）</span>
      </TableCell>
    );
  }
  return (
    <TableCell className={`${STYLE.tableCell} min-w-[160px] whitespace-nowrap ${className}`}>
      <span className={C.text}>{staff}</span>
    </TableCell>
  );
}
