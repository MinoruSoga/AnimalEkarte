// P4-5(試行10): `/_internal/migrate` の認証・レスポンス整形ロジック。
// `Container.exec()` 呼び出し自体(DOインスタンスの`this.ctx`が必要)は index.ts の
// AnimalEkarteApiContainer クラス側に置き、ここでは Worker の fetch() から呼べる
// 純粋関数のみを置く(unit test容易性のため分離)。

export interface MigrateExecResult {
  exitCode: number;
  stdout: string;
  stderr: string;
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

/**
 * migrate exec は Container 起動 env を継承しない。DB_* に加え、ログイン seed が
 * 読む APP_ENV だけを足す。JWT/SMTP は渡さない。
 */
export function attachLoginSeedMigrateEnv(
  dbEnv: Record<string, string>,
  appEnv: string | undefined,
): Record<string, string> {
  const migrateEnv: Record<string, string> = { ...dbEnv };
  if (appEnv !== undefined && appEnv !== "") {
    migrateEnv.APP_ENV = appEnv;
  }
  return migrateEnv;
}
