import { describe, it, expect } from "vitest";

import {
  IMAGE_UPLOAD_ACCEPT,
  isAcceptedUploadFile,
  validateUploadFiles,
  MAX_FILE_SIZE_BYTES,
  MAX_UPLOAD_BATCH_BYTES,
  MAX_UPLOAD_FILES,
} from "./image-upload-files";

function makeFile(name: string, size: number, type = "image/jpeg"): File {
  // size を実体バイトなしで差し替え（50MiB 級の ArrayBuffer 確保を避ける）
  const file = new File([""], name, { type });
  Object.defineProperty(file, "size", { value: size });
  return file;
}

describe("isAcceptedUploadFile", () => {
  it("IMAGE_UPLOAD_ACCEPT の MIME タイプは全て受け付ける", () => {
    for (const mime of IMAGE_UPLOAD_ACCEPT.split(",")) {
      expect(isAcceptedUploadFile(makeFile("a.bin", 1, mime))).toBe(true);
    }
  });

  it("非対応の MIME タイプは拒否する", () => {
    expect(isAcceptedUploadFile(makeFile("a.txt", 1, "text/plain"))).toBe(false);
    expect(isAcceptedUploadFile(makeFile("a.zip", 1, "application/zip"))).toBe(false);
    expect(isAcceptedUploadFile(makeFile("a.webp", 1, "image/webp"))).toBe(false);
  });

  it("file.type が空の環境では拡張子にフォールバックする", () => {
    expect(isAcceptedUploadFile(makeFile("scan.JPG", 1, ""))).toBe(true);
    expect(isAcceptedUploadFile(makeFile("doc.pdf", 1, ""))).toBe(true);
    expect(isAcceptedUploadFile(makeFile("memo.txt", 1, ""))).toBe(false);
  });
});

describe("validateUploadFiles — SEC-CS-F08 共通バリデーション", () => {
  it("空配列は ok", () => {
    expect(validateUploadFiles([])).toEqual({ ok: true });
  });

  it("件数上限超過は whole-batch 拒否", () => {
    const files = Array.from({ length: MAX_UPLOAD_FILES + 1 }, (_, i) =>
      makeFile(`img-${i}.jpg`, 1024),
    );
    const result = validateUploadFiles(files);
    expect(result.ok).toBe(false);
    if (!result.ok) {
      expect(result.message).toBe(
        `一度にアップロードできるファイルは${MAX_UPLOAD_FILES}件までです`,
      );
    }
  });

  it("合計バイト上限超過は拒否", () => {
    const half = Math.floor(MAX_UPLOAD_BATCH_BYTES / 2) + 1;
    const result = validateUploadFiles([makeFile("a.jpg", half), makeFile("b.jpg", half)]);
    expect(result.ok).toBe(false);
    if (!result.ok) {
      expect(result.message).toBe("合計ファイルサイズが上限（50MB）を超えています");
    }
  });

  it("1件でも 10MB 超過があれば拒否しファイル名を含める", () => {
    const result = validateUploadFiles([
      makeFile("ok.jpg", 1024),
      makeFile("huge.jpg", MAX_FILE_SIZE_BYTES + 1),
    ]);
    expect(result.ok).toBe(false);
    if (!result.ok) {
      expect(result.message).toContain("huge.jpg");
    }
  });

  it("非対応形式を含む batch を拒否する（D&D 経路の非画像拒否）", () => {
    const result = validateUploadFiles([
      makeFile("ok.jpg", 1024),
      makeFile("note.txt", 512, "text/plain"),
    ]);
    expect(result.ok).toBe(false);
    if (!result.ok) {
      expect(result.message).toContain("対応していないファイル形式");
      expect(result.message).toContain("note.txt");
    }
  });

  it("全ての条件を満たす batch は ok", () => {
    expect(
      validateUploadFiles([makeFile("ok.jpg", 1024), makeFile("doc.pdf", 2048, "application/pdf")]),
    ).toEqual({ ok: true });
  });
});
