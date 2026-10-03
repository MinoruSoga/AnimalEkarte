/**
 * ChatHistoryPage — ヘルプチャットの質問+回答ペア一覧ページ（全スタッフ・全医院に公開）
 *
 * /settings/chat-history（権限ゲートなし — バグ報告ボードと同じ全医院共有の
 * 製品判断。質問傾向の横断分析が目的）。一覧はペアを新しい順で返し、
 * 医院/スタッフ列で provenance を識別できる。
 * 行クリックで詳細ダイアログ（質問・回答全文と参照したマニュアル記事）。
 */
import { useState } from "react";
import { MessagesSquare } from "lucide-react";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { EmptyState, ErrorFallback, LoadingFallback } from "@/components/shared/DataStates";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { formatJSTDate, formatJSTTime } from "@/lib/jst-date";

import { useGetSupportChatExchanges } from "../api/get-support-chat-exchanges";
import type { SupportChatExchange } from "../types";

interface ChatExchangeDetailDialogProps {
  exchange: SupportChatExchange | null;
  onClose: () => void;
}

function ChatExchangeDetailDialog({ exchange, onClose }: ChatExchangeDetailDialogProps) {
  return (
    <Dialog open={exchange !== null} onOpenChange={(open) => (open ? undefined : onClose())}>
      {exchange !== null ? (
        <DialogContent className="sm:max-w-2xl max-h-[85vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className={`text-base ${C.text}`}>チャット履歴</DialogTitle>
            <DialogDescription className="text-left">
              {exchange.clinic_name || "―"} ・ {exchange.staff_name || "―"} ・{" "}
              {formatJSTDate(exchange.created_at)} {formatJSTTime(exchange.created_at)}
            </DialogDescription>
          </DialogHeader>

          <dl className="grid grid-cols-[96px_1fr] gap-x-3 gap-y-1.5 text-sm">
            <dt className={C.text50}>質問</dt>
            <dd className={`whitespace-pre-wrap break-words ${C.text}`}>{exchange.question}</dd>
            <dt className={C.text50}>回答</dt>
            <dd className={`whitespace-pre-wrap break-words ${C.text70}`}>{exchange.answer}</dd>
            {exchange.sources && exchange.sources.length > 0 ? (
              <>
                <dt className={C.text50}>参照記事</dt>
                <dd className={C.text70}>
                  <ul className="list-disc pl-4 space-y-0.5">
                    {exchange.sources.map((s) => (
                      <li key={`${s.category}-${s.slug}`}>
                        {s.title}
                        <span className={C.text50}>（{s.category}）</span>
                      </li>
                    ))}
                  </ul>
                </dd>
              </>
            ) : null}
          </dl>

          <div className="flex justify-end pt-1">
            <Button variant="outline" size="sm" onClick={onClose}>
              閉じる
            </Button>
          </div>
        </DialogContent>
      ) : (
        <DialogContent className="sm:max-w-2xl" aria-label="チャット履歴の詳細" />
      )}
    </Dialog>
  );
}

export function ChatHistoryPage() {
  const { data: exchanges, isLoading, isError } = useGetSupportChatExchanges();
  const [selected, setSelected] = useState<SupportChatExchange | null>(null);

  return (
    <PageLayout
      title="チャット履歴"
      description="ヘルプチャットで送信された質問と回答の一覧です（全医院・全スタッフに公開）"
      icon={<MessagesSquare className={`${ICON.page} ${C.text}`} />}
      maxWidth="max-w-5xl"
    >
      {isLoading ? (
        <LoadingFallback />
      ) : isError ? (
        <ErrorFallback message="チャット履歴の取得に失敗しました。時間をおいて再度お試しください。" />
      ) : exchanges === undefined || exchanges.length === 0 ? (
        <EmptyState message="チャット履歴はまだありません" />
      ) : (
        <div className={STYLE.tableContainer}>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[130px]">日時</TableHead>
                <TableHead className="w-[110px]">医院</TableHead>
                <TableHead className="w-[110px]">スタッフ</TableHead>
                <TableHead>質問</TableHead>
                <TableHead>回答</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {exchanges.map((exchange) => (
                <TableRow key={exchange.id} className={`${C.hoverBgPageHalf} h-14`}>
                  <TableCell className={C.text70}>
                    {formatJSTDate(exchange.created_at)} {formatJSTTime(exchange.created_at)}
                  </TableCell>
                  <TableCell className={C.text70}>{exchange.clinic_name || "―"}</TableCell>
                  <TableCell className={C.text70}>{exchange.staff_name || "―"}</TableCell>
                  <TableCell>
                    <DataTableRowButton
                      aria-label={`詳細: ${exchange.question}`}
                      onClick={() => setSelected(exchange)}
                    >
                      <span className="line-clamp-2">{exchange.question}</span>
                    </DataTableRowButton>
                  </TableCell>
                  <TableCell className={C.text60}>
                    <span className="line-clamp-2">{exchange.answer}</span>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      <ChatExchangeDetailDialog exchange={selected} onClose={() => setSelected(null)} />
    </PageLayout>
  );
}
