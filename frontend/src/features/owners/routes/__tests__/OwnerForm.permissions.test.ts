import { describe, it, expect } from "vitest";
import { readFileSync } from "fs";
import { resolve } from "path";

/**
 * OwnerForm — accounting:view 権限ガード（会計履歴セクション）
 *
 * 会計履歴は閲覧専用なので `accounting:view` を持たないユーザーには
 * セクション全体を表示しない。OwnerForm 単体は loader / queryClient /
 * auth context のセットアップが大きいため、ここでは render テスト
 * ではなくソースコード解析で実装意図を固定する。
 *
 * 同じパターンは AccountingDetail.test.tsx (#20 Print Performance) でも採用済み。
 */
describe("OwnerForm — accounting:view permission guard", () => {
  const sourceCode = readFileSync(
    resolve(__dirname, "../OwnerForm.tsx"),
    "utf-8",
  );

  it("usePermission(\"accounting\") から canView を取得している", () => {
    expect(sourceCode).toMatch(
      /canView\s*:\s*canViewAccounting\s*\}\s*=\s*usePermission\(\s*["']accounting["']\s*\)/,
    );
  });

  it("OwnerAccountingHistory の描画が canViewAccounting で gating されている", () => {
    // canViewAccounting がガード式に含まれている JSX の条件を検出
    expect(sourceCode).toMatch(
      /isEdit\s*&&\s*ownerId\s*&&\s*canViewAccounting\s*\?[\s\S]*<OwnerAccountingHistory/,
    );
  });

  it("OwnerAccountingHistory は単独条件（権限ガードなし）では描画されない", () => {
    // canViewAccounting を経由せず OwnerAccountingHistory が描画される
    // パスが他に無いことを保証する。lazy 経由の宣言行は許容するが、
    // JSX 描画行 (<OwnerAccountingHistory ...) は 1 箇所のみ。
    const jsxOccurrences = sourceCode.match(/<OwnerAccountingHistory/g) ?? [];
    expect(jsxOccurrences).toHaveLength(1);
  });

  it("会計履歴セクションのコメントに権限制御の意図が明記されている", () => {
    // 設計意図のドキュメンテーション。コードレビューで意図を読み取れることを保証。
    expect(sourceCode).toContain("accounting:view");
  });
});
