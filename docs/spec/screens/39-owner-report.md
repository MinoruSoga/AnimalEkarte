# 飼主カルテレポート 仕様書 (Owner Report)

## 概要
- **画面の目的**: 飼主単位で全ペットの診療サマリー（既往歴・投薬歴・ワクチン・トリミング等）を 1 画面で俯瞰する。カルテ入力中や受付対応中に別ウィンドウで参照する読み取り専用ビュー（GitHub #158/#159/#178 由来）。
- **URLパターン**: `/owners/:id/report`（`?petId=` クエリで初期表示ペットを指定可能）
- **アクセス権限**: サイドバー付き Layout の外に登録されるスタンドアロンルート（`app-routes.tsx`）のため、認証ガードと権限ゲートを `OwnerReport` 自身が持つ。
  1. 未認証ならログインページへリダイレクト（`useAuth`）。
  2. `RequirePermission` + `ResourceMedicalRecords`（デフォルトアクション view）でページ全体をガード。権限なしは「アクセス権限がありません」表示。
  3. さらに各履歴セクションが `usePermission` で自リソースの view 権限を個別チェックし、権限がないセクションだけ「閲覧権限がありません」に縮退する（ページ全体は落とさない）。

---

## 1. 画面構成

### 1.1 ヘッダー（sticky 固定）
- **飼主情報ストリップ (`OwnerReportPanel`)**: 飼主No・氏名・ふりがな・電話・会員区分に加え、値が存在する場合のみ住所・メール・勤務先・勤務先TEL・DM 区分を横並び表示する。DM 区分は未設定（null）の場合は行ごと非表示（`formatDMPreference` が「未設定」と「不要」を区別し、既存飼主を「不要」で埋めない）。
- **ペット切替タブ (`PetSwitcher`)**: 飼主に紐づく全ペット（同居ペット）をタブとして表示。切替はページ遷移せず state 更新 + URL `?petId=` 同期（replace）。`?petId=` が無効値なら先頭ペットへフォールバック。

### 1.2 レポート本文（9 パネルグリッド）
選択中ペット 1 頭分を xl: 3列×3行 / lg: 2列×5行 / それ未満: 1列 で敷き詰める。

| # | セクション | コンポーネント | データソース |
|:---|:---|:---|:---|
| ① | ペット詳細 | `PetDetailSection` | `useGetPets` のペット属性 + `useGetPetFirstVisit`（初診日 = 最古の有効カルテ date 由来の派生値） |
| ② | 予防接種履歴 | `VaccinationHistorySection` | `useGetPetVaccinations`（接種日・ワクチン・次回予定、日付降順） |
| ③ | 健康診断（検査）履歴 | `ExaminationHistorySection` | `useGetPetExaminations`（下書き = 依頼中/検査中を除外。検査ごとに項目・結果・基準値の小テーブル） |
| ④ | 健診（パッケージ）履歴 | `CheckupHistorySection` | `useGetPetCheckupResults`（#211。checkup 単位にグルーピングし型付きフィールドを表示） |
| ⑤ | 投薬履歴 | `TreatmentHistorySection`（filter=medicine） | `useGetPetTreatmentHistory`（薬剤・投与経路・数量） |
| ⑥ | 麻酔処置履歴 | `TreatmentHistorySection`（filter=procedure + anesthesiaOnly） | 同上（#159。麻酔種別を日本語ラベルで併記） |
| ⑦ | 手術処置履歴 | `TreatmentHistorySection`（filter=procedure + isSurgery） | 同上（#159。麻酔列付き） |
| ⑧ | 治療履歴 | `TreatmentHistorySection`（filter=all） | 同上（投薬・処置を含む全 treatments。⑤〜⑦ と重複掲載） |
| ⑨ | トリミング履歴 | `TrimmingHistorySection` | `useGetPetTrimmingHistory`（`selectCompletedTrimmingHistory` が status「完了」= 施術実施済みのみを実施日降順で返す。予約・進行中・キャンセルは除外） |

- 治療系の診療日・予防接種の接種日/次回予定日は、いずれも `toJSTWallDate` により絶対時刻を JST 壁日付に変換してから整形する（ブラウザのローカル TZ に依存すると日付が 1 日ずれ得るため。SD-19）。
- 各セクションのヘッダー右に件数バッジを表示（閲覧可・取得済み・非エラー時のみ）。
- 履歴系 API（検査・投薬/麻酔/手術/治療・トリミング）の取得上限は `HISTORY_FETCH_LIMIT`（100 件）。バックエンド `total` が実際に取得した件数を上回る場合、該当セクション本文の先頭に「直近100件を表示しています」の打ち切り注記（`ReportSection` の `isTruncated`）を表示する（SD-18。ページング実装はスコープ外）。それでも古い履歴自体は表示されない（予防接種・健診（パッケージ）は limit 未指定で対象外）。

### 1.3 レイアウト挙動（密集ワークスペース）
- lg 以上ではページ全体をスクロールさせず、各パネル本文（`ReportPanel`）だけが内部スクロールする。履歴テーブルの列見出しは `HistoryTable` が sticky 化しスクロール中も保持。
- lg 未満（タブレット/モバイル）はパネルが自然高さで縦積みされ、ページスクロールに委ねる。ヘッダーは常に sticky。

---

## 2. 主要な機能

### 2.1 起動導線（別ウィンドウ）
共有ヘルパー `openOwnerReport`（`owner-report-window.ts`）が URL を構築し、セキュリティのため必ず noopener,noreferrer 付きで別ウィンドウを開く。導線は 2 箇所:
- **飼主・ペット一覧 (`OwnersList`)**: 行操作からペットを指定して起動。
- **カルテ画面ヘッダー (`MedicalRecordStickyHeader`)**: 診療中のペットを初期選択して起動。medical-records の view 権限がない場合はボタン自体を表示しない。

ウィンドウタイトルは `useTitle` で「飼主レポート - {飼主名}」となり、複数ウィンドウのタブ識別に使う。

### 2.2 同居ペットの扱い
レポートは飼主スコープで、`useGetPets`（owner_id 絞り込み）が返す同一飼主の全ペットがタブに並ぶ。表示は常に選択中の 1 頭分で、複数ペットを合算・マージしたビューは持たない。ペット切替のたびに当該ペットの履歴クエリが走る（React Query キャッシュ共有）。

### 2.3 印刷・共有導線
**画面閲覧専用（印刷非対応）**。印刷ボタン・印刷用スタイル・共有機能は存在せず、追加の予定もない（SD-17: gh issue #158 原文を確認した結果、スコープは「飼主横断の投薬歴一覧表示」のみで印刷要件は無いと確定。過去版の本セクションにあった「印刷で長い履歴が切れる制約」という記述は doc 執筆時の推測混入であり撤回する）。本画面は lg 以上でパネル内部スクロールの固定ビューポートを採用しており、ブラウザ標準の印刷機能を使った場合は長い履歴が切れ得るが、印刷自体が要件外のため対応しない。印刷が必要になった場合は、業務上の目的と責任者（個人名）を明記した新規要望として起票すること。

---

## 3. 臨床安全・アクセシビリティ

### 3.1 臨床安全
- **値を捏造しない**: 未設定値は "-" 表示。初診日は受診歴なし/読み込み中は "-"（推定値を出さない）。年齢のみ birthDate からの導出値（`formatPetAge`）で、レガシー EMR の併記仕様を算出で再現。
- **異常値の強調**: 検査履歴は基準値外項目（status が normal 以外）を Destructive Red で表示。健診履歴は要注意フィールドに ⚠（aria-label「要注意」）+ 赤色表示。
- **下書き・未実施の除外**: 検査は依頼中/検査中を除外、トリミングは「完了」のみ。未確定情報を実施履歴として誤読させない。
- **読み取り専用**: 本画面に入力・保存操作はなく、データ保護（離脱ガード）は不要。

### 3.2 アクセシビリティ
- ペット切替は WAI-ARIA APG「Tabs（手動アクティベーション）」準拠: roving tabindex、矢印キーはフォーカスのみ移動（データ再取得を伴う選択は Enter/Space/クリックで発火）、Home/End 対応、tab と tabpanel を aria-controls / aria-labelledby で関連付け。
- 各パネル本文は見出しで命名された region ランドマーク + tabIndex=0 で、キーボードだけで内部スクロール可能（WCAG 4.1.2）。
- ローディング表示は role=status + aria-live=polite。

---

## 4. 技術仕様

### 使用コンポーネント
- **`OwnerReport`**: ルートコンポーネント。認証・権限ガードとペット選択 state を持つ。
- **`ReportSection`**: 権限なし / ローディング / エラー / 空 の縮退表示を一元化する共通シェル。
- **`ReportPanel`** / **`HistoryTable`**: 境界内スクロールパネルと sticky ヘッダー付き履歴テーブル。
- **feature 内 API フック**: `useGetPetFirstVisit` / `useGetPetExaminations` / `useGetPetTreatmentHistory` / `useGetPetTrimmingHistory`（`features/owner-report/api`）。予防接種・健診はグローバル共有フック（`useGetPetVaccinations` / `useGetPetCheckupResults`）を流用。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/owners/:id` | 飼主基本情報の取得 | `owners` | `view` |
| GET | `/api/v1/pets?owner_id=` | 飼主に紐づく全ペット（同居ペット）一覧 | `owners` | `view` |
| GET | `/api/v1/pets/:id/first-visit` | 初診日（最古の有効カルテ date）の取得 | `medical-records` | `view` |
| GET | `/api/v1/pets/:id/treatment-history` | 投薬・麻酔・手術・治療履歴の取得（item_type / anesthesia_only / is_surgery で絞り込み） | `medical-records` | `view` |
| GET | `/api/v1/vaccinations?pet_id=` | 予防接種履歴の取得 | `vaccinations` | `view` |
| GET | `/api/v1/examinations?pet_id=` | 検査履歴の取得 | `examinations` | `view` |
| GET | `/api/v1/checkups/field-results?pet_id=` | 健診（パッケージ）結果の取得 | `checkups` | `view` |
| GET | `/api/v1/trimmings?pet_id=` | トリミング履歴の取得 | `trimming` | `view` |

clinic 隔離・論理削除除外はバックエンド（`pet_handler.go` ほか）が担保する。

---
