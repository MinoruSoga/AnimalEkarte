/**
 * HelpChat — LLM ヘルプチャット
 *
 * ユーザーの質問でマニュアルを検索し、上位記事の抜粋をコンテキストとして
 * POST /v1/support/chat に送信。回答は抜粋のみを根拠に生成され（グラウンディング）、
 * 参照記事が sources リンクとして表示される。
 *
 * 会話履歴は useHelpChatHistory（SupportWidget 常駐の provider）に保持され、
 * パネルの開閉・タブ切替・記事リンク遷移でアンマウントされても残る。
 * 送信成功したやり取りはサーバー（support_chat_messages）にも保存され、
 * 初回マウント時に GET /v1/support/chat/history で復元する。
 * 空状態では質問例チップを表示し、送信失敗ターンは「もう一度送信」で再送できる。
 */
import { useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router";
import { BookOpen, Loader2, RotateCcw, Send } from "lucide-react";

import { paths } from "@/config/paths";
import { C } from "@/lib/design-tokens";
import { createManualSearcher, type ManualArticle } from "@/lib/manual-index";

import { useClearSupportChatHistory } from "../api/clear-support-chat-history";
import { useGetSupportChatHistory } from "../api/get-support-chat-history";
import { useSendSupportChat } from "../api/send-support-chat";
import { buildChatContext } from "../lib/chat-context";
import { useHelpChatHistory } from "../lib/help-chat-history";
import type { SupportChatHistoryRecord, SupportChatTurn } from "../types";

/** 1回の送信でコンテキストに含めるマニュアル記事数 */
const CONTEXT_TOP_K = 3;
/** 各記事の抜粋文字数（バックエンド上限 6000 内で 3 件に収まるサイズ） */
const CONTEXT_EXCERPT_LENGTH = 1800;
/** 送信する会話履歴の最大件数（バックエンド上限 16 より余裕を持たせる） */
const HISTORY_MAX_MESSAGES = 12;

const SEND_ERROR_MESSAGE =
  "送信に失敗しました。時間をおいて再度お試しください。マニュアル検索で代わりに調べることもできます。";
const RESET_ERROR_MESSAGE = "履歴の削除に失敗しました。時間をおいて再度お試しください。";

/** 保存済みメッセージを表示用ターンに変換する（エラー/再送情報は保存対象外） */
function toHistoryTurn(record: SupportChatHistoryRecord): SupportChatTurn {
  return { role: record.role, content: record.content, sources: record.sources };
}

/** 空状態で提示する質問例（クリックでそのまま送信される） */
const SUGGESTED_QUESTIONS = ["レジ締めの手順は？", "予約の変更方法は？", "カルテの作成方法は？"];

interface HelpChatProps {
  articles: ManualArticle[];
  onClose: () => void;
}

export function HelpChat({ articles, onClose }: HelpChatProps) {
  const { turns, setTurns, hydrated, markHydrated } = useHelpChatHistory();
  const [input, setInput] = useState("");
  const [resetError, setResetError] = useState(false);
  const search = useMemo(() => createManualSearcher(articles), [articles]);
  const send = useSendSupportChat();
  const historyQuery = useGetSupportChatHistory();
  const clearHistory = useClearSupportChatHistory();
  const listRef = useRef<HTMLDivElement>(null);

  // 保存済み履歴を初回だけローカルへ反映する。ロード中に送信されたターンが
  // 既にある場合やロード失敗時はローカル状態を優先し、以後は再適用しない。
  useEffect(() => {
    if (hydrated) return;
    if (historyQuery.isSuccess) {
      if (turns.length === 0) {
        setTurns(historyQuery.data.map(toHistoryTurn));
      }
      markHydrated();
    } else if (historyQuery.isError) {
      markHydrated();
    }
  }, [
    hydrated,
    historyQuery.isSuccess,
    historyQuery.isError,
    historyQuery.data,
    turns.length,
    setTurns,
    markHydrated,
  ]);

  // 新しいターンが追加されたら末尾へスクロール
  useEffect(() => {
    const el = listRef.current;
    if (el !== null) {
      el.scrollTop = el.scrollHeight;
    }
  }, [turns.length, send.isPending]);

  /**
   * 質問を送信する。priorTurns は history 構築に使う「この質問より前のターン列」で、
   * 既定は現在の turns。再送時は失敗したやり取りを除いた列を渡す。
   */
  const sendMessage = (text: string, priorTurns: SupportChatTurn[] = turns) => {
    if (text.length === 0 || send.isPending || !hydrated) return;

    const context = buildChatContext(search, text, CONTEXT_TOP_K, CONTEXT_EXCERPT_LENGTH);
    const history = priorTurns
      .filter((t) => !t.isError)
      .slice(-HISTORY_MAX_MESSAGES)
      .map((t) => ({ role: t.role, content: t.content }));

    setTurns((prev) => [...prev, { role: "user", content: text }]);
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
              content: SEND_ERROR_MESSAGE,
              isError: true,
              retryMessage: text,
            },
          ]);
        },
      },
    );
  };

  const handleSend = () => {
    const text = input.trim();
    if (text.length === 0 || send.isPending || !hydrated) return;
    setInput("");
    sendMessage(text);
  };

  /**
   * サーバー側の履歴を削除してからローカルを空にする。失敗時はローカルを
   * 残したままエラーを表示する（見た目だけ消えると再ロードで復活するため）。
   */
  const handleReset = () => {
    if (clearHistory.isPending) return;
    clearHistory.mutate(undefined, {
      onSuccess: () => {
        setTurns([]);
        setResetError(false);
      },
      onError: () => setResetError(true),
    });
  };

  /**
   * 失敗した質問を送り直す。エラーターンと直前の失敗ユーザーターンを取り除き、
   * 同じ質問を末尾から新規送信する（history に失敗した質問が重複して残らない）。
   */
  const handleRetry = (errorIndex: number) => {
    const retryMessage = turns[errorIndex]?.retryMessage;
    if (retryMessage === undefined || send.isPending) return;
    const failedUserIndex = turns[errorIndex - 1]?.role === "user" ? errorIndex - 1 : -1;
    const remaining = turns.filter((_, i) => i !== errorIndex && i !== failedUserIndex);
    setTurns(remaining);
    sendMessage(retryMessage, remaining);
  };

  return (
    <div className="flex flex-1 min-h-0 flex-col gap-2 p-3">
      <div
        ref={listRef}
        className="flex flex-1 min-h-48 flex-col gap-2 overflow-y-auto"
        aria-live="polite"
        aria-label="会話履歴"
      >
        {!hydrated ? (
          <div
            className={`self-start flex items-center gap-2 rounded-lg px-3 py-2 text-sm ${C.bgMuted} ${C.textMuted}`}
          >
            <Loader2 className="size-4 animate-spin" aria-hidden="true" />
            履歴を読み込み中…
          </div>
        ) : turns.length === 0 ? (
          <div className="flex flex-col gap-3 px-1 py-2">
            <p className={`text-sm ${C.text50}`}>
              使い方を聞いてください。取扱説明書の内容をもとに回答します。
            </p>
            <div role="group" aria-label="質問の例" className="flex flex-wrap gap-1.5">
              {SUGGESTED_QUESTIONS.map((question) => (
                <button
                  key={question}
                  type="button"
                  disabled={send.isPending}
                  onClick={() => sendMessage(question)}
                  className={`rounded-full border ${C.borderLight} px-3 py-1.5 text-xs ${C.textActionPrimary} ${C.hoverBgLight} transition-colors disabled:opacity-40`}
                >
                  {question}
                </button>
              ))}
            </div>
          </div>
        ) : (
          // mt-auto スペーサーで会話を下端（入力欄直上）へ寄せる。
          // 件数が溢れた場合は margin が0になり上から通常スクロールする
          // （justify-end だと先頭メッセージがスクロール圏外に切れるため）
          <>
            <div className="mt-auto" aria-hidden="true" />
            {turns.map((turn, i) => (
              <div
                key={i}
                role={turn.isError === true ? "alert" : undefined}
                className={
                  turn.role === "user"
                    ? `self-end max-w-[85%] rounded-lg rounded-br-xxs ${C.bgActionPrimary} ${C.textOnActionPrimary} px-3 py-2 text-sm whitespace-pre-wrap break-words`
                    : `self-start max-w-[85%] rounded-lg rounded-bl-xxs px-3 py-2 text-sm whitespace-pre-wrap break-words ${
                        turn.isError === true
                          ? `border ${C.borderDanger} ${C.danger}`
                          : `${C.bgMuted} ${C.text}`
                      }`
                }
              >
                {turn.content}
                {turn.isError === true && turn.retryMessage !== undefined ? (
                  <button
                    type="button"
                    onClick={() => handleRetry(i)}
                    disabled={send.isPending}
                    className={`mt-1.5 flex items-center gap-1 text-xs font-semibold ${C.textActionPrimary} underline underline-offset-2 transition-opacity disabled:opacity-40`}
                  >
                    <RotateCcw className="size-3" aria-hidden="true" />
                    もう一度送信
                  </button>
                ) : null}
                {turn.sources !== undefined && turn.sources.length > 0 ? (
                  <div className={`mt-2 flex flex-col gap-0.5 border-t ${C.borderLight} pt-1.5`}>
                    {turn.sources.map((s) => (
                      <Link
                        key={`${s.category}/${s.slug}`}
                        to={paths.manual.article.getHref(
                          s.category as "screens" | "workflows",
                          s.slug,
                        )}
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
            ))}
          </>
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
        className="flex items-end gap-2"
        onSubmit={(e) => {
          e.preventDefault();
          handleSend();
        }}
      >
        <textarea
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => {
            // Enter で送信、Shift+Enter で改行。IME 変換中の確定 Enter は送信しない。
            if (e.key === "Enter" && !e.shiftKey && !e.nativeEvent.isComposing) {
              e.preventDefault();
              handleSend();
            }
          }}
          autoFocus
          rows={1}
          placeholder="例: レジ締めの手順は？"
          aria-label="使い方を質問"
          maxLength={2000}
          className={`field-sizing-content max-h-32 flex-1 resize-none rounded-xxs border ${C.borderMedium} px-3 py-2 text-sm ${C.text} focus:outline-none ${C.focusBorderAccent}`}
        />
        <button
          type="submit"
          disabled={input.trim().length === 0 || send.isPending || !hydrated}
          aria-label="送信"
          className={`rounded-xxs ${C.bgActionPrimary} ${C.textOnActionPrimary} ${C.hoverBgActionPrimary} p-2 transition-colors disabled:opacity-40`}
        >
          <Send className="size-4" aria-hidden="true" />
        </button>
      </form>
      <div className="flex items-end justify-between gap-2">
        <p className={`text-2xs ${C.textMuted}`}>
          患者・飼主の氏名など個人情報は入力しないでください(外部AIへ送信されます)。
        </p>
        {turns.length > 0 ? (
          <button
            type="button"
            onClick={handleReset}
            disabled={clearHistory.isPending}
            className={`shrink-0 text-2xs ${C.textMuted} ${C.hoverText} underline underline-offset-2 transition-colors disabled:opacity-40`}
          >
            会話をリセット
          </button>
        ) : null}
      </div>
      {resetError ? (
        <p role="alert" className={`text-2xs ${C.danger}`}>
          {RESET_ERROR_MESSAGE}
        </p>
      ) : null}
    </div>
  );
}
