/**
 * staff 資格情報の任意項目バリデーション。
 *
 * backend は account を持たない staff を許容する（`001_init.sql` の
 * `staffs.account_id` は nullable、seed `003_demo/staffs.csv` は37件中20件が空、
 * `resource` 種別は部屋・機材等の非人物リソースで原理的にログインを持ち得ない）。
 * したがって email / password を必須化すると正当な業務データを作成できなくなる。
 *
 * 本モジュールは「未入力なら検証しない。入力されたら形式を検証する」契約を提供する。
 * password の要件は backend の `auth.ValidatePassword`
 * （`backend/internal/auth/http_password.go`）と同一に保つこと。
 */

/** 入力された場合のみ形式を検証する。未入力は null（エラーなし）。 */
export function validateOptionalEmail(email: string): string | null {
  if (email === "") return null;
  // ローカル部・ドメイン部ともに空白を許さず、ドメインにドットを要求する。
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
    return "メールアドレスの形式が正しくありません";
  }
  return null;
}

const PASSWORD_MIN_LENGTH = 8;
const PASSWORD_MAX_BYTES = 72;

/**
 * 入力された場合のみ強度を検証する。未入力は null（エラーなし）。
 * backend `auth.ValidatePassword` と同契約: 8文字以上（コードポイント数）・
 * 72バイト以下・英字と数字の両方を含む。
 */
export function validateOptionalPassword(password: string): string | null {
  if (password === "") return null;

  // backend は utf8.RuneCountInString で数えるため、サロゲートペアを1文字として
  // 扱う [...password] を使う（password.length は UTF-16 単位で過大に数える）。
  if ([...password].length < PASSWORD_MIN_LENGTH) {
    return `パスワードは${PASSWORD_MIN_LENGTH}文字以上で入力してください`;
  }
  if (new TextEncoder().encode(password).length > PASSWORD_MAX_BYTES) {
    return `パスワードは${PASSWORD_MAX_BYTES}バイト以下で入力してください`;
  }

  const hasLetter = /\p{L}/u.test(password);
  const hasDigit = /\p{Nd}/u.test(password);
  if (!hasLetter || !hasDigit) {
    return "パスワードは英字と数字の両方を含めてください";
  }
  return null;
}
