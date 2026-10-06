# 定期バッチ（Scheduler Jobs）一覧と有効化判断

> 目的: 本番 go-live 時にどのバッチを有効化するかを判断するための一覧票。
> 操作手順（status/pause/resume/catch-up）は [SCHEDULER_OPERATIONS.md](./SCHEDULER_OPERATIONS.md) 参照。

最終更新: 2026-10-03（実装コード照合済み）

## アーキテクチャ概要

```
Cloudflare cron (UTC)
  → Worker scheduled handler (animalekarte-scheduler-v1)
    → Durable Object coordinator（ledger + lease + revision CAS）
      → Go Container 内部 endpoint（/_internal/scheduled-jobs/:jobAction）
        → domain use case（no_show / delivery / dormant / plane_sync）
```

- cron 式は UTC 解釈。許可された式以外は fail-closed で拒否
- Worker→Go 間は `X-Scheduler-Token`（`SCHEDULER_INTERNAL_TOKEN`）で相互認証。secret 未設定時は全要求 401（fail-closed）
- lease（150秒）は二重実行防止。開始済みの Go 側副作用は取消しないため domain 側の冪等性も安全境界
- 実行結果は `Outcome: success/partial/failed` + 処理/成功/失敗件数を ledger に記録

## スケジュール（production）

| Cron (UTC) | JST | 実行順 |
|---|---|---|
| `0 1 * * *` | 10:00 | no_show → delivery |
| `0 6,11 * * *` | 15:00 / 20:00 | no_show |
| `0 17 * * *` | 翌 02:00 | dormant |

- plane_sync は独立 job（`scheduler.JobPlaneSync`）。手動実行・ops 経由で起動可能
- STG の `wrangler.jsonc` には第4 trigger `*/4 0-9 * * *`（UTC）があるが、これは keep-alive で scheduler job ではない

## バッチ別仕様と有効化判断

### no_show — 予約の無断キャンセル検出

| 項目 | 内容 |
|---|---|
| 検出条件 | 予約終了時刻 + 4時間経過しても来院処理されていない予約（rule: `end-time-plus-4h-v1`） |
| 書き込み | `reservations` のステータス遷移（no-show）+ 遷移監査ログ（旧status・評価時刻・runID・rule version） |
| 実行単位 | 1予約ごとにトランザクション + 監査。失敗はその行だけスキップして継続 |
| 実行回数 | 1日3回（10:00/15:00/20:00 JST） |
| 前提 | 予約ステータス体系（来院済み=no-show 対象外）が業務フローと一致していること |
| 止める影響 | no-show が自動記録されないだけ。手動で記録する運用なら不要 |
| 判断点 | 医院が「無断キャンセルの自動記録」を使うか？使わないなら停止で害なし |

### delivery — LINE/Lステップ配信トリガー

| 項目 | 内容 |
|---|---|
| 動作 | 10:00 JST のみ判定（他時刻は no-op）。全院の配信トリガー条件を評価して発火 |
| 書き込み | 発火したトリガーの記録 + 監査 |
| 前提 | **Lステップ連携が有効であること**（EMR-38: 先方 API 有効化未確認・送信 gate OFF） |
| 止める影響 | LINE リマインド等が送られない。LINE 非有効化中は実行しても実質何もしない |
| 判断点 | LINE 連携 go-live まで**停止が自然**。有効化後に resume |

### dormant — 休眠飼い主タグ同期

| 項目 | 内容 |
|---|---|
| 検出条件 | 最終来院から **180日以上** 経過した飼主（医院ごとの閾値設定 `GetDormantThresholds` を参照） |
| 書き込み | 飼主への休眠タグ同期（`SyncDormantTagsWithThresholds`）。閾値は監査メタデータに記録 |
| 実行単位 | カーソルページネーションで全件スキャン。タグ同期失敗はその行だけスキップ |
| 実行回数 | 1日1回（02:00 JST） |
| 止める影響 | 休眠タグが更新されない。画面で休眠タグを使わないなら実害なし |
| 判断点 | 休眠タグを業務で使うか？低リスクのため「動かしておいて害なし」でも可 |

### plane_sync — Plane チケット状態同期

| 項目 | 内容 |
|---|---|
| 動作 | アプリ内サポート報告（Plane 起票済み）で、Plane 側が completed になったものを resolved に揃える**片方向**同期 |
| 前提 | support 機能（アプリ内バグ報告→Plane 起票）が本番で有効・Plane reader 設定済み |
| 止める影響 | Plane で Done にしてもアプリ内報告が resolved に更新されない。運用上の齟齬のみ |
| 判断点 | サポート機能を本番で有効化するか（EMR-263 で情報区分を決め中）に連動 |

## 管理用エンドポイント（/_internal/scheduler/*）

| メソッド・パス | 操作 |
|---|---|
| `GET /_internal/scheduler/status?limit=N` | control.paused・active lease・直近 run・操作履歴の確認 |
| `PUT /_internal/scheduler/control` | バッチ全体の pause / resume |
| `POST /_internal/scheduler/runs` | 手動実行・catch-up |

- 認証: `SCHEDULER_OPS_SECRET` Bearer（CLI 用）または Cloudflare Access JWT（人間用）
- **決定済み（2026-10-03）**: 本番では Cloudflare Access で追加保護する（Zero Trust policy + TEAM_DOMAIN/AUDIENCE を 10/20 構築時に発行）

## 関連 secrets

| 名前 | 用途 |
|---|---|
| `SCHEDULER_OPS_SECRET` | ops CLI の Bearer（32byte+、Bearer token） |
| `SCHEDULER_INTERNAL_TOKEN` | Worker→Go の `X-Scheduler-Token` 共有秘密 |
| `SCHEDULER_ACCESS_TEAM_DOMAIN` / `SCHEDULER_ACCESS_AUDIENCE` | Access JWT 検証（JWKS/issuer/audience/exp/nbf） |
| `SCHEDULER_ALERT_*` | 失敗時 Webhook 通知先 |

## Go-live 時の有効化チェックリスト

| バッチ | 有効化の前提 | 判断 |
|---|---|---|
| no_show | 医院が no-show 自動記録を使うと回答 | 未定 |
| delivery | EMR-38（先方有効化）+ LINE 送信ゲート ON | 未定（LINE 未開始なら停止） |
| dormant | 休眠タグの利用有無 | 未定 |
| plane_sync | support 機能の本番有効化（EMR-263） | 未定 |

> 停止方法: `PUT /_internal/scheduler/control` で pause、または `wrangler.production.jsonc` の cron trigger から該当式を外す。削除ではなく停止維持を基本とし、後から有効化できる。
