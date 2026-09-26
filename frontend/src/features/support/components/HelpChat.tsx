/**
 * HelpChat — LLM ヘルプチャット
 *
 * ユーザーの質問でマニュアルを検索し、上位記事の抜粋をコンテキストとして
 * POST /v1/support/chat に送信。回答は抜粋のみを根拠に生成され（グラウンディング）、
 * 参照記事が sources リンクとして表示される。
 */
import { useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router";
import { BookOpen, Loader2, Send } from "lucide-react";

import { paths } from "@/config/paths";
import { C } from "@/lib/design-tokens";
import { createManualSearcher, type ManualArticle } from "@/lib/manual-index";

import { useSendSupportChat } from "../api/send-support-chat";
import { buildChatContext } from "../lib/chat-context";
import type { SupportChatRole, SupportChatSource } from "../types";

/** 1回の送信でコンテキストに含めるマニュアル記事数 */
const CONTEXT_TOP_K = 3;
/** 各記事の抜粋文字数（バックエンド上限 6000 内で 3 件に収まるサイズ） */
const CONTEXT_EXCERPT_LENGTH = 1800;
/** 送信する会話履歴の最大件数（バックエンド上限 16 より余裕を持たせる） */
const HISTORY_MAX_MESSAGES = 12;

interface ChatTurn {
  role: SupportChatRole;
  content: string;
  sources?: SupportChatSource[];
  isError?: boolean;
}

interface HelpChatProps {
  articles: ManualArticle[];
  onClose: () => void;
}

export function HelpChat({ articles, onClose }: HelpChatProps) {
  const [turns, setTurns] = useState<ChatTurn[]>([]);
  const [input, setInput] = useState("");
  const search = useMemo(() => createManualSearcher(articles), [articles]);
  const send = useSendSupportChat();
  const listRef = useRef<HTMLDivElement>(null);

  // 新しいターンが追加されたら末尾へスクロール
  useEffect(() => {
    const el = listRef.current;
    if (el !== null) {
      el.scrollTop = el.scrollHeight;
    }
  }, [turns.length, send.isPending]);

  const handleSend = () => {
    const text = input.trim();
    if (text.length === 0 || send.isPending) return;

    const context = buildChatContext(search, text, CONTEXT_TOP_K, CONTEXT_EXCERPT_LENGTH);
    const history = turns
      .filter((t) => !t.isError)
      .slice(-HISTORY_MAX_MESSAGES)
      .map((t) => ({ role: t.role, content: t.content }));

    setTurns((prev) => [...prev, { role: "user", content: text }]);
    setInput("");
    send.mutate(
      { message: text, history, context },
      {
        onSuccess: (res) => {
          setTurns((prev) => [
            ...prev,
            { role: "assistant", content: res.reply, sources: res.sources },
          ]);
        },
        onError: () => {
          setTurns((prev) => [
            ...prev,
            {
              role: "assistant",
              content:
                "送信に失敗しました。時間をおいて再度お試しください。マニュアル検索で代わりに調べることもできます。",
              isError: true,
            },
          ]);
        },
      },
    );
  };

  return (
    <div className="flex flex-1 min-h-0 flex-col gap-2 p-3">
      <div
        ref={listRef}
        className="flex flex-1 min-h-48 flex-col gap-2 overflow-y-auto"
        aria-live="polite"
        aria-label="会話履歴"
      >
        {turns.length === 0 ? (
          <p className={`px-1 py-2 text-sm ${C.text50}`}>
            使い方を聞いてください。取扱説明書の内容をもとに回答します。
          </p>
        ) : (
          turns.map((turn, i) => (
            <div
              key={i}
              className={
                turn.role === "user"
                  ? `self-end max-w-[85%] rounded-lg rounded-br-xxs ${C.bgActionPrimary} ${C.textOnActionPrimary} px-3 py-2 text-sm whitespace-pre-wrap break-words`
                  : `self-start max-w-[85%] rounded-lg rounded-bl-xxs px-3 py-2 text-sm whitespace-pre-wrap break-words ${
                      turn.isError ? `border ${C.borderDanger} ${C.danger}` : `${C.bgMuted} ${C.text}`
                    }`
              }
            >
              {turn.content}
              {turn.sources !== undefined && turn.sources.length > 0 ? (
                <div className={`mt-2 flex flex-col gap-0.5 border-t ${C.borderLight} pt-1.5`}>
                  {turn.sources.map((s) => (
                    <Link
                      key={`${s.category}/${s.slug}`}
                      to={paths.manual.article.getHref(s.category as "screens" | "workflows", s.slug)}
                      onClick={onClose}
                      className={`flex items-center gap-1 text-2xs ${C.textActionPrimary} underline underline-offset-2`}
                    >
                      <BookOpen className="size-3 shrink-0" aria-hidden="true" />
                      {s.title}
                    </Link>
                  ))}
                </div>
              ) : null}
            </div>
          ))
        )}
        {send.isPending ? (
          <div
            className={`self-start flex items-center gap-2 rounded-lg px-3 py-2 text-sm ${C.bgMuted} ${C.textMuted}`}
          >
            <Loader2 className="size-4 animate-spin" aria-hidden="true" />
            回答を生成中…
          </div>
        ) : null}
      </div>

      <form
        className="flex items-center gap-2"
        onSubmit={(e) => {
          e.preventDefault();
          handleSend();
        }}
      >
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="例: レジ締めの手順は？"
          aria-label="使い方を質問"
          maxLength={2000}
          className={`flex-1 rounded-xxs border ${C.borderMedium} px-3 py-2 text-sm ${C.text} focus:outline-none ${C.focusBorderAccent}`}
        />
        <button
          type="submit"
          disabled={input.trim().length === 0 || send.isPending}
          aria-label="送信"
          className={`rounded-xxs ${C.bgActionPrimary} ${C.textOnActionPrimary} ${C.hoverBgActionPrimary} p-2 transition-colors disabled:opacity-40`}
        >
          <Send className="size-4" aria-hidden="true" />
        </button>
      </form>
    </div>
  );
}
