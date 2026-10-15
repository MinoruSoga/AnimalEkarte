# 飼主カルテレポート 仕様書 (Owner Report)

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | page.owners.clinical-report／D21 業務・画面仕様 |
| 決定状態 | 既存記載を継承。明示済み合意と現行実装説明を分け、2026-10-15整理節の提案・未決は未承認 |
| 実装状態 | 既存記載を継承（今回の再監査なし）。既知gapを保持 |
| 実機検証状態 | 今回未実施。テスト名・過去証拠の掲載は再実行を意味しない |
| 文書確認日 | 2026-10-15。本文・参照・FR/AC整理。独立レビュー前 |


## 概要

- **目的**: 選択中ペットの「診療前に確認する情報」「今日の来院」「次の行動」「前回診療」「基本情報」「種類別履歴」を、ページ切替やページスクロールなしで俯瞰する読み取り専用画面。
- **URL**: `/owners/:id/report?petId=`。`petId` が同一飼主のペットでない場合は先頭ペットへフォールバックする。
- **表示形態**: サイドバー付き Layout 外で別ウィンドウ表示するスタンドアロンルート。
- **ページ権限**: `medical-records:view` と `owners:view` の両方が必要。未認証時はログインへ遷移し、権限不足時はキャッシュ済みの飼主・ペット情報も描画しない。
- **個別権限**: 検査・予防接種・健診・トリミング・予約は各リソースの `view` を確認し、権限がない情報だけ「閲覧権限なし」に縮退する。

```mermaid
flowchart TB
    A["別ウィンドウで /owners/:id/report?petId= を開く"] --> G{"ページ権限: medical-records:view と owners:view の両方が必要"}
    G -->|未認証| L["ログインへ遷移"]
    G -->|不足| NG["キャッシュ済みの飼主・ペット情報も描画しない"]
    G -->|両方あり| R["固定ヘッダーと全領域を同じビューポートに描画"]
    R --> P{"各リソースの view 権限を確認"}
    P -->|あり| S["領域に情報を表示"]
    P -->|なし| D["その情報だけ「閲覧権限なし」に縮退"]
```

## 1. 画面構成

### 1.1 固定ヘッダー

- **飼主情報 (`OwnerReportPanel`)**: 飼主No・氏名・ふりがな・電話だけをコンパクトに表示する。住所・勤務先・DM 等の詳細は「基本情報」に集約する。
- **選択中ペット (`SelectedPetContext`)**: ペット名、種別・品種、年齢、性別、体重を表示する。
- **ペット切替 (`PetSwitcher`)**: タブではなくラベル付きネイティブ select を使う。選択時はページ遷移せず、URL の `?petId=` を replace で同期する。

### 1.2 診療ブリーフィング（6領域）

`OwnerClinicalBriefing` は次の6領域を同じビューポート内へ配置する。順序は業務上の確認順に合わせる。

| # | 領域 | 優先情報 |
|:---|:---|:---|
| ① | 診療前の確認 | ペット備考、血液型、直近の薬・処方、最新検査、予防予定 |
| ② | 今日の来院 | 本日の予約時刻、初診/再診、受付状態、予約内容、担当 |
| ③ | 次の行動 | 次回予約（確定・実予約）、次回予定（tentative appointment の意図日・時間未定）、予防予定 |
| ④ | 前回診療 | 最新の確定カルテにある診療日、担当、主訴、評価、方針 |
| ⑤ | 基本情報 | ペット・飼主属性、初診日、前回来院、保険、住所、勤務先、DM |
| ⑥ | 種類別履歴 | 縦を情報種類、横を新しい日付順とした履歴マトリクス |

### 1.3 種類別履歴

- 縦行は「診療」「検査」「薬・処方」「予防接種」「処置」「ケア」の6種類で固定し、薬・予防接種・処置を混在させない。
- 横列は取得済みデータの全日付を新しい順に並べ、同日の複数種類を同じ日付列へ配置する。画面幅を超える列は履歴領域内の横スクロールで参照する。これは日付による俯瞰であり、別APIのデータを同一診療レコードとして結合したことを意味しない。
- 治療履歴は `filter=all` で1回だけ取得し、`itemType=medicine` を「薬・処方」、それ以外を「処置」に表示する。旧画面の重複した4クエリ・重複掲載は行わない。
- 検査は依頼中・検査中を履歴から除外する。検査結果明細がAPIから返らない場合は「異常なし」と断定せず「結果明細なし」と表示する。
- トリミングは status「完了」の施術だけを「ケア」として表示する。
- 履歴系APIの取得上限は `HISTORY_FETCH_LIMIT`（100件）。上限超過を検知した場合は件数末尾に `+` を付け、打ち切り注記を表示する。
- 上限超過の検知はバックエンド `total` と累積取得 raw 行数の比較で行い、クライアント側フィルタ（検査の下書き除外・トリミングの完了のみ抽出）後の可視件数では判定しない。
- 追加読み込み: 「検査」「薬・処方」「処置」「ケア」の4行は、取得元（検査 API・治療履歴 API・トリミング API の3系統）に残ページがある限り行見出しに「続きを読む」を表示する。押下でのみ次ページを取得して累積表示し、初回マウントで page>1 の自動取得は行わない。「薬・処方」と「処置」は同一治療履歴ソースを共有するため同じ制御を両行に出す。「診療」「予防接種」行は対象外。健診結果は「検査」行に合算表示されるが追加取得対象ではない。
- 治療・予防接種の日付はJST壁日付へ正規化してから表示する。

### 1.4 レイアウト

- デスクトップ・タブレット・モバイルを問わず、ルートをビューポートへ固定し、ページ全体をスクロールさせない。
- 6領域の見出しを常に同時提示する。情報が領域を超えた場合だけ、各パネル本文の内部スクロールを許可する。
- 幅840px以下は6領域を1列6行へ再配分するが、自然高さの縦積みにはせず、利用可能な高さを比率配分する。
- 高さ600px以下は2〜3列の短辺向けグリッドへ切り替え、6領域を2〜3段に収める。
- 種類別履歴は内部だけ横スクロールを許可し、日付列見出しと種類行見出しを sticky に保つ。
- フォントサイズは `docs/spec/design-system.md` の役割に従い、見出し20px、主要値15px、ラベル14px、補助情報12pxを基本とする。メディアクエリで本文文字を縮小しない。

## 2. 主要機能

### 2.1 起動導線

共有ヘルパー `openOwnerReport` が noopener,noreferrer 付きで別ウィンドウを開く。飼主・ペット一覧とカルテ画面ヘッダーから起動できる。ウィンドウタイトルは `飼主レポート - {飼主名}` とする。

### 2.2 同居ペット

`useGetOwnerReportPets` が返す同一飼主のペットだけを select に表示する。画面は常に1頭分で、複数ペットを合算しない。切替後は対象ペットの各クエリをReact Queryキャッシュ経由で取得する。

### 2.3 印刷

画面閲覧専用で、印刷ボタン・印刷用スタイル・共有機能は設けない。固定ビューポートと内部スクロールのためブラウザ標準印刷では長い履歴が切れ得るが、印刷は要件外とする。

## 3. 臨床安全・アクセシビリティ

### 3.1 臨床安全

- 未設定値は「記録なし」または `-` とし、APIに存在しない情報を他の項目から代用しない。
- 年齢だけは生年月日から `formatPetAge` で導出する。初診日は専用API、前回来院はペットAPIの値を使う。
- 最新検査の異常件数だけを診療前確認へ表示し、過去の異常を現在の要確認事項として累積しない。
- 権限なし・読み込み中・取得失敗を0件や「予定なし」として扱わない。
- 前回診療と診療履歴は `status=finalized` のカルテだけを対象とし、下書きを診療済みとして扱わない。
- 本画面は読み取り専用で、入力・保存操作を持たない。

### 3.2 アクセシビリティ

- ペット切替は accessible name「ペット切替」のネイティブ select。tab / tablist / tabpanel は使用しない。
- 6領域は各見出しで命名された region とし、本文は `tabIndex=0` でキーボードによる内部スクロールを可能にする。
- 種類別履歴は table、日付は `scope=col`、種類は `scope=row` とし、読み上げでも縦横の対応を保つ。
- ローディングは `role=status`、取得失敗は `role=alert` で通知する。

## 4. 技術仕様

### 使用コンポーネント

- `OwnerReport`: 認証・ページ権限・ペット選択・URL同期。
- `OwnerReportPanel` / `SelectedPetContext` / `PetSwitcher`: 固定コンテキスト。
- `OwnerClinicalBriefing`: データを一度ずつ取得し、6領域向けに派生する。
- `ClinicalBriefingPanel`: 見出し付き境界内スクロール領域。
- `ClinicalHistoryMatrix`: 縦=種類、横=日付の履歴表。残ページのある行は行見出しに「続きを読む」を出す。
- `useGetPetExaminations` / `useGetPetTreatmentHistory` / `useGetPetTrimmingHistory`: `useInfiniteQuery` ベースの履歴3系統。ページは `fetchNextPage` で手動追加され、消費側には累積済み `{items, isTruncated}` を返す。

### API連携

| メソッド | エンドポイント | 用途 | 必須権限 |
|:---|:---|:---|:---|
| GET | `/api/v1/owners/:id` | 飼主基本情報 | `owners:view` |
| GET | `/api/v1/owners/:id/report/pets` | Owner Report用に絞った同居ペット一覧 | `owners:view` |
| GET | `/api/v1/pets/:id/first-visit` | 初診日 | `medical-records:view` |
| GET | `/api/v1/medical-records?pet_id=&status=finalized&sort=date&order=desc` | 前回診療・診療履歴（最大100件） | `medical-records:view` |
| GET | `/api/v1/pets/:id/treatment-history?page=` | 薬・処方と処置（filter=all・手動追加読み込み） | `medical-records:view` |
| GET | `/api/v1/vaccinations?pet_id=` | 予防予定・予防接種履歴 | `vaccinations:view` |
| GET | `/api/v1/examinations?pet_id=&page=` | 最新検査・検査履歴（手動追加読み込み） | `examinations:view` |
| GET | `/api/v1/checkups/field-results?pet_id=` | 健診履歴 | `checkups:view` |
| GET | `/api/v1/trimmings?pet_id=&page=` | ケア履歴（手動追加読み込み） | `trimming:view` |
| GET | `/api/v1/reservations?pet_id=&start_date=&end_date=` | 今日の来院・次回予約（今日から731日（現行2年予定窓）） | `reservations:view` |

clinic隔離・論理削除除外は各バックエンドAPIが担保する。

<a id="owner-report-document-contract"></a>
## D21の読取専用・患者context契約整理（2026-10-15）

### 目的・対象範囲と状態

診療前の院内スタッフ向け複合読取画面。飼主本人向け公開、共有、印刷、臨床記録のwriteを追加しない。本文のページ権限・section別権限・1頭context・6領域・履歴の取得制約を既存記載として継承する。現コード再監査／実機再受入は行っていない。

### 操作・境界

<a id="requirement-fr-owner-report-page-section-auth"></a>

<a id="requirement-fr-owner-report-single-pet-context"></a>

<a id="requirement-fr-owner-report-history-meaning"></a>

<a id="requirement-fr-owner-report-partial-data"></a>

<a id="requirement-fr-owner-report-manual-pagination"></a>

<a id="requirement-fr-owner-report-read-only"></a>

| 要件ID | 操作・契約 | 根拠と決定状態 |
|:---|:---|:---|
| FR-owner-report-page-section-auth | medical-records:viewとowners:viewの両方をページ前提とし、各sectionは各リソースviewで縮退する | 本文概要・§3 |
| FR-owner-report-single-pet-context | 同一飼主の1頭だけを選択表示し、URLのpetIdをreplace同期する。範囲外petIdは先頭ペットへfallbackし、複数頭を合算しない | 本文概要・§1〜2 |
| FR-owner-report-history-meaning | 診療はfinalized、検査は依頼中/検査中を除外、ケアは完了のみ。日付で横断表示しても同一診療IDへ統合した意味にしない | 本文§1.3、§3.1 |
| FR-owner-report-partial-data | 権限なし/読込中/失敗/0件/明細なし/取得上限を区別し、異常なしや予定なしへ代用しない | 本文§1.3、§3 |
| FR-owner-report-manual-pagination | 対象3sourceだけ利用者操作で追加取得し、薬/処置は同一sourceを共有する。上限検知はraw行数とtotalで判定する | 本文§1.3 |
| FR-owner-report-read-only | 別ウィンドウの閲覧に限定し、保存・印刷・共有を新設しない | 本文§2〜3 |

### 受入条件

| 条件ID | Given（前提） | When（操作） | Then（期待結果） | 独立確認・成功条件・確認手順／証跡 |
|:---|:---|:---|:---|:---|
| AC-owner-report-page-denied | ページ必須権限の片方がない | 開く | キャッシュ済み飼主/ペット情報も描画しない | 本文概要の両権限条件を照合 |
| AC-owner-report-section-denied | ページ権限はあるが検査viewがない | 6領域を見る | 検査部分だけ閲覧権限なしとし他の許可情報を継続表示する | 各section権限対応を照合 |
| AC-owner-report-pet-switch | 同居ペットが複数いる | 1頭を切り替える | 選択中contextとURLがその1頭に対応し複数頭を合算しない | 本文§1.1/2.2を照合 |
| AC-owner-report-outside-pet | petIdが当該飼主に属さない | 開く | 同一飼主ペットの先頭へfallbackする | 本文概要/OwnerReport.tsx。空配列は専用0ペット表示、詳細query対象なし |
| AC-owner-report-finalized | 下書きと確定カルテが混在する | 前回診療/履歴を見る | finalizedだけを診療済みとして使う | 本文§3.1 |
| AC-owner-report-no-result-detail | 検査明細が返らない | 最新検査/履歴を見る | 結果明細なしと表示し異常なしを捏造しない | 本文§1.3 |
| AC-owner-report-truncation | totalがraw取得行数より多い | 履歴を表示する | フィルタ後件数と混同せず打切りを示す | 本文§1.3のraw/可視の区別 |
| AC-owner-report-more | 追加取得対象sourceに次pageがある | 続きを読むを押す | 対応sourceを追加し、薬/処置は共通制御。初回自動page>1取得をしない | 本文§1.3 |
| AC-owner-report-errors | 一部sourceが取得失敗 | 表示する | 失敗を0件/予定なしへ置き換えない | 本文§3.1とrole=alert |
| AC-owner-report-readonly | レポートを開く | 操作を確認する | 閲覧・患者切替・追加読取だけで、臨床writeや新しい公開操作がない | 本文§2.3/3.1。ブラウザ標準印刷の品質は対象外 |

### 保存・再取得・遅延応答と引継ぎ

画面自身の保存・取消は該当なし（writeを所有しない）。訂正はD03/07/08/09等の正本で行い、本画面は再取得して表示する。表示中の他画面変更や権限変更を常時即時反映する保証は本文にないため創作しない。

技術整理提案：医院・飼主・ペット・権限contextをquery identityとして、切替前の遅延応答を切替後患者の情報へ混入させず、失敗時に前患者を現在患者として残さない。既存の1頭context・認可・臨床安全を満たす受け手契約であり、新しい業務操作を追加するものではない。現行owner-report-pets query keyはowner IDのみで、共通query-key方針は医院切替時clear＋reload。0ペット時は専用空状態表示がある。遅延応答と全関連queryの実装適合は実機確認対象であり、query keyに医院IDがあると誤記しない。


### 読取操作の静的照合（2026-10-15）

`frontend/src/features/owner-report/routes/OwnerReport.tsx`、`hooks/use-owner-clinical-briefing-data.ts`、`api/get-owner-report-pets.ts`、`frontend/src/lib/query-keys.ts` に基づく。

| 操作 | 現行契約 |
|:---|:---|
| 初期表示 | 認証loading中は表示保留、未認証はredirect。owner/pets読取loadingとerrorを分離する |
| 患者選択 | URL petが現在ownerのpetsに含まれれば選択、なければ先頭へfallback。手動選択はURL replace。fallback自体が必ずURLを書換えるとはしない |
| 0ペット | 「この飼主に登録されたペットがありません」を表示。pet切替/臨床取得対象を作らない |
| カルテ読取 | finalizedのみ、取得上限100。未取得がない全履歴とは表示しない |
| 他領域読取 | 領域viewなしはpet ID undefined/取得disabledとして対応queryを実行しない。画面のviewを各領域viewへ昇格させない |
| 予約読取 | 当日から731日の窓。従来本文365日は現行2年予定窓へ訂正。D01予定の存在は確定予約や来院実績と同義でない |
| 追加取得 | 対象3履歴の手動追加読込を個別に扱い、1件の失敗で他取得済セクションを空成功にしない |
| cache境界 | owner-report-petsはowner ID key、共通方針は医院切替clear＋reload。患者切替前の結果を別患者へ描画しない契約の実機確認は別途 |

| 条件ID | 前提・操作 | 期待結果 | 根拠 |
|:---|:---|:---|:---|
| AC-owner-report-no-pets | ownerにpetsなしで開く | 専用空状態、架空patientを選ばない | OwnerReport.tsx |
| AC-owner-report-invalid-pet | 別owner/不存在のpet queryで開く | 現ownerの先頭petへfallback、指定IDの他患者記録を表示しない | OwnerReport.tsx |
| AC-owner-report-finalized-only | draftとfinalizedカルテがある | report queryはfinalized・上限100で取得 | clinical briefing hook |
| AC-owner-report-section-query-denied | 予約等のviewなし | 該当queryを実行せず情報を漏らさない | 同hookのenabled/undefined |
