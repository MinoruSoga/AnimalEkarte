import { describe, it, expect } from "vitest";

import {
  validateOptionalEmail,
  validateOptionalPassword,
} from "./validate-credentials";

// R-1③ の裁定: backend は account を持たない staff を許容する（`resource` 種別は
// 部屋・機材等の非人物リソースで原理的にログインを持ち得ない）。したがって
// email / password は「未入力なら検証しない」が正であり、必須化してはならない。
describe("validateOptionalEmail", () => {
  it("未入力は許可する（アカウント無しstaffを作成できる）", () => {
    expect(validateOptionalEmail("")).toBeNull();
  });

  it("空白のみは未入力として扱わず形式エラーにする", () => {
    expect(validateOptionalEmail("   ")).not.toBeNull();
  });

  it("形式が正しいメールアドレスを許可する", () => {
    expect(validateOptionalEmail("staff@noavet.jp")).toBeNull();
  });

  it.each(["noavet.jp", "staff@", "@noavet.jp", "staff @noavet.jp", "staff@noavet"])(
    "形式不正 %s を拒否する",
    (invalid) => {
      expect(validateOptionalEmail(invalid)).not.toBeNull();
    },
  );
});

// backend の auth.ValidatePassword と同契約にする:
// 8文字以上（rune 数）/ 72バイト以下 / 英字と数字の両方を含む。
describe("validateOptionalPassword", () => {
  it("未入力は許可する（アカウント無しstaffを作成できる）", () => {
    expect(validateOptionalPassword("")).toBeNull();
  });

  it("8文字以上で英数字を含むものを許可する", () => {
    expect(validateOptionalPassword("Fe12pass1")).toBeNull();
  });

  it("8文字未満を拒否する", () => {
    expect(validateOptionalPassword("Ab3xyz9")).not.toBeNull();
  });

  it("rune数で数えるため全角8文字は長さ要件を満たす", () => {
    // 長さは満たすが英字/数字を含まないため、別の理由で拒否される。
    expect(validateOptionalPassword("ぱすわーどです")).not.toBeNull();
  });

  it("数字のみを拒否する", () => {
    expect(validateOptionalPassword("123456789")).not.toBeNull();
  });

  it("英字のみを拒否する", () => {
    expect(validateOptionalPassword("abcdefghi")).not.toBeNull();
  });

  it("72バイトを超えるものを拒否する", () => {
    expect(validateOptionalPassword(`a1${"b".repeat(71)}`)).not.toBeNull();
  });

  it("マルチバイトはバイト数で上限判定する", () => {
    // 「あ」は UTF-8 で3バイト。24文字 = 72バイト + "a1" で上限超過。
    expect(validateOptionalPassword(`a1${"あ".repeat(24)}`)).not.toBeNull();
  });
});
