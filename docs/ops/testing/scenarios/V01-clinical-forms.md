# V01: 臨床系フォーム検証（入力・更新・DB整合）

> **目的**: カルテ（本体・治療・バイタル・健診・予防接種・画像・追記・検査取込・見積）・検査・予防接種・定期健診・入院・トリミングの全臨床フォームについて、入力検証（必須・形式・境界）・更新の永続化・DB 整合（FK 選択肢・一意制約・不在 ID）が実機ブラウザ経由で正しく機能することを納品前に証明する。
> **所要目安**: 120分 / **深度**: 深い + **項目単位 F プロトコル**
> **フォーム数**: 18
> **項目単位**: [FIELD-LEVEL-PROTOCOL.md](FIELD-LEVEL-PROTOCOL.md) + [FORM-FIELD-INVENTORY.md](FORM-FIELD-INVENTORY.md) §V01。C1 は入口。**全 fieldKey に F0–F6 を適用**して完了とする（代表 1 項目のみでは未完了）。
> **仕様正本**: [screens/06-medical-records-form.md](../../../spec/screens/06-medical-records-form.md)・[screens/13-examinations-form.md](../../../spec/screens/13-examinations-form.md)・[screens/15-vaccinations-form.md](../../../spec/screens/15-vaccinations-form.md)・[screens/25-checkups-list.md](../../../spec/screens/25-checkups-list.md)・[screens/08-hospitalization-detail.md](../../../spec/screens/08-hospitalization-detail.md)・[screens/09-hospitalization-form.md](../../../spec/screens/09-hospitalization-form.md)・[screens/17-trimming-form.md](../../../spec/screens/17-trimming-form.md)

## 前提条件

- 環境: ローカル（seed 003_demo）。ログイン: admin ロール（カルテ/検査の create 権限・明細行の削除権限を含む全権限）。
- 対象ペット: ペット検索で「ステータス=生存」の犬または猫を 1 頭選ぶ（体重登録済みの個体 — §2 の薬量自動計算に必要）。作成するデータ・マスタは名前に「V01」を含め、終了時に削除（削除不可なら無効化）する。C3-1 用のマスタ追加は設定画面（screens/20 系）から行う。
- スコープ外: クロステナント隔離（BE isolation テスト正本）。finalized カルテの 409 ロック・audit_logs は [S06](S06-record-lock-audit-trail.md)、検査異常値ハイライト・確定ロックは [S02](S02-exam-abnormal-highlight-lock.md)、ワクチン次回予定の自動計算は [S03](S03-vaccination-next-due-autocalc.md)、入院サイクル・退院会計は [S05](S05-hospitalization-cycle.md) が正本。

## 共通チェック手順

各フォームのセクションで (C1)(C2)(C3) を参照する。加えて **F プロトコルを inventory 全項目に適用**する。フィールド・境界値は各セクションおよび [FORM-FIELD-INVENTORY.md](FORM-FIELD-INVENTORY.md) の指定に従う。

**C1 入力チェック**

| # | 操作 | 期待結果 |
|:--|:--|:--|
| C1-1 | 必須欄を空のまま保存 | 保存されず、エラーが表示される |
| C1-2 | 代表的な形式違反（型/文字数/日付/金額）を入力して保存 | 拒否され、保存されない |
| C1-3 | 境界値 1 件（各セクション指定） | 境界内は受理、境界外は拒否 |

**C2 更新チェック**

| # | 操作 | 期待結果 |
|:--|:--|:--|
| C2-1 | 既存レコードを編集して保存 | 詳細・一覧に変更が反映される |
| C2-2 | ページを再読込（F5） | 変更が永続している |
| C2-3 | 編集フォームを再オープン | 保存した値が初期表示される |

**C3 DB 存在チェック**

| # | 操作 | 期待結果 |
|:--|:--|:--|
| C3-1 | FK 選択肢の元マスタに新規レコードを追加し、フォームを開き直す | 追加したレコードが選択肢に現れる（選択肢がマスタ実データ由来） |
| C3-2 | 一意制約に重複する値で登録 | 拒否され、保存されない |
| C3-3 | 存在しない ID の URL を直叩き | 404/エラー画面が表示される（白画面・無限ロードにならない） |

**代表確認（各 1 フォームのみ — BE テストで網羅済みのため）**

- PATCH 部分更新が他フィールドを消さないこと → §1 medical-record-form でのみ確認。
- 削除（無効化）済みマスタ参照の挙動 → §12 trimming-form でのみ確認（#228 が仕様正本）。
- C3-2: 本領域のフォームに UI から入力可能な一意制約カラムはない（`medical_records(clinic_id, record_no)` は自動採番、`clinic_id+name` 一意は参照先マスタ側の制約でマスタ設定フォームのシナリオが担当）— 全セクションで該当なし。

## 手順と期待結果

§1〜§12 を順に実行する。各節の (C1)(C2)(C3) は上記の共通チェック手順を指す。

### 1. カルテ本体 — SOAP・問診・診察/治療プラン (medical-record-form)

- ルート: `/medical-records/new?petId=xxx`（新規）・`/medical-records/:id`（編集）。保存はタブ別 PATCH（問診=`/inquiries`、診察/治療プラン=`/clinical-plan`+本体）。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | (C1-1) petId なしで `/medical-records/new` を直接開く | ペット選択画面（`/medical-records/select-pet`）へリダイレクトされる（`use-medical-record-form.ts` の `shouldRedirectToSelectPet` → `MedicalRecordForm` Navigate） |
| 2 | 新規カルテで SOAP を空のまま保存し、問診タブに主訴・主訴区分・治療方針を入力して保存 → (C2) 一覧反映・再読込・再オープン | SOAP 空でも保存成功（必須ではない）。問診は C2-1〜C2-3 のとおり永続・初期表示される |
| 3 | 【代表 PATCH】診察/治療プランタブで診断第1（区分+名称）を保存後、問診タブのみ変更して保存 | 診断・SOAP・治療明細が消えずに保持される（タブ別 PATCH 部分更新） |
| 4 | (C3-1) 診断第1の区分を選択して名称の選択肢を確認 | 選択した区分に属する診断名のみが出る（マスタ実データ由来。区分不一致は BE が invalid input で拒否 — BE テスト正本） |
| 5 | ヘッダーの担当医・来院種別を変更（保存ボタンを押さない）→ 再読込 | 保存ボタンを押さずに即時 PATCH され、詳細キャッシュ invalidate 後の再読込でも永続する（[06 §2.2](../../../spec/screens/06-medical-records-form.md)） |
| 6 | 問診に入力後、保存せず一覧へ遷移 | 未保存警告（NavigationBlocker）が表示される |
| 7 | (C3-3) `/medical-records/<存在しない ID>` を直叩き | エラー画面が表示される |

### 2. カルテ 治療明細タブ — 処置・処方 (medical-record-treatments-tab)

- `/medical-records/:id` 「治療」タブ。行単位で即時 POST/PATCH（メイン保存と独立）。処方（medicine）もこのタブに統合。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | (C1-2/C1-3) 行追加で quantity に 0 → 保存、1 → 保存。別行で unit_price に -1 → 保存 | quantity 0 は拒否、1 は受理。unit_price -1 は拒否、0 円は受理 |
| 2 | (C3-1) 薬剤マスタに用量パラメータ（dose_per_kg 等）付きの「V01薬剤」を追加 → 行の項目選択肢を確認 | 追加した薬剤が選択肢に現れる |
| 3 | 「V01薬剤」を体重登録済みの犬/猫のカルテで選択 | quantity に自動計算値がプリフィルされる（#201 薬量自動計算: 種別+体重+マスタ dose params） |
| 4 | 数量を安全域外へ変更して保存 | 絶対上限超過は理由がインライン表示され、確認ダイアログによる解除経路なしで保存をブロックする。下限未満または推奨値からの大幅乖離は警告・監査対象だが保存は継続できる（[06 §2.3](../../../spec/screens/06-medical-records-form.md)） |
| 5 | (C2) 既存行の数量・単価を編集して保存 → 再読込 | 行単位で即時保存され永続する |

### 3. カルテ バイタル (medical-record-vitals)

- `/medical-records/:id`。記録フォームの実体は `VitalsModal`（[99-medical-record-flow.md](../../../spec/screens/99-medical-record-flow.md) §2 の通り）。バイタルタブ（`VitalsTab`）は記録一覧の表示とモーダル起動を担う — 実装確認済み（`frontend/src/features/medical-records/components/VitalsTab/`・`VitalsModal.tsx`）。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | (C1-1) 記録日時を空で保存 / (C1-2) 未来日時を指定して保存 | 空はエラー。未来日時は「未来の日時は入力できません」で拒否される |
| 2 | (C1-3) 体温に 45.0 → 保存、45.1 → 保存 | 45.0 は受理、45.1 は拒否（FE 範囲 30〜45℃） |
| 3 | 記録日時のみ・全計測値を空で保存 | 「体温・心拍数・呼吸数・体重のいずれかを入力してください」で拒否される |
| 4 | (C2) 既存バイタルの体重（Kg/g 切替含む）を編集して保存 → 再読込・再オープン | 永続・初期表示される（JST 日時が化けないこと） |

### 4. カルテ 定期健診タブ (medical-record-checkups-tab)

- `/medical-records/:id` 「定期健診」タブ。確定済みカルテでの 409 拒否は S06 正本のため省略。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | (C1-1) 実施日・健診種別を空のまま追加 | それぞれインラインエラーで保存されない |
| 2 | (C3-1) 健診種別マスタに動的フィールド定義付きの「V01健診」を追加 → タブの種別選択肢を確認 | 「V01健診」が現れ、選択すると定義した動的フィールド（checkup_type_fields）が表示される |
| 3 | 動的フィールド・結果・次回日を入力して保存 → (C2) 再読込・再オープン | 永続・初期表示される |

### 5. カルテ 予防接種タブ (medical-record-vaccination-tab)

- `/medical-records/:id` 「予防接種」タブ（VaccinationForm）。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | (C1-1) ワクチン未選択で追加 | 保存されない。FE インライン「ワクチン種別を選択してください」（独立画面と同じ。無音失敗ではない） |
| 2 | (C3-1) ワクチンマスタ実データ由来の選択肢から接種日・ワクチン・LOT 番号（4 つまで）を入力して保存 → (C2) 再読込 | 永続・初期表示される。左ペイン一覧は `GET /vaccinations?pet_id=` で更新され、空の EmptyState のまま残らない |
| 3 | 「次回予防接種予定設定」ラジオを切り替える。続けて補助説明を入力し、次回予定日を手入力して保存 → 独立画面から同じ接種記録を再オープン | ラジオ切替で次回日付は `calculateNextDate` により変わる（既定 type は `4weeks`）。手入力後は type が `other` になり得る。補助説明・type・次回予定日が永続する |

### 6. カルテ サブフォーム群 — 画像・追記・検査取込・見積タブ

- 対象: medical-record-image-upload / medical-record-addendum / medical-record-examination-import / medical-record-estimate-tab。前提: 追記には確定済み（finalized）カルテ 1 件（検索条件: ステータス=確定済み。無ければ S06 実行後のカルテを流用）。検査取込には対象ペットのカルテ未紐付け検査（無ければ §7 で作成した検査を使用）。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 画像: jpeg/png/gif/pdf をアップロード → 再読込。次に 10 MB 超のファイルを選択 | 対応形式は画像タブ一覧に表示され永続する（写真・PDF 資料のアップロード管理 — [06-medical-records-form.md](../../../spec/screens/06-medical-records-form.md)）。バックエンド上限は 10 MB。【要実測】フロントの選択制限と上限超過時のエラー表示 |
| 2 | 追記: 確定済みカルテで追記モーダルを開き、修正内容・修正理由を空で登録 | 「修正内容は必須です」「修正理由は必須です」で保存されない。片方だけ入力して失敗しても入力済み側の値は消えない（controlled） |
| 3 | 追記: (C1-3) 修正理由に 500 文字 → 登録、501 文字 → 登録 | 500 は受理、501 は拒否（[06 §2.3](../../../spec/screens/06-medical-records-form.md)。BE は 500 文字・rune 数判定） |
| 4 | 追記: 未確定カルテで追記導線を探す | 追記セクションごと非表示（`isMedicalRecordFinalizedStatus`）。BE も finalized 以外を拒否する |
| 5 | 検査取込: 何も選択せずに取込を実行 | 取込ボタンは `selectedIds.size === 0` で disabled |
| 6 | 検査取込: 対象ペットの既存検査を選択して取込 → 再読込 | 検査タブに紐付き表示され永続する（検査ごとに PATCH で medical_record_id 設定） |
| 7 | 見積タブ: 件名と明細を入力してメイン保存 → 再読込 | 件名・明細行・金額が同一 tx で永続する（必須欄なし。postSave 経由で POST/PATCH /estimates が `items` を送る）。独立画面（/estimates）はヘッダ金額のみで本シナリオ対象外 |

### 7. 検査入力/結果登録 (examination-form)

- ルート: `/examinations/new?petId=xxx`（新規・create 権限必須）・`/examinations/:id`（編集）。異常値ハイライトの網羅・確定ロックは S02 正本。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | (C1-1) 検査種別・担当医を空で保存 | それぞれエラーが表示され保存されない（両者は保存時の必須項目 — [13-examinations-form.md](../../../spec/screens/13-examinations-form.md)） |
| 2 | (C3-1) 検査種別マスタに項目定義付きの「V01検査」を追加 → 新規フォームの種別選択肢を確認 | 選択肢に現れ、選択すると定義した検査項目（exam_type_fields）が測定値入力テーブルに表示される |
| 3 | 測定値（数値）を入力して保存 → (C2) 再読込・再オープン | 永続・初期表示される。基準値マスタ範囲外の値は H/L 判定（高値=赤・低値=status-blue）が再読込後も表示される（判定は BE 導出） |
| 4 | (C3-3) `/examinations/<存在しない ID>` 直叩き | エラー画面が表示される |
| 5 | `examination-unconfirm:edit` があるアカウントで確定済み検査を開く / 無いアカウントで開く | あるときだけ確定解除ダイアログ。無い医院の default-deny では出ない |

### 8. 予防接種 独立画面 (vaccination-form)

- ルート: `/vaccinations/new?petId=xxx`（新規）・`/vaccinations/:id`（編集）。次回予定日の自動算出（3週/4週/1年）・再計算は S03 正本。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | (C1-1) ワクチン種別・接種日を空で保存 | 「ワクチン種別を選択してください」「接種日を入力してください」で保存されない |
| 2 | (C1-2) 接種日に明日（未来日）を指定して保存 | 「接種日は今日以前の日付を入力してください」で拒否される |
| 3 | (C1-3) 次回予定日に接種日と同日 → 保存、接種日の翌日以降 → 保存 | 同日は拒否（次回予定日は接種日より後 — 実装ガード: `use-vaccination-form.ts` のコメント参照）、翌日以降は受理（新規時は本日以降も必須 — 同ファイルの実装ガード） |
| 4 | LOT 番号を 4 つ入力して保存 → (C2) 一覧反映・再読込・再オープン | 4 つとも永続・初期表示される（最大 4 並行登録）。(C3-1) ワクチン選択肢は vaccines マスタ由来のマスタ有効行（固定 2 択ではない） |
| 5 | (C3-3) `/vaccinations/<存在しない ID>` 直叩き | エラー画面が表示される |
| 6 | ヘッダーの動物種表示を確認 | `formatPatientPetDetails` により実データ（例: 犬）が先頭に出る。「不明」は年齢・性別・去勢避妊の欠損時だけ |

### 9. 定期健診 独立登録（クイック登録）(checkup-form)

- ルート: `/checkups/new?petId=xxx`（新規のみ）。登録時にカルテを自動生成してから checkup をサブリソース登録する 2 段階処理。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | (C1-1) 健診種別・実施日を空で保存 | 「健診種別を選択してください」「実施日を入力してください」で保存されない |
| 2 | (C1-2/日付) 実施日の DatePicker で明日を選択しようとする | 未来日は選択不可（disabledDays・JST 基準の物理ブロック） |
| 3 | (C3-1) 種別に「V01健診」（§4 で作成）+ 当日で登録し、編集導線を探す | 登録成功。実施日を診察日とするカルテが自動生成され、そのカルテの定期健診タブに本記録が表示される。独立編集ルートは無い。一覧の編集は `/medical-records/:id?tab=定期健診&checkupId=` へ遷移する |
| 4 | ヘッダーの動物種表示を確認 | `formatPatientPetDetails` で種が先頭。欠損の年齢・性別・去勢避妊だけ「不明」 |

### 10. 入院/ホテル 登録・編集 (hospitalization-form)

- ルート: `/hospitalization/new?petId=xxx`（新規）・`/hospitalization/:id/edit`（編集）。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | (C1-1) ペット未選択で保存 | 「ペットを選択してください」で保存されず、エラーフィールドへ自動フォーカスされる（useActionState） |
| 2 | 入院タイプ・入院予定日を空のまま（ペットのみ選択）で保存 | 初期値はタイプ「入院」、開始は当日。空ケージは必須（BUG-037）。BE create は `hospitalization_type` / `start_date` required |
| 3 | (C3-1) ケージマスタに「V01ケージ」・治療プラン項目名マスタに「V01プラン」を追加 → フォームの選択肢確認 | それぞれ選択肢に現れる（ケージは空き状況フィルタなし = 使用中でも表示される） |
| 4 | 主訴・担当医・保険適用 ON + 保険会社名/保険番号、治療プラン明細を入力して保存 | 主訴は `owner_request`、担当医はヘッダー `doctor_id`。一覧の主訴・担当医列に載る。治療プランは登録後読み取り専用 |
| 5 | (C2) 既存入院の日付・メモ・明細を編集して保存 → 一覧反映・再読込・再オープン | 永続・初期表示される |
| 6 | 検索条件「ステータス=退院済み」の入院を開いて編集保存（無ければ S05 側で確認しスキップ可） | 通常 PATCH は discharged を理由に拒否しない（退院**遷移時**だけ監査）。再保存は通る |
| 7 | (C3-3) `/hospitalization/<存在しない ID>/edit` 直叩き | エラー画面が表示される |
| 8 | ヘッダーの動物種表示を確認 | `formatPatientPetDetails` で種が先頭 |

### 11. 入院詳細 インラインフォーム群（ケアプラン・デイリー記録）

- 対象: hospitalization-care-plan / hospitalization-daily-vitals / hospitalization-daily-care-logs / hospitalization-daily-staff-notes（`/hospitalization/:id` の各タブ）。共通手順を 4 フォームに適用する。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | (C1-1) 下表の必須フィールドを空にする | 保存されない。ケアプランは trim 空で追加/保存 disabled。デイリーバイタル/申し送りは空時刻で submit no-op |
| 2 | 必須+任意フィールドを入力して追加 → (C2) 一覧反映・再読込 | 永続する。時刻は `HH:mm` 入力が `HH:mm:00` で保存される |
| 3 | （デイリー 3 フォームのみ）同一日にバイタル・ケア記録・申し送りを続けて追加 | エラーなく同一日の記録欄にまとまる（daily_records は入院×日付で親 1 件 — `(hospitalization_id, date)` 一意） |

フォーム別差分表（共通手順は上表を参照）:

| フォーム (id) | ルート | 必須フィールド | 一意制約 | 特記チェック |
|:--|:--|:--|:--|:--|
| ケアプラン項目 (hospitalization-care-plan) | `/hospitalization/:id` ケアプランタブ | name（trim 後非空） | なし | type Select・timing（朝/昼/夜等トグル）。編集は EditRow でインライン — 編集後の永続も確認 |
| デイリーバイタル (hospitalization-daily-vitals) | `/hospitalization/:id` デイリー記録タブ | time | 親 daily_records(hospitalization_id, date) | 体温 50 は範囲チェックなしで保存される（カルテバイタルの 30〜45 とは別） |
| デイリーケア記録 (hospitalization-daily-care-logs) | `/hospitalization/:id` デイリー記録タブ | time・type（既定値あり） | 親 daily_records(hospitalization_id, date) | value/notes は任意 |
| スタッフ申し送り (hospitalization-daily-staff-notes) | `/hospitalization/:id` デイリー記録タブ | time・content（trim 後非空） | 親 daily_records(hospitalization_id, date) | 記録者スタッフが表示に紐づく |

### 12. トリミング登録・編集 (trimming-form)

- ルート: `/trimming/new?petId=xxx`（新規）・`/trimming/:id`（編集）。前提: トリミング予約区分（reservation_types）が設定済みであること（未設定環境では新規登録が行えずエラーが表示される — システム前提）。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | (C1-1) 担当者・コース未選択で保存 | 「担当者を選択してください」「コースを選択してください」のインラインエラー+トーストで保存されない（実装ガード: `use-trimming-form-validation.ts` 参照） |
| 2 | (C3-1) トリミングコースマスタに「V01コース」（価格付き）・オプションマスタに「V01オプション」を追加 → フォームの選択肢確認 | いずれも選択肢に現れ、コースには価格が表示される |
| 3 | 新規登録: 登録時ステータス「予約」（#233）+ コース + オプション複数 + スタイル要望 + 体重（Kg/g 切替）+ 画像 2 種（希望スタイル/仕上がり、image/*）を入力して保存 | 予約として登録される（ステータス選択は新規のみ表示で編集画面には出ない）。(C2) 一覧反映・再読込・再オープンで永続・初期表示される |
| 4 | 【代表・無効化マスタ】手順 3 のレコードに紐づく「V01コース」をマスタで無効化 → (a) 新規フォーム (b) 当該レコードの編集を開く | (a) 選択肢から除外される。(b) 「（無効）」表記で選択が維持される（#228） |
| 5 | (C3-3) `/trimming/<存在しない ID>` 直叩き | エラー画面が表示される |
| 6 | 同一担当スタッフで同日に別ペットの record_shortcut 新規を続けて登録 | 固定 10:00 ではなく一意な JST 時刻が付く。`uk_appointment_staff_time` の 409 で無関係な 2 件目がブロックされない |

## 確認観点

- 既存の機械テストとの分担: FE component/hook test（use-medical-record-form・TreatmentsTab/dose gate・CheckupsTab・use-examination-form・use-vaccination-form・use-hospitalization-form・use-trimming-form-validation 等）と BE service/validator test（treatment/dose_validators・vital・checkup・vaccination・clinical_plan・hospitalization・trimming 各 service）が単体レベルの入力検証を網羅済み。**E2E は全臨床フローで表示・遷移スモークのみで、フォーム送信〜永続化を検証する E2E は存在しない — 本シナリオはブラウザ → API → DB を貫く受け入れ時の実機フォーム検証である。**
- FE 側にしかガードが無い（または有無が未確定の）項目（バイタルの体温範囲・未来日時、健診/ワクチンの未来日、トリミング必須）はブラウザ経由確認が必須。逆に治療明細の quantity/price・追記 500 字・カルテタブのワクチン必須は BE 拒否が期待線で、FE 側のエラー提示の有無を観察する。
- クロステナント隔離はスコープ外（BE isolation テスト正本）。finalized ロック・監査証跡は S06、異常値ハイライトは S02、次回予定自動計算は S03、入院サイクルは S05 へ委譲。NG 項目は [`todo.md` 受入バグ](../../../../todo.md) へ `### BUG-XXX` 節として起票する（ローカル連番 最大+1・[README.md](README.md) のルール）。
- 本文中の具体的なエラー文言・プリフィル挙動のうち画面仕様書に記載のないもの（薬量プリフィル+dose gate・追記/健診/トリミングの必須文言・入院差分表の必須断定・トリミングの予約区分前提）は、FE hook / BE バリデータの実装を確認済み（2026-07-16）。文言はリファクタで変わりうるため、合格基準は「該当エラーが表示され保存されないこと」であり文言の完全一致ではない。

## 異常系（臨床安全に直結する独立確認）

手順表に埋め込まれた C1/C3 異常系のうち、臨床安全に直結する以下は単体でも必ず実施する（索引）:

| 参照 | 内容 |
|:--|:--|
| §2 治療明細 | 薬量 dose hard gate — 自動計算値と乖離した手入力が確認ステップなしで保存されないこと |
| §6 追記 | 確定後カルテへの変更が追記（addendum）経由のみで、修正内容・理由が必須であること |
| §8 ワクチン（独立画面） | 未来日接種の拒否・次回予定日の境界（接種日と同日は拒否） |
| §11 入院デイリー | `(hospitalization_id, date)` 一意 — 同一日の重複親レコードが作られないこと |
| 全フォーム共通 | (C3-3) 存在しない ID の URL 直叩きがエラー画面になり、空フォームとして開かないこと |

## 実装突合
- 突合日: 2026-08-07
- HEAD: 844e43f69
- 変更:
  - 異常系索引の節番号を現行構成に修正（ワクチンは §8、§10 は入院登録のため誤参照だった）
  - §1 petId なし新規の select-pet リダイレクトを実装突合済みに昇格（要実測を解除）
  - 臨床ルート（`/medical-records`・`/examinations`・`/vaccinations`・`/checkups`・`/hospitalization`・`/trimming` および `select-pet`/`new`/`:id`/`edit`）を `frontend/src/config/paths.ts` と一致確認
