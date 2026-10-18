# S19: カルテ編集画面 — 1366×625 での全タブ表示適合

> **目的**: カルテ編集画面が 1366×625 CSS px の狭い viewport でも、最新仕様に従う統合後の全タブで必須情報・保存・主要操作へ見切れず到達できることを納品前に証明する。
> **所要目安**: 20分 / **深度**: 中
> **仕様正本**: [domains/medical-records/pages/medical-record-form.md §1](../../../spec/domains/medical-records/pages/medical-record-form.md)。検証キュー: `UAT-R2-CHART-FIT`（[todo-verification.md](../../../../todo.md#verification-ledger)）。

## 最新仕様との適用関係

[問診と治療プランの統合](../../../spec/domains/medical-records/lifecycle.md#integrated-plan-20261017)を優先し、独立した問診タブを維持する受入条件にはしない。治療プランへ移した入力・飼主由来問診の別欄・過去カルテ履歴も同時参照と操作到達の対象に含む。下記の9タブという数・旧コード構成の記述は従来実装を照合した時点の参照であり、現在の目標は統合後に存在する全タブを確認する。新しいタブ数・配置方式は本票で指定しない。既存の保存先・飼主原文保護・同時編集権・未保存入力保護も最新正本へ従う。編集権を保持し既存条件を満たす端末の正規の自動保存を、Enter誤送信防止や本票の手順で停止しない。本票を実機合格の証拠とはしない。

## 前提条件

- 対象 viewport は **1366×625 CSS px**。sidebar（`w-[220px]`）の開閉両状態で確認する。実機の CSS 表示領域・ブラウザ版・100%/既報 125% ズームを別記し、最新 Chromium の合格を旧 Chrome 対応済みとしない。
- 長文の所見・多数の治療明細・多数の検査項目を持つ fixture と、空の新規カルテの両方を用意する。
- actor は medical-records の編集権限を持つ attached account。開くだけでは閲覧で編集権を取得しない。入力前に「編集を開始」で取得し、同じカルテのタブ移動では維持する（[編集権の正本](../../../spec/domains/medical-records/lifecycle.md#record-edit-right-lifecycle-20261018)）。
- 依存シナリオ: S17・S18 と同じ fixture/run を使ってよい。会計タブの並行更新は S21、画像は S26 を参照する。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | カルテ編集を 1366×625 で開き、タブバーを確認する | 統合後に存在する全タブが到達可能。旧問診入力と過去履歴は治療プラン内で参照・操作できる。タブバーが画面外に押し出されない |
| 2 | sidebar 開・閉の両状態で主要タブを切り替える | どちらの状態でもタブ切替・表示・操作が破綻しない。sidebar 幅が本文を潰さない |
| 3 | 統合後の治療プラン: スタッフ記録の長文入力・飼主問診の別欄参照・過去履歴の同時参照を行い、所定の保存操作まで進む | 入力欄と保存ボタンへスクロールで到達できる。sticky ヘッダーと本文が重なって読めない/押せない状態にならない |
| 4 | 治療タブ: 多数明細の一覧を最下行まで操作する | 一覧がコンテナ内でスクロールし、追加・数量・削除・保存操作へ到達できる（`UnifiedTabsRoot` の `flex-1 min-h-0` 配下で内側スクロール） |
| 5 | 検査・画像タブを開く | 検査項目テーブル・画像ギャラリーが viewport 内で操作可能。モーダル/ドロップダウンが画面外へ逃げない |
| 6 | 未保存変更がある状態でタブ切替・離脱を試みる | `NavigationBlocker` が離脱確認を出し、確認なしに入力を失わない。キャンセルで元の編集に留まれる |
| 7 | 保存ボタン・患者ヘッダー（飼主/ペット名・死亡マーク）の到達と可読性を確認する | 長い名前はツールチップ付きで truncate、必須操作は常に到達可能。header の固定で本文先頭が隠れない |

画面構成とスクロールの分担:

```text
カルテ編集画面（1366×625 CSS px・sidebar 開閉両状態）
├─ 患者ヘッダー … sticky。飼主/ペット名は truncate + ツールチップ、死亡マーク表示
├─ タブバー     … 全タブが到達可能・画面外へ押し出されない
└─ タブ本文     … flex-1 min-h-0 配下で内側スクロール（max-height ではない）
   ├─ 治療プラン … 統合した入力・飼主原文の別欄・過去履歴と所定の保存へ到達
   ├─ 治療  … 明細一覧がコンテナ内スクロール・追加/数量/削除/保存へ到達
   └─ 未保存変更での離脱 → NavigationBlocker が確認ダイアログ
```

## 確認観点

- タブ構成は `MEDICAL_RECORD_TAB_ITEMS`（9 タブ、`medical-record-form-model.ts`）、レイアウトは `MedicalRecordFormReadyPanels` の `UnifiedTabsRoot`（`flex-1 min-h-0`）。高さ制約は直接の `max-height` ではなく flex/min-h-0 による内側スクロールで実現する点に注意し、レンダリング結果で判断する。
- 1366×625 で「見切れず操作できる」は `UAT-R2-CHART-FIT` の受入条件。統合後の全タブ・sidebar 両状態・長文/長一覧/ダイアログの組み合わせを最低カバーする。
- sidebar の開閉は `Sidebar.tsx` 内のローカル `useState`＋`matchMedia` リスナーで制御され、幅は `design-tokens.ts` の `LAYOUT.sidebar.expanded`（`w-[220px]`）/`collapsed`（`w-[56px]`）（2026-09-29 訂正: 「`use-sidebar` / `SidebarProvider`」→ そのようなフック/Provider は現行コードに存在せず、`Sidebar.tsx` のローカル state が実体）。開閉でタブや保存が隠れる場合は FAIL。
- 患者ヘッダーの truncate・死亡マークは S25（MC 番号）・S01（死亡ガード）と連携。ここでは viewport 適合の観点で確認する。
- 見切れ・操作不能が出たら対象端末・ズーム・タブを特定して FAIL 記録。環境側の要因（非対象ブラウザ等）は BLOCKED として分ける。

## 実装突合

- 変更サマリ:
  - `MEDICAL_RECORD_TAB_ITEMS` の 9 タブと `UnifiedTabsRoot` の `flex-1 min-h-0` レイアウトを現行コードと突合
  - `UAT-R2-CHART-FIT` の「1366×625・全 9 タブ・sidebar 両状態・長文/長一覧/ダイアログで必須情報・保存・フォーカス到達」を手順 1–7 に対応づけ
  - `NavigationBlocker`（未保存保護）と sticky ヘッダーの併存を手順 6–7 に組み込み
  - 2026-09-29 訂正: sidebar 開閉の実体を「`use-sidebar`/`SidebarProvider`」→ `Sidebar.tsx` ローカル `useState`+`matchMedia`・`LAYOUT.sidebar`（expanded `w-[220px]` / collapsed `w-[56px]`）へ訂正
