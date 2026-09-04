import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { C } from "@/lib/design-tokens";
import { formatJSTDateTimeLocal } from "@/lib/jst-date";
import { formatCurrency } from "@/lib/format/number";
import type { CashRegisterClose } from "../api/get-cash-register-closes";
import { PERIOD_LABELS } from "../lib/constants";
import { summarizeCategoryTotals } from "../lib/category-breakdown";
import { diffClass, formatDiff } from "../lib/cash-register-history-model";

interface CashRegisterHistoryDetailDialogProps {
  selectedClose: CashRegisterClose | null;
  onOpenChange: (open: boolean) => void;
}

export function CashRegisterHistoryDetailDialog({
  selectedClose,
  onOpenChange,
}: CashRegisterHistoryDetailDialogProps) {
  const detailSubtotals = selectedClose
    ? summarizeCategoryTotals(selectedClose.categoryBreakdown)
    : [];
  const detailDiff = selectedClose
    ? (selectedClose.actualCash ?? 0) - (selectedClose.theoreticalCash ?? 0)
    : 0;

  return (
    <Dialog open={selectedClose !== null} onOpenChange={onOpenChange}>
      <DialogContent>
        {selectedClose ? (
          <>
            <DialogHeader>
              <DialogTitle>
                {selectedClose.closeDate.slice(0, 10)} {PERIOD_LABELS[selectedClose.period]}{" "}
                の締め詳細
              </DialogTitle>
              <DialogDescription>
                この締めレコードの集計内訳と差額を表示しています。
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4">
              <dl className="grid grid-cols-2 gap-x-4 gap-y-2 text-base">
                <dt className={C.text60}>理論現金</dt>
                <dd className={`text-right ${C.text}`}>
                  {formatCurrency(selectedClose.theoreticalCash ?? 0)}
                </dd>
                <dt className={C.text60}>実際の現金</dt>
                <dd className={`text-right ${C.text}`}>
                  {formatCurrency(selectedClose.actualCash ?? 0)}
                </dd>
                <dt className={C.text60}>差額</dt>
                <dd className={`text-right font-medium ${diffClass(detailDiff)}`}>
                  {formatDiff(detailDiff)}
                </dd>
                <dt className={C.text60}>担当者</dt>
                <dd className={`text-right ${C.text}`}>{selectedClose.closedByStaffName ?? "—"}</dd>
                <dt className={C.text60}>締め日時</dt>
                <dd className={`text-right ${C.text}`}>
                  {selectedClose.closedAt
                    ? formatJSTDateTimeLocal(selectedClose.closedAt).replace("T", " ")
                    : "—"}
                </dd>
              </dl>

              <div>
                <p className={`text-sm font-medium ${C.text70} mb-1`}>メモ</p>
                <p className={`text-base ${C.text}`}>
                  {selectedClose.memo ? selectedClose.memo : "—"}
                </p>
              </div>

              <div>
                <p className={`text-sm font-medium ${C.text70} mb-2`}>部門別内訳</p>
                {detailSubtotals.length > 0 ? (
                  <dl className="space-y-1">
                    {detailSubtotals.map((row) => (
                      <div key={row.label} className="flex justify-between text-base">
                        <dt className={C.text60}>{row.label}</dt>
                        <dd className={C.text}>
                          {formatCurrency(row.total)}
                          {row.count !== undefined ? (
                            <span className={`ml-2 ${C.text60}`}>
                              {row.count === null ? "記録なし" : `${row.count}件`}
                            </span>
                          ) : null}
                        </dd>
                      </div>
                    ))}
                  </dl>
                ) : (
                  <p className={`text-base ${C.text50}`}>内訳データなし</p>
                )}
              </div>
            </div>
          </>
        ) : null}
      </DialogContent>
    </Dialog>
  );
}
