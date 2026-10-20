/**
 * ManualEditor — マニュアル本文の編集エディタ（EMR-513 既存記事編集フロー）
 *
 * - textarea ベースのシンプルなマークダウン編集
 * - 編集 / 分割 / プレビュー の 3 モード
 * - 編集内容のクリップボードコピー・.md ファイルダウンロード
 *
 * 保存フロー（固定API契約・backend/docs/api.yaml 採用SHA 30538608）:
 *   1. マウント時に GET .../edit で編集元（CAS版 version と current 内容版）を取得する。
 *      CAS版と内容版（id/content_revision/content_digest）は別物として扱う。
 *   2. current=NULL / legacy_unknown / digest・revision欠落 は架空補完せず
 *      「元内容を確認」状態で保存を停止する。
 *   3. PUT は operation_key（UUID）付きの全置換content。IDは正数十進文字列のまま往復する。
 *   4. 成功表示は receipt照合（対象/版一致）を通過したときだけ。編集の保存は公開版を変更しない。
 *   5. 409はcodeで分岐（version_conflictは最新edit viewを別取得して比較・再確認。自動再送しない）。
 *   6. 通信断/5xx は同一 operation_key で結果照会し、unknown中は入力保持して停止する
 *      （新しい operation_key での再保存はしない）。
 *
 * 詳細は業務フロー「マニュアルの編集依頼方法」(workflows/27-manual-edit-request.md) を参照。
 */

import { useLayoutEffect, useMemo, useRef, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Copy, Download, X, FileText, Eye, Columns2, Save, Loader2 } from "lucide-react";
import { toast } from "sonner";

import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { C } from "@/lib/design-tokens";
import { SUPPORT_WIDGET_LAYOUT } from "@/constants/support-widget-layout";
import { extractApiErrorMessage } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

import {
  getManualArticleEditView,
  getManualEditOperationResult,
  manualArticleEditViewQueryKey,
  resolveManualEditSource,
  verifyManualEditReceipt,
  ManualArticleSaveError,
  ManualEditConflictError,
  ManualReceiptMismatchError,
  type ManualArticleEditView,
  type ManualEditBase,
  type ManualEditBlockedReason,
  type ManualEditContent,
  type ManualEditOperationResult,
  type ManualEditReceiptExpectation,
} from "../api/manual-article-edit";
import { useUpsertManualArticle } from "../api/upsert-manual-article";
import type { ManualArticle } from "@/lib/manual-index";
import { ManualContent } from "./ManualContent";

/** コピー完了表示を戻すまでの待ち時間 (FE5-6) */
const COPY_FEEDBACK_RESET_MS = 2000;
const DISCARD_TITLE = "編集内容が保存されていません";
const DISCARD_DESCRIPTION = "破棄して閉じますか？";

type EditorMode = "edit" | "preview" | "split";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";
const EDIT_FETCH_CONTEXT = "マニュアル編集元の取得";
const SAVE_CONTEXT = "マニュアル保存";
const RECEIPT_MISMATCH_NOTICE =
  "保存結果の照合に失敗しました。入力は保持されています。編集を閉じて開き直し、最新状態を確認してください。";
const UNKNOWN_NOTICE =
  "保存結果を確認できていません（unknown）。入力は保持されています。新しい保存は行われません。編集を閉じて開き直し、最新状態を確認してください。";
const SOURCE_UNRESOLVED_CONFLICT_NOTICE =
  "保存が競合しました（source_unresolved）。編集元の内容が解決できなくなっています。入力は保持されています。元内容を確認してください。";
const TITLE_SECTION_EMPTY_MESSAGE =
  "タイトルとセクションが空のため保存できません。元内容を確認してください。";
const EDIT_SOURCE_LOADING_MESSAGE = "編集元を取得しています…";

const UNRESOLVED_MESSAGES: Record<ManualEditBlockedReason, string> = {
  current_null:
    "この記事の編集元（現在内容版）が未解決のため保存できません。元内容を確認してください。",
  legacy_unknown:
    "この記事の出所が未解決（legacy_unknown）のため保存できません。元内容を確認してください。",
  content_unidentified:
    "この記事の内容版のdigestまたは改訂番号が欠落しているため保存できません。元内容を確認してください。",
  invalid_target:
    "この記事の編集対象の識別が契約どおりではないため保存できません。元内容を確認してください。",
};

const CONFLICT_STOP_MESSAGES: Record<"operation_key_reused" | "target_mismatch", string> = {
  operation_key_reused:
    "保存操作が重複しています（operation_key_reused）。入力は保持されています。編集を開き直して最新状態を確認してください。",
  target_mismatch:
    "保存対象の特定に失敗しました（target_mismatch）。入力は保持されています。編集を開き直して最新状態を確認してください。",
};

interface ManualEditorProps {
  article: ManualArticle;
  onClose: () => void;
  canEdit?: boolean;
}

/** 編集元取得の状態（query結果から派生させる。setState in effect を避ける） */
type EditSourceState =
  | { phase: "loading" }
  | { phase: "error"; message: string }
  | { phase: "unresolved"; reason: ManualEditBlockedReason }
  | { phase: "ready"; base: ManualEditBase };

/** 直前の保存要求（409時の比較・結果照会時のreceipt照合に使う） */
interface LastSaveRecord {
  expectation: Omit<ManualEditReceiptExpectation, "operationKey">;
  content: ManualEditContent;
  previousBase: ManualEditBase;
}

/** frontmatter + 本文を結合した完全な MD 文字列を返す */
function buildFullMarkdown(article: ManualArticle, body: string): string {
  return `---\ntitle: ${article.title}\norder: ${article.order}\nsection: ${article.section}\n---\n\n${body}`;
}

export function ManualEditor({ article, onClose, canEdit = false }: ManualEditorProps) {
  const [mode, setMode] = useState<EditorMode>("split");
  /** 編集本文。null は未編集（編集元の initialBody をそのまま表示） */
  const [body, setBody] = useState<string | null>(null);
  /** 結果照会でも確定できなかった等、この編集セッションで保存を停止した状態 */
  const [saveStopped, setSaveStopped] = useState(false);
  const [notice, setNotice] = useState<{ kind: "error" | "info"; text: string } | null>(null);
  const [copyStatus, setCopyStatus] = useState<"idle" | "copied">("idle");
  const [discardOpen, setDiscardOpen] = useState(false);

  const queryClient = useQueryClient();
  const saveMutation = useUpsertManualArticle();
  const canEditRef = useRef(canEdit);
  const lastSaveRef = useRef<LastSaveRecord | null>(null);
  useLayoutEffect(() => {
    canEditRef.current = canEdit;
  }, [canEdit]);

  // 編集開始時（マウント時）に編集元を取得する。editorは編集再開のたびに
  // マウントし直すため、gcTime 0 で旧receipt/CASを混ぜずに毎回取り直す。
  const editViewQuery = useQuery({
    queryKey: manualArticleEditViewQueryKey(article.category, article.slug),
    queryFn: () => getManualArticleEditView(article.category, article.slug),
    retry: false,
    refetchOnWindowFocus: false,
    staleTime: 0,
    gcTime: 0,
    meta: { silentError: true, errorContext: EDIT_FETCH_CONTEXT },
  });

  const source = useMemo<EditSourceState>(() => {
    if (editViewQuery.isPending) return { phase: "loading" };
    if (editViewQuery.isError) {
      return {
        phase: "error",
        message: extractApiErrorMessage(editViewQuery.error, EDIT_FETCH_CONTEXT),
      };
    }
    const resolution = resolveManualEditSource(editViewQuery.data);
    if (!resolution.ok) return { phase: "unresolved", reason: resolution.reason };
    return { phase: "ready", base: resolution.base };
  }, [editViewQuery.data, editViewQuery.isPending, editViewQuery.isError, editViewQuery.error]);

  const base = source.phase === "ready" ? source.base : null;
  // 編集元が解決するまではバンドル（表示中）の内容を閲覧・コピー用に表示する。
  // 編集対象はあくまで current 内容版で、解決後に差し替わる。
  const initialBodyForDisplay = base?.initialBody ?? article.content;
  const currentBody = body ?? initialBodyForDisplay;
  const isDirty = base !== null && body !== null && body !== base.initialBody;
  const saveDisabled =
    !isDirty || saveMutation.isPending || source.phase !== "ready" || saveStopped;

  const handleVerifiedSuccess = () => {
    setNotice(null);
    toast.success("マニュアルを保存しました");
    queryClient.invalidateQueries({ queryKey: queryKeys.manualArticles.all() });
    onClose();
  };

  /** 409 version_conflict: 最新edit viewを別取得して比較し、再確認を促す（自動再送しない） */
  const adoptLatestEditViewAfterConflict = async () => {
    const previous = lastSaveRef.current?.previousBase ?? null;
    let fresh: ManualArticleEditView;
    try {
      fresh = await getManualArticleEditView(article.category, article.slug);
    } catch (fetchErr) {
      setNotice({
        kind: "error",
        text: `保存が競合しました（version_conflict）。最新の編集元の取得に失敗しました。${extractApiErrorMessage(
          fetchErr,
          EDIT_FETCH_CONTEXT,
        )}`,
      });
      return;
    }
    queryClient.setQueryData<ManualArticleEditView>(
      manualArticleEditViewQueryKey(article.category, article.slug),
      fresh,
    );
    const resolution = resolveManualEditSource(fresh);
    if (!resolution.ok) {
      setNotice({
        kind: "error",
        text: `保存が競合しました（version_conflict）。${UNRESOLVED_MESSAGES[resolution.reason]}`,
      });
      return;
    }
    const freshBase = resolution.base;
    const changed: string[] = [];
    if (previous !== null && freshBase.initialBody !== previous.initialBody) {
      changed.push("本文");
    }
    if (
      previous !== null &&
      (freshBase.title !== previous.title ||
        freshBase.section !== previous.section ||
        freshBase.orderValue !== previous.orderValue)
    ) {
      changed.push("タイトル・セクション・表示順");
    }
    const detail =
      changed.length > 0
        ? `編集元の${changed.join("・")}が更新されています。`
        : previous === null
          ? ""
          : "編集元の内容に変更はありません。";
    setNotice({
      kind: "error",
      text: `保存が競合しました（version_conflict）。他の操作で編集元が進んでいます。最新の編集元を取得しました。${detail}入力は保持されています。内容を確認のうえ、再度保存してください（自動再送はしません）。`,
    });
  };

  /** 通信断/5xx: 同一 operation_key で結果照会する。unknown中は新keyでの再保存をしない */
  const resolveUnknownOperation = async (operationKey: string) => {
    const last = lastSaveRef.current;
    let result: ManualEditOperationResult;
    try {
      result = await getManualEditOperationResult(article.category, article.slug, operationKey);
    } catch (queryErr) {
      setSaveStopped(true);
      setNotice({
        kind: "error",
        text: `保存結果の照会に失敗しました。${extractApiErrorMessage(
          queryErr,
          "マニュアル保存結果の照会",
        )}`,
      });
      return;
    }
    if (result.state === "unknown" || result.receipt === null || last === null) {
      setSaveStopped(true);
      setNotice({ kind: "error", text: UNKNOWN_NOTICE });
      return;
    }
    // 照会結果もreceipt照合を通す
    try {
      verifyManualEditReceipt(result.receipt, {
        operationKey,
        articleId: last.expectation.articleId,
        expectedVersion: last.expectation.expectedVersion,
        expectedContentVersionId: last.expectation.expectedContentVersionId,
      });
    } catch {
      setSaveStopped(true);
      setNotice({ kind: "error", text: RECEIPT_MISMATCH_NOTICE });
      return;
    }
    handleVerifiedSuccess();
  };

  const handleSaveFailure = async (err: unknown) => {
    if (err instanceof ManualEditConflictError) {
      if (err.conflict.code === "version_conflict") {
        await adoptLatestEditViewAfterConflict();
        return;
      }
      if (err.conflict.code === "source_unresolved") {
        setSaveStopped(true);
        setNotice({ kind: "error", text: SOURCE_UNRESOLVED_CONFLICT_NOTICE });
        return;
      }
      setSaveStopped(true);
      const serverMessage = err.conflict.error ? `（${err.conflict.error}）` : "";
      setNotice({
        kind: "error",
        text: `${CONFLICT_STOP_MESSAGES[err.conflict.code]}${serverMessage}`,
      });
      return;
    }
    if (err instanceof ManualReceiptMismatchError) {
      setSaveStopped(true);
      setNotice({ kind: "error", text: RECEIPT_MISMATCH_NOTICE });
      return;
    }
    if (err instanceof ManualArticleSaveError) {
      if (err.status === undefined || err.status >= 500) {
        // 応答喪失・5xxは結果不明。元keyで結果照会し新keyの再保存で救済しない
        await resolveUnknownOperation(err.operationKey);
        return;
      }
      // 400/401/403/404: 確定的な拒否。入力保持し自動再試行しない
      setNotice({
        kind: "error",
        text: extractApiErrorMessage(err.requestError ?? err, SAVE_CONTEXT),
      });
      return;
    }
    setNotice({ kind: "error", text: extractApiErrorMessage(err, SAVE_CONTEXT) });
  };

  const runSaveFlow = async (editBase: ManualEditBase) => {
    const content: ManualEditContent = {
      title: editBase.title,
      order_value: editBase.orderValue,
      section: editBase.section,
      body_markdown: currentBody,
    };
    lastSaveRef.current = {
      expectation: {
        articleId: editBase.articleId,
        expectedVersion: editBase.casVersion,
        expectedContentVersionId: editBase.contentVersionId,
      },
      content,
      previousBase: editBase,
    };
    try {
      // receipt照合はadapter内で行い、通過した保存だけがresolveする
      await saveMutation.mutateAsync({
        category: article.category,
        slug: article.slug,
        articleId: editBase.articleId,
        expectedVersion: editBase.casVersion,
        expectedContentVersionId: editBase.contentVersionId,
        content,
      });
      handleVerifiedSuccess();
    } catch (err) {
      await handleSaveFailure(err);
    }
  };

  const handleSave = () => {
    if (canEditRef.current !== true) {
      toast.error(PERMISSION_DENIED_MESSAGE);
      return;
    }
    if (base === null) {
      // ボタンdisabledで到達しない。fail-closedの再確認
      toast.error(
        source.phase === "unresolved"
          ? UNRESOLVED_MESSAGES[source.reason]
          : source.phase === "error"
            ? source.message
            : EDIT_SOURCE_LOADING_MESSAGE,
      );
      return;
    }
    if (base.title === "" || base.section === "") {
      // 既存serviceと同じ空文字拒否（trim・文字数上限は追加しない。body_markdown空文字は許可）
      toast.error(TITLE_SECTION_EMPTY_MESSAGE);
      return;
    }
    setNotice(null);
    void runSaveFlow(base);
  };

  const handleCopy = async () => {
    const full = buildFullMarkdown(article, currentBody);
    try {
      await navigator.clipboard.writeText(full);
      setCopyStatus("copied");
      window.setTimeout(() => setCopyStatus("idle"), COPY_FEEDBACK_RESET_MS);
    } catch {
      // クリップボード権限が無い等のフォールバック: 何もしない
    }
  };

  const handleDownload = () => {
    const full = buildFullMarkdown(article, currentBody);
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
    () => ({ ...article, content: currentBody, searchText: currentBody }),
    [article, currentBody],
  );

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
    <div className="flex-1 flex flex-col overflow-hidden" aria-busy={source.phase === "loading"}>
      {/* ツールバー */}
      <div
        className={`flex flex-wrap items-center gap-2 px-3 py-2 border-b ${C.borderDivider} ${C.bgWhite} no-print`}
      >
        <span className={`text-sm font-semibold ${C.text}`}>編集中</span>
        <span className={`text-xs ${C.text50} truncate`}>{article.title}</span>
        {isDirty ? (
          <span className={`text-2xs px-1.5 py-0.5 rounded-xxs ${C.bgNotice} ${C.textBadgeYellow}`}>
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
            className={`flex min-h-11 items-center gap-1 px-2.5 py-1.5 text-sm ${C.text} ${mode === "edit" ? C.bgPrimary5 : ""}`}
          >
            <FileText className="size-4" />
            <span className="hidden sm:inline">編集</span>
          </button>
          <button
            type="button"
            onClick={() => setMode("split")}
            aria-pressed={mode === "split"}
            className={`flex min-h-11 items-center gap-1 px-2.5 py-1.5 text-sm border-l ${C.borderDivider} ${C.text} ${mode === "split" ? C.bgPrimary5 : ""}`}
          >
            <Columns2 className="size-4" />
            <span className="hidden sm:inline">分割</span>
          </button>
          <button
            type="button"
            onClick={() => setMode("preview")}
            aria-pressed={mode === "preview"}
            className={`flex min-h-11 items-center gap-1 px-2.5 py-1.5 text-sm border-l ${C.borderDivider} ${C.text} ${mode === "preview" ? C.bgPrimary5 : ""}`}
          >
            <Eye className="size-4" />
            <span className="hidden sm:inline">プレビュー</span>
          </button>
        </div>

        {/* DB 保存ボタン: 管理者権限が必要。receipt照合（対象/版一致）成功時のみ閉じる。
            編集元未解決・結果不明中は保存を停止する（workflow状態ゲート）。 */}
        <button
          type="button"
          onClick={handleSave}
          disabled={saveDisabled}
          className={`flex min-h-11 items-center gap-1 px-2.5 py-1.5 text-sm rounded-xxs border ${C.borderDivider} ${
            saveDisabled
              ? `opacity-50 cursor-not-allowed ${C.text}`
              : `${C.bgActionPrimarySolid} ${C.textOnActionPrimary} ${C.hoverBgActionPrimarySolid} ${C.hoverTextOnActionPrimary} ${C.activeBgActionPrimarySolid}`
          }`}
          title="DB に編集を保存（管理者権限が必要・公開版は変更されません）"
        >
          {saveMutation.isPending ? (
            <Loader2 className="size-4 animate-spin" />
          ) : (
            <Save className="size-4" />
          )}
          <span className="hidden md:inline">{saveMutation.isPending ? "保存中..." : "保存"}</span>
        </button>

        <button
          type="button"
          onClick={handleCopy}
          className={`flex min-h-11 items-center gap-1 px-2.5 py-1.5 text-sm rounded-xxs border ${C.borderDivider} ${C.hoverBgLight} ${C.text}`}
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
          className={`flex min-h-11 items-center gap-1 px-2.5 py-1.5 text-sm rounded-xxs border ${C.borderDivider} ${C.hoverBgLight} ${C.text}`}
          title=".md ファイルとしてダウンロード"
        >
          <Download className="size-4" />
          <span className="hidden md:inline">ダウンロード</span>
        </button>

        <button
          type="button"
          onClick={handleCloseRequest}
          aria-label="編集を終了"
          className={`min-h-11 min-w-11 flex items-center justify-center rounded-xxs border ${C.borderDivider} ${C.hoverBgLight}`}
        >
          <X className="size-4" />
        </button>
      </div>

      {/* 注意バナー */}
      <div
        className={`px-4 py-2 text-xs border-b ${C.borderDivider} ${C.bgWarning50} ${C.textBadgeYellow}`}
      >
        💡 「<strong>保存</strong>」ボタン: 編集内容を <strong>DB に保存</strong>
        （管理者権限が必要）。保存は編集の確定のみで、
        <strong>公開版は変更されません</strong>（公開は別の明示操作です）／ 「
        <strong>コピー</strong>」「<strong>ダウンロード</strong>」: 編集案を IT
        担当者に手動で渡す場合に使用。詳細は業務フロー「
        <strong>マニュアルの編集依頼方法</strong>」を参照。
      </div>

      {/* 編集元の取得状態・保存結果の通知 */}
      {source.phase === "loading" ? (
        <div
          role="status"
          className={`flex items-center gap-1 px-4 py-2 text-xs border-b ${C.borderDivider} ${C.bgWarning50} ${C.textBadgeYellow}`}
        >
          <Loader2 className="size-3 animate-spin" aria-hidden />
          <span>{EDIT_SOURCE_LOADING_MESSAGE}</span>
        </div>
      ) : null}
      {source.phase === "unresolved" ? (
        <div
          role="alert"
          className={`px-4 py-2 text-xs border-b ${C.borderDivider} ${C.bgDanger10} ${C.danger}`}
        >
          {UNRESOLVED_MESSAGES[source.reason]}
        </div>
      ) : null}
      {source.phase === "error" ? (
        <div
          role="alert"
          className={`px-4 py-2 text-xs border-b ${C.borderDivider} ${C.bgDanger10} ${C.danger}`}
        >
          {source.message}
        </div>
      ) : null}
      {notice !== null ? (
        <div
          role="alert"
          className={`px-4 py-2 text-xs border-b ${C.borderDivider} ${C.bgDanger10} ${C.danger}`}
        >
          {notice.text}
        </div>
      ) : null}

      {/* 編集 + プレビュー本体 */}
      <div className="flex-1 flex overflow-hidden">
        {mode === "edit" || mode === "split" ? (
          <textarea
            value={currentBody}
            onChange={(e) => setBody(e.target.value)}
            readOnly={source.phase !== "ready"}
            className={`flex-1 p-4 font-mono text-sm resize-none outline-none focus-visible:ring-2 focus-visible:ring-inset ${C.focusRingAccent40} ${C.bgSubtle} ${mode === "split" ? `border-r ${C.borderDivider}` : ""}`}
            spellCheck={false}
            aria-label="マニュアル本文編集"
          />
        ) : null}
        {mode === "preview" || mode === "split" ? (
          <div
            className={`flex-1 overflow-y-auto relative ${SUPPORT_WIDGET_LAYOUT.scrollBottomClearance}`}
          >
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
