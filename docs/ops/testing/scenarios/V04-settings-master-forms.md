# V04: 設定マスタ系フォーム検証（入力・更新・DB整合）

> **目的**: `/settings` 配下のマスタ・設定系フォーム（マスタ SidePanel 群・締め時間・シフトパターン・Lステップ連携・LINE 予約ページ設定）について、入力検証（必須・形式・境界）・更新の永続化・DB 整合（FK 選択肢・一意制約）が実機ブラウザ経由で正しく機能することを納品前に証明する。
> **所要目安**: 120分 / **深度**: フォーム検証（V 系固有の深度 — S 系テンプレの薄い/深い二値の拡張）
> **仕様正本**: [screens/settings/ 配下各文書](../../../spec/screens/settings/README.md)（差分表・各セクションに個別文書を明記）・[screens/31-lstep-integration.md](../../../spec/screens/31-lstep-integration.md)

## 前提条件

- 環境: ローカル（seed 003_demo）。ログイン: admin ロール（マスタ管理・会計・支払方法・Lステップ設定を含む全権限）。
- 本シナリオで作成するデータは名前に「V04」を含め、終了時に削除（削除不可なら無効化）する。
- スコープ外: **クロステナント隔離（BE isolation テストが正本）**。スタッフ・権限グループ・医院マスタは [V03](V03-owner-pet-staff-forms.md)、認証・LINE 連携操作系は [V05](V05-auth-line-forms.md) の対象。マスタ削除時の FK 保護（参照中削除の拒否等）は既存シナリオ・BE テスト済み前提で対象外。
- 共通アーキテクチャ: /settings 配下マスタは MasterCRUDPage/MasterTabPage + SidePanel（新規/編集兼用）+ useMasterSave（FE 必須チェック）+ D&D 並び順 + isActive トグルの共通パターン。SidePanel 起動式で個別詳細 URL を持たないため、**C3-3（ID 直叩き）は §5 予約可能枠を除き全フォーム該当なし**。

## 共通チェック手順

各セクション・差分表の行で (C1)(C2)(C3) を参照する。フィールド・境界値・一意制約は各所の指定に従う。

**C1 入力チェック**

| # | 操作 | 期待結果 |
|:--|:--|:--|
| C1-1 | 必須欄を空のまま保存 | 保存されず、エラーが表示される |
| C1-2 | 代表的な形式違反（型/文字数/日付/金額）を入力して保存 | 拒否され、保存されない |
| C1-3 | 境界値 1 件（各所指定） | 境界内は受理、境界外は拒否 |

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

**代表確認（ドメイン内 各 1 フォームのみ — BE テストで網羅済みのため）**

- PATCH 部分更新が他フィールドを消さないこと → §3 薬剤マスタでのみ確認。
- 削除済みマスタを参照するレコードの挙動 → §1 トリミングコース（コース種別 FK）でのみ確認。

**一意制約の共通形**: ほぼ全マスタが `(clinic_id, name)` UNIQUE（削除済み行を除く部分 Index）。重複時は FE 事前チェックなしで BE が UNIQUE 違反を返すため、C3-2 では「エラーが表示され保存されない（無音失敗・白画面にならない）」ことを確認する。

## 1. 標準マスタ SidePanel 群（フォーム別差分表）

実行方法: 各行で **C1-1（必須欄空→エラー）→ 新規「V04〇〇」作成 → C2-1〜C2-3 → 一意制約列が「—」以外なら C3-2** を実施し、特記チェック列の指示を追加実行する。画面正本は [screens/settings/](../../../spec/screens/settings/README.md) 配下の各文書 — 同名文書がある形（例: `master-cage.md`）のほか、診断カテゴリ/病名→`master-diagnosis.md`、問診テンプレート→`master-interview.md`、物販→`master-merchandise.md`、キャンペーン→`master-campaigns.md`、支払方法→`payment-methods.md`、トリミングコース/オプション→`master-trimming.md`、予約区分グループ→`master-reservation-type.md`。

| フォーム (id) | ルート | 必須 (C1-1) | 一意制約 (C3-2) | 特記チェック |
|:--|:--|:--|:--|:--|
| 動物種類 (master-animal-species) | /settings/animal-species | 動物種類名 | name **グローバル一意**（clinic_id なし・WHERE is_active=true） | 無効化すると一意から外れ同名を再登録できる（部分 Index）。D&D 並び順が再読込後も永続 |
| 診断カテゴリ (master-diagnosis-type) | /settings/diagnosis?tab=diagnosis_type | name | (clinic_id,name) | 配下に診断病名が 1 件でもあるカテゴリの削除は「この診断カテゴリには診断名が登録されているため削除できません」の競合エラーで拒否される（service 層ガード。DDL の ON DELETE CASCADE は soft delete のため発火しない）。空カテゴリのみ削除可 |
| 診断病名 (master-diagnosis-name) | /settings/diagnosis?tab=diagnosis_name | name・所属カテゴリ | —（migration 上一意制約なし） | (C3-1) 「V04カテゴリ」追加→カテゴリ選択肢に反映。カテゴリ未選択で保存→「カテゴリを選択してください」エラー（C1-1 の所属カテゴリ分） |
| 主訴種別 (master-chief-complaint) | /settings/interview/chief-complaint | name | (clinic_id,name) | — |
| 問診・定型文テンプレート (master-interview-template) | /settings/inquiry-templates | title・category | —（一意制約なし） | content（本文）は任意 — 空のまま保存できる |
| 予約区分グループ (master-reservation-type-group) | /settings/reservation-type | name | —（一意制約なし） | カラーはピッカー+テキスト両入力 — どちらで入れても保存値が一致。グループ未所属の区分は「未分類」集約表示 |
| 入院・宿泊プラン (master-hospitalization-plan) | /settings/hospitalization | name | (clinic_id,name) | 体格 bodySize・課金単位 billingUnit は空許容 — 未選択のまま保存できる |
| ケージ (master-cage) | /settings/cage | ケージ名 | (clinic_id,name) | 種別 icu/dog/cat/general・サイズ small/medium/large は選択式（BE enum 検証あり — 不正値 400） |
| 物販・商品 (master-merchandise-item) | /settings/merchandise-items | 品目名 | (clinic_id,name) **WHERE is_active=true** | 無効化すると同名を再登録できる（is_active 条件付き一意）。税率 8% の選択が永続する |
| 保険 (master-insurance) | /settings/insurance | name | (clinic_id,name) | (C1-3) 補償率 100→受理、101→拒否（BE 0〜100 範囲検証） |
| 職種 (master-occupation) | /settings/occupations | name | (clinic_id,name) | 追加した「V04職種」が §4 予約区分の職種セクション選択肢に反映される（C3-1） |
| トリミングコース (master-trimming-course) | /settings/trimming?tab=course | name | (clinic_id,name) | (C3-1) 「V04種別」追加→コース種別選択肢に反映。**【代表・削除済みマスタ】**下記注記参照 |
| トリミングオプション (master-trimming-option) | /settings/trimming?tab=option | name | (clinic_id,name) | 併用可 combinable トグル（既定 ON）を OFF にして保存→再オープンで保持 |
| トリミングコース種別 (master-trimming-course-type) | /settings/trimming-course-type | name | (clinic_id,name) | name+isActive のみの最小構成（空文字保存不可 — master-trimming-course-type.md） |
| 割引キャンペーン (master-campaign) | /settings/campaigns | キャンペーン名 | —（name 一意なし） | (C3-1) 対象商品選択肢が物販マスタ実データ由来。【要実測】終了日 < 開始日の拒否（doc は開始日以降必須と規定・FE validate は name のみで実装位置未確認） |
| 支払方法 (master-payment-method) | /settings/payment-methods | name | (clinic_id,name) | 【要実測】システム既定行（現金等 system_key 保持行）の名称変更・無効化の可否（ADR-003 関連・FE に system_key 参照なし） |

- 【代表・削除済みマスタ】: 「V04コース」にコース種別「V04種別」を設定して保存 → コース種別マスタから「V04種別」を削除 → コース編集を再オープン。【要実測】削除済み種別の表示挙動（保持表示か空欄か）と、そのまま保存してもエラーにならないか。
- 割引キャンペーンの権限は ResourceAccounting（会計）、支払方法は ResourcePaymentMethod — admin 以外で実行する場合は権限に注意。

## 2. 診療項目マスタ 5 タブ (master-treatment-item)

- ルート: `/settings/treatment-items?tab=consultation|examination|procedure|vaccine|checkup`。正本: [master-treatment.md](../../../spec/screens/settings/master-treatment.md)（検査タブは [master-examinations.md](../../../spec/screens/settings/master-examinations.md)）。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | (C1-1) 診察タブで名称空のまま保存 | 保存されずエラー（5 タブ同一 validate のため代表 1 タブで実施し、他 4 タブは「V04」項目の新規保存が通ることのみ確認） |
| 2 | (C1-2) 価格に負値 −100 を入力して保存 | 拒否される（BE 非負価格検証 — validators_test.go） |
| 3 | (C2) 処置タブの「V04処置」の価格・説明を変更 | C2-1〜C2-3 のとおり反映・永続・初期表示 |
| 4 | (C3-2) 同一タブ内で同名登録 / 別タブに同名登録 | 同一タブ内は拒否（タブごとに別テーブルで (clinic_id,name) 一意）。別タブへの同名は受理される |
| 5 | 親子階層: 処置タブで「V04処置」を親に指定して子項目を作成 | 親セレクタは同タブ実データ由来（C3-1）。一覧にツリーで親子表示される |
| 6 | 子を持つ root 行「V04処置」を編集 | 親変更セレクタが非表示（親変更不可 — E2E master-crud.spec.ts と同根拠） |
| 7 | 定期健診タブで新規保存 | 保存され一覧に反映される（checkup-types API 経由 — 他タブと API が異なるため個別確認） |

## 3. 薬剤マスタ+投与量パラメータ (master-medicine) — 代表 PATCH

- ルート: `/settings/medicine`。正本: [master-medicine.md](../../../spec/screens/settings/master-medicine.md)。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | (C1-1) 名称空のまま保存 | 保存されずエラー |
| 2 | 新規「V04薬剤」: 剤形・単位・価格・製品含量・服用回数/日・既定日数を入力して保存 | 保存され、(C2) 再読込・再オープンで全フィールドが初期表示される |
| 3 | 【代表 PATCH】価格のみ変更して保存 → 再オープン | 剤形・含量・服用回数・説明・計算方式（#201）が消えずに保持される |
| 4 | (C1-2) 製品含量 strength に数値以外（`abc`）を入力して保存 | 【要実測】拒否またはクリアされる（文字列入力→送信時 parse のため黙殺されないかを確認） |
| 5 | 投与量パラメータ: 上限（mg/kg・mg とも）未入力で追加 | 拒否される（過量防止のため上限いずれか必須 — validateDoseParamForm） |
| 6 | (C1-3) 下限 > 上限で追加 / 下限 = 上限で追加 | 下限 > 上限は拒否、下限 = 上限は受理（下限≦上限検証） |
| 7 | (C3-2) 同一薬剤×同一対象種でパラメータ 2 件目を追加 | 拒否される（uq_dose_params_med_species）。薬剤名自体の同名登録も拒否される（(clinic_id,name)） |

- パラメータ変更時の既存処方の再検証（dose_revalidation）は BE テスト正本のため対象外。

## 4. 予約区分 (master-reservation-type)

- ルート: `/settings/reservation-type`。正本: [master-reservation-type.md](../../../spec/screens/settings/master-reservation-type.md)。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | (C1-1) 名称空のまま保存 | 保存されずエラー |
| 2 | 所要時間に数値以外を入力して保存 | 15 分にフォールバックして保存される（FE 仕様 — 拒否ではない点に注意） |
| 3 | (C3-1) グループ選択肢を確認 | §1 で作成した「V04グループ」が選択肢に現れる |
| 4 | (C2) 短縮名・LINE 表示・院内専用・LINE 表示名・コメントを設定して保存 | 再読込・再オープンで全フィールド保持。【要実測】LINE 表示名を空欄にした場合に区分名称が代用されるか（仕様文書・FE 実装とも根拠未確認） |
| 5 | 職種セクション: §1 の「V04職種」を担当職種に追加 | 選択肢は職種マスタ実データ由来（C3-1）。同一職種の重複追加は拒否される（(reservation_type_id,occupation_id) 一意） |
| 6 | (C3-2) 同名の予約区分を新規登録 | 拒否される |

- パネル内の予約可能枠セクションは §5 と同一コンポーネントのためどちらか一方で確認すればよい。BE validator（availability/staff capability）はテスト済み。

## 5. 予約可能枠設定 (reservation-type-available-slots)

- ルート: `/line-reservation/slots`（+ §4 パネル内セクション）。画面固有の spec 文書なし — BE validator とコンポーネントテストが正本。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 対象区分（リーフ）を選択し weekly モードで曜日×時刻スロットを追加 | スロットが表示され、再読込後も永続する |
| 2 | 同一曜日×同一時刻を重複追加 | 拒否される（(reservation_type_id,day_of_week,start_time) 一意） |
| 3 | specific モードで特定日×時刻を追加/重複追加 | 追加は永続し、重複は拒否される（specific_date 側 partial index） |
| 4 | スロットを削除して再読込 | 削除が永続する |
| 5 | 区分セレクタを確認 | リーフ区分のみ選択可（親区分は選択肢に出ない）。選択肢は予約区分マスタ実データ由来（C3-1） |
| 6 | (C3-3 相当) `?typeId=` に無効な値を付けて直叩き | フォールバックで既定区分が表示され、白画面・無限ロードにならない |

## 6. 締め時間設定 3 フォーム (closing-standard-time / closing-holiday / closing-special-period)

- ルート: `/settings/closing-time`。正本: [closing-time-settings.md](../../../spec/screens/settings/closing-time-settings.md)。締め境界の業務的検証は [S09](S09-closing-time-boundaries.md) が正本 — 本節はフォームの入力・永続のみ。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 標準時刻: AM/PM 境界・平日終了・日曜終了を変更して保存 | 保存され、AM/PM/EMG の派生レンジ表示が連動更新される（EMG は翌 am_start までの越日表示 — #215）。(C2) 再読込で永続 |
| 2 | 標準時刻: (C1-1) 時刻欄を空にして保存 | 保存されずエラー（3 欄とも必須） |
| 3 | 標準時刻: 境界逆転（平日終了 < AM/PM 境界）で保存 | 【要実測】BE 挙動未確認 — 拒否されるか、保存されてレンジ表示が破綻しないか |
| 4 | 休診日: (C1-1) 日付空のまま追加 | 追加されない（date 必須。理由は任意） |
| 5 | 休診日: 日付+理由で追加 → (C2) 再読込 | 一覧に反映され永続する |
| 6 | 休診日: (C3-2) 同一日付を再登録 | 拒否される（uk_clinic_holidays_clinic_date UNIQUE） |
| 7 | 特別期間: 期間+AM/PM 境界+終了時刻で追加 → (C2) 再読込 | 一覧に反映され永続する |
| 8 | 特別期間: (C1-3) 開始日 > 終了日 / 境界 >= 終了時刻で追加 | 拒否される（DB CHECK: start_date<=end_date・am_pm_boundary<pm_end）。【要実測】エラー応答の形（500 系画面でなくエラー表示で収まるか） |
| 9 | 特別期間: 既存と重複する期間を追加 | 【要実測】一意制約なしのため受理される可能性 — 受理時の締め計算への影響確認は S09 の領域 |

## 7. シフトパターン (master-shift-template)

- ルート: `/settings/shift-templates`。正本: [master-shift-template.md](../../../spec/screens/settings/master-shift-template.md)。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | (C1-1) 名称空のまま保存 | 保存されずエラー |
| 2 | 勤務種別に「休み(off)」「有給(paid_leave)」を選択 | 時刻入力が非表示になり、時刻なしで保存できる（isShiftTemplateTimeHidden） |
| 3 | 勤務系の種別で開始/終了時刻を空にして保存 | 拒否される（勤務系は時刻必須） |
| 4 | 休憩 2 件（開始/終了）を追加して保存 → (C2) 再オープン | 休憩複数件が保持・初期表示される |
| 5 | 勤務時間外の休憩（勤務 9:00–18:00 に休憩 19:00–20:00）を保存 | 【要実測】検証位置未確認 — 拒否か受理かを実測する |
| 6 | (C3-2) 同名テンプレートを登録 | 拒否される（uk_shift_templates_clinic_name） |

## 8. Lステップ連携 4 フォーム (lstep-settings / lstep-tag-config / lstep-tag-code-mappings / lstep-trigger-priority)

- ルート: `/settings/integrations/lstep`。正本: [31-lstep-integration.md](../../../spec/screens/31-lstep-integration.md)。ローカルでは実配信・実 API 疎通は発生しない前提で、フォームの保存・永続のみ確認する。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 連携設定: シークレット 3 種（API キー・チャネルトークン・チャネルシークレット）を空のまま他項目を変更して保存 | 保存成功し、シークレットは既存値維持（空=未変更扱いのマスク運用） |
| 2 | 連携設定: LIFF ID を空にして保存 | クリアされる（空文字でクリアできる唯一の項目） |
| 3 | 連携設定: 数値項目（休眠予防閾値等）に 0 / 負値を入力して保存 → 再読込 | 送信されず既存値のまま表示される（parseInt>=1 のみ送信 — **黙って無視される仕様**であることの確認。DB 側にも CHECK >=1） |
| 4 | 連携設定: (C1-3) 数値項目に 1 を入力して保存 → 再読込 | 受理され永続する |
| 5 | タグ設定: 自動管理プレフィックス/条件タグ対応/送信目的プレフィックスの 3 フォームに各 1 件追加・行削除 | 追加・削除とも一覧に反映される |
| 6 | タグ設定: (C3-2) 既存と同一キー（prefix / conditionCode / purpose）で追加 | 拒否される（各列 UNIQUE — **グローバル・clinic 無関係**のため既存行全体と衝突し得る） |
| 7 | タグ設定: (C1-2) 条件コードに 51 文字を入力して追加 | 拒否される（VARCHAR(50)。prefix/purpose/タグ系は 100・category は 20） |
| 8 | コードマッピング: タグ 1 件を編集モードにし entries を追加して保存 | 「{タグ名} を保存しました」トーストが表示され、(C2) 再読込で永続 |
| 9 | コードマッピング: entries を全削除して保存 | 【要実測】置換型 PUT のため空置換の挙動（全削除されるか、拒否されるか） |
| 10 | 配信優先順位: 並び替え → 保存 → 再読込 | 順序が永続する（入力フィールドなしの順序永続化フォーム） |
| 11 | 配信優先順位: 保存直後に続けて並び替え → 保存 | 2 回目の保存も正しく反映される（保存後 baseline 更新の回帰観点） |

## 9. LINE 予約ページ設定 (line-reservation-page-editor)

- ルート: `/line-reservation/page-editor`（/settings 外）。正本: [28-line-reservation.md §2](../../../spec/screens/28-line-reservation.md)（旧 master-pages.md は 2026-07-16 に統合・削除）。本棚卸しでは深掘り未実施のため、実行時に文書と突合して確認する（LINE 予約系シナリオと重複していれば重複分はスキップ可）。

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 【要実測】28-line-reservation.md §2 と突合しながら C1（必須項目）・C2（保存 → 再読込永続）を実施 | 必須項目・保存動線は文書正本に従う（本棚卸しでは必須項目・一意制約とも未確認） |
| 2 | (C3-1) 表示対象の予約区分選択肢を確認 | 予約区分マスタ実データ由来（§4 の「V04」区分が反映される） |

## 確認観点

- 既存の機械テストとの分担: 共通フック単体（use-master-save / use-master-crud）、E2E settings-crud.spec.ts（動物種 CRUD+検索・薬剤新規保存・診断病名パネル表示）、master-crud.spec.ts（主訴ナビ・診療項目の親子階層と 5 タブ — arm64 では skip）、settings-smoke.spec.ts（全設定ページの表示）、component test（予約区分パネル・予約可能枠 3 本・締め 3 セクション・Lステップ 4 セクション・ケージ・薬剤 model 2 本）、BE validators_test.go（RequiredName/TaxType/NonNegativePrice/CageType/CageSize/CoverageRate）+ dose / availability / staff capability 各 validator テストが単体レベルを網羅済み。**本シナリオはブラウザ → API → DB を通した受け入れ時の実機フォーム検証**であり、特に機械テスト未カバーの「一意制約違反時のエラー表示」「更新の永続化」「FK 選択肢のマスタ由来」を対象とする。
- 重複登録は FE 事前チェックなしで BE の UNIQUE 違反頼み — 全マスタ共通で「無音失敗・白画面にならない」ことが最重点の確認事項。
- animal_species と Lステップタグ 3 テーブルは clinic 無関係のグローバル一意 — 変更が他クリニックにも見える点に注意（それ以外の clinic_id 隔離検証はスコープ外 — BE isolation テスト正本）。
- NG 項目は todo.md「バグ台帳」節へ BUG-XXX として起票する（[README.md](README.md) のルールに従う）。
