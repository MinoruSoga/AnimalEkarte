/**
 * chat-markdown-components.tsx — チャット吹き出し用の react-markdown components 上書き
 *
 * アシスタント回答はマークダウンで返るため、吹き出しのサイズに合わせたコンパクトな
 * components 上書きを HelpChat（送信画面）と ChatHistoryPage（履歴閲覧）で共有する。
 * ManualContent とは別のサイズ体系。モジュール定数として保持（rerender 対策）。
 * react-refresh/only-export-components のため component ファイルではなく lib/ に置く。
 */
import type ReactMarkdown from "react-markdown";

import { C, STYLE } from "@/lib/design-tokens";
import { getSafeMarkdownHref } from "@/lib/safe-markdown-href";

export const CHAT_MARKDOWN_COMPONENTS: Parameters<typeof ReactMarkdown>[0]["components"] = {
  h1: ({ children }) => <p className="font-semibold mt-2 mb-1">{children}</p>,
  h2: ({ children }) => <p className="font-semibold mt-2 mb-1">{children}</p>,
  h3: ({ children }) => <p className="font-semibold mt-2 mb-1">{children}</p>,
  h4: ({ children }) => <p className="font-semibold mt-2 mb-1">{children}</p>,
  h5: ({ children }) => <p className="font-semibold mt-2 mb-1">{children}</p>,
  h6: ({ children }) => <p className="font-semibold mt-2 mb-1">{children}</p>,
  p: ({ children }) => <p className="my-1.5 first:mt-0 last:mb-0">{children}</p>,
  ul: ({ children }) => <ul className="list-disc pl-6 my-1.5 space-y-0.5">{children}</ul>,
  ol: ({ children }) => <ol className="list-decimal pl-6 my-1.5 space-y-0.5">{children}</ol>,
  li: ({ children }) => <li>{children}</li>,
  strong: ({ children }) => <strong className="font-semibold">{children}</strong>,
  a: ({ href, children }) => {
    const safeHref = getSafeMarkdownHref(href);
    if (!safeHref) {
      return <span className="underline">{children}</span>;
    }
    return (
      <a
        href={safeHref}
        target="_blank"
        rel="noopener noreferrer"
        className={`underline ${C.textActionPrimary} hover:opacity-70`}
      >
        {children}
      </a>
    );
  },
  code: ({ className, children }) =>
    className ? (
      <code className={`block ${className}`}>{children}</code>
    ) : (
      <code className={`px-1 py-0.5 rounded-xxs text-[0.875em] ${C.bgHoverMd}`}>{children}</code>
    ),
  pre: ({ children }) => (
    <pre className={`my-1.5 p-2 rounded-xxs overflow-x-auto text-2xs ${C.bgHoverMd}`}>
      {children}
    </pre>
  ),
  blockquote: ({ children }) => (
    <blockquote className={`my-1.5 pl-3 border-l-2 italic ${C.text65} ${C.borderDivider}`}>
      {children}
    </blockquote>
  ),
  hr: () => <hr className={`my-2 border-t ${C.borderDivider}`} />,
  table: ({ children }) => (
    <div className="my-1.5 overflow-x-auto">
      <table className={`w-full border ${C.borderDivider}`}>{children}</table>
    </div>
  ),
  th: ({ children }) => (
    <th className={`${STYLE.tableHeaderCell} text-left border ${C.borderDivider}`}>{children}</th>
  ),
  td: ({ children }) => (
    <td className={`${STYLE.tableCell} border ${C.borderDivider}`}>{children}</td>
  ),
};
