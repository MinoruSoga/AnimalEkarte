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

実行順の制約: S01 を最初に、S10 は S08 の後。S13 は独立。それ以外は任意順。

**#254 close 条件の再配置（結果は書かない）**: [UAT-254-CLOSE-CHECKLIST.md](UAT-254-CLOSE-CHECKLIST.md)。local PASS では close しない。実施レーンは BRT-68。

## V シリーズ — フォーム検証（入力・更新・DB 整合 + **項目単位**）

業務フロー横断の S シリーズと別軸で、全永続化フォームの入力・更新・DB 整合を検証する。

| 文書 | 役割 |
|:--|:--|
| [FIELD-LEVEL-PROTOCOL.md](FIELD-LEVEL-PROTOCOL.md) | **項目単位**チェック F0–F6 の定義（必須） |
| [FORM-FIELD-INVENTORY.md](FORM-FIELD-INVENTORY.md) | フォーム× fieldKey の棚卸し（カバー範囲） |
| V01〜V05 | フォーム単位の手順・業務固有チェック・C1〜C3 |

各 V ファイル冒頭の「共通チェック手順」（C1 / C2 / C3）に加え、**inventory の全項目に F プロトコルを適用**して完了とする。代表 1 項目だけの C1 では受入完了とみなさない。

| ID | ファイル | 対象ドメイン | フォーム数 |
|:---|:---|:---|:---|
| [V01](V01-clinical-forms.md) | 臨床系 | 臨床 | 18 |
| [V02](V02-accounting-reservation-forms.md) | 会計・予約・受付・シフト・**在庫** | 会計/予約/在庫 | **12** |
| [V03](V03-owner-pet-staff-forms.md) | 飼主・ペット・スタッフ・権限・医院 | 顧客/組織 | 7 |
| [V04](V04-settings-master-forms.md) | /settings マスタ | マスタ | 30 |
| [V05](V05-auth-line-forms.md) | 認証・LIFF・LINE・Lステップ | 認証/LINE | 18 |
| **合計** | | | **85** |

（旧 84 = 在庫フォーム未計上。2026-08-14 に `inventory-form` を V02 §12 へ追加。）

## 実行と記録のルール

- **環境**: [../UAT-ENV-SETUP.md](../UAT-ENV-SETUP.md)。ローカル（seed 003_demo）または STG（004_staging）。前提データは検索条件で指定（ID 直書き禁止）。
- **実行記録はシナリオファイルに書かない**。証跡は gitignore の `reports/uat-YYYY-MM-DD/`（results は `formId.fieldKey.Fx` 推奨）。
- **製品 FAIL は `bug.md` 必須**（確認済みのみ · 見出し重複禁止 · env/権限 BLOCKED は書かない）。PARTIAL は bug.md にしない。Linear Issue 化は後続レーン。
- **S シリーズは core 受入**: local では S01→S13 を実施し FINAL を書く。V シリーズは項目単位の別軸（inventory 全 fieldKey）。「全て実施」は少なくとも core S の実行完了を指し、FAIL/PARTIAL/BLOCKED が残る場合は「全て PASS」と言わない。
- **AI 実行**: browser-test + Chrome DevTools MCP、または Playwright MCP / 再現スクリプト。
- **【要実測】**: 初回実測後、正しければ期待結果へ昇格。
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
