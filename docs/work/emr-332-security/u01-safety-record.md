# U01 — Safety Record (安全条件固定・検証境界)

Campaign: emr-332-security-defense / Unit: U01-SAFETY-BASELINE / Attempt: att-u01-1
Source: EMR-332 チケット本文 (sources/emr-332-ticket-body.txt) 工程0+1
Task branch: `feat/emr-332-security-defense` at base `de28e9765a29091339a872a85f1875a947e07e5b`
Recorded by: receiver-emr332-u01-01 (static/documentary only — 本ユニットはステージングへの実通信を一切行わない)

本記録は EMR-332 の全ケース実施に先立つ検証境界 (verification boundary) を固定する。
ここに未確定と記録された条件を満たさない後段ユニットの live チェックは BLOCKED とする。

---

## 1. 対象環境 (owned staging origins)

| 項目 | 値 | 根拠 |
|------|-----|------|
| Frontend origin | `https://animalekarte-stg-frontend.baritech-soga.workers.dev` | frontend/wrangler.jsonc `name: "animalekarte-stg-frontend"`, `workers_dev: true`; SECURITY.md Deployment 前提 (2026-10-03 live 確認済み・200) |
| API origin | `https://animalekarte-stg-api.baritech-soga.workers.dev` | backend/wrangler.jsonc `name: "animalekarte-stg-api"`, `workers_dev: true`; SECURITY.md 同 (`/health`・`/health/db` 200 実測) |
| Custom routes (参考値・非活性) | `stg.noah-karte.com/*` (frontend), `api.stg.noah-karte.com/*` (API) | 両 wrangler.jsonc `routes` — NS 切替未完了のため現時点ではドーマント。検証対象経路は workers.dev origin のみ |

所有権根拠 (ownership basis): 対象は依頼者所有の AnimalEkarte ステージング環境 (Cloudflare アカウント subdomain `baritech-soga`、Worker 名は上記 wrangler 設定と一致)。本番 (`production`)、他テナント、実ユーザーの環境・データは対象外。`noah-karte.com` カスタムドメイン側は route が未活性のため今回の実通信対象に含めない。

## 2. 稼働リビジョン (running staging revision) の確定

- **判定方法**: staging が serve する稼働リビジョンを、タスクブランチ `feat/emr-332-security-defense` のベース `de28e9765a29091339a872a85f1875a947e07e5b` と照合する。照合はデプロイ記録・Worker バージョン識別・変更対象ファイルの既知差分 (本ユニット実施時点ではコード差分なし — documentary unit) で行う。
- **判定規則**: staging-served revision と task branch の内容が一致しない、または稼働リビジョンが確定できない (undetermined) 場合、当該ケースの **live チェックは BLOCKED**。静的・模擬チェックはコード側証拠として別途実施可能だが、staging-observed 判定には丸めない。
- **現時点の状態**: 本ユニットは実通信を行わないため稼働リビジョンの照合は未実施。後段ユニットが staging 実測を行う前にこの照合を完了させること。不一致の場合は差分範囲を記録し、影響ケースのみ限定実施するか全 live チェックを BLOCKED とする。
- チケット記載のローカル参照 HEAD `40384aa3bb92eaa3ea950d56a9a0bf604ff08f2c` と staging 稼働リビジョンの一致は起票時点で未確認 (ticket 本文「現状・確認済みの根拠」より)。

## 3. 単一の許可テナント (permitted tenant identifier)

- **状態: 未プロビジョン (not provisioned)**。検証専用の単一 clinic (tenant) 識別子は本ユニット時点で確定・払い出しされていない。
- **規則**: 許可テナント識別子がプロビジョンされない限り、認証済みのテナントスコープ live チェック (C03, C04, C13 等の認証前提ケース) は **BLOCKED**。他テナント境界は架空属性・DB モック・静的レビューで扱い、実アクセスしない (ticket 共通安全条件)。
- プロビジョンされる場合の形式: 単一の staging clinic ID (数値) を本節へ追記し、その clinic に属さない対象へのアクセスを全て禁止とする。

## 4. 副作用のない対象経路 (side-effect-free target paths)

実通信を許可するのは **未認証の GET / HEAD / OPTIONS のみ**。POST / PUT / PATCH / DELETE は送信しない。対象は以下に限定する (route 詳細は u01-route-middleware-map.md 参照):

| 経路 | メソッド | 備考 |
|------|---------|------|
| `/health`, `/api/v1/health`, `/health/db` | GET | 公開 health。応答は `{"status": ...}` のみで詳細を含まない (base_routes.go:18-46) |
| 公開静的・公開 API 経路の OPTIONS | OPTIONS | Worker edge が CORS preflight を即応答 (worker/index.ts:540-542)。Container は起こさない |
| `/api/liff/:clinicId/settings` | GET | 認証不要の公開設定経路・レートリミット付き (reservation/routes.go:181)。架空 clinicId のみ使用 |
| 認証必須経路への未認証 GET | GET | C01 用: 認証要求/拒否応答 (401) の確認のみ。本文の内容・データは取得しない |

- `/uploads` は STORAGE_TYPE=s3 (staging) のため StaticFS はマウントされない (base_routes.go:81-87)。`/api/v1/uat/*` は環境・Host・Bearer secret の条件付き surface であり本検証の対象外。`/_internal/*` は Worker が外部へ転送しない (404)。
- 副作用確認対象は DB だけでなく session 作成・更新、監査ログ、キャッシュ、キュー、ファイル、メール、外部テレメトリを含む (ticket 共通安全条件)。遮断・観測を保証できない実行単位は BLOCKED。

## 5. ログ観測方法 (log observation)

- **方式**: メモリ上のマーカー捕捉 (memory-only marker capture)。テスト内で構造化ログの出力先をメモリバッファへ差し替え、検証マーカーが記録されること、および秘密扱いの架空値が記録されないことを確認する。
- **禁止**: staging/本番の永続ログ (Wrangler tail、observability 基盤、DB audit_logs の実データ) の実読取は行わない。永続ログを生む実通信は実施しない (ticket: ログ自体の永続化も禁止対象の場合メモリ捕捉に限定)。
- ログ観測で取得してよいのはマーカー・時刻・判定・理由・経路・相関 ID。本文、資格情報、Cookie、トークン、実ユーザー情報は取得しない。

## 6. リクエスト上限 (request bounds)

staging への実通信 (後段ユニットが実施する場合) は以下の上限をすべて満たす場合にのみ許可する:

| 上限 | 値 |
|------|-----|
| 総リクエスト数 | 最大 20 回/検証セッション |
| 間隔 | 最小 3 秒 (逐次実行、並列送信なし) |
| リクエストボディ | 最大 1 KiB (GET/HEAD/OPTIONS のみのため通常 0) |
| メソッド | 未認証 GET / HEAD / OPTIONS のみ |
| マーカーヘッダ | `X-Sec-Chk: SECCHK20261009*` — 値は `SECCHK20261009` + ケース/試行識別の英数字 (例: `SECCHK20261009C01R01`)。実行構文・攻撃ペイロードを含めない |

**停止条件 (stop conditions)**: 以下のいずれかを検知した時点で即停止し、追加読取・再現継続・昇格権限の利用を行わない (ticket「停止・例外・ロールバック」):
1. 想定外の保存・更新・外部送信の兆候 (2xx による作成、副作用呼出し回数 > 0)
2. 対象外通信・対象外ホストへの誘導
3. 個人情報・実データの応答混入
4. 権限付与・認証状態の変化
5. サービス劣化 (応答遅延・429/5xx の連続)
6. 共有リミッター・アカウントロックへの影響が排除できないと判明した場合

## 7. 専用テストアカウント (dedicated test account)

- **状態: 未プロビジョン (not provisioned)**。検証専用の staff アカウント・LIFF アカウントは本ユニット時点で払い出しされていない。
- **規則**: 専用アカウントがプロビジョンされない限り、認証済み live チェックは全て **BLOCKED**。認証が前提のケースは静的・模擬 (メモリ上の検証器・httptest) で代替し、staging-observed には丸めない。
- STG デモカタログ共有パスワードの存在は SECURITY.md で確認済みだが、検証専用アカウントではないため本検証の認証手段として使用しない (安易な共有アカウント利用は tenant 境界の証明にならない)。

## 8. ロールバック (rollback)

- **方式**: 検証用ブラウザーを閉じる・メモリ上の試験状態を破棄する のみ。データ復旧は不要 (書込みを伴わないため)。
- **禁止**: 共有サービス・他セッション・監査ログの停止・削除・改変。独断の削除・DB 復元・資格情報変更による復旧。予期せぬ実環境変更を検知した場合は、事象概要 (機微情報を除く) を記録して依頼者/環境所有者の承認済み復旧手順へ移す (ticket「停止・例外・ロールバック」)。

## 9. 現時点の成立状況まとめ

| 境界条件 | 状態 |
|---------|------|
| 所有 staging origins | 確定 (§1) |
| 稼働リビジョン照合 | 未確定 → 照合完了まで live チェック BLOCKED (§2) |
| 許可テナント識別子 | 未プロビジョン → 認証テナントスコープ live チェック BLOCKED (§3) |
| 副作用のない対象経路 | 確定 (§4) — 未認証 GET/HEAD/OPTIONS のみ |
| ログ観測方法 | 確定 (§5) — メモリ上のマーカー捕捉のみ |
| リクエスト上限・停止条件 | 確定 (§6) |
| 専用テストアカウント | 未プロビジョン → 認証 live チェック BLOCKED (§7) |
| ロールバック | 確定 (§8) — ブラウザー終了/メモリ破棄のみ |

本ユニット自身は staging/production への通信、DB アクセス、資格情報の取得・変更を一切行っていない (documentary unit)。
