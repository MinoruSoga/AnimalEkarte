/**
 * ManualHelpTab — 「使い方を聞く」タブ
 *
 * LLM チャットが有効な環境（SUPPORT_LLM_API_KEY 設定済み）では会話 UI、
 * 無効な環境では Fuse.js による検索型 UI を表示する。
 * どちらも共有のマニュアル索引（@/lib/manual-index）+ DB オーバーライドを使う。
 */
import { useDeferredValue, useMemo, useState } from "react";
import { Link } from "react-router";
import { ChevronRight, Search } from "lucide-react";

import { paths } from "@/config/paths";
import { useGetManualArticleOverrides } from "@/hooks/use-manual-article-overrides";
import { useManualSearch } from "@/hooks/use-manual-search";
import { usePermission } from "@/hooks/use-permission";
import { C, STYLE } from "@/lib/design-tokens";
import { applyOverrides, screenArticles, workflowArticles, type ManualArticle } from "@/lib/manual-index";

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
  const { data: fetchedOverrides } = useGetManualArticleOverrides(canViewOverrides);
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

  if (chatStatus?.enabled) {
    return <HelpChat articles={articles} onClose={onClose} />;
  }
  return <ManualSearchView articles={articles} onClose={onClose} />;
}

function ManualSearchView({ articles, onClose }: { articles: ManualArticle[]; onClose: () => void }) {
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
                  <span className={`shrink-0 rounded-xxs px-1.5 py-0.5 text-2xs ${C.bgMuted} ${C.textMuted}`}>
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
        <p className={`px-1 text-2xs ${C.text40}`}>
          ヒント: 「レジ締め」「予防接種」「在庫」などのキーワードで検索できます。
        </p>
      )}
    </div>
  );
}
