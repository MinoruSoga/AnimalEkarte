# Animal Ekarte Backend

動物病院電子カルテシステムの Go/Gin API。

## Technology

- Go 1.25
- Gin v1.12
- GORM v1.31.2
- PostgreSQL 18
- OpenAPI 3.0: [`docs/api.yaml`](docs/api.yaml)
- `slog` structured logging
- golangci-lint

## Architecture policy

Go/Gin公式は特定の application architecture や folder tree を規定しない。この backend では、[ADR-006](../docs/architecture/adr/006-backend-domain-package-boundaries.md)によりdomain/capability-firstのmodular monolithをproject decisionとして採用し、packageを凝集性、利用者、依存方向、変更単位で設計する。BE9移行は2026-07-24にcode complete（release pending）となり、旧`internal/handler`は削除済み、旧`internal/service`と`internal/repository`はtest-onlyでproduction実装は0件である。旧3layerは新規production codeの追加先ではなく、境界の正本はADR-006と[boundary map](../docs/architecture/be9-2a-boundary-map.md)とする。

正本:

- [Go/Gin Backend Guidelines](../.claude/rules/go-gin-backend-guidelines.md)
- [Backend Coding Rules](CODING_RULES.md)
- [Backend Application Invariants](../.claude/refs/backend-application-invariants.md)
- [ADR-006: Backend Domain Package Boundaries](../docs/architecture/adr/006-backend-domain-package-boundaries.md)
- [Architecture Overview](../docs/architecture/overview.md)

## Module organization

- `cmd/`: executable commands
- `internal/`: module 外から import させない application code
- `migrations/`: versioned SQL migrations
- `docs/api.yaml`: API contract

`internal/`配下の固定3層treeを新設する規約はない。新規production実装はADR-006のtarget domain packageへ置き、domain内のsubpackageは実際に独立した責務・利用者・依存方向が確認できる場合だけ追加する。

## Development

project root から Docker 経由で実行する。host で Go command を直接実行しない。

```bash
# development containers
make build

# API logs
make logs-api

# restart API
make restart-api
```

全 project の test/lint は高出力のため自動実行しない。変更に合わせて scoped verification を使う。

```bash
docker compose exec backend go test ./internal/<package>/...
docker compose exec backend golangci-lint run ./internal/<package>/...
```

詳細な実行制約は [`.claude/CLAUDE.md`](../.claude/CLAUDE.md) を参照する。

## API changes

endpoint を追加・変更するときは、folder の作成順ではなく contract と request lifecycle を基準に進める。

1. authentication、authorization、ownership、tenant boundary を定義する。
2. request/response/status/error contract を `docs/api.yaml` に反映する。
3. route group と middleware scope を設計する。
4. 凝集した package に、必要最小限の type と dependency を実装する。
5. DB schema 変更があれば versioned migration を追加する。startup `AutoMigrate` に依存しない。
6. `httptest`、integration test、cross-tenant test を risk に応じて追加する。
7. cancellation、error mapping、log の機密情報、graceful shutdown への影響を確認する。

Handler → Service → Repository、repository interface、DTO の固定配置、特定 helper は Go/Gin公式の追加手順ではない。

## Runtime configuration

主な環境変数は `PORT`、database connection、`GIN_MODE`、`LOG_LEVEL`。credential や secret を code/README に固定せず、起動時に必須値を検証する。実際の deployment 設定は infra/ops の正本を参照する。

## Production

- HTTPS と明示的な trusted proxy/CORS 設定を使う。
- `http.Server` に workload に合う timeout/limit を設定する。
- SIGINT/SIGTERM から timeout 付き graceful shutdown を行う。
- DB/worker 等の resource を安全な順序で close する。
- unknown error の内部情報と個人情報を response/log に出さない。
