import { C, BADGE } from "@/lib/design-tokens";
import { formatJSTDateTimeLocal } from "@/lib/jst-date";
import { useGetLineSendHistory } from "../api/get-line-send-history";
import type { LineSendHistoryItem } from "../api/get-line-send-history";

const TYPE_LABEL: Record<LineSendHistoryItem["message_type"], string> = {
  text: "テキスト",
  pdf_url: "PDF",
  image_url: "画像",
};

const STATUS_BADGE: Record<LineSendHistoryItem["status"], string> = {
  sent: BADGE.green,
  failed: BADGE.red,
  pending: BADGE.yellow,
};

const STATUS_LABEL: Record<LineSendHistoryItem["status"], string> = {
  sent: "送信済",
  failed: "失敗",
  pending: "送信中",
};

function formatSentAt(sentAt: string): string {
  return formatJSTDateTimeLocal(sentAt).slice(5).replace("T", " ");
}

interface LineSendHistoryProps {
  ownerId: string;
}

export function LineSendHistory({ ownerId }: LineSendHistoryProps) {
  const { data, isLoading, isError } = useGetLineSendHistory(ownerId);

  const recent = (data ?? []).slice(0, 5);

  if (isLoading) {
    return (
      <div className={`h-4 w-40 rounded ${C.bgSkeleton} animate-pulse`} />
    );
  }

  if (isError) {
    return (
      <p className={`text-xs ${C.danger}`}>送信履歴の取得に失敗しました</p>
    );
  }

  if (recent.length === 0) {
    return (
      <p className={`text-xs ${C.text40}`}>送信履歴はありません</p>
    );
  }

  return (
    <ul className="flex flex-col gap-1.5">
      {recent.map((item) => (
        <li
          key={item.id}
          className={`flex items-center gap-2 text-xs ${C.text70} px-2 py-1.5 rounded-xxs border ${C.borderLight}`}
        >
          <span className={`text-xs ${C.text50} font-mono shrink-0`}>
            {formatSentAt(item.sent_at)}
          </span>
          <span className={`shrink-0 ${C.text70}`}>
            {TYPE_LABEL[item.message_type]}
          </span>
          {item.content_summary !== null ? (
            <span className={`flex-1 truncate ${C.text60}`}>{item.content_summary}</span>
          ) : null}
          <span
            className={`shrink-0 text-xs px-1.5 py-0.5 rounded-xxs border font-medium ${STATUS_BADGE[item.status]}`}
          >
            {STATUS_LABEL[item.status]}
          </span>
        </li>
      ))}
    </ul>
  );
}
