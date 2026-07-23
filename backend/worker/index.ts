// P4-2〜P4-4: Cloudflare Worker エントリポイント。
//
// 役割は薄いプロキシのみ: 全 HTTP リクエストを Container(Go/Gin API, :8080) へフォワードする。
// ビジネスロジックは一切持たない（Container内のGoアプリケーション側が担当）。
//
// DB接続方針(docs/ops/infra/_archive/migration-cloudflare.md 試行9): Hyperdrive は Container 内非対応
// (https://github.com/cloudflare/containers/issues/97 — Container は通常の Linux プロセスであり
// Workers runtime 固有の Hyperdrive バインディングを直接利用できない)。そのため Container 内の
// Go API は PlanetScale へ直結する(DB_HOST 等を envVars で直接注入。sslmode=require)。
// Worker 側の HYPERDRIVE バインディングは Phase 4 では未使用(将来 Worker 自身が直接 DB を
// 触るユースケースが増えた場合のために wrangler.jsonc の binding 自体は残す)。
import { Container, getContainer } from "@cloudflare/containers";
import { env } from "cloudflare:workers";
import { isAuthorizedMigrateRequest, toMigrateResponse, type MigrateExecResult } from "./migrate-exec";
import { dispatchScheduledEvent } from "./scheduled-handler";
import {
  SchedulerCoordinator,
  runScheduledPlan,
  type SchedulerControl,
  type ScheduledRunResult,
} from "./scheduler-coordinator";
import {
  SCHEDULER_NAME,
  isScheduledJobsInternalPath,
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
    // 接続プール上限(wrangler.jsonc vars 参照 — スロット枯渇防止のため CF では低値必須)
    DB_MAX_OPEN_CONNS: env.DB_MAX_OPEN_CONNS,
    DB_MAX_IDLE_CONNS: env.DB_MAX_IDLE_CONNS,

    JWT_SECRET: env.JWT_SECRET,
    INTEGRATION_ENCRYPTION_KEY: env.INTEGRATION_ENCRYPTION_KEY,

    // H2: Worker→Container 経路の信頼プロキシ CIDR(rate-limit bypass 防止)。
    // 値の根拠は docs/ops/infra/_archive/migration-cloudflare.md 試行9(実測ログに基づき決定)参照。
    TRUSTED_PROXY_CIDR: env.TRUSTED_PROXY_CIDR,

    CORS_ALLOWED_ORIGIN: env.CORS_ALLOWED_ORIGIN,
    FRONTEND_URL: env.FRONTEND_URL,

    // SMTP(空文字許容 = 送信無効)
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
    await this.startAndWaitForPorts(this.defaultPort);

    const rawContainer = this.ctx.container;
    if (!rawContainer) {
      throw new Error("container is not available on DurableObjectState");
    }

    // 実測判明(試行10): 低レベル exec() は起動時 envVars を継承しない(docker exec と異なり
    // 新規プロセスは素の環境で起動される)。migrate バイナリが読む DB_* のみを明示的に渡す
    // (code-reviewer指摘 LOW — this.envVars 全体ではなく最小権限の subset に絞る。
    // JWT_SECRET/SMTP等の非DB値をmigrateプロセスに渡す必要はない)。
    const migrateEnv: Record<string, string> = {
      DB_HOST: this.envVars.DB_HOST,
      DB_PORT: this.envVars.DB_PORT,
      DB_USER: this.envVars.DB_USER,
      DB_PASSWORD: this.envVars.DB_PASSWORD,
      DB_NAME: this.envVars.DB_NAME,
      DB_SSL_MODE: this.envVars.DB_SSL_MODE,
    };

    const proc = await rawContainer.exec(["/app/migrate"], {
      env: migrateEnv,
      stdout: "pipe",
      stderr: "pipe",
    });

    const timeoutMs = AnimalEkarteApiContainer.MIGRATE_TIMEOUT_MS;
    const timeout = new Promise<never>((_, reject) => {
      setTimeout(() => reject(new Error(`migrate exec timed out after ${timeoutMs}ms`)), timeoutMs);
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
      stderr: decoder.decode(output.stderr),
    };
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
      ),
    );
  }

  // 外部 HTTP からは公開しない。Workers RPC binding 経由の運用制御だけを許可し、
  // 新規シークレットや既存シークレットの用途流用を避ける。
  async setScheduledJobsPaused(paused: boolean): Promise<SchedulerControl> {
    if (typeof paused !== "boolean") {
      throw new Error("paused must be a boolean");
    }
    return new SchedulerCoordinator(this.ctx.storage).setPaused(paused);
  }

  async getScheduledJobsControl(): Promise<SchedulerControl> {
    return new SchedulerCoordinator(this.ctx.storage).getControl();
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
    if (isScheduledJobsInternalPath(url.pathname)) {
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
    } catch (err) {
      // Container 起動失敗(イメージ・メモリ等)・タイムアウト時、Workers既定の500本文では
      // フロントエンドがJSONエラーとして解釈できないため、明示的なフォールバックを返す。
      console.error("container fetch failed", err);
      return new Response(JSON.stringify({ error: "service_unavailable" }), {
        status: 503,
        headers: { "Content-Type": "application/json" },
      });
    }
  },

  async scheduled(controller: ScheduledController, env: Env): Promise<void> {
    const coordinator = getContainer(env.API_CONTAINER, SCHEDULER_NAME);
    try {
      await dispatchScheduledEvent(controller, (cron, scheduledTime) =>
        coordinator.runScheduledJobs(cron, scheduledTime),
      );
    } catch {
      // cron と scheduledTime は Cloudflare 設定由来で PII/secret を含まない。
      // Go 応答本文や例外詳細は記録せず、失敗種別は永続 run ledger で確認する。
      console.error("scheduled invocation failed", {
        scheduler: SCHEDULER_NAME,
        cron: controller.cron,
        scheduled_time: controller.scheduledTime,
      });
      throw new Error("scheduled invocation failed");
    }
  },
};

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
    return toMigrateResponse(result);
  } catch (err) {
    console.error("migrate exec failed", err);
    return new Response(JSON.stringify({ error: "migrate_exec_failed" }), {
      status: 500,
      headers: { "Content-Type": "application/json" },
    });
  }
}
