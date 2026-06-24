import { memo } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { C } from "@/lib/design-tokens";
import { useGetOwnerUnpaidBalance } from "../api/get-owner-unpaid-balance";

interface OwnerUnpaidBalanceCardProps {
  ownerId: string;
}

/**
 * #182: 会計画面に飼主の未納残高（未入金 billing の合計）を表示する。
 *
 * レジがカルテシステムと非連動の環境で、会計時に「この飼主は他にも未納がある」ことを
 * 把握できるようにする。残高定義は未納者一覧(#120)/月次繰越(#114)と同一
 * （status=waiting の billings.total_amount 合計）で、各画面の値が乖離しない。
 *
 * 未納が無い／取得失敗／読み込み中は何も表示しない（情報過多を避ける）。
 */
export const OwnerUnpaidBalanceCard = memo(function OwnerUnpaidBalanceCard({
  ownerId,
}: OwnerUnpaidBalanceCardProps) {
  const { data, isLoading, isError } = useGetOwnerUnpaidBalance(ownerId);

  if (isLoading || isError || !data || data.unpaidCount === 0) return null;

  return (
    <Card data-testid="owner-unpaid-balance">
      <CardContent className="p-4 flex items-center justify-between">
        <div className="flex flex-col gap-0.5">
          <span className={`text-sm font-medium ${C.text60}`}>未納残高</span>
          <span className={`text-xs ${C.text40}`}>未入金 {data.unpaidCount} 件</span>
        </div>
        <span className={`text-xl font-bold ${C.danger}`}>
          ¥{data.unpaidTotal.toLocaleString()}
        </span>
      </CardContent>
    </Card>
  );
});
