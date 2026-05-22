# 推測追加機能 監査レポート

**作成日**: 2026-05-13  
**更新日**: 2026-05-14（`canceledDate time.Time` 削除完了 / 最終監査完了 — 推測追加ゼロ確認）  
**対象**: FEAT-372 を中心とした明示要求外の実装洗い出し  
**実装・コード変更**: FEAT-372 Phase B 削除（commit `40d745f4`）、HLTH_専門検診候補スケルトン一式削除（commit `ecb9a8ad`）

---

## 結論（先出し）

- **削除済み（2026-05-13, commit `40d745f4`）**: 1 件（FEAT-372 Phase B — 受付ページの健診アラートカード）→ **対応済み**
- **削除済み（2026-05-13, commit `ecb9a8ad`）**: 1 件（`SyncSpecialCheckupCandidateTag` noop + FE tag 定数）→ **対応済み**
- **現仕様外 / 実装対象外**: 3 件（ワクチン/在庫/入院アラートカード）
- **廃止確定済み（WONTFIX・コードなし）**: 5 件（FEAT-373/370/371/381-3/382-3）
- **明示要求に基づいて残すべき**: FEAT-372 Phase A・C・D（API・一覧バッジ・フィルタ）

---

## 対応済み — 削除・文書整理完了

### FEAT-372 Phase B — 受付ページの健診アラートカード ✅ 削除済み

| 項目 | 内容 |
|------|------|
| **機能名** | `CheckupAlertCard.tsx`（受付ページ上部の健診アラート 2 枚カード） |
| **実装箇所** | `frontend/src/features/reception/components/CheckupAlertCard.tsx`（削除済み）<br>`frontend/src/features/reception/routes/Reception.tsx`（アラートブロック削除済み） |
| **実装 commit** | `7565ce9c`（2026-05-12） |
| **削除 commit** | `40d745f4`（2026-05-13） |
| **削除理由** | `SPECIFICATION.md §3.1` に受付ページへのカード表示記述なし。カンバン操作専用スコープ |
| **状態** | **対応済み。受付ページへの再追加は PO の仕様書明記が前提条件** |

### 文書整理バッチ（2026-05-13 完了） ✅

| 文書 | 対応内容 |
|------|---------|
| `docs/testing/FULL_DOMAIN_SCENARIO_TEST_GUIDE.md §3.9` | 在庫アラート全 4 行に「現仕様外 / 実装対象外」注記 + 警告ブロック追加 |
| `docs/tasks/closed/ux/BUG-CLEAR-BUTTON-COLOR-INCONSISTENCY.md` | ステータスを `🔴 未修正` → `✅ 修正済み（2026-05-13 確認）` に更新 |
| `docs/tasks/closed/accounting/BUG-404_closing-report-missing-tax-rate-breakdown.md` | 起票根拠ブロック追加（FEAT-368 Q10 PO 明示回答ベースと明記） |

---

## 削除済み — HLTH_専門検診候補スケルトン一式（commit `ecb9a8ad`）

| 削除対象 | 内容 |
|---------|------|
| `SyncSpecialCheckupCandidateTag` | `lstep_health_tag_sync.go` のインターフェースメソッドと noop 実装を削除 |
| `HLTH_専門検診候補` FE 定数 | `segment-tags.ts` / `tag-names.ts` から削除 |
| 関連テスト | `lstep_health_tag_sync_test.go` / `lstep_tag_sync_service_test.go` 等から削除 |
| seed データ | `002_seed_master.sql` から HLTH_専門検診候補エントリを削除 |

SPEC-002 Q6 確定前にスケルトン実装を入れたこと自体が推測追加と判断し削除確定。  
Q6 が PO 確定した場合は、その時点で改めて実装を起票すること。

---

## 保留（テストガイド記載あり・未実装）

以下はコードが存在しないため削除作業は不要。ただし、将来実装の起票根拠に使われるリスクがある。

| 機能 | テストガイド該当行 | 明示要求 | 判定 |
|------|-------------------|---------|------|
| ワクチン期限アラートカード（受付ページ） | §3.3 L1566「ワクチン期限アラート件数カード表示」 | SPECIFICATION.md に記述なし | **現仕様外 / 実装対象外** |
| 在庫アラートカード（受付ページ） | §3.9（注記済み）+ §3.3 L1574「在庫アラート件数更新確認」 | 同上 | **現仕様外 / 実装対象外** |
| 入院中件数カード（受付ページ） | §3.3 L1572「別タブで入院登録 → 件数増加」 | 同上 | **現仕様外 / 実装対象外** |

> **注記**: §3.3 / §3.9 ともに「現仕様外 / 実装対象外」注記を 2026-05-13 に追加済み。
> 将来の追記時も同様の注記パターンを維持すること。

---

## 廃止確定済み一覧（WONTFIX・コードなし）

既に整理済みの記録として掲載。

| タスク | 廃止理由 |
|--------|---------|
| FEAT-373 来院未確認ペットへの再リマインダー | テストガイド §4.124 派生・PO 要求記録なし・推測ベース起票 |
| FEAT-370 検査異常値の高度判定（重症度段階 + 急変フラグ） | PO 明示要求なし・EXAM-001 で代替可能・マスタ整備コスト膨大 |
| FEAT-371 検査統計ダッシュボード | FEAT-370 連動廃止・データ蓄積前で価値ゼロ |
| FEAT-381-3 受付・診療入力 Phase 3 BE（来院動機・紹介元・拒否理由） | SPEC-002 参照ミス・機能重複（recommendation_reason で対応済）・PO 要求なし |
| FEAT-382-3 受付・診療入力 Phase 3 FE | 親 FEAT-381-3 廃止に連動 |

---

## 明示要求に基づいて残すべき機能一覧

| 機能 | 根拠 |
|------|------|
| FEAT-372 Phase A — `GET /v1/checkups/alerts` BE API | SPECIFICATION.md §4.1「定期健診: 健康状態の定期的な確認」+ FULL_DOMAIN_SCENARIO_TEST_GUIDE §4.22「Tab5 次回健診アラート」がタスク起票より前から存在 |
| FEAT-372 Phase C — 健診一覧 alertStatus フィルタ + バッジ | 同上（CheckupsList は健診機能画面そのもの） |
| FEAT-372 Phase D-1 — CheckupsTab バッジ（カルテ Tab5） | テストガイド §4.22 L2204「次回健診アラート（期限切れ）」は健診 feature 内の動作記述 |
| FEAT-372 Phase D-2 — JST タイムゾーン整合化 | Phase A/C の品質修正（バグ修正）。独立した仕様追加ではない |

---

## 補足

### FEAT-372 Phase B が不要と判断できる理由の整理

1. **配置の問題**: 受付ページ（`/`）は予約カンバン専用画面。健診アラートを置く業務文脈がない（スタッフは当日受付業務に集中しており、健診期限確認は別タスク）
2. **PO 確認の範囲逸脱**: SPEC-001 Q6 は「checkup_reminders テーブルを個別か共通か」という設計質問であり、「受付ページにカード表示するか」とは別問題。Q6 が確定したからといって Phase B の UI 設計が承認されたわけではない
3. **起票根拠の循環**: FEAT-372 task が「テストガイド §3.3 L1567 に定期健診アラート件数カード表示がある」と背景に書いているが、該当テストガイド行が FEAT-372 実装中に追記された場合、自己参照的正当化になる

### 残課題（次に整理すべき文書）

| 文書 | 対応状況 | 残作業 |
|------|---------|--------|
| `FULL_DOMAIN_SCENARIO_TEST_GUIDE.md §3.9` | ✅ 2026-05-13 注記済み | なし |
| `FULL_DOMAIN_SCENARIO_TEST_GUIDE.md §3.3` | ✅ 2026-05-13 注記済み（全行「現仕様外 / 実装対象外」+「削除済み」に更新済み） | なし |
| `docs/SPECIFICATION.md §3.1` | ✅ 2026-05-13 注記済み（受付画面カンバン専用スコープ・アラートカード対象外を明記済み） | なし |
| SPEC-002 Q6 | 推測実装削除済み | スケルトン削除済み (`ecb9a8ad`)。PO 明示要求が来た場合に動的設定ベースで再起票 |

---

## 最終監査（2026-05-14）

### 追加削除済み（commit `87a177d7`）

| 削除対象 | 内容 |
|---------|------|
| `SyncCancellationTag(... canceledDate time.Time)` | インターフェース・実装・呼び出し側・mock 4箇所 計6ファイル |
| `_ = canceledDate` 行 | 「将来的に `canceled_visit_YYYY-MM-DD` 形式に拡張する場合に使用」コメント付き廃棄 |

`SyncReservationTag` / `SyncNoShowTag` は日付付きタグ（`reserved_YYYY-MM-DD` / `no_show_YYYY-MM-DD`）を実際に生成するため `time.Time` パラメータを保持。
`SyncCancellationTag` は固定タグ `canceled_visit` のみ使用しており、仕様書に日付付きキャンセルタグの記述なし。

### 最終掃引結果（2026-05-14）

全層（frontend / backend / seed / docs）を再点検した結果、**追加削除候補なし**。

| 確認項目 | 判定 |
|---------|------|
| frontend 補助 UI・件数カード・バッジ | 在庫アラートは `docs/screens/18-inventory-list.md:10` 明記。CheckupAlertBadge は FEAT-372 Phase C PO 合意済み。削除対象なし |
| backend noop・未使用引数・将来コメント | `CountUsageByInquiryTemplateID` stub は P10 パターンの意図的プレースホルダ（AUDIT-2026-05-06 既知）。`将来` コメントは実装なしのメモ行のみ。削除対象なし |
| seed / constants | 追加削除対象なし |
| docs/line / docs/testing | 追加削除対象なし |

### 保留（削除対象ではない）

| 項目 | 理由 |
|------|------|
| `LstepTagCodeMappingsSection.tsx` の SPEC-002 Q5 プレースホルダー文言 | 判定材料は設定画面/DBで動的に管理する。固定コード一覧の投入待ちにはしない。PO 明示要求が来た場合もコード直書きではなく動的設定として扱う |

### 方針（以降の実装ルール）

> **明示要求が来た時だけ実装する。** 仕様書・PO の明示指示のない機能・引数・コメント付き拡張余地は追加しない。
> 「将来拡張しやすいように」という理由は実装根拠にならない。
