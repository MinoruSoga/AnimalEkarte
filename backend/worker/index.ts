// P4-2〜P4-4: Cloudflare Worker エントリポイント。
//
// 役割は薄いプロキシのみ: 全 HTTP リクエストを Container(Go/Gin API, :8080) へフォワードする。
// ビジネスロジックは一切持たない（Container内のGoアプリケーション側が担当）。
//
// DB接続方針(docs/ops/infra/_archive/migration-cloudflare.md 試行9): Hyperdrive は Container 内非対応
// (https://github.com/cloudflare/containers/issues/97 — Container は通常の Linux プロセスであり
// Workers runtime 固有の Hyperdrive バインディングを直接利用できない)。そのため Container 内の
// Go API は PlanetScale へ直結する(DB_HOST 等を envVars で直接注入。
// sslmode=verify-full + sslrootcert=system)。
// Worker 側の HYPERDRIVE バインディングは Phase 4 では未使用(将来 Worker 自身が直接 DB を
// 触るユースケースが増えた場合のために wrangler.jsonc の binding 自体は残す)。
import { Container, getContainer } from "@cloudflare/containers";
import { env } from "cloudflare:workers";
import { isAuthorizedMigrateRequest, toMigrateResponse, type MigrateExecResult } from "./migrate-exec";
import { dispatchScheduledEvent } from "./scheduled-handler";
import {
  SchedulerCoordinator,
  runScheduledPlan,
  type SchedulerControlCommand,
  type SchedulerControlOperation,
  type SchedulerManualCommand,
  type SchedulerManualOperation,
  type SchedulerStatus,
  type ScheduledRunResult,
} from "./scheduler-coordinator";
import {
  SCHEDULER_OPS_PREFIX,
  handleSchedulerOpsRequest,
  isInternalProxyPath,
  notifySchedulerFailures,
  type SchedulerAlertConfig,
  type SchedulerOpsAuthConfig,
} from "./scheduler-ops";
import {
  SCHEDULER_NAME,
  runScheduledJobRequest,
} from "./scheduled-jobs";

export class AnimalEkarteApiContainer extends Container<Env> {
  defaultPort = 8080;
  // AC-5: scale-to-zero 検証用。アイドル10分でコンテナを停止する
  // (docs/ops/infra/_archive/migration-cloudflare.md の想定コスト・「通常操作 10 分間程度」の負荷スモーク方針に合わせる)。
  sleepAfter = "10m";

  // Container 起動時に注入する環境変数。Go 側の config.Load()/main.go が読む
  // os.Getenv キーと1:1で対応させる(対応表は wrangler.jsonc のコメント参照)。
  // 値そのものはここに書かず、必ず Worker の vars/secrets(env.*)経由で渡す。
  envVars = {
    PORT: "8080",
    GIN_MODE: env.GIN_MODE,
    LOG_LEVEL: env.LOG_LEVEL,

    // DB — PlanetScale 直結(Hyperdrive 非経由。上記コメント参照)
    DB_HOST: env.DB_HOST,
    DB_PORT: env.DB_PORT,
    DB_USER: env.DB_USER,
    DB_PASSWORD: env.DB_PASSWORD,
    DB_NAME: env.DB_NAME,
    DB_SSL_MODE: env.DB_SSL_MODE,
    DB_SSL_ROOT_CERT: env.DB_SSL_ROOT_CERT,
    // 接続プール上限(wrangler.jsonc vars 参照 — スロット枯渇防止のため CF では低値必須)
    DB_MAX_OPEN_CONNS: env.DB_MAX_OPEN_CONNS,
    DB_MAX_IDLE_CONNS: env.DB_MAX_IDLE_CONNS,

    JWT_SECRET: env.JWT_SECRET,
    INTEGRATION_ENCRYPTION_KEY: env.INTEGRATION_ENCRYPTION_KEY,
    // DEC-36 / CMD-02: Go requireSchedulerInternalToken の expected 値。
    // 3 ホップ必須: wrangler secrets.required → 本 allowlist → scheduled-jobs.ts ヘッダ。
    // Cutover: secret put を先、デプロイを最後にする。Go の requireSchedulerInternalToken は
    // expected が空だと Worker が何を送っても全リクエストを 401 にする(fail-closed on missing
    // config)ため、secret 不在のままデプロイすると全院バッチが停止する。
    SCHEDULER_INTERNAL_TOKEN: env.SCHEDULER_INTERNAL_TOKEN,

    // H2: Worker→Container 経路の信頼プロキシ CIDR(rate-limit bypass 防止)。
    // 値の根拠は docs/ops/infra/_archive/migration-cloudflare.md 試行9(実測ログに基づき決定)参照。
    TRUSTED_PROXY_CIDR: env.TRUSTED_PROXY_CIDR,

    CORS_ALLOWED_ORIGIN: env.CORS_ALLOWED_ORIGIN,
    FRONTEND_URL: env.FRONTEND_URL,

    // SMTP(releaseではaccount recoveryを成立させるため全項目必須)
    SMTP_HOST: env.SMTP_HOST,
    SMTP_PORT: env.SMTP_PORT,
    SMTP_USER: env.SMTP_USER,
    SMTP_PASS: env.SMTP_PASS,
    SMTP_FROM: env.SMTP_FROM,

    // R2(S3互換)。P2-3(試行8)で発行済みのR2 API Token由来。aws-sdk-go-v2の既定クレデンシャル
    // チェーンに合わせて AWS_* の名前で渡す(R2固有の名前ではない。config.go/s3_*.go参照)。
    STORAGE_TYPE: env.STORAGE_TYPE,
    S3_BUCKET: env.S3_BUCKET,
    S3_REGION: env.S3_REGION,
    S3_SHARED_BUCKET: env.S3_SHARED_BUCKET,
    S3_SHARED_REGION: env.S3_SHARED_REGION,
    S3_ENDPOINT: env.S3_ENDPOINT,
    S3_PUBLIC_BASE_URL: env.S3_PUBLIC_BASE_URL,
    AWS_ACCESS_KEY_ID: env.AWS_ACCESS_KEY_ID,
    AWS_SECRET_ACCESS_KEY: env.AWS_SECRET_ACCESS_KEY,
  };

  // migrate exec のハングアップ対策(code-reviewer指摘 MEDIUM)。pg_advisory_lock が
  // 別プロセスに握られたまま解放されない等の異常時、exec が無期限に応答を待たないようにする。
  // ECS版もタイムアウト付き(backend-deploy.yml `timeout-minutes: 15`)だが、単発execなので短く設定。
  static readonly MIGRATE_TIMEOUT_MS = 120_000;

  // P4-5(試行10): ECS `animalekarte-stg-migrate` one-shot task 相当。
  // `@cloudflare/containers`(0.3.7)の Container ラッパーは exec() を公開していないため、
  // DurableObjectState が保持する低レベル container(`this.ctx.container`)を直接使う。
  // Worker 側からは RPC(DurableObjectStub のメソッド呼び出し)として `container.runMigrate()`
  // を叩ける(Container は DurableObject を継承しており、fetch/alarm以外の public メソッドも
  // 標準の Workers RPC で呼び出し可能)。
  //
  // DB_RESET は渡す env に含めていないため、Go側 os.Getenv("DB_RESET") は常に空文字 = false
  // (このexecでも上書きしない。破壊的操作は本エンドポイントでは不可能)。
  async runMigrate(): Promise<MigrateExecResult> {
    const rawContainer = this.ctx.container;
    if (!rawContainer) {
      throw new Error("container is not available on DurableObjectState");
    }

    // API 本体が起動失敗していても migrate を走らせるため、
    // まず sleep でコンテナを上げてから exec する（port wait しない）。
    try {
      await this.start({
        entrypoint: ["sleep", "infinity"],
      });
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      return {
        exitCode: -1,
        stdout: "",
        stderr: `container_start_failed: ${message}`,
      };
    }

    const migrateEnv: Record<string, string> = {
      DB_HOST: this.envVars.DB_HOST,
      DB_PORT: this.envVars.DB_PORT,
      DB_USER: this.envVars.DB_USER,
      DB_PASSWORD: this.envVars.DB_PASSWORD,
      DB_NAME: this.envVars.DB_NAME,
      DB_SSL_MODE: this.envVars.DB_SSL_MODE,
      DB_SSL_ROOT_CERT: this.envVars.DB_SSL_ROOT_CERT,
    };

    // 診断用: secret 値は出さず、有無と長さだけ
    const diag = [
      `DB_HOST_set=${Boolean(migrateEnv.DB_HOST)}`,
      `DB_HOST_len=${(migrateEnv.DB_HOST || "").length}`,
      `DB_USER_set=${Boolean(migrateEnv.DB_USER)}`,
      `DB_USER_len=${(migrateEnv.DB_USER || "").length}`,
      `DB_PASSWORD_set=${Boolean(migrateEnv.DB_PASSWORD)}`,
      `DB_PASSWORD_len=${(migrateEnv.DB_PASSWORD || "").length}`,
      `DB_NAME=${migrateEnv.DB_NAME || ""}`,
      `DB_PORT=${migrateEnv.DB_PORT || ""}`,
      `DB_SSL_MODE=${migrateEnv.DB_SSL_MODE || ""}`,
    ].join(" ");

    try {
      const proc = await rawContainer.exec(["/app/migrate"], {
        env: migrateEnv,
        stdout: "pipe",
        stderr: "pipe",
      });

      const timeoutMs = AnimalEkarteApiContainer.MIGRATE_TIMEOUT_MS;
      const timeout = new Promise<never>((_, reject) => {
        setTimeout(
          () => reject(new Error(`migrate exec timed out after ${timeoutMs}ms`)),
          timeoutMs,
        );
      });

      let output;
      try {
        output = await Promise.race([proc.output(), timeout]);
      } catch (err) {
        proc.kill();
        throw err;
      }

      const decoder = new TextDecoder();
      return {
        exitCode: output.exitCode,
        stdout: decoder.decode(output.stdout),
        stderr: `${diag}\n${decoder.decode(output.stderr)}`,
      };
    } finally {
      try {
        await this.stop();
      } catch {
        // ignore stop errors after migrate
      }
    }
  }

  // BE9-3: Cron 専用の named DO からのみ呼ばれる RPC。
  // Durable Object storage が pause・global lease・fence・run ledger を保持し、
  // container の scale-to-zero や Worker の再起動後も重複実行を防ぐ。
  async runScheduledJobs(
    cron: string,
    scheduledTime: number,
  ): Promise<readonly ScheduledRunResult[]> {
    const coordinator = new SchedulerCoordinator(this.ctx.storage);
    return runScheduledPlan(coordinator, cron, scheduledTime, (request) =>
      runScheduledJobRequest(
        (internalRequest) => this.containerFetch(internalRequest),
        request,
        this.envVars.SCHEDULER_INTERNAL_TOKEN,
      ),
    );
  }

  async getScheduledJobsStatus(limit: number): Promise<SchedulerStatus> {
    return new SchedulerCoordinator(this.ctx.storage).getStatus(limit);
  }

  async consumeScheduledJobsOpsRateLimit(
    actorPrincipal: string,
    now: number,
  ) {
    return new SchedulerCoordinator(this.ctx.storage).consumeOpsRateLimit(
      actorPrincipal,
      now,
    );
  }

  async setScheduledJobsControl(
    command: SchedulerControlCommand,
  ): Promise<SchedulerControlOperation> {
    return new SchedulerCoordinator(this.ctx.storage).setControl(command);
  }

  async runScheduledJobManually(
    command: SchedulerManualCommand,
  ): Promise<SchedulerManualOperation> {
    const coordinator = new SchedulerCoordinator(this.ctx.storage);
    return coordinator.runManual(command, (scheduledRequest) =>
      runScheduledJobRequest(
        (internalRequest) => this.containerFetch(internalRequest),
        scheduledRequest,
        this.envVars.SCHEDULER_INTERNAL_TOKEN,
      ),
    );
  }

  /**
   * API バイナリが数秒以内に exit するか確認する（値は出さずログ文字列のみ）。
   * sleep コンテナ上で /app/api を短時間 exec する。
   */
  async diagnoseApiBoot(): Promise<{ ok: boolean; detail: string }> {
    const rawContainer = this.ctx.container;
    if (!rawContainer) {
      return { ok: false, detail: "container unavailable" };
    }
    try {
      await this.start({ entrypoint: ["sleep", "infinity"] });
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      return { ok: false, detail: `start_failed: ${message.slice(0, 300)}` };
    }

    const env: Record<string, string> = {};
    for (const [key, value] of Object.entries(this.envVars)) {
      if (typeof value === "string") {
        env[key] = value;
      }
    }
    // 長さだけ（値なし）
    const lens = [
      `JWT_len=${(env.JWT_SECRET || "").length}`,
      `DB_HOST_len=${(env.DB_HOST || "").length}`,
      `DB_USER_len=${(env.DB_USER || "").length}`,
      `DB_PASS_len=${(env.DB_PASSWORD || "").length}`,
      `INTEG_len=${(env.INTEGRATION_ENCRYPTION_KEY || "").length}`,
      `AWS_KEY_len=${(env.AWS_ACCESS_KEY_ID || "").length}`,
      `AWS_SEC_len=${(env.AWS_SECRET_ACCESS_KEY || "").length}`,
      `SMTP_HOST_len=${(env.SMTP_HOST || "").length}`,
      `SMTP_USER_len=${(env.SMTP_USER || "").length}`,
      `SMTP_PASS_len=${(env.SMTP_PASS || "").length}`,
      `SCHED_INT_len=${(env.SCHEDULER_INTERNAL_TOKEN || "").length}`,
      `GIN_MODE=${env.GIN_MODE || ""}`,
      `STORAGE=${env.STORAGE_TYPE || ""}`,
    ].join(" ");

    try {
      const proc = await rawContainer.exec(["/app/api"], {
        env,
        stdout: "pipe",
        stderr: "pipe",
      });
      const outputOrTimeout = await Promise.race([
        proc.output().then((output) => ({ kind: "exit" as const, output })),
        new Promise<{ kind: "running" }>((resolve) => {
          setTimeout(() => resolve({ kind: "running" }), 5000);
        }),
      ]);
      if (outputOrTimeout.kind === "running") {
        try {
          proc.kill();
        } catch {
          // ignore
        }
        return { ok: true, detail: `api_still_running_after_5s ${lens}` };
      }
      const decoder = new TextDecoder();
      const stdout = decoder.decode(outputOrTimeout.output.stdout).slice(0, 800);
      const stderr = decoder.decode(outputOrTimeout.output.stderr).slice(0, 800);
      return {
        ok: false,
        detail: `api_exited code=${outputOrTimeout.output.exitCode} ${lens} stdout=${stdout} stderr=${stderr}`,
      };
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      return { ok: false, detail: `api_exec_failed: ${message.slice(0, 400)} ${lens}` };
    } finally {
      try {
        await this.stop();
      } catch {
        // ignore
      }
    }
  }
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    // P4-5(試行10): /_internal/* は Worker でのみ処理し、Container の Gin へは
    // ルーティングしない(通常プロキシ経路より前に分岐させる)。
    const url = new URL(request.url);
    if (url.pathname === "/_internal/migrate") {
      return handleMigrateRequest(request, env);
    }
    if (
      url.pathname === SCHEDULER_OPS_PREFIX ||
      url.pathname.startsWith(`${SCHEDULER_OPS_PREFIX}/`)
    ) {
      const coordinator = getContainer(env.API_CONTAINER, SCHEDULER_NAME);
      return handleSchedulerOpsRequest(
        request,
        schedulerOpsAuthConfig(env),
        coordinator,
        Date.now(),
        async (operation) => {
          if (operation.result !== undefined) {
            await notifySchedulerFailures(
              [operation.result],
              schedulerAlertConfig(env),
            );
          }
        },
      );
    }
    if (isInternalProxyPath(url.pathname)) {
      return new Response(JSON.stringify({ error: "not_found" }), {
        status: 404,
        headers: { "Content-Type": "application/json" },
      });
    }

    // H2/AC-2: containerFetch は既定では X-Forwarded-For を注入しない(試行9の実測で確認—
    // Cloudflare公式 "Env Vars and Secrets" 例に倣い、CF-Connecting-IP を明示的に転記する)。
    // これが無いと Go側 c.ClientIP() は常に内部プロキシIP(実測: 10.1.0.0)を返し、
    // レート制限が全ユーザー共有バケットになる(セキュリティ上のバイパスではないが機能バグ)。
    const headers = new Headers(request.headers);
    const connectingIP = request.headers.get("CF-Connecting-IP");
    if (connectingIP) {
      headers.set("X-Forwarded-For", connectingIP);
    } else {
      // security-reviewer指摘(L-2): CF-Connecting-IPが無い場合にクライアント制御のXFFを
      // そのまま転送しない(defense-in-depth。通常のCloudflareエッジ経由では常に付与される)。
      headers.delete("X-Forwarded-For");
    }
    const forwardedRequest = new Request(request, { headers });

    const container = getContainer(env.API_CONTAINER);
    try {
      return await container.fetch(forwardedRequest);
    } catch {
      // Container 起動失敗(イメージ・メモリ等)・タイムアウト時、Workers既定の500本文では
      // フロントエンドがJSONエラーとして解釈できないため、明示的なフォールバックを返す。
      // 例外本文・stack は外部応答や機密値を含み得るためログへ出さない。
      console.error("container fetch failed", {
        event: "container_fetch_failed",
        failure_code: "container_unavailable",
      });
      return new Response(JSON.stringify({ error: "service_unavailable" }), {
        status: 503,
        headers: { "Content-Type": "application/json" },
      });
    }
  },

  async scheduled(controller: ScheduledController, env: Env): Promise<void> {
    const coordinator = getContainer(env.API_CONTAINER, SCHEDULER_NAME);
    try {
      await dispatchScheduledEvent(
        controller,
        async (cron, scheduledTime) => {
          const results = await coordinator.runScheduledJobs(
            cron,
            scheduledTime,
          );
          await notifySchedulerFailures(results, schedulerAlertConfig(env));
          return results;
        },
      );
    } catch {
      // cron と scheduledTime は Cloudflare 設定由来で PII/secret を含まない。
      // Go 応答本文や例外詳細は記録せず、失敗種別は永続 run ledger で確認する。
      console.error("scheduled invocation failed", {
        event: "scheduler_invocation_failed",
        scheduler: SCHEDULER_NAME,
        cron: controller.cron,
        scheduled_time: controller.scheduledTime,
        failure_code: "scheduled_invocation_failed",
      });
      throw new Error("scheduled invocation failed");
    }
  },
};

function schedulerAlertConfig(env: Env): SchedulerAlertConfig {
  return {
    environment: env.SCHEDULER_ENVIRONMENT || "unconfigured",
    webhookURL: env.SCHEDULER_ALERT_WEBHOOK_URL,
    webhookSecret: env.SCHEDULER_ALERT_WEBHOOK_SECRET,
    allowedHost: env.SCHEDULER_ALERT_ALLOWED_HOST,
  };
}

function schedulerOpsAuthConfig(env: Env): SchedulerOpsAuthConfig {
  return {
    automationSecret: env.SCHEDULER_OPS_SECRET,
    accessTeamDomain: env.SCHEDULER_ACCESS_TEAM_DOMAIN,
    accessAudience: env.SCHEDULER_ACCESS_AUDIENCE,
  };
}

// P4-5(試行10): migrate one-shot 管理エンドポイント。POST + Bearer secret必須。
// GET/その他メソッドは405、secret不一致・未設定は401(存在の有無を分けない — enumeration対策)。
async function handleMigrateRequest(request: Request, env: Env): Promise<Response> {
  if (request.method !== "POST") {
    return new Response(JSON.stringify({ error: "method_not_allowed" }), {
      status: 405,
      headers: { "Content-Type": "application/json", Allow: "POST" },
    });
  }

  if (!isAuthorizedMigrateRequest(request, env.MIGRATE_RUN_SECRET)) {
    return new Response(JSON.stringify({ error: "unauthorized" }), {
      status: 401,
      headers: { "Content-Type": "application/json" },
    });
  }

  const container = getContainer(env.API_CONTAINER);
  try {
    const result = await container.runMigrate();
    // migrate 成功後も API が起動できないケースを切り分ける（secret 値は出さない）
    let apiDiag: { ok: boolean; detail: string } | undefined;
    try {
      apiDiag = await container.diagnoseApiBoot();
    } catch (diagErr) {
      const message = diagErr instanceof Error ? diagErr.message : String(diagErr);
      apiDiag = { ok: false, detail: `diag_threw: ${message.slice(0, 400)}` };
    }
    const response = toMigrateResponse(result);
    if (result.exitCode === 0) {
      const body = {
        ...result,
        api_boot: apiDiag,
      };
      return new Response(JSON.stringify(body), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    }
    return response;
  } catch (err) {
    const message = err instanceof Error ? err.message : "unknown";
    // 秘密値は含めない前提の例外メッセージのみ返す
    console.error("migrate exec failed", {
      event: "migrate_exec_failed",
      failure_code: "migrate_exec_failed",
      message,
    });
    return new Response(
      JSON.stringify({
        error: "migrate_exec_failed",
        message: message.slice(0, 500),
      }),
      {
        status: 500,
        headers: { "Content-Type": "application/json" },
      },
    );
  }
}
