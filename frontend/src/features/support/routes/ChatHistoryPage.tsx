/**
 * ChatHistoryPage — ヘルプチャット履歴の会話ルーム一覧ページ（全スタッフ・全医院に公開）
 *
 * /settings/chat-history（権限ゲートなし — バグ報告ボードと同じ全医院共有の
 * 製品判断。質問傾向の横断分析が目的）。
 * 一般的なチャットアプリ型の2ペイン構成: 左サイドバーに会話ルーム一覧、
 * 右ペインに選択中ルームの会話を吹き出しで表示する。
 * ルームは DB 上の境界ではなく、サーバーが同一 clinic×staff の発言ストリームを
 * 一定の無発言時間で区切って導出した疑似セッション（最終発言の新しい順）。
 */
import { useState } from "react";
import { Link } from "react-router";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { MessagesSquare } from "lucide-react";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { EmptyState, ErrorFallback, LoadingFallback } from "@/components/shared/DataStates";
import { paths } from "@/config/paths";
import { C, ICON } from "@/lib/design-tokens";
import { formatJSTDate, formatJSTTime } from "@/lib/jst-date";

import { useGetSupportChatRooms } from "../api/get-support-chat-rooms";
import { CHAT_MARKDOWN_COMPONENTS } from "../lib/chat-markdown-components";
import type { SupportChatRoom, SupportChatRoomTurn } from "../types";

/** ルーム一覧のラベルに使う先頭の質問（user 発言） */
function firstUserQuestion(room: SupportChatRoom): string {
  const firstUserTurn = room.turns.find((turn) => turn.role === "user");
  return firstUserTurn?.content ?? room.turns[0]?.content ?? "";
}

function RoomListItem({
  room,
  isSelected,
  onSelect,
}: {
  room: SupportChatRoom;
  isSelected: boolean;
  onSelect: (id: string) => void;
}) {
  return (
    <button
      type="button"
      onClick={() => onSelect(room.id)}
      aria-current={isSelected}
      className={`w-full rounded-md border p-3 text-left transition-colors ${
        isSelected
          ? `${C.borderActionPrimary} ${C.bgActionPrimaryLight}`
          : `${C.borderLight} ${C.hoverBgPage60}`
      }`}
    >
      <p className={`line-clamp-1 text-sm font-medium ${C.text}`}>{firstUserQuestion(room)}</p>
      <p className={`mt-1 text-xs ${C.text50}`}>
        {room.clinic_name || "―"} ・ {room.staff_name || "―"}
      </p>
      <p className={`mt-0.5 text-xs ${C.text50}`}>
        {formatJSTDate(room.last_message_at)} {formatJSTTime(room.last_message_at)} ・{" "}
        {room.turns.length}件の発言
      </p>
    </button>
  );
}

function RoomTurnBubble({ turn }: { turn: SupportChatRoomTurn }) {
  const isUser = turn.role === "user";
  return (
    <li className={`flex flex-col ${isUser ? "items-end" : "items-start"}`}>
      <div
        className={
          isUser
            ? `max-w-[85%] rounded-lg rounded-br-xxs ${C.bgActionPrimarySolid} ${C.textOnActionPrimary} px-3 py-2 text-sm whitespace-pre-wrap break-words`
            : `max-w-[85%] rounded-lg rounded-bl-xxs ${C.bgMuted} ${C.text} px-3 py-2 text-sm break-words`
        }
      >
        {isUser ? (
          turn.content
        ) : (
          <ReactMarkdown remarkPlugins={[remarkGfm]} components={CHAT_MARKDOWN_COMPONENTS}>
            {turn.content}
          </ReactMarkdown>
        )}
        {turn.sources !== undefined && turn.sources.length > 0 ? (
          <div className={`mt-2 flex flex-col gap-0.5 border-t ${C.borderLight} pt-1.5`}>
            {turn.sources.map((source) => (
              <Link
                key={`${source.category}/${source.slug}`}
                to={paths.manual.article.getHref(
                  source.category as "screens" | "workflows",
                  source.slug,
                )}
                className={`flex items-center gap-1 text-2xs ${C.textActionPrimary} underline underline-offset-2`}
              >
                {source.title}
              </Link>
            ))}
          </div>
        ) : null}
      </div>
      <time dateTime={turn.created_at} className={`mt-0.5 text-2xs ${C.text50}`}>
        {formatJSTTime(turn.created_at)}
      </time>
    </li>
  );
}

export function ChatHistoryPage() {
  const { data: rooms, isLoading, isError } = useGetSupportChatRooms();
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const selectedRoom = rooms?.find((room) => room.id === selectedId) ?? rooms?.[0];

  return (
    <PageLayout
      title="チャット履歴"
      description="ヘルプチャットの会話履歴です（全医院・全スタッフに公開。同一スタッフの連続した発言を会話単位にまとめて表示します）"
      icon={<MessagesSquare className={`${ICON.page} ${C.text}`} />}
      maxWidth="max-w-6xl"
    >
      {isLoading ? (
        <LoadingFallback />
      ) : isError ? (
        <ErrorFallback message="チャット履歴の取得に失敗しました。時間をおいて再度お試しください。" />
      ) : rooms === undefined || rooms.length === 0 ? (
        <EmptyState message="チャット履歴はまだありません" />
      ) : (
        <div className="flex flex-col gap-4 md:flex-row">
          <nav aria-label="チャットルーム一覧" className="md:w-80 md:shrink-0">
            <ul className="flex flex-col gap-1.5">
              {rooms.map((room) => (
                <li key={room.id}>
                  <RoomListItem
                    room={room}
                    isSelected={room.id === selectedRoom?.id}
                    onSelect={setSelectedId}
                  />
                </li>
              ))}
            </ul>
          </nav>

          {selectedRoom !== undefined ? (
            <section
              aria-label="選択中の会話"
              className={`min-w-0 flex-1 rounded-md border ${C.borderLight}`}
            >
              <header className={`border-b ${C.borderLight} px-4 py-3`}>
                <p className={`text-sm font-medium ${C.text}`}>
                  {selectedRoom.clinic_name || "―"} ・ {selectedRoom.staff_name || "―"}
                </p>
                <p className={`mt-0.5 text-xs ${C.text50}`}>
                  {formatJSTDate(selectedRoom.started_at)} {formatJSTTime(selectedRoom.started_at)}{" "}
                  開始 ・ {selectedRoom.turns.length}件の発言
                </p>
              </header>
              <ol className="flex max-h-[70vh] flex-col gap-3 overflow-y-auto p-4">
                {selectedRoom.turns.map((turn) => (
                  <RoomTurnBubble key={turn.id} turn={turn} />
                ))}
              </ol>
            </section>
          ) : null}
        </div>
      )}
    </PageLayout>
  );
}
