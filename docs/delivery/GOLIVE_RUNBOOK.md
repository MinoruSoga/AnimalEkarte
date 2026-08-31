# Go-live 手順書 — 本番切替（相対 timeline / gate-driven HOLD）

> **対象 Issue**: #257 ／ **状態**: ドラフト — **実行 HOLD**（全 **pre-window prerequisite** green かつ USER が新 window を記入するまで fail-closed。day-of gate は当日タイムライン内で別判定）
> **次回切替日（新 window）**: （確定待ち — 下記「新 window 記入欄」に USER が一箇所記入する。本 runbook は日付を発明しない）
> **履歴（historical No-Go）**: 予定 window **2026-08-03** は期限超過・未実施。当該 window は **historical No-Go** であり、**実行可能な current window ではない**。延期履歴: 7/18 → 7/25 → 7/27 → 8/3（当初 PO 裁定 2026-07-15 は 7/18。8/3 は 2026-08-01 の USER 決定）。旧 timeline 表記 `2026-07-18` は失効済みの絶対日であり、当日手順として実行しない。
> **経路**: Cloudflare 経路で納品（PO 決定 2026-07-15。現行構成: [architecture.md](../ops/infra/architecture.md)。移行実施記録は git 履歴）
> **原則**: 診療を止めない。問題発生時は事前合意した基準で業務を Access へ退避し、Cloudflare 正系統を復旧する。
> **重要**: 旧 AWS ECS/RDS 経路は 2026-07-20 に廃止済みで、技術的な切り戻し先として利用できない。**旧 AWS 系を rollback 先として復活させない。**

### 新 window 記入欄（USER 一箇所・値は発明しない）

| 項目 | 値 |
|---|---|
| 切替日 T 日（カレンダー日付） | （確定待ち） |
| Go/No-Go authority（named owner） | （確定待ち: 担当者名） |
| Support primary（named owner） | （確定待ち: 担当者名・対応時間帯） |
| Rollback owner（named owner） | （確定待ち: 担当者名） |

**記入条件（fail-closed）**: 下記 §1 の全 **pre-window prerequisite** が non-secret evidence で green になり、Go/No-Go authority・support・rollback owner が named で揃った後に限る（DEC-60 / #257）。最終 production import と突合は day-of gate のため、この時点では要求しない。未記入のまま当日手順を実行しない。

---

## 1. Pre-window 前提チェックリスト（切替日前に全項目 ✅ であること）

未完了・未確定の項目が 1 つでも残る間は **Go 判定不可（HOLD）**。

| # | 前提 | 対応 Issue / 正本 | 完了条件 | 状態 |
|---|---|---|---|---|
| 1 | STG Cloudflare 移行 Phase 7（NS 切替・並行稼働）完遂 | [現行構成](../ops/infra/architecture.md) | staging デプロイ 2 回連続 green → 画像移行 → データ投入 → NS 切替 → フルスモーク | ✅ 2026-07-17 完了 |
| 2 | credential / provider residual（secret manager 状態・実値は repo 外） | #89 ／ #97 ／ #98 ／ #99 | 非機密 evidence で residual 解消または明示的 HOLD 解除条件が揃っていること。**実 credential 値は本書に書かない** | （確定待ち） |
| 3 | 本番 Cloudflare 環境・配信契約・CI/backup/restore/rollback | #253 ／ [production/setup.md](../ops/infra/production/setup.md) ／ [production/runbook.md](../ops/infra/production/runbook.md) ／ [CI-CD-PIPELINE.md](../ops/deploy/CI-CD-PIPELINE.md) §0 | (a) 本番 CF 基盤構築 (b) `production` Environment + **Required reviewers** (c) production workflow 適用 (d) STG は main→staging 自動・本番は無承認開始不可 (e) **CF-only** rollback 手順 (f) backup/restore rehearsal 記録 (g) latest main CI green（当日確認した GitHub run URL / ID・commit・確認時刻を記録） | （確定待ち: 実行時 receipt 未記入。過去の CI / billing 状態を流用しない） |
| 4 | Access データ移行の事前準備 | #250 | リハーサル PASS、最終 import 手順・担当・入力停止・backup/rollback・突合方法が承認済み。**最終 production import と突合 PASS は day-of gate** | （確定待ち: 当日実行前の準備 evidence） |
| 5 | 全業務シナリオ通し確認済み（authenticated UAT） | #254 | 全シナリオ PASS、または FAIL 項目が「納品後対応合意済みリスト」に隔離済み | （確定待ち） |
| 6 | スタッフアカウント発行・権限設定済み | #255 | 全スタッフに個人アカウント発行・所属院スコープ・役割別権限設定済み | （確定待ち: スタッフ一覧の先方提供がブロッカー） |
| 7 | フロントエンド CSP の最終確認 | [architecture.md](../ops/infra/architecture.md) | `frontend/index.html` の CSP `connect-src` に本番 API オリジンが含まれている | （確定待ち） |
| 8 | 監視・通知の有効化 | #253 ／ [production/runbook.md](../ops/infra/production/runbook.md) §4 ／ `infra/cloudflare/notifications.tf` | ゾーン 5xx 通知が有効・送信先メール検証済み（PROD 専用ポリシーは二重通知のため追加しない） | （確定待ち: 通知先供給・アドレス事前検証） |
| 9 | 切り戻し体制・authority / support / rollback owner の合意 | 本書 §4・冒頭「新 window 記入欄」 ／ [production/runbook.md](../ops/infra/production/runbook.md) §3 | 判断者・判断基準・連絡経路が先方と合意済み。**ECS 切り戻しは選択肢に含めない**（#99） | （確定待ち） |

---

## 2. 当日タイムライン（相対: T 日 = 切替日）

**T 日** は冒頭「新 window 記入欄」に USER が記入した切替日のみを指す。未記入の間は本節を実行しない。

相対オフセットの基準は **T+0:00 = Access 入力停止宣言**。壁時計（09:00 等）は目安の対応表であり、確定待ちの各医院診療開始時刻に合わせて USER が調整する。旧絶対日 `2026-07-18` は実行指示ではない。

| 相対時刻 | 壁時計目安（確定待ち） | 作業 | 実施者 | 完了確認 |
|---|---|---|---|---|
| T-1 日 夕方 | — | pre-window 前提チェックリスト §1 の最終確認・Go/No-Go 事前判定 | 開発側 + 先方管理者 | pre-window 全項目 ✅ |
| T+0:00 | 09:00 | **旧システム（Access）の入力停止**を先方へ宣言 | 先方管理者 | 全スタッフへ周知済み |
| T+0:15 | 09:15 | Access 最終データ抽出（#250 手順） | 開発側 | 抽出ファイル受領 |
| T+0:30 | 09:30 | 本番 DB の事前バックアップ取得（§3.1 / [production/runbook.md §5.1](../ops/infra/production/runbook.md#4-backuprestore-rehearsal) の承認済み方式） | 開発側 | owner 管理の repo 外 artifact / managed-backup receipt、size・checksum、保持期限、restore rehearsal PASS を記録 |
| T+0:45 | 09:45 | 最終データ移行の実行（#250 変換ツール） | 開発側 | ジョブログ・エラー行 0 件 |
| T+1:15 | 10:15 | **Day-of gate**: 突合検証（テーブル別件数・clinic_id 別件数・金額合計・サンプル目視） | 開発側 | 最終 import と検証レポートが PASS。未達なら T+3:00 は No-Go |
| T+1:45 | 10:45 | 本番 DNS レコードの最終確認（§3。**zone-wide NS 変更禁止**） | 開発側 | 実行時の registrar / `dig NS` / 本番レコード確認結果・時刻・receipt を記録 |
| T+2:00 | 11:00 | 疎通確認: `curl -fsS https://api.noah-karte.com/health` → 200 `{"status":"ok"}`、フロント `https://noah-karte.com` 表示 | 開発側 | ヘルスチェック PASS |
| T+2:15 | 11:15 | スモーク: ログイン → 受付 → カルテ → 会計 → 締め（テストデータ）＋ LINE 予約疎通 | 開発側 | 全操作正常・テストデータ削除記録 |
| T+3:00 | 12:00 | **Go/No-Go 最終判定**（§4 の基準） | 判断者（§4・冒頭記入欄） | Go 判定を記録 |
| T+3:15 | 12:15 | 先方へ利用開始を宣言・新システムでの業務開始 | 先方管理者 | — |
| T+日中〜終業 | 〜終業 | 集中監視（§5）。初回レジ締めの立ち会い（リモート可） | 開発側 | 締め完了確認 |

> **NS 安全規則**: repo 履歴は zone-wide NS 切替を 2026-07-17 完了として記録している。当日は NS を変更しない。現在の registrar / NS / DNS 応答は外部状態なので、実行時に読み取り確認し、時刻と receipt を残す。記録と実測が違う場合は HOLD とし、この runbook から NS 操作を行わない。

---

## 3. DNS 最終確認（変更禁止）

本節は Go-live 当日の**読み取り確認だけ**を定める。repo 記録だけで現在の registrar、nameserver、DNS 応答を確定しない。

### 3.1 本番バックアップ gate

正本は [production/runbook.md §5.1](../ops/infra/production/runbook.md#4-backuprestore-rehearsal)。Go 判定前に、次をすべて満たす。

1. production runbook に operator 承認済みの `pg_dump` コマンド、または managed-backup の取得手順を記載する。資格情報や出力先の秘密値は本書へ書かない。
2. 出力は owner 管理の **repo 外**保存先に限定する。artifact / receipt ID、取得時刻、size、SHA-256 checksum、保持期限を当日ログへ記録する。
3. artifact が「存在する」だけでは成功にしない。承認済み隔離環境で restore rehearsal を完了し、復元後の件数・`clinic_id` 別件数・金額合計の検証結果と所要時間を記録する。
4. 上記の手順、owner、retention、restore rehearsal receipt のいずれかが未記入なら HOLD とする。

### 3.2 DNS / 疎通の読み取り確認

| 確認 | コマンド / 方法 | 期待値 / 記録 |
|---|---|---|
| Registrar | 承認済み provider 画面または WHOIS/RDAP を読み取り確認 | provider 名、確認時刻、receipt。repo の過去記載を流用しない |
| Nameserver | `dig NS noah-karte.com +short` | Cloudflare NS 2 件。値・確認時刻を記録し、変更しない |
| 本番 DNS | `dig api.noah-karte.com` / `dig noah-karte.com` | 承認済み production 構成との一致。差異は HOLD |
| API ヘルスチェック | `curl -fsS https://api.noah-karte.com/health` | HTTP success と `{"status":"ok"}`。必要なら `curl -sS -o /tmp/health.json -w '%{http_code}' ...` で status を receipt 化 |
| フロントエンド表示 | ブラウザで `https://noah-karte.com` | ログイン画面表示・証明書有効 |
| ログイン（Cookie） | 実ブラウザでログイン → 画面遷移 | 正常（httpOnly Cookie 発行） |
| CSP | ブラウザ DevTools Console | CSP 違反エラーなし |
| LINE 予約導線 | LINE から予約ページを開く | 表示・予約作成可 |

### 3.3 歴史資料（実行禁止）

repo 履歴では、Cloudflare zone-wide NS 切替は 2026-07-17 に完了した。検証値として記録されている nameserver は次の 2 件である。

- `melissa.ns.cloudflare.com`
- `yadiel.ns.cloudflare.com`

これは過去の移行記録であり、Vercel / Cloudflare の画面で nameserver を変更する手順ではない。現在値が異なる場合も、この runbook から変更せず HOLD / escalation とする。

---

## 4. 切り戻し基準と手順

### 4.1 判断体制

| 役割 | 担当 | 備考 |
|---|---|---|
| 切り戻し判断者 | （確定待ち: 担当者名）— 冒頭「新 window 記入欄」と同一 | Go/No-Go 判定・切り戻し発動の最終決定権 |
| 技術実施者 | （確定待ち: 担当者名） | Cloudflare再デプロイ・DB互換性確認・必要時のスナップショット復元 |
| 先方連絡窓口 | （確定待ち: 担当者名） | 現場への周知 |

**判断時刻**: タイムライン上の **T+3:00 Go/No-Go 行**を第 1 判断点、**T 日終業時**を第 2 判断点とする。以降は §5 のサポート期間中に随時（確定待ち: 時刻の最終合意）。

### 4.2 切り戻し発動基準（いずれか 1 つで発動）

| # | 症状 | 判定方法 |
|---|---|---|
| 1 | `/health` が継続的に非応答（5 分以上） | curl で HTTP 200 が返らない |
| 2 | ログイン不能（複数スタッフ・複数端末で再現） | 実ブラウザ確認 |
| 3 | カルテ・会計の保存が失敗する（業務停止級） | 現場報告 + 再現確認 |
| 4 | データ破損・想定外削除・医院間データ混在の検知 | スモーク・監査ログ確認 |
| 5 | 移行データの突合検証 FAIL（件数・金額不一致） | #250 検証レポート |

軽微な不具合（表示崩れ・特定画面のみのエラー等）は切り戻さず、納品後対応リスト（#254 の合意リスト）で管理する。

### 4.3 切り戻し手順

**原則: AWS の hot standby は存在しない（#99）。DNS/NS の変更を復旧手段にせず、Cloudflare 正系統を復旧する。ECS workflow・IAM・旧 task definition を復活させない。**

1. 判断者が復旧対応を宣言し、先方連絡窓口が必要に応じて現場へ「旧運用（Access）へ一時退避する」ことを周知する。
2. デプロイ直後の障害は、直前に正常稼働したコミットを特定し、DB schema との互換性を確認したうえで Cloudflare へ再デプロイする。STG の具体手順は [STG運用Runbook](../ops/infra/staging/runbook.md)、本番は [PROD運用Runbook](../ops/infra/production/runbook.md) §3（CF-only）を正本とする。
3. provider または基盤障害で再デプロイできない場合は、§3.1 の検証済み DB バックアップと IaC から環境を再建する。production runbook にこの手順と責任者が確定するまでは Go 判定しない。
4. Access へ退避した期間の新規入力と、新システム側で既に受け付けた操作の範囲（時刻・操作者・対象）を記録し、復旧後に差分突合する。
5. データ破損時の DB 復元は、**T+0:30 行で取得し §3.1 の size・checksum・restore rehearsal を満たした**バックアップを使い、#250 の切り戻し手順と承認境界に従う。
6. インシデント記録を作成し（原因・影響範囲・再発防止）、業務復帰時刻を判断者が決定する。credential / PHI は記録に含めない。

---

## 5. 切替直後サポート体制

| 項目 | 内容 |
|---|---|
| 集中サポート期間 | **T+0 〜 T+N 日**（N 確定待ち; 目安 14 日 / 2 週間） |
| 問い合わせ窓口 | （確定待ち: 連絡手段 — 電話 / LINE / メールの別と宛先） |
| 一次対応者 | （確定待ち: 担当者名・対応時間帯）— 冒頭「新 window 記入欄」Support primary と同一 |
| エスカレーション先 | （確定待ち: 担当者名） |
| 日次確認 | エラーログ（Workers Logs）・課金実績・現場からの問い合わせ内容を日次でレビュー |
| 監視 | Cloudflare 通知ポリシー（5xx 率）+ ヘルスチェック。詳細は [DELIVERY_PACKAGE.md](DELIVERY_PACKAGE.md) §4 |

旧AWS環境の廃止は2026-07-20に完了済みであり、サポート期間中も切り戻し先として利用しない。

---

## 6. 関連ドキュメント

- [現行インフラ構成](../ops/infra/architecture.md) — Cloudflare構成の正本
- [infra/cloudflare/README.md](../../infra/cloudflare/README.md) — Cloudflare Terraform / CI デプロイ手順
- [docs/ops/deploy/README.md](../ops/deploy/README.md) — 環境一覧・ロールバック判定フレームワーク
- [docs/ops/deploy/CI-CD-PIPELINE.md](../ops/deploy/CI-CD-PIPELINE.md) — デプロイ契約・パイプライン（#253 §0）
- [docs/ops/infra/production/runbook.md](../ops/infra/production/runbook.md) — 本番運用・CF-only rollback・監視/backup
- [docs/ops/infra/production/setup.md](../ops/infra/production/setup.md) — 本番 CF 事前構築手順
- [DELIVERY_PACKAGE.md](DELIVERY_PACKAGE.md) — 納品ドキュメント（システム構成・運用手順）
- [OPERATION_MANUAL.md](OPERATION_MANUAL.md) — 現場スタッフ向け操作マニュアル
