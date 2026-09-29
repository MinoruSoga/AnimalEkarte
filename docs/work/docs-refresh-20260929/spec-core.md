# DOCS-REFRESH-20260929 — spec-core

**パーティション**: `docs/spec/*.md` 直下、`docs/spec/line/**`、`docs/spec/screens/00*`–`19*`、`docs/spec/screens/README.md`、`docs/spec/screens/CLAUDE.md`
**基準**: worktree `/private/tmp/ae-dr29-spec-core`、branch `docs-refresh-20260929/spec-core`、base `main@4722f4db7`
**作業日**: 2026-09-29（静的照合のみ。runtime / E2E / design-audit / DB / STG の再実行は行っていない）

## 1. ファイル一覧と分類

| ファイル | 分類 | 備考 |
|---|---|---|
| `docs/spec/specification.md` | corrected | Linear→Plane 訂正、テーブル数 128→130、末尾の迷子 `）` 除去 |
| `docs/spec/ui-design-compliance.md` | corrected | 製品ルート 86→87（全所）、`/settings/bug-reports` 行追加、E2E inventory gap 訂正 |
| `docs/spec/screens/README.md` | corrected | product leaf 86→87、Linear→Plane 訂正（41 ファイル数は変更なし） |
| `docs/spec/line/README.md` | corrected | BRT-50 を Linear 閉鎖済み事実へ訂正、`EMR-132` 関連項目を明記 |
| `docs/spec/README.md` | kept | 索引のみ・数値/Linear 言及なし |
| `docs/spec/cash-register.md` | kept | 12 分類・退院時 CreateAccounting 選択式・append-only 締め・post-close edit 権限は vault 12/13 確定仕様と実装 (`medical_record_crud.go` / `hospitalization_discharge_tx.go`) と一致 |
| `docs/spec/customer-aggregation.md` | kept | 集計仕様に現行との矛盾を検出せず |
| `docs/spec/design-system.md` | kept | 製品色 `#038B94` / ink 4 段 / nav canvas-soft は `design-tokens.ts` と一致 |
| `docs/spec/reservation-to-record-flow.md` | kept | `appointments` SoT・明示 staff の選択時刻 shift 未検証 gap・`available-staffs` 不在の記述は実装と一致 |
| `docs/spec/line/architecture.md` | kept | Write API 停止既定・暗号化・`clinic_integrations` 記述は実装と一致 |
| `docs/spec/line/setup.md` | kept | LSTEP_WRITE_API_ENABLED 既定 OFF・ordered gates・connection-test 既知 gap は `lstep_settings_connection.go` / `lstep-settings.ts` と一致 |
| `docs/spec/line/lstep-integration.md` | kept | 配信トリガー・CPM 記述に「未決を確定扱い」を検出せず |
| `docs/spec/line/reservation-spec.md` | kept | LIFF 予約フロー記述に矛盾を検出せず |
| `docs/spec/line/cost-analysis.md` | kept | docs-only 課金試算。実装主張なし |
| `docs/spec/line/CLAUDE.md` | kept | 執筆ルールのみ |
| `docs/spec/screens/CLAUDE.md` | kept | 執筆ルールのみ |
| `docs/spec/screens/00-pet-selection.md` | kept | 死亡ペット選択可・作成は死亡ゲート拒否・不明 status は fail-closed。§2.1 不変式と一致 |
| `docs/spec/screens/01-reception.md` | kept | カンバン・権限・API 表記に矛盾を検出せず |
| `docs/spec/screens/02-reservations.md` | kept | 一般/トリミング予約分離・`staff_reservation_capabilities` 記述は実装と一致 |
| `docs/spec/screens/03-owners-list.md` | kept | ClinicScopeFilter・危険表示は実装と一致 |
| `docs/spec/screens/04-owners-form.md` | kept | Lステップ個別送信・副飼主の記述に矛盾を検出せず |
| `docs/spec/screens/05-medical-records-list.md` | kept | 検索・確定表示は実装と一致 |
| `docs/spec/screens/06-medical-records-form.md` | kept | 9 タブ・確定ロック・#235 DnD 未決（vault 12 の「制約確定・実装未了」と一致） |
| `docs/spec/screens/07-hospitalization-list.md` | kept | ケージ稼働・ボード表示に矛盾を検出せず |
| `docs/spec/screens/08-hospitalization-detail.md` | kept | care plan・バイタル時系列に矛盾を検出せず |
| `docs/spec/screens/09-hospitalization-form.md` | kept | 登録時治療プラン・料金スナップショットに矛盾を検出せず |
| `docs/spec/screens/10-accounting-list.md` | kept | 12 分類・検索は vault 12 (#251) と実装と一致 |
| `docs/spec/screens/11-accounting-detail.md` | kept | 保険窓口精算・レシート/ドロア未実装（宿題 F10 と一致）・append-only 記述と一致 |
| `docs/spec/screens/12-examinations-list.md` | kept | 検査一覧・進捗管理に矛盾を検出せず |
| `docs/spec/screens/13-examinations-form.md` | kept | 基準値判定・未紐付け受信に矛盾を検出せず |
| `docs/spec/screens/14-vaccinations-list.md` | kept | 一覧・次回予定に矛盾を検出せず |
| `docs/spec/screens/15-vaccinations-form.md` | kept | lot1–4・`useGetAllVaccinesMaster`・`NEXT_SCHEDULE_OPTIONS`（3週後/4週後/1年後/手動）実在確認 |
| `docs/spec/screens/16-trimming-list.md` | kept | `billing_item_repository` UNION・push でない会計連携記述は実装と一致 |
| `docs/spec/screens/17-trimming-form.md` | kept | 画像プレビュー未配線・`defaultRecordShortcutTimes`・`uk_appointment_staff_time` 記述は実装と一致（`migrations/001_init.sql:2737`） |
| `docs/spec/screens/18-inventory-list.md` | kept | ステータス導出（旧 SD-4）・サーバページ・削除 UI 未実装は `backend/internal/inventory/`・FE と一致 |
| `docs/spec/screens/19-clinic-settings.md` | kept | `CompanyInvoiceSection`・`useGetCompany`/`useUpdateCompany`・`scope=all` 実在確認 |

**propose-delete**: なし。

## 2. 主要訂正（旧主張 → 新主張）

| 場所 | 旧主張 | 新主張 | 根拠 |
|---|---|---|---|
| `specification.md:8` | release gate の作業入口は Linear `BRT-4` | Linear は 2026-09-16 閉鎖。SoT は Plane `baritechllc`/`EMR`/hub `EMR-1`/case `BRT-4` | vault `AGENTS.md` PLANE-CASE-ROUTE、`docs/work/plane-md-migration-20260923-receipt.md` |
| `specification.md:22` | 全 128 テーブル | 全 130 テーブル | `ls backend/migrations/*.sql` 直下 `CREATE TABLE` 実測 130（`scripts/check-docs-symbol-drift.sh` ゲート値） |
| `specification.md:61` | スキーマは 128 テーブル | 130 テーブル | 同上 |
| `screens/README.md:7` | `route-inventory` の product leaf = 86 | 87（`/settings/bug-reports` 追加） | `frontend/src/app/routes/route-inventory.test.tsx:57`（`expect(pages).toHaveLength(87)`）、`settings-routes.tsx:463`、コミット `48ec94c7`（2026-09-26） |
| `screens/README.md:103` | release state は Linear hub `BRT-4` | Plane `EMR` / hub `EMR-1` | 同上 Linear 閉鎖 |
| `ui-design-compliance.md`（序文・C12・C14・監査範囲・§2・脚注） | 本体 86 製品ルート | 本体 87 製品ルート | 同上 |
| `ui-design-compliance.md` §2 表 | `/settings/bug-reports` 行なし | `BugReportsPage` 行追加（静的 ✅ / runtime pending） | `frontend/src/features/support/routes/BugReportsPage.tsx:126` が `PageLayout` 使用（:131）、`ResourceHospitalSettings` ゲート |
| `ui-design-compliance.md` E2E gap | E2E は 85 製品 route で `/examinations/new` のみ欠落 | 85 route template（public 3 + protected 82）で `/examinations/new` と `/settings/bug-reports` を欠く | `frontend/e2e/ui-design-compliance-readonly.spec.ts` 実測（`bug-reports` grep 0 件） |
| `line/README.md:61` | Linear BRT-50 の現在値は今回取得していない | 2026-09-06 時点で未取得の記述を維持しつつ、Linear 閉鎖で今後も取得不可・関連 Plane 項目は `EMR-132`（Blocked）を明記 | `docs/work/plane-md-migration-20260923-receipt.md:70`（`QA-UAT-LSTEP-REAL` → `EMR-132`） |

## 3. 未解決の外部依存 / BLOCKED

| 項目 | 状態 | 根拠 |
|---|---|---|
| Lステップ Write API 再開 + cron 実配線 | **BLOCKED_EXTERNAL / 外部依存** — 先方 enable、STG 少数 live-send、cron fire/stop 実測待ち。`LSTEP_WRITE_API_ENABLED` 既定 OFF 維持 | `docs/spec/line/README.md` §3、`docs/ops/deploy/LSTEP_WRITE_API_PAUSE.md`、GitHub `#259`（2026-09-06 時点 OPEN）、Plane `EMR-132`（Blocked） |
| 健診パッケージ要件（vault `14_健診パッケージ_クライアント要件_2026-08-24.md`） | **未承認** — 臨床 + PO 承認待ち。spec 側は実装済み/確定として記述していないことを確認（番号付き仕様 `25-checkups-list` は spec-detail パーティション） | vault 14、宿題 F14「仕様承認は 14 / 臨床+PO · 記録のみ」 |
| `/examinations/new`・`/settings/bug-reports` の E2E inventory 欠落 + scoped rerun | **別 task** — docs-only 変更では未対応 | `ui-design-compliance.md` 脚注 |
| `/identity-links`・新規 route の route 単位 runtime 証跡 | **pending（未実行）** — runtime E2E は再実行していない | 同上 |
| トリミング画像アップロード | 未配線（FE のプレビューのみ）。実装判断は別 task | `screens/17-trimming-form.md` §1.2 |
| 予約の明示 staff 選択時刻 shift 再検証 | **source gap** — backend 未実装。FE shift 絞り込みを安全根拠にしない | `reservation-to-record-flow.md`、`line/README.md` §2 |
| Lステップ connection-test の既知 gap（非401/403 を成功扱い・HTTP200+失敗フラグ・FE が component result を評価しない） | **source gap** — docs-only では未修正 | `line/setup.md` §5 |
| 八王子 `TBL_KNJO_DATA` 破損・Dr.わん現地疎通・端末棚卸し 等 | **BLOCKED_EXTERNAL / WAIT** — spec-core の記述とは非矛盾（移行/運用側の宿題） | vault `02_宿題と未決.md` H1/D1/T1 |

## 4. 検証

| コマンド | 結果 |
|---|---|
| `git diff --check -- docs/spec` | 0（OK） |
| `bash scripts/check-docs-symbol-drift.sh` | 0 — `OK ドリフトなし（検査トークン 606 件）`。130 テーブル宣言と migration 実測が一致 |
| `git status --porcelain` | 変更は allowlist 内 5 ファイルのみ（spec 4 ファイル + 本レポート）。allowlist 外の差分なし |

**スコープ外確認**: runtime / E2E / `pnpm design-audit` / ビルド / DB / STG / Plane 書き込みは一切未実施（禁止事項どおり）。
