/**
 * ManualContent — Markdown 本体レンダラー
 *
 * react-markdown + remark-gfm でテーブル/タスクリスト等の GFM 拡張を有効化。
 * 画像参照は ../content/images/ 配下に解決。Vite の URL 解決を使わず、
 * 公開ディレクトリ風の相対パスを許容するため、imgRenderer で書き換える。
 */

import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { memo } from "react";
import { Link } from "react-router";
import { ChevronLeft, ChevronRight } from "lucide-react";

import { C } from "@/lib/design-tokens";
import { paths } from "@/config/paths";

import {
  screenArticles,
  workflowArticles,
  type ManualArticle,
} from "../lib/manual-index";

export function getSafeMarkdownHref(href: unknown): string | undefined {
  if (typeof href !== "string") return undefined;
  const trimmed = href.trim();
  if (trimmed === "") return undefined;

  const decoded = (() => {
    try {
      return decodeURIComponent(trimmed);
    } catch {
      return trimmed;
    }
  })();

  // 制御文字（NULLバイト等）でスキーム判定を回避する攻撃を防ぐため意図的に制御文字を除去する
  // eslint-disable-next-line no-control-regex
  const compact = decoded.replace(/[\u0000-\u001F\u007F\s]+/g, "");
  const schemeMatch = /^([a-z][a-z0-9+.-]*):/i.exec(compact);
  if (schemeMatch) {
    const protocol = `${schemeMatch[1].toLowerCase()}:`;
    return protocol === "http:" || protocol === "https:" || protocol === "mailto:"
      ? trimmed
      : undefined;
  }

  if (trimmed.startsWith("//")) {
    return undefined;
  }

  return trimmed;
}

interface ManualContentProps {
  article: ManualArticle | undefined;
}

/** 同カテゴリ内の前後記事を取得 */
function getAdjacent(article: ManualArticle): {
  prev: ManualArticle | undefined;
  next: ManualArticle | undefined;
} {
  const list = article.category === "screens" ? screenArticles : workflowArticles;
  const idx = list.findIndex(
    (a) => a.category === article.category && a.slug === article.slug,
  );
  if (idx < 0) return { prev: undefined, next: undefined };
  return {
    prev: idx > 0 ? list[idx - 1] : undefined,
    next: idx < list.length - 1 ? list[idx + 1] : undefined,
  };
}

// Vite 静的 import の URL マッピング。バンドル時にハッシュ付き URL に置換される。
const imageUrls = import.meta.glob("../content/images/*", {
  eager: true,
  query: "?url",
  import: "default",
}) as Record<string, string>;

function resolveImageSrc(src: string | undefined): string | undefined {
  if (!src) return undefined;
  // 既に絶対 URL or data URL なら素通し
  if (/^(https?:|data:|\/)/.test(src)) return src;
  // `images/foo.png` or `./images/foo.png` を ../content/images/ に正規化
  const name = src.replace(/^\.?\/?images\//, "");
  const key = `../content/images/${name}`;
  return imageUrls[key] ?? src;
}

// react-markdown の components オーバーライドはレンダー毎に生成する必要がない
// モジュール定数（js-cache-function-results / rerender-memo 対策）。
const MARKDOWN_COMPONENTS: Parameters<typeof ReactMarkdown>[0]["components"] = {
  h1: ({ children }) => (
    <h1 className={`text-heading-2 font-bold mb-6 pb-3 border-b ${C.borderDivider} ${C.text}`}>
      {children}
    </h1>
  ),
  h2: ({ children }) => (
    <h2 className={`text-heading-3 font-bold mt-8 mb-3 ${C.text}`}>{children}</h2>
  ),
  h3: ({ children }) => (
    <h3 className={`text-xl font-semibold mt-6 mb-2 ${C.text}`}>{children}</h3>
  ),
  h4: ({ children }) => (
    <h4 className={`text-base font-semibold mt-4 mb-2 ${C.text}`}>{children}</h4>
  ),
  p: ({ children }) => (
    <p className={`my-3 leading-relaxed ${C.text}`}>{children}</p>
  ),
  ul: ({ children }) => (
    <ul className={`list-disc pl-6 my-3 space-y-1 ${C.text}`}>{children}</ul>
  ),
  ol: ({ children }) => (
    <ol className={`list-decimal pl-6 my-3 space-y-1 ${C.text}`}>{children}</ol>
  ),
  li: ({ children }) => <li className="leading-relaxed">{children}</li>,
  a: ({ href, children }) => {
    const safeHref = getSafeMarkdownHref(href);
    if (!safeHref) {
      return (
        <span className={`underline ${C.textBrand} hover:opacity-70`}>{children}</span>
      );
    }
    return (
      <a
        href={safeHref}
        target="_blank"
        rel="noopener noreferrer"
        className={`underline ${C.textBrand} hover:opacity-70`}
      >
        {children}
      </a>
    );
  },
  table: ({ children }) => (
    <div className="my-4 overflow-x-auto">
      <table className={`w-full text-sm border ${C.borderDivider}`}>{children}</table>
    </div>
  ),
  thead: ({ children }) => (
    <thead className="bg-black/5">{children}</thead>
  ),
  th: ({ children }) => (
    <th className={`px-3 py-2 text-left font-semibold border ${C.borderDivider} ${C.text}`}>
      {children}
    </th>
  ),
  td: ({ children }) => (
    <td className={`px-3 py-2 border ${C.borderDivider} ${C.text65}`}>{children}</td>
  ),
  blockquote: ({ children }) => (
    <blockquote
      className={`my-4 pl-4 border-l-2 italic ${C.text65} ${C.borderLBrand}`}
    >
      {children}
    </blockquote>
  ),
  code: ({ className, children }) => {
    const isInline = !className;
    if (isInline) {
      return (
        <code
          className={`px-1.5 py-0.5 rounded-xxs text-[0.875em] ${C.text} ${C.bgHoverMd}`}
        >
          {children}
        </code>
      );
    }
    return (
      <code className={`block ${className ?? ""}`}>{children}</code>
    );
  },
  pre: ({ children }) => (
    <pre
      className={`my-4 p-4 rounded-xxs overflow-x-auto text-sm ${C.text} ${C.bgPrimary10}`}
    >
      {children}
    </pre>
  ),
  img: ({ src, alt }) => (
    <img
      src={resolveImageSrc(typeof src === "string" ? src : undefined)}
      alt={alt ?? ""}
      className={`my-4 rounded-xxs border ${C.borderDivider} max-w-full`}
      loading="lazy"
    />
  ),
  hr: () => <hr className={`my-6 border-t ${C.borderDivider}`} />,
};

/** 前後記事へのナビゲーション（同カテゴリ内の隣接記事） */
const ArticleAdjacentNav = memo(function ArticleAdjacentNav({
  article,
}: {
  article: ManualArticle;
}) {
  const { prev, next } = getAdjacent(article);
  if (!prev && !next) return null;
  return (
    <nav
      aria-label="前後の項目"
      className={`mt-12 pt-6 border-t ${C.borderDivider} grid grid-cols-1 md:grid-cols-2 gap-3 no-print`}
    >
      {prev ? (
        <Link
          to={paths.manual.article.getHref(prev.category, prev.slug)}
          className={`flex items-center gap-2 p-3 rounded-xxs border ${C.borderDivider} ${C.hoverBgLight} transition-colors`}
        >
          <ChevronLeft className={`size-5 shrink-0 ${C.text50}`} />
          <div className="min-w-0">
            <div className={`text-2xs ${C.text40} uppercase`}>前の項目</div>
            <div className={`truncate ${C.text}`}>{prev.title}</div>
          </div>
        </Link>
      ) : (
        <div />
      )}
      {next ? (
        <Link
          to={paths.manual.article.getHref(next.category, next.slug)}
          className={`flex items-center gap-2 p-3 rounded-xxs border ${C.borderDivider} ${C.hoverBgLight} transition-colors text-right md:justify-end`}
        >
          <div className="min-w-0 flex-1">
            <div className={`text-2xs ${C.text40} uppercase`}>次の項目</div>
            <div className={`truncate ${C.text}`}>{next.title}</div>
          </div>
          <ChevronRight className={`size-5 shrink-0 ${C.text50}`} />
        </Link>
      ) : (
        <div />
      )}
    </nav>
  );
});

export const ManualContent = memo(function ManualContent({ article }: ManualContentProps) {
  if (!article) {
    return (
      <div className="flex-1 flex items-center justify-center p-8">
        <p className={C.text50}>左メニューからマニュアル項目を選択してください</p>
      </div>
    );
  }

  return (
    <article
      className={`flex-1 overflow-y-auto px-4 md:px-8 py-6 max-w-[860px] mx-auto manual-article pt-16 md:pt-6 ${C.bgWhite}`}
    >
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={MARKDOWN_COMPONENTS}>
        {article.content}
      </ReactMarkdown>

      <ArticleAdjacentNav article={article} />
    </article>
  );
});
