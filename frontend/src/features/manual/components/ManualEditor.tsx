/**
 * ManualEditor — マニュアル本文の編集エディタ（FE only）
 *
 * - textarea ベースのシンプルなマークダウン編集
 * - 編集 / 分割 / プレビュー の 3 モード
 * - 編集内容のクリップボードコピー・.md ファイルダウンロード
 *
 * バックエンド保存は持たない。スタッフは編集後 .md をダウンロードし、
 * IT 担当・開発者に渡してリポジトリに反映してもらう運用。
 *
 * 詳細は業務フロー「マニュアルの編集依頼方法」(workflows/27-manual-edit-request.md) を参照。
 */

import { useLayoutEffect, useMemo, useRef, useState } from "react";
import { Copy, Download, X, FileText, Eye, Columns2, Save, Loader2 } from "lucide-react";
import { toast } from "sonner";

import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { C } from "@/lib/design-tokens";

import { useUpsertManualArticle } from "../api/upsert-manual-article";
import type { ManualArticle } from "../lib/manual-index";
import { ManualContent } from "./ManualContent";

/** コピー完了表示を戻すまでの待ち時間 (FE5-6) */
const COPY_FEEDBACK_RESET_MS = 2000;
const DISCARD_TITLE = "編集内容が保存されていません";
const DISCARD_DESCRIPTION = "破棄して閉じますか？";

type EditorMode = "edit" | "preview" | "split";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

interface ManualEditorProps {
  article: ManualArticle;
  onClose: () => void;
  canEdit?: boolean;
}

/** frontmatter + 本文を結合した完全な MD 文字列を返す */
function buildFullMarkdown(article: ManualArticle, body: string): string {
  return `---\ntitle: ${article.title}\norder: ${article.order}\nsection: ${article.section}\n---\n\n${body}`;
}

export function ManualEditor({ article, onClose, canEdit = false }: ManualEditorProps) {
  const [mode, setMode] = useState<EditorMode>("split");
  const [content, setContent] = useState(article.content);
  const [copyStatus, setCopyStatus] = useState<"idle" | "copied">("idle");
  const [discardOpen, setDiscardOpen] = useState(false);
  const upsertMutation = useUpsertManualArticle();
  const canEditRef = useRef(canEdit);
  useLayoutEffect(() => {
    canEditRef.current = canEdit;
  }, [canEdit]);

  const handleSave = () => {
    if (canEditRef.current !== true) {
      toast.error(PERMISSION_DENIED_MESSAGE);
      return;
    }
    upsertMutation.mutate(
      {
        category: article.category,
        slug: article.slug,
        title: article.title,
        order_value: article.order,
        section: article.section,
        body_markdown: content,
      },
      {
        onSuccess: () => {
          onClose();
        },
      },
    );
  };

  const handleCopy = async () => {
    const full = buildFullMarkdown(article, content);
    try {
      await navigator.clipboard.writeText(full);
      setCopyStatus("copied");
      window.setTimeout(() => setCopyStatus("idle"), COPY_FEEDBACK_RESET_MS);
    } catch {
      // クリップボード権限が無い等のフォールバック: 何もしない
    }
  };

  const handleDownload = () => {
    const full = buildFullMarkdown(article, content);
    const blob = new Blob([full], { type: "text/markdown;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${article.slug}.md`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  // プレビュー用に編集中 content を反映した一時 article を組み立てる。
  // rerender-memo: ManualContent は memo。毎レンダー新規オブジェクトを渡すと
  // 無効化されるため content/article が変化した時のみ再生成する。
  const previewArticle: ManualArticle = useMemo(
    () => ({ ...article, content, searchText: content }),
    [article, content],
  );

  const isDirty = content !== article.content;

  // 編集中（dirty）でエディタを閉じようとする時は離脱確認
  const handleCloseRequest = () => {
    if (isDirty) {
      setDiscardOpen(true);
      return;
    }
    onClose();
  };

  const handleDiscardConfirm = () => {
    setDiscardOpen(false);
    onClose();
  };

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* ツールバー */}
      <div
        className={`flex flex-wrap items-center gap-2 px-3 py-2 border-b ${C.borderDivider} ${C.bgWhite} no-print`}
      >
        <span className={`text-sm font-semibold ${C.text}`}>編集中</span>
        <span className={`text-xs ${C.text50} truncate`}>{article.title}</span>
        {isDirty ? (
          <span className={`text-2xs px-1.5 py-0.5 rounded-xxs ${C.bgWarning50} ${C.textWarning}`}>
            未保存
          </span>
        ) : null}
        <div className="flex-1 min-w-2" />

        {/* モード切替 */}
        <div className={`inline-flex rounded-xxs border ${C.borderDivider}`}>
          <button
            type="button"
            onClick={() => setMode("edit")}
            aria-pressed={mode === "edit"}
            className={`flex items-center gap-1 px-2.5 py-1.5 text-sm ${C.text} ${mode === "edit" ? "bg-black/5" : ""}`}
          >
            <FileText className="size-4" />
            <span className="hidden sm:inline">編集</span>
          </button>
          <button
            type="button"
            onClick={() => setMode("split")}
            aria-pressed={mode === "split"}
            className={`flex items-center gap-1 px-2.5 py-1.5 text-sm border-l ${C.borderDivider} ${C.text} ${mode === "split" ? "bg-black/5" : ""}`}
          >
            <Columns2 className="size-4" />
            <span className="hidden sm:inline">分割</span>
          </button>
          <button
            type="button"
            onClick={() => setMode("preview")}
            aria-pressed={mode === "preview"}
            className={`flex items-center gap-1 px-2.5 py-1.5 text-sm border-l ${C.borderDivider} ${C.text} ${mode === "preview" ? "bg-black/5" : ""}`}
          >
            <Eye className="size-4" />
            <span className="hidden sm:inline">プレビュー</span>
          </button>
        </div>

        {/* DB 保存ボタン: 管理者権限が必要。サーバへ送信し成功時にエディタを閉じる。 */}
        <button
          type="button"
          onClick={handleSave}
          disabled={!isDirty || upsertMutation.isPending}
          className={`flex items-center gap-1 px-2.5 py-1.5 text-sm rounded-xxs border ${C.borderDivider} ${
            !isDirty || upsertMutation.isPending
              ? `opacity-50 cursor-not-allowed ${C.text}`
              : `${C.bgBrand} ${C.textOnBrand} hover:opacity-90`
          }`}
          title="DB に保存（管理者権限が必要）"
        >
          {upsertMutation.isPending ? (
            <Loader2 className="size-4 animate-spin" />
          ) : (
            <Save className="size-4" />
          )}
          <span className="hidden md:inline">
            {upsertMutation.isPending ? "保存中..." : "保存"}
          </span>
        </button>

        <button
          type="button"
          onClick={handleCopy}
          className={`flex items-center gap-1 px-2.5 py-1.5 text-sm rounded-xxs border ${C.borderDivider} ${C.hoverBgLight} ${C.text}`}
          title="編集内容をクリップボードにコピー"
        >
          <Copy className="size-4" />
          <span className="hidden md:inline">
            {copyStatus === "copied" ? "コピー済" : "コピー"}
          </span>
        </button>

        <button
          type="button"
          onClick={handleDownload}
          className={`flex items-center gap-1 px-2.5 py-1.5 text-sm rounded-xxs border ${C.borderDivider} ${C.hoverBgLight} ${C.text}`}
          title=".md ファイルとしてダウンロード"
        >
          <Download className="size-4" />
          <span className="hidden md:inline">ダウンロード</span>
        </button>

        <button
          type="button"
          onClick={handleCloseRequest}
          aria-label="編集を終了"
          className={`size-9 flex items-center justify-center rounded-xxs border ${C.borderDivider} ${C.hoverBgLight}`}
        >
          <X className="size-4" />
        </button>
      </div>

      {/* 注意バナー */}
      <div
        className={`px-4 py-2 text-xs border-b ${C.borderDivider} ${C.bgWarning50} ${C.textWarning}`}
      >
        💡 「<strong>保存</strong>」ボタン: 変更を <strong>DB に保存</strong>
        （管理者権限が必要、全スタッフに即時反映）／ 「<strong>コピー</strong>」「
        <strong>ダウンロード</strong>」: 編集案を IT 担当者に手動で渡す場合に使用。
        詳細は業務フロー「<strong>マニュアルの編集依頼方法</strong>」を参照。
      </div>

      {/* 編集 + プレビュー本体 */}
      <div className="flex-1 flex overflow-hidden">
        {mode === "edit" || mode === "split" ? (
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            className={`flex-1 p-4 font-mono text-sm resize-none outline-none focus-visible:ring-2 focus-visible:ring-inset ${C.focusRingAccent40} ${C.bgSubtle} ${mode === "split" ? `border-r ${C.borderDivider}` : ""}`}
            spellCheck={false}
            aria-label="マニュアル本文編集"
          />
        ) : null}
        {mode === "preview" || mode === "split" ? (
          <div className="flex-1 overflow-y-auto">
            <ManualContent article={previewArticle} />
          </div>
        ) : null}
      </div>

      <ConfirmDialog
        open={discardOpen}
        onClose={() => setDiscardOpen(false)}
        onConfirm={handleDiscardConfirm}
        title={DISCARD_TITLE}
        description={DISCARD_DESCRIPTION}
      />
    </div>
  );
}
