# scenarios/ — 納品前受け入れテストシナリオ

> **目的**: 納品前に主要業務が実データ相当環境で通ることを証明する受け入れシナリオの索引を提供する。
> **読者**: 検証実施者（AI エージェント / 人間どちらでも実行可能）・PO。
> **タイミング**: 納品前検証・大きなリリース前。
> **アーキテクチャ正本**: [../TEST_ARCHITECTURE.md](../TEST_ARCHITECTURE.md)（L4 受入層）

## 既存テストとの分担（重複させない）

| 領域 | 正本 |
|:---|:---|
| テスト層・実行優先・記録 | [../TEST_ARCHITECTURE.md](../TEST_ARCHITECTURE.md) |
| 受入環境準備 | [../UAT-ENV-SETUP.md](../UAT-ENV-SETUP.md) |
| 外来1件サイクル・会計計算・CRM タグ（重点手動） | [../SECTION_14_MANUAL_TEST_GUIDE.md](../SECTION_14_MANUAL_TEST_GUIDE.md) |
| 画面表示・遷移・検索・マスタ CRUD の回帰 | Playwright E2E（`frontend/e2e/`） |
| **上記が覆わないギャップ（臨床安全・LIFF・入院・フォーム項目単位）** | **本ディレクトリ** |

## シナリオ索引（S01 から順に実行）

| ID | シナリオ | 分類 | 深度 |
|:---|:---|:---|:---|
| [S01](S01-deceased-pet-guard.md) | 死亡ペット誤操作の物理ブロック | 臨床安全 | 深い |
| [S02](S02-exam-abnormal-highlight-lock.md) | 検査異常値ハイライトと確定ロック | 臨床安全 | 深い |
| [S03](S03-vaccination-next-due-autocalc.md) | ワクチン接種→次回予定自動計算 | 臨床安全 | 深い |
| [S04](S04-liff-reservation-journey.md) | LIFF 飼い主予約ジャーニー通し | 顧客体験 | 薄い+境界 |
| [S05](S05-hospitalization-cycle.md) | 入院サイクル（ケア記録→退院会計） | 入院 | 深い |
| [S06](S06-record-lock-audit-trail.md) | カルテ確定 Lock と監査証跡 | 臨床安全 | 深い |
| [S07](S07-estimate-status-control.md) | 見積ステータス制御 | 会計 | 深い |
| [S08](S08-accounting-corrections.md) | 会計訂正系（クレジット訂正・未収金） | 会計 | 深い |
| [S09](S09-closing-time-boundaries.md) | 締め境界（AM/PM/EMG・越日） | 会計 | 深い |
| [S10](S10-customer-aggregation-consistency.md) | 顧客集計ダッシュボード整合 | 経営 | 薄い |
| [S11](S11-trimming-combined-accounting.md) | トリミング業務と診察併用精算 | 会計/トリミング | 深い |
| [S12](S12-liff-pet-health.md) | LIFF ペットヘルスとアカウント連携 | 顧客体験 | 薄い |
| [S13](S13-identity-links-manual-correction.md) | 同一飼主・ペット連携 — 手動訂正 | 顧客/組織 | 中 |
| [S14](S14-search-and-multi-word.md) | 飼主・ペット検索 — 複数語 AND と医院スコープ | 顧客/検索 | 中 |
| [S15](S15-insurance-rate-preservation.md) | 会計保険負担割合 — 新規 50/70 と既存 90/100 の保持 | 会計 | 中 |
| [S16](S16-interview-history-navigation.md) | 問診抜粋 → カルテ詳細への履歴導線 | 臨床 | 薄い |
| [S17](S17-treatment-quantity-commit.md) | 治療数量入力 — Enter 2 回・Blur・Escape・IME の確定操作 | 臨床 | 深い |
| [S18](S18-treatment-search-dialog-height.md) | 治療項目検索ダイアログ — 一覧の可視範囲と操作到達 | 臨床/UI | 薄い |
| [S19](S19-medical-record-viewport-fit.md) | カルテ編集画面 — 1366×625 での全タブ表示適合 | 臨床/UI | 中 |
| [S20](S20-master-to-accounting-path.md) | マスタ登録 → 会計明細への経路（価格・税区分・請求不能） | 会計/マスタ | 深い |
| [S21](S21-accounting-concurrency-idempotency.md) | 会計確定の並行更新 — 冪等リプレイ・409・確定後保護 | 会計 | 深い |
| [S22](S22-reservation-conflict-and-staff.md) | 予約作成 — 409 理由の表示と担当者候補のフェイルクローズ | 予約 | 中 |
| [S23](S23-multi-clinic-scope.md) | 複数医院所属 — 拠点スコープと他院操作の境界 | 組織/認可 | 深い |
| [S24](S24-vital-signs-latest-chips.md) | カルテヘッダーのバイタルサイン — 最新値のみ・バックフィルなし | 臨床 | 薄い |
| [S25](S25-microchip-header-display.md) | 患者ヘッダーのマイクロチップ番号表示 | 臨床 | 薄い |
| [S26](S26-medical-record-image-capture.md) | カルテ画像 — 撮影・アップロード入力と同一ファイル再選択 | 臨床 | 薄い |
| [S27](S27-chief-complaint-blank-persistence.md) | 主訴 — 区分空欄許可・本文保持・意図的解除 | 臨床 | 中 |
| [S28](S28-manual-urine-examination.md) | 手動尿検査 — 手入力項目・文字列保持・機器結果との非混在 | 臨床(検査) | 中 |
| [S29](S29-same-day-multiple-vaccinations.md) | 同一ペット同日の複数予防接種登録 | 臨床(予防) | 中 |
| [S30](S30-shift-occupation-filter.md) | シフトカレンダー — 職種フィルタと未設定スタッフ | 組織 | 薄い |
| [S31](S31-detail-route-direct-access.md) | 詳細画面への直接到達（会計・入院・在庫） | 回帰 | 薄い |
| [S32](S32-owner-search-modal-fit.md) | 飼主検索モーダル — 多数候補の可視範囲と操作到達 | 顧客/UI | 薄い |
| [S33](S33-receipt-print-pdf.md) | 明細兼領収書 — プレビュー → 印刷 → PDF 保存 | 会計 | 中 |
| [S34](S34-overlay-stack-fit.md) | オーバーレイの積層 — モーダル内ポップアップ・入れ子ダイアログ・トースト | UI | 中 |
| [S35](S35-extreme-content-fit.md) | 極端コンテンツ — 長文・空・NULL・大桁数・全角/絵文字の表示適合 | UI | 中 |
| [S36](S36-table-kanban-horizontal-fit.md) | 一覧テーブル・カンバン — 横方向の収まりと操作到達 | UI | 中 |
| [S37](S37-print-documents-layout.md) | 印刷帳票のレイアウト — 領収書以外の印刷面 | UI/帳票 | 中 |
| [S38](S38-liff-mobile-viewport.md) | LIFF モバイル画面 — スマホ viewport での表示適合 | UI/LIFF | 中 |
| [S39](S39-state-feedback-visibility.md) | 状態フィードバック — loading・空・エラー・disabled・フォーカスの可視性 | UI/a11y | 中 |

実行順の制約: S01 を最初に、S10 は S08 の後。S13・S14〜S39 は独立。それ以外は任意順。S21 は `billing-schema-readiness` の DB 制約ゲート、S28 は医院承認の項目表、S23 は複数医院 fixture が前提 — 前提未充足のシナリオは BLOCKED として扱い、スキップを PASS にしない。

機器クライアントの独立確認票: [LAB_DEVICE_CLIENT_UAT.md](LAB_DEVICE_CLIENT_UAT.md)（NX600/AU10V。未確定機器を PASS にしない）。

**#254 close 条件の再配置（結果は書かない）**: [UAT-254-CLOSE-CHECKLIST.md](UAT-254-CLOSE-CHECKLIST.md)。local PASS では close しない。実施レーンは BRT-68。

## V シリーズ — フォーム検証（入力・更新・DB 整合 + **項目単位**）

業務フロー横断の S シリーズと別軸で、inventory に exact field key が収録された永続化フォームの入力・更新・DB 整合を検証する。inventory 再構築完了までは「全フォーム」を主張しない。

| 文書 | 役割 |
|:--|:--|
| [FIELD-LEVEL-PROTOCOL.md](FIELD-LEVEL-PROTOCOL.md) | **項目単位**チェック F0–F6 の定義（必須） |
| [FORM-FIELD-INVENTORY.md](FORM-FIELD-INVENTORY.md) | フォーム× fieldKey の棚卸し（カバー範囲） |
| V01〜V05 | フォーム単位の手順・業務固有チェック・C1〜C3 |

各 V ファイル冒頭の「共通チェック手順」（C1 / C2 / C3）に加え、**inventory の全項目に F プロトコルを適用**して完了とする。代表 1 項目だけの C1 では受入完了とみなさない。

| ID | ファイル | 対象ドメイン | フォーム数 |
|:---|:---|:---|:---|
| [V01](V01-clinical-forms.md) | 臨床系 | 臨床 | **算定保留** |
| [V02](V02-accounting-reservation-forms.md) | 会計・予約・受付・シフト・**在庫** | 会計/予約/在庫 | **算定保留** |
| [V03](V03-owner-pet-staff-forms.md) | 飼主・ペット・スタッフ・権限・医院 | 顧客/組織 | **算定保留** |
| [V04](V04-settings-master-forms.md) | /settings マスタ | マスタ | **算定保留** |
| [V05](V05-auth-line-forms.md) | 認証・LIFF・LINE・Lステップ | 認証/LINE | **算定保留** |
| **合計** | | | **算定保留** |

フォーム inventory は再構築中であり、全フォーム・全項目の完了や一意フォーム総数はまだ主張できない。route inventory の 86 product pages はフォーム数とは別の指標。

## 実行と記録のルール

- **環境**: 自動投入されるのは `002_master` の参照マスタだけ。各シナリオに記載した合成 fixture を、承認済み UAT skeleton/import 手順でローカルの使い捨て clinic に作成する。手順の正本が `../UAT-ENV-SETUP.md` で修正済みであることを実行前に確認する。共有 STG、既存患者・飼主、固定 ID は使用しない。
- **実行記録はシナリオファイルに書かない**。証跡は gitignore の `reports/uat-YYYY-MM-DD/`（results は `formId.fieldKey.Fx` 推奨）。
- **製品 FAIL は `todo.md#product-bugs` 必須**（確認済みのみ · 見出し重複禁止 · env/権限 BLOCKED は書かない）。PARTIAL は todo.md#product-bugs にしない。Linear Issue 化は後続レーン。
- **S シリーズは core 受入**: local では S01→S39 を実施し FINAL を書く。V シリーズは項目単位の別軸（inventory 全 fieldKey）。「全て実施」は少なくとも core S の実行完了を指し、FAIL/PARTIAL/BLOCKED が残る場合は「全て PASS」と言わない。S14〜S39 は各ファイル冒頭の前提条件（fixture・権限・ゲート）を満たした環境でのみ実行する。
- **AI 実行**: browser-test + Chrome DevTools MCP、または Playwright MCP / 再現スクリプト。
- **【要実測】**: 観測結果を実装/テスト/承認済み仕様と照合する。観測だけで期待結果へ昇格せず、仕様判断が残る場合は PARTIAL/BLOCKED。
- **クレデンシャル禁止**: パスワード・トークンを本ディレクトリに書かない。アカウントはロール名。認証は `E2E_LOGIN_*`。

## シナリオの構造（S テンプレート）

```markdown
# S{NN}: <業務フロー名>
> **目的** / **所要目安** / **深度** / **仕様正本**
## 前提条件
## 手順と期待結果
## 確認観点
## 異常系（深度=深いのみ）
```
