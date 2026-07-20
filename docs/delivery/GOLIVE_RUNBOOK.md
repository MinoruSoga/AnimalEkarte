# Go-live 手順書 — 本番切替（2026-07-18）

> **対象 Issue**: #257 ／ **切替日**: 2026-07-18（PO 裁定 2026-07-15: 納品日 = 7/18）／ **状態**: ドラフト
> **経路**: Cloudflare 経路で納品（PO 決定 2026-07-15。正本: [migration-cloudflare.md](../../migration-cloudflare.md) 「現況サマリ」節）
> **原則**: 診療を止めない。問題発生時は事前合意した基準で旧運用へ戻す。

---

## 1. 前提チェックリスト（切替当日までに全項目 ✅ であること）

| # | 前提 | 対応 Issue / 正本 | 完了条件 | 状態 |
|---|---|---|---|---|
| 1 | STG Cloudflare 移行 Phase 7（NS 切替・並行稼働）完遂 | [migration-cloudflare.md](../../migration-cloudflare.md) P5-5〜P7 ／人間タスク正本 `q&a.html`「Cloudflare 移行」節 | staging デプロイ 2 回連続 green（P5-5）→ 画像移行（P2-4/5）→ データ投入（P3-6/7）→ NS 切替（P1-2）→ フルスモーク（P7-3） | （確定待ち: PR #186 マージ後に進行） |
| 2 | 本番 Cloudflare 環境構築済み | #253 | 本番向け `backend-deploy.yml`（production 対応）で main → 本番反映が自動化・失敗通知あり・ロールバック手順文書化 | （確定待ち） |
| 3 | Access データの本番投入済み | #250 | リハーサル移行 PASS → 本番 DB へ最終移行 → 突合検証（件数・clinic_id 別件数・金額合計）PASS | （確定待ち: 最終移行は当日タイムライン内で実施） |
| 4 | 全業務シナリオ通し確認済み | #254 | 全シナリオ PASS、または FAIL 項目が「納品後対応合意済みリスト」に隔離済み | （確定待ち） |
| 5 | スタッフアカウント発行・権限設定済み | #255 | 全スタッフに個人アカウント発行・所属院スコープ・役割別権限設定済み | （確定待ち: スタッフ一覧の先方提供がブロッカー） |
| 6 | フロントエンド CSP の最終確認 | [migration-cloudflare.md](../../migration-cloudflare.md) §9 リスク登録簿 | `frontend/index.html` の CSP `connect-src` に本番 API オリジンが含まれている | （確定待ち） |
| 7 | 監視・通知の有効化 | #253 ／ `infra/cloudflare/notifications.tf` | 通知ポリシー apply 済み・送信先メール検証済み | （確定待ち: `notification_email` 供給・アドレス事前検証） |
| 8 | 切り戻し体制の合意 | 本書 §4 | 判断者・判断基準・連絡経路が先方と合意済み | （確定待ち） |

---

## 2. 当日タイムライン（2026-07-18）

時刻は目安。診療開始時刻に合わせて確定する（確定待ち: 各医院の当日診療開始時刻）。

| 時刻（目安） | 作業 | 実施者 | 完了確認 |
|---|---|---|---|
| T-1 日 夕方 | 前提チェックリスト §1 の最終確認・Go/No-Go 事前判定 | 開発側 + 先方管理者 | 全項目 ✅ |
| 09:00 | **旧システム（Access）の入力停止**を先方へ宣言 | 先方管理者 | 全スタッフへ周知済み |
| 09:15 | Access 最終データ抽出（#250 手順） | 開発側 | 抽出ファイル受領 |
| 09:30 | 本番 DB の事前スナップショット取得（pg_dump。切り戻し用） | 開発側 | ダンプファイル保管確認 |
| 09:45 | 最終データ移行の実行（#250 変換ツール） | 開発側 | ジョブログ・エラー行 0 件 |
| 10:15 | 突合検証（テーブル別件数・clinic_id 別件数・金額合計・サンプル目視） | 開発側 | 検証レポート PASS |
| 10:45 | NS/DNS 切替（§3。※STG Phase 7 で切替済みの場合は本番 DNS レコードの最終確認のみ） | 開発側 | `dig NS noah-karte.com` で Cloudflare NS を確認 |
| 11:00 | 疎通確認: `curl https://api.noah-karte.com/health` → 200 `{"status":"ok"}`、フロント `https://noah-karte.com` 表示 | 開発側 | ヘルスチェック PASS |
| 11:15 | スモーク: ログイン → 受付 → カルテ → 会計 → 締め（テストデータ）＋ LINE 予約疎通 | 開発側 | 全操作正常・テストデータ削除記録 |
| 12:00 | **Go/No-Go 最終判定**（§4 の基準） | 判断者（§4） | Go 判定を記録 |
| 12:15 | 先方へ利用開始を宣言・新システムでの業務開始 | 先方管理者 | — |
| 〜終業 | 集中監視（§5）。初回レジ締めの立ち会い（リモート可） | 開発側 | 締め完了確認 |

> **注意（DNS 伝播）**: NS 切替はインターネット全体への伝播に最大 24〜48 時間かかることがあります。STG Phase 7 で事前に NS 切替を済ませておくことで、当日の DNS 起因リスクを排除する計画です（前提 #1）。当日に NS 切替を行う判断をした場合は、伝播待ち時間をタイムラインに追加してください。

---

## 3. NS/DNS 切替手順

ドメイン `noah-karte.com` のレジストラは **Vercel** です。ゾーンは Cloudflare に作成済み（`infra/cloudflare/zone.tf`。DNS レコードは棚卸し・移設済み）。

### 3.1 切替（Vercel → Cloudflare NS）

1. Cloudflare 側の DNS レコードが現行の全レコードと一致していることを最終確認する（`infra/cloudflare/zone.tf` が正本。`terraform plan` で差分ゼロを確認）。
2. Vercel ダッシュボード → Domains → `noah-karte.com` → Nameservers を **Custom Nameservers** に変更し、Cloudflare 指定の NS を設定する:
   - `melissa.ns.cloudflare.com`
   - `yadiel.ns.cloudflare.com`
3. Cloudflare ダッシュボード（またはメール通知）でゾーンが **Active** になったことを確認する。
4. 伝播確認: `dig NS noah-karte.com +short` が上記 2 件を返すこと。`dig api.noah-karte.com` / `dig noah-karte.com` が期待値を返すこと。

> NS 切替は **Terraform の管理対象外**（`infra/cloudflare/README.md` 安全ルール 4）。人手で実施し、実施前後の状態を `migration-cloudflare.md` に記録すること。
> NS 切替は `noah-karte.com` 配下の**全サブドメイン（STG 含む）に一括で影響**します。STG（Phase 7）で先行切替済みであれば、本番 Go-live 当日に NS 操作は発生しません。

### 3.2 切替後の確認

| 確認 | コマンド / 方法 | 期待値 |
|---|---|---|
| API ヘルスチェック | `curl -s https://api.noah-karte.com/health` | 200 `{"status":"ok"}` |
| フロントエンド表示 | ブラウザで `https://noah-karte.com` | ログイン画面表示・証明書有効 |
| ログイン（Cookie） | 実ブラウザでログイン → 画面遷移 | 正常（httpOnly Cookie 発行） |
| CSP | ブラウザ DevTools Console | CSP 違反エラーなし |
| LINE 予約導線 | LINE から予約ページを開く | 表示・予約作成可 |

---

## 4. 切り戻し基準と手順

### 4.1 判断体制

| 役割 | 担当 | 備考 |
|---|---|---|
| 切り戻し判断者 | （確定待ち: 担当者名） | Go/No-Go 判定・切り戻し発動の最終決定権 |
| 技術実施者 | （確定待ち: 担当者名） | NS 戻し・DB リストアの実操作 |
| 先方連絡窓口 | （確定待ち: 担当者名） | 現場への周知 |

**判断時刻**: 切替当日 12:00 の Go/No-Go 判定を第 1 判断点、当日終業時を第 2 判断点とする。以降は §5 のサポート期間中に随時（確定待ち: 時刻の最終合意）。

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

**原則: NS を Vercel 既定に戻すだけで旧経路が復活する。AWS 側（ECS/RDS/CloudFront）は切替当日も無変更のまま温存されているため、AWS 側の操作は不要。**

1. 判断者が切り戻しを宣言し、先方連絡窓口が現場へ「旧運用（Access）へ戻る」ことを周知する。
2. Vercel ダッシュボード → Domains → `noah-karte.com` → Nameservers を **Vercel 既定の NS に戻す**（Custom Nameservers を解除）。
3. 伝播確認: `dig NS noah-karte.com +short` が Vercel NS を返すこと。
4. 新システム側で入力済みのデータがあれば、範囲（時刻・操作内容）を記録する（再切替時の差分取り込みのため）。
5. Access データ移行後に問題が発覚した場合の DB 巻き戻しは、当日 09:30 取得の pg_dump スナップショットからのリストアで行う（#250 の切り戻し手順が正本）。
6. インシデント記録を作成し（原因・影響範囲・再発防止）、再切替日を判断者が決定する。

> 補足: NS 切替の伝播ラグにより、切り戻し直後は新旧が混在して見える時間帯が発生し得ます。現場周知は「URL でなく Access を使う」という業務指示で行い、DNS の状態に依存させないこと。

---

## 5. 切替直後サポート体制

| 項目 | 内容 |
|---|---|
| 集中サポート期間 | 切替日から（確定待ち: 期間。目安 2 週間）— `migration-cloudflare.md` Phase 7 の並行稼働・監視期間（1〜2 週間）と同期 |
| 問い合わせ窓口 | （確定待ち: 連絡手段 — 電話 / LINE / メールの別と宛先） |
| 一次対応者 | （確定待ち: 担当者名・対応時間帯） |
| エスカレーション先 | （確定待ち: 担当者名） |
| 日次確認 | エラーログ（Workers Logs）・課金実績・現場からの問い合わせ内容を日次でレビュー |
| 監視 | Cloudflare 通知ポリシー（5xx 率）+ ヘルスチェック。詳細は [DELIVERY_PACKAGE.md](DELIVERY_PACKAGE.md) §4 |

AWS 側リソースの廃止（migration-cloudflare.md Phase 8）は、本サポート期間で安定を確認した後にのみ実施する。

---

## 6. 関連ドキュメント

- [migration-cloudflare.md](../../migration-cloudflare.md) — 移行計画・実施記録・リスク登録簿（正本）
- [infra/cloudflare/README.md](../../infra/cloudflare/README.md) — Cloudflare Terraform / CI デプロイ手順
- [docs/ops/deploy/README.md](../ops/deploy/README.md) — 環境一覧・ロールバック判定フレームワーク
- [docs/ops/deploy/CI-CD-PIPELINE.md](../ops/deploy/CI-CD-PIPELINE.md) — デプロイパイプライン
- [DELIVERY_PACKAGE.md](DELIVERY_PACKAGE.md) — 納品ドキュメント（システム構成・運用手順）
- [OPERATION_MANUAL.md](OPERATION_MANUAL.md) — 現場スタッフ向け操作マニュアル
