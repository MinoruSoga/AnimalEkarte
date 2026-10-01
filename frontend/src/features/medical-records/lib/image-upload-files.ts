/**
 * カルテ画像アップロードの受付ルール（file input と ドラッグ&ドロップで共用する正本）。
 * SEC-CS-F08: 件数・1件あたりサイズ・合計バイトを fail-closed で制限する。
 */

/** アップロードボタン経路・ドロップ経路共通の受付 MIME タイプ（input の accept 属性値でもある） */
export const IMAGE_UPLOAD_ACCEPT = "image/jpeg,image/png,image/gif,application/pdf";
/** 撮影専用 input の accept。カメラは PDF を生成しないため画像のみ。 */
export const IMAGE_CAPTURE_ACCEPT = "image/jpeg,image/png,image/gif";

/** 1ファイルあたりの上限（MB） */
const MAX_FILE_SIZE_MB = 10;
export const MAX_FILE_SIZE_BYTES = MAX_FILE_SIZE_MB * 1024 * 1024;
/** SEC-CS-F08: 1回の選択で受け付ける最大ファイル数 */
export const MAX_UPLOAD_FILES = 10;
/** SEC-CS-F08: 1回の選択で受け付ける合計バイト上限（50MiB） */
export const MAX_UPLOAD_BATCH_BYTES = 50 * 1024 * 1024;

const ACCEPTED_MIME_TYPES = new Set(IMAGE_UPLOAD_ACCEPT.split(","));
const ACCEPTED_EXTENSIONS = [".jpg", ".jpeg", ".png", ".gif", ".pdf"];

/**
 * accept 属性が効かないドロップ経路用の形式チェック。
 * 環境によって file.type が空になるケースへ拡張子でフォールバックする。
 */
export function isAcceptedUploadFile(file: File): boolean {
  if (file.type !== "") return ACCEPTED_MIME_TYPES.has(file.type);
  const lowerName = file.name.toLowerCase();
  return ACCEPTED_EXTENSIONS.some((ext) => lowerName.endsWith(ext));
}

export type UploadFilesValidation = { ok: true } | { ok: false; message: string };

/**
 * file input / ドラッグ&ドロップ共通のアップロードバリデーション。
 * SEC-CS-F08-R1: 一部だけを受け付ける部分アップロードはせず、
 * 1件でも違反があれば whole-batch で拒否して message を返す。
 */
export function validateUploadFiles(files: readonly File[]): UploadFilesValidation {
  if (files.length > MAX_UPLOAD_FILES) {
    return {
      ok: false,
      message: `一度にアップロードできるファイルは${MAX_UPLOAD_FILES.toLocaleString()}件までです`,
    };
  }
  const totalBytes = files.reduce((sum, f) => sum + f.size, 0);
  if (totalBytes > MAX_UPLOAD_BATCH_BYTES) {
    return { ok: false, message: "合計ファイルサイズが上限（50MB）を超えています" };
  }
  const oversized = files.filter((f) => f.size > MAX_FILE_SIZE_BYTES);
  if (oversized.length > 0) {
    const names = oversized.map((f) => f.name).join(", ");
    return {
      ok: false,
      message: `ファイルサイズが上限（${MAX_FILE_SIZE_MB}MB）を超えています: ${names}`,
    };
  }
  const rejected = files.filter((f) => !isAcceptedUploadFile(f));
  if (rejected.length > 0) {
    const names = rejected.map((f) => f.name).join(", ");
    return {
      ok: false,
      message: `対応していないファイル形式です: ${names}`,
    };
  }
  return { ok: true };
}
