/**
 * ManualHelpTab — 「使い方を聞く」タブ
 *
 * LLM チャットが有効な環境（SUPPORT_LLM_API_KEY 設定済み）では会話 UI、
 * 無効な環境では Fuse.js による検索型 UI を表示する。
 * どちらも共有のマニュアル索引（@/lib/manual-index）+ DB オーバーライドを使う。
 */
import { useDeferredValue, useLayoutEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router";
import { ChevronRight, Search } from "lucide-react";

import { EmptyState } from "@/components/shared/DataStates";
import { paths } from "@/config/paths";
import { useGetManualArticleOverrides } from "@/hooks/use-manual-article-overrides";
import { useManualSearch } from "@/hooks/use-manual-search";
import { usePermission } from "@/hooks/use-permission";
import { C, STYLE } from "@/lib/design-tokens";
import {
  applyOverrides,
  screenArticles,
  workflowArticles,
  type ManualArticle,
} from "@/lib/manual-index";

import { useGetSupportChatStatus } from "../api/get-support-chat-status";

import { HelpChat } from "./HelpChat";

const MAX_RESULTS = 8;

const CATEGORY_LABEL = { screens: "画面別", workflows: "業務フロー" } as const;

interface ManualHelpTabProps {
  onClose: () => void;
}

export function ManualHelpTab({ onClose }: ManualHelpTabProps) {
  const { data: chatStatus } = useGetSupportChatStatus();
  // DB オーバーライドは manual-edit 権限保持者のみ取得（ManualPage と同じ条件）
  const { canView: canViewOverrides } = usePermission("manual-edit");
  const {
    data: fetchedOverrides,
    refetch: refetchOverrides,
    isPending: overridesPending,
    isError: overridesError,
  } = useGetManualArticleOverrides(canViewOverrides, true);
  const overrides = canViewOverrides ? fetchedOverrides : undefined;

  const articles = useMemo(
    () => [
      ...applyOverrides(
        screenArticles,
        (overrides ?? []).filter((o) => o.category === "screens"),
      ),
      ...applyOverrides(
        workflowArticles,
        (overrides ?? []).filter((o) => o.category === "workflows"),
      ),
    ],
    [overrides],
  );

  // EMR-410: 409（manual_revision 不一致）時に overrides を再取得して
  // 最新の実効 snapshot を返す。権限チェックを再評価し、refetch 失敗や
  // data 欠落は throw する（stale base で回答しない）。
  const canViewOverridesRef = useRef(canViewOverrides);
  const overridesPendingRef = useRef(overridesPending);
  const overridesErrorRef = useRef(overridesError);
  const articlesRef = useRef<ManualArticle[]>([]);
  useLayoutEffect(() => {
    canViewOverridesRef.current = canViewOverrides;
    overridesPendingRef.current = overridesPending;
    overridesErrorRef.current = overridesError;
    articlesRef.current = articles;
  }, [canViewOverrides, overridesPending, overridesError, articles]);
  const refetchArticles = async (): Promise<ManualArticle[]> => {
    const initiallyPermitted = canViewOverridesRef.current;
    const res = initiallyPermitted ? await refetchOverrides() : { data: undefined, isError: false };
    // await 後に権限を再評価。剥奪されたら base のみ返し、付与後初回は stale 回避で throw。
    if (!canViewOverridesRef.current) {
      return [...screenArticles, ...workflowArticles];
    }
    if (!initiallyPermitted) {
      throw new Error("permission granted during refetch; retry required");
    }
    if (res.isError || res.data === undefined) {
      throw new Error("failed to refetch manual overrides");
    }
    const next = res.data;
    return [
      ...applyOverrides(
        screenArticles,
        next.filter((o) => o.category === "screens"),
      ),
      ...applyOverrides(
        workflowArticles,
        next.filter((o) => o.category === "workflows"),
      ),
    ];
  };

  // 権限ありで strict 取得が pending/error の間は stale base で回答させない。
  const canSubmit =
    !canViewOverrides || (!overridesPending && !overridesError && fetchedOverrides !== undefined);

  const getCurrentArticles = (): ManualArticle[] => {
    if (canViewOverridesRef.current && (overridesPendingRef.current || overridesErrorRef.current)) {
      throw new Error("manual overrides are not ready");
    }
    return articlesRef.current;
  };

  if (chatStatus?.enabled) {
    return (
      <HelpChat
        // 権限境界の変化で旧スナップショット由来の進行中送信を必ず中断する。
        key={canViewOverrides ? "overrides" : "base"}
        articles={articles}
        onClose={onClose}
        onRevisionConflict={refetchArticles}
        canSubmit={canSubmit}
        getCurrentArticles={getCurrentArticles}
      />
    );
  }
  // 非 chat path も bundled 記事へ黙って fallback しない（ManualPage fix-9b と同じ
  // wording）。data の有無ではなく query state で gate する。
  if (canViewOverrides && overridesPending) {
    return (
      <div className="flex-1 flex items-center justify-center p-8">
        <EmptyState message="マニュアル記事を読み込んでいます" />
      </div>
    );
  }
  if (canViewOverrides && overridesError) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center gap-4 p-8">
        <EmptyState
          message="マニュアル記事の取得に失敗しました"
          description="通信状態を確認して再試行してください"
        />
        <button
          type="button"
          onClick={() => refetchOverrides()}
          className={`rounded-xxs border px-4 py-2 ${C.borderDivider} ${C.bgWhite} ${C.hoverBgLight}`}
        >
          再試行
        </button>
      </div>
    );
  }
  return <ManualSearchView articles={articles} onClose={onClose} />;
}

function ManualSearchView({
  articles,
  onClose,
}: {
  articles: ManualArticle[];
  onClose: () => void;
}) {
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const results = useManualSearch(deferredQuery, articles);
  const isSearching = deferredQuery.trim().length > 0;
  const visible = isSearching ? results.slice(0, MAX_RESULTS) : [];

  return (
    <div className="flex flex-col gap-3 p-3">
      <p className={`text-sm ${C.text60}`}>
        知りたい操作や画面名を入力すると、取扱説明書の関連項目を表示します。
      </p>
      <div className="relative">
        <Search className={STYLE.searchIcon} aria-hidden="true" />
        <input
          type="search"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="例: 会計、予約、カルテ"
          aria-label="マニュアルを検索"
          className={STYLE.searchInput}
        />
      </div>

      {isSearching ? (
        visible.length > 0 ? (
          <ul className="flex flex-col gap-1" aria-label="検索結果">
            {visible.map((article) => (
              <li key={`${article.category}/${article.slug}`}>
                <Link
                  to={paths.manual.article.getHref(article.category, article.slug)}
                  onClick={onClose}
                  className={`flex items-center gap-2 rounded-xxs px-2 py-2 ${C.hoverBgLight} transition-colors`}
                >
                  <span
                    className={`shrink-0 rounded-xxs px-1.5 py-0.5 text-2xs ${C.bgMuted} ${C.textMuted}`}
                  >
                    {CATEGORY_LABEL[article.category]}
                  </span>
                  <span className={`flex-1 min-w-0 truncate text-sm ${C.text}`}>
                    {article.title}
                  </span>
                  <ChevronRight className={`size-4 shrink-0 ${C.text35}`} aria-hidden="true" />
                </Link>
              </li>
            ))}
          </ul>
        ) : (
          <p className={`px-1 py-4 text-sm ${C.text50}`}>
            該当する項目が見つかりませんでした。別のキーワードを試すか、
            <Link
              to={paths.manual.getHref()}
              onClick={onClose}
              className={`${C.textActionPrimary} underline underline-offset-2`}
            >
              マニュアル
            </Link>
            を直接ご覧ください。
          </p>
        )
      ) : (
        <p className={`px-1 text-2xs ${C.text60}`}>
          ヒント: 「レジ締め」「予防接種」「在庫」などのキーワードで検索できます。
        </p>
      )}
    </div>
  );
}
