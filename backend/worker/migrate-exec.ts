// P4-5(試行10): `/_internal/migrate` の認証・レスポンス整形ロジック。
// `Container.exec()` 呼び出し自体(DOインスタンスの`this.ctx`が必要)は index.ts の
// AnimalEkarteApiContainer クラス側に置き、ここでは Worker の fetch() から呼べる
// 純粋関数のみを置く(unit test容易性のため分離)。

export interface MigrateExecResult {
  exitCode: number;
  stdout: string;
  stderr: string;
}

// migrate 専用の named Container インスタンス。既定の singleton は通常
// トラフィックを捌くため、そこでコンテナを再起動すると in-flight リクエストを
// 落とす。専用インスタンスに分離すれば stale image 検出時の再起動が
// トラフィックへ影響しない(scheduler の SCHEDULER_NAME と同じ分離パターン)。
export const MIGRATE_RUNNER_NAME = "animalekarte-migrate-runner-v1" as const;

// deploy パイプライン(cf-run-migrate.sh)がリポジトリ上の最新 migration ファイル名を
// 送るヘッダ。warm コンテナが旧イメージのまま残ると新しい migration ファイルが
// イメージ内に存在しないため、exec 前の `test -f` プローブで stale を検出する。
export const EXPECTED_MIGRATION_HEADER = "X-Expected-Migration";

// マイグレーションファイル名規約: NNN_name.sql (001_init.sql 等)。
// exec のシェル引数に埋め込むためパス区切り・空白・`..` を構造的に排除する。
const MIGRATION_FILENAME_PATTERN = /^[0-9]{3}_[0-9A-Za-z][0-9A-Za-z_.-]{0,120}\.sql$/;

export class InvalidExpectedMigrationError extends Error {
  constructor() {
    super("invalid_expected_migration_header");
    this.name = "InvalidExpectedMigrationError";
  }
}

/**
 * `X-Expected-Migration` ヘッダを検証して返す。未指定・空は null(旧互換:
 * プローブなしで exec する)。形式不正は InvalidExpectedMigrationError —
 * 黙って無視すると CI 側のミス設定を隠すため 400 で失敗させる。
 */
export function expectedMigrationFromRequest(request: Request): string | null {
  const raw = request.headers.get(EXPECTED_MIGRATION_HEADER);
  if (raw === null || raw.trim() === "") {
    return null;
  }
  const value = raw.trim();
  if (!MIGRATION_FILENAME_PATTERN.test(value) || value.includes("..")) {
    throw new InvalidExpectedMigrationError();
  }
  return value;
}

/**
 * 定数時間文字列比較。Cloudflare Workers ランタイムが SubtleCrypto に独自追加している
 * `crypto.subtle.timingSafeEqual`(Web標準APIではないCloudflare拡張)を使う
 * (code-reviewer指摘 LOW — 自前XORループより実装依存のリスクが低い)。
 * 長さが異なる場合は即 false を返す(その時点で長さの違いは秘密の値自体を漏らさないため
 * 許容する — MIGRATE_RUN_SECRET は固定長運用)。
 */
export function timingSafeEqual(a: string, b: string): boolean {
  const enc = new TextEncoder();
  const aBytes = enc.encode(a);
  const bBytes = enc.encode(b);
  if (aBytes.length !== bBytes.length) {
    return false;
  }
  return crypto.subtle.timingSafeEqual(aBytes, bBytes);
}

/**
 * `Authorization: Bearer <MIGRATE_RUN_SECRET>` を検証する。
 * secret が未設定・空・UTF-8 バイト長 32 未満の場合は常に false
 * (誤って認証をバイパスしない。長さは JS 文字列長ではなく TextEncoder の UTF-8 バイト)。
 */
export function isAuthorizedMigrateRequest(request: Request, secret: string | undefined): boolean {
  if (secret === undefined || new TextEncoder().encode(secret).length < 32) {
    return false;
  }
  const authHeader = request.headers.get("Authorization") ?? "";
  const expected = `Bearer ${secret}`;
  return timingSafeEqual(authHeader, expected);
}

/**
 * exec結果を ECS版(exit code 検証 → 失敗ならabort)と同等の意味論で HTTP レスポンスに変換する。
 * exitCode 0 → 200、非0 → 500(呼び出し元シェルスクリプトが `set -e` 相当で気付けるように)。
 */
export function toMigrateResponse(result: MigrateExecResult): Response {
  return new Response(JSON.stringify(result), {
    status: result.exitCode === 0 ? 200 : 500,
    headers: { "Content-Type": "application/json" },
  });
}

/** Sanitize thrown migrate errors for CI logs (no env/secrets). */
export function sanitizeMigrateFailureMessage(err: unknown): string {
  const raw = err instanceof Error ? err.message : String(err);
  return raw
    .replace(/Bearer\s+\S+/gi, "Bearer [redacted]")
    .replace(/password[=:]\S+/gi, "password=[redacted]")
    .slice(0, 500);
}

export function toMigrateExecFailedResponse(err: unknown): Response {
  return new Response(
    JSON.stringify({
      error: "migrate_exec_failed",
      failure_code: "migrate_exec_failed",
      message: sanitizeMigrateFailureMessage(err),
    }),
    {
      status: 500,
      headers: { "Content-Type": "application/json" },
    },
  );
}

export interface LoginSeedOperatorEnv {
  email?: string;
  name?: string;
  password?: string;
}

function copyNonEmptyEnv(
  target: Record<string, string>,
  key: string,
  value: string | undefined,
): void {
  if (value !== undefined && value !== "") {
    target[key] = value;
  }
}

/**
 * migrate exec は Container 起動 env を継承しない。DB_* に加え、ログイン seed が
 * 読む APP_ENV と任意の SEEDLOGIN_OPERATOR_* だけを足す。JWT/SMTP は渡さない。
 * オペレータ変数が空なら載せない（Go 側は未設定としてスキップする）。
 */
export function attachLoginSeedMigrateEnv(
  dbEnv: Record<string, string>,
  appEnv: string | undefined,
  operatorEnv: LoginSeedOperatorEnv = {},
): Record<string, string> {
  const migrateEnv: Record<string, string> = { ...dbEnv };
  if (appEnv !== undefined && appEnv !== "") {
    migrateEnv.APP_ENV = appEnv;
  }
  copyNonEmptyEnv(migrateEnv, "SEEDLOGIN_OPERATOR_EMAIL", operatorEnv.email);
  copyNonEmptyEnv(migrateEnv, "SEEDLOGIN_OPERATOR_NAME", operatorEnv.name);
  copyNonEmptyEnv(migrateEnv, "SEEDLOGIN_OPERATOR_PASSWORD", operatorEnv.password);
  return migrateEnv;
}
