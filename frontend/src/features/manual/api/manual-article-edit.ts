/**
 * manual-article-edit.ts — 既存マニュアル記事編集の契約wire adapter（EMR-513）
 *
 * 固定API契約（backend/docs/api.yaml・採用SHA 30538608）に固定する:
 * - GET /v1/manual/articles/{category}/{slug}/edit             → ManualArticleEditView
 * - PUT /v1/manual/articles/{category}/{slug}                  → 200 ManualEditReceipt / 409 ManualEditConflict / 400,401,403,404,500 ErrorResponse{error}
 * - GET /v1/manual/articles/{category}/{slug}/operations/{key} → ManualEditOperationResult
 *
 * wire規約:
 * - 記事ID/内容版IDは正数十進「文字列」のままstate/請求/照合へ往復する（JS Numberへ変換しない。int64誤差防止）。
 * - digestはserver算出の64hexを透過保持する（client再計算・正規化をしない）。
 * - CAS版(view.version)と内容版(current.id/content_revision/content_digest)は別物として扱う。
 * - 編集保存は全置換content（title/section/body_markdown/order_value）。部分更新は送らない。
 * - 型名はwire契約のschemaに対応させるが、旧生成型
 *   （`@/types/generated/models.ts` の ManualArticle / ManualArticleVersion: number ID）
 *   および `@/lib/manual-index` の ManualArticle（バンドルMD用）とは別物のため、
 *   ManualArticleVersion系は ManualEditContentVersion というwire専用名を使う。
 */

import { isAxiosError } from "axios";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";

import type { ManualCategory } from "@/lib/manual-index";

// ── wire DTO（backend/docs/api.yaml components/schemas 対応） ─────────────

/** wire: ManualArticleContent（表示/履歴内容の値object。編集要求 ManualArticleEditContent の基底） */
export interface ManualEditContent {
  title: string;
  order_value: number;
  section: string;
  body_markdown: string;
}

/** wire: ManualArticleVersion.source_kind のenum。旧NULLはserver側で legacy_unknown に正規化済み */
export type ManualContentSourceKind =
  "static" | "override" | "edited" | "restored" | "legacy_unknown";

/** wire: ManualArticleVersion — 応答専用の不変内容版（wire専用名。旧生成型と混同しない） */
export interface ManualEditContentVersion {
  /** 内容版ID。先頭ゼロなし正数十進文字列。Number変換禁止 */
  id: string;
  article_id: string;
  /** 記事内の内容改訂番号。NULLは未解決出所（架空補完しない） */
  content_revision: number | null;
  /** server算出の32byte digest hex表現（64hex）。NULLは未解決出所 */
  content_digest: string | null;
  source_kind: ManualContentSourceKind;
  restored_from_id: string | null;
  content: ManualEditContent;
  edited_by_staff_id: number | null;
  edited_at: string;
}

/** wire: ManualArticleEditView — 編集権限者限定の応答DTO */
export interface ManualArticleEditView {
  /** 記事ID。正数十進文字列 */
  id: string;
  category: ManualCategory;
  slug: string;
  /** 記事状態のCAS版（編集・公開・復元・非表示等で進む）。current内容版とは別物 */
  version: number;
  /** 現在の編集内容版。NULLは元内容未解決（編集送信停止） */
  current: ManualEditContentVersion | null;
  published_version_id: string | null;
  hidden: boolean;
}

/** wire: UpsertManualArticleRequest — 既存記事edit専用（不存在の新規upsertはしない。404） */
export interface UpsertManualArticleRequest {
  operation_key: string;
  article_id: string;
  /** 記事CAS版（編集開始時の view.version） */
  expected_version: number;
  /** 編集開始時の current.id */
  expected_content_version_id: string;
  content: ManualEditContent;
}

/** wire: ManualEditReceipt — 業務変更と原子的に確定したreceipt。resulting_version=previous_version+1 */
export interface ManualEditReceipt {
  operation_key: string;
  article_id: string;
  intent: "edit";
  previous_version: number;
  resulting_version: number;
  content_version_id: string;
  content_digest: string;
  published_version_id: string | null;
}

/** wire: ManualEditConflict.code のenum */
export const MANUAL_EDIT_CONFLICT_CODES = [
  "version_conflict",
  "operation_key_reused",
  "target_mismatch",
  "source_unresolved",
] as const;
export type ManualEditConflictCode = (typeof MANUAL_EDIT_CONFLICT_CODES)[number];

/** wire: ManualEditConflict */
export interface ManualEditConflict {
  error: string;
  code: ManualEditConflictCode;
}

/** wire: ManualEditOperationResult — succeededはreceipt非NULL、unknownはNULL */
export type ManualEditOperationState = "succeeded" | "unknown";
export interface ManualEditOperationResult {
  operation_key: string;
  state: ManualEditOperationState;
  receipt: ManualEditReceipt | null;
}

/** wire: ErrorResponse{error} */
export interface ManualErrorBody {
  error: string;
}

// ── 編集元の解決（current=NULL / legacy_unknown / digest・revision欠落の停止判定） ──

/** 編集保存の対象として解決した編集元（CAS版と内容版を別fieldで保持） */
export interface ManualEditBase {
  /** 記事ID（view.id。正数十進文字列のまま） */
  articleId: string;
  /** 記事CAS版（view.version） */
  casVersion: number;
  /** 編集開始時の内容版ID（current.id。正数十進文字列のまま） */
  contentVersionId: string;
  contentRevision: number;
  /** server算出digestの透過保持 */
  contentDigest: string;
  title: string;
  section: string;
  orderValue: number;
  initialBody: string;
}

export type ManualEditBlockedReason =
  "current_null" | "legacy_unknown" | "content_unidentified" | "invalid_target";

export type ManualEditSourceResolution =
  { ok: true; base: ManualEditBase } | { ok: false; reason: ManualEditBlockedReason };

const POSITIVE_DECIMAL_STRING_RE = /^[1-9][0-9]*$/;
const SHA256_HEX_RE = /^[0-9a-f]{64}$/;

export const isPositiveDecimalString = (value: unknown): value is string =>
  typeof value === "string" && POSITIVE_DECIMAL_STRING_RE.test(value);

/**
 * edit view を「編集保存の元として使える状態」へ解決する。
 * current=NULL / legacy_unknown / digest・revision欠落 / ID・digestの契約形式不一致は
 * 架空補完せず未解決扱いにして停止させる（ticket AC 2）。
 */
export function resolveManualEditSource(view: ManualArticleEditView): ManualEditSourceResolution {
  if (!isPositiveDecimalString(view.id)) {
    return { ok: false, reason: "invalid_target" };
  }
  const current = view.current;
  if (current === null) {
    return { ok: false, reason: "current_null" };
  }
  if (current.source_kind === "legacy_unknown") {
    return { ok: false, reason: "legacy_unknown" };
  }
  if (
    typeof current.content_revision !== "number" ||
    !Number.isInteger(current.content_revision) ||
    current.content_revision < 1
  ) {
    return { ok: false, reason: "content_unidentified" };
  }
  if (typeof current.content_digest !== "string" || !SHA256_HEX_RE.test(current.content_digest)) {
    return { ok: false, reason: "content_unidentified" };
  }
  if (!isPositiveDecimalString(current.id)) {
    return { ok: false, reason: "invalid_target" };
  }
  return {
    ok: true,
    base: {
      articleId: view.id,
      casVersion: view.version,
      contentVersionId: current.id,
      contentRevision: current.content_revision,
      contentDigest: current.content_digest,
      title: current.content.title,
      section: current.content.section,
      orderValue: current.content.order_value,
      initialBody: current.content.body_markdown,
    },
  };
}

// ── receipt照合 ─────────────────────────────────────────────────────────

/** receipt照合の期待値（保存要求の内容と1対1に対応する） */
export interface ManualEditReceiptExpectation {
  operationKey: string;
  articleId: string;
  /** 送信した expected_version（= receipt.previous_version の期待値） */
  expectedVersion: number;
  /** 編集開始時の current.id。receiptはこの版とは別の新内容版を指す */
  expectedContentVersionId: string;
}

/** receipt照合（対象/版/内容版の一致）の失敗。成功表示をしてはいけない */
export class ManualReceiptMismatchError extends Error {
  readonly reason: string;

  constructor(reason: string) {
    super(`manual edit receipt mismatch: ${reason}`);
    this.name = "ManualReceiptMismatchError";
    this.reason = reason;
  }
}

/**
 * receipt照合: 「対象/版一致」のときだけ成功扱いにする。
 * - 対象一致: article_id / operation_key / intent
 * - 版一致: previous_version === 送信した expected_version、resulting_version === previous_version + 1
 * - 内容版一致: content_version_id は新内容版（開始時の current.id と異なる正数十進文字列）、
 *   content_digest はserver算出の64hex（client再計算・正規化はしない）
 */
export function verifyManualEditReceipt(
  receipt: ManualEditReceipt,
  expectation: ManualEditReceiptExpectation,
): void {
  if (receipt.operation_key !== expectation.operationKey) {
    throw new ManualReceiptMismatchError("operation_key mismatch");
  }
  if (receipt.article_id !== expectation.articleId) {
    throw new ManualReceiptMismatchError("article_id mismatch");
  }
  if (receipt.intent !== "edit") {
    throw new ManualReceiptMismatchError("intent mismatch");
  }
  if (receipt.previous_version !== expectation.expectedVersion) {
    throw new ManualReceiptMismatchError("previous_version mismatch");
  }
  if (receipt.resulting_version !== receipt.previous_version + 1) {
    throw new ManualReceiptMismatchError("resulting_version is not previous_version + 1");
  }
  if (!isPositiveDecimalString(receipt.content_version_id)) {
    throw new ManualReceiptMismatchError("content_version_id is not a positive decimal string");
  }
  if (receipt.content_version_id === expectation.expectedContentVersionId) {
    throw new ManualReceiptMismatchError("content_version_id is not a new content version");
  }
  if (typeof receipt.content_digest !== "string" || !SHA256_HEX_RE.test(receipt.content_digest)) {
    throw new ManualReceiptMismatchError("content_digest is not a server-computed 64hex digest");
  }
}

// ── 保存エラーの型付け ───────────────────────────────────────────────────

/**
 * 保存要求の結果がクライアントから確定できない場合のエラー。
 * operation_keyを保持し、結果照会（GET operations/{operation_key}）に同じkeyを使う。
 */
export class ManualArticleSaveError extends Error {
  readonly operationKey: string;
  /** HTTPステータス。応答喪失（通信断等）ではundefined */
  readonly status?: number;
  /** 送出元のaxios error（メッセージ抽出用） */
  readonly requestError?: unknown;

  constructor(params: {
    operationKey: string;
    message: string;
    status?: number;
    requestError?: unknown;
  }) {
    super(params.message);
    this.name = "ManualArticleSaveError";
    this.operationKey = params.operationKey;
    this.status = params.status;
    this.requestError = params.requestError;
  }
}

/** 409 ManualEditConflict。codeで分岐する（version_conflict / operation_key_reused / target_mismatch / source_unresolved） */
export class ManualEditConflictError extends ManualArticleSaveError {
  readonly conflict: ManualEditConflict;

  constructor(params: { operationKey: string; conflict: ManualEditConflict }) {
    super({
      operationKey: params.operationKey,
      status: 409,
      message: `manual edit conflict: ${params.conflict.code}`,
    });
    this.name = "ManualEditConflictError";
    this.conflict = params.conflict;
  }
}

// ── operation_key ───────────────────────────────────────────────────────

/** operation_key はUUID文字列。crypto.randomUUID が無い非セキュアコンテキスト用の v4 形式フォールバック */
export function newManualOperationKey(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  const hex = (): string => Math.floor(Math.random() * 16).toString(16);
  const seq = (length: number): string => Array.from({ length }, hex).join("");
  const variant = "89ab"[Math.floor(Math.random() * 4)];
  return `${seq(8)}-${seq(4)}-4${seq(3)}-${variant}${seq(3)}-${seq(12)}`;
}

// ── adapter ─────────────────────────────────────────────────────────────

/** feature内で完結するquery key（共有 lib/query-keys は触らない） */
export const manualArticleEditViewQueryKey = (category: ManualCategory, slug: string) =>
  [...queryKeys.manualArticles.all(), "edit-view", category, slug] as const;

/** GET /manual/articles/{category}/{slug}/edit — 編集元とCAS版の取得（manual-edit:edit必須） */
export async function getManualArticleEditView(
  category: ManualCategory,
  slug: string,
): Promise<ManualArticleEditView> {
  const { data } = await axios.get<ManualArticleEditView>(
    `/v1/manual/articles/${category}/${slug}/edit`,
  );
  return data;
}

/** GET /manual/articles/{category}/{slug}/operations/{operation_key} — 保存結果の照会 */
export async function getManualEditOperationResult(
  category: ManualCategory,
  slug: string,
  operationKey: string,
): Promise<ManualEditOperationResult> {
  const { data } = await axios.get<ManualEditOperationResult>(
    `/v1/manual/articles/${category}/${slug}/operations/${operationKey}`,
  );
  return data;
}

export interface SaveManualArticleEditParams {
  category: ManualCategory;
  slug: string;
  /** 対象記事ID（正数十進文字列。Number変換しない） */
  articleId: string;
  /** 記事CAS版（編集開始時の view.version） */
  expectedVersion: number;
  /** 編集開始時の current.id */
  expectedContentVersionId: string;
  /** 全置換content（title/section空文字はserver 400。clientでも同じ条件で拒否する） */
  content: ManualEditContent;
}

export interface ManualArticleSaveOutcome {
  operationKey: string;
  /** receipt照合（対象/版一致）を通過した確定receipt */
  receipt: ManualEditReceipt;
}

function readErrorBody(body: unknown): ManualErrorBody | null {
  if (body !== null && typeof body === "object" && "error" in body) {
    const error = (body as { error: unknown }).error;
    if (typeof error === "string") {
      return { error };
    }
  }
  return null;
}

function parseConflictCode(value: unknown): ManualEditConflictCode | null {
  if (
    typeof value === "string" &&
    (MANUAL_EDIT_CONFLICT_CODES as readonly string[]).includes(value)
  ) {
    return value as ManualEditConflictCode;
  }
  return null;
}

/**
 * PUT /manual/articles/{category}/{slug} — 既存記事の編集保存。
 * operation_keyはここで生成し、応答喪失・5xx時も同じkeyで結果照会できるよう
 * エラーに保持する。200応答はreceipt照合を通過したときだけ成功として返す。
 */
export async function saveManualArticleEdit(
  params: SaveManualArticleEditParams,
): Promise<ManualArticleSaveOutcome> {
  const operationKey = newManualOperationKey();
  const body: UpsertManualArticleRequest = {
    operation_key: operationKey,
    article_id: params.articleId,
    expected_version: params.expectedVersion,
    expected_content_version_id: params.expectedContentVersionId,
    content: params.content,
  };
  try {
    const { data } = await axios.put<ManualEditReceipt>(
      `/v1/manual/articles/${params.category}/${params.slug}`,
      body,
    );
    verifyManualEditReceipt(data, {
      operationKey,
      articleId: params.articleId,
      expectedVersion: params.expectedVersion,
      expectedContentVersionId: params.expectedContentVersionId,
    });
    return { operationKey, receipt: data };
  } catch (err) {
    if (err instanceof ManualReceiptMismatchError) {
      throw err;
    }
    if (isAxiosError(err)) {
      const status = err.response?.status;
      if (status === 409) {
        const parsed = readErrorBody(err.response?.data);
        throw new ManualEditConflictError({
          operationKey,
          conflict: {
            error: parsed?.error ?? "",
            // wire違反（code欠落等）の409は自動再取得につながらない停止系codeへ倒す
            code:
              parseConflictCode((err.response?.data as { code?: unknown } | undefined)?.code) ??
              "target_mismatch",
          },
        });
      }
      throw new ManualArticleSaveError({
        operationKey,
        status,
        message: `manual article save failed: ${status ?? "no response"}`,
        requestError: err,
      });
    }
    throw new ManualArticleSaveError({
      operationKey,
      message: "manual article save request could not be sent",
      requestError: err,
    });
  }
}
