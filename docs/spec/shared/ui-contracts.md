# 共通UI・選択部品の契約（C01）

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | shared.ui-contracts／共通契約の整理候補 |
| 決定状態 | 2026-10-17追加決定は末尾と進捗入口を参照。従来：既存記載の継承と追加の死亡後専用導線の承認を区別。死亡後の生前分処理はC03の個別承認に従い、その他の根拠未照合は未確認のまま |
| 実装状態 | 既存資料の記載を継承（コード全体の再監査なし）。選定仕様と既存挙動の差を明記 |
| 実機検証状態 | 未実施 |
| 文書確認日・基準 | 2026-10-15／PR #670 commit 41fcf3eb0a8befb7c0b5f613d1df749a60e30be8 |
| 対象 | 本体画面UI、7 primitive、非自明な共有部品、7機能の患者選択 |

本書は既存正本への対応と受入観点を束ねる。数値・語彙・業務条件の第二正本は作らず、リンク先の有効な個別契約を優先する。FR/ACの識別子付与は追加承認や実装済みを表さない。未確認の不足を対象外へ移して完了としない。

## 責務・読む順

[色・適用範囲](../design-system.md) → [寸法等](../../../DESIGN.md) → [状態](../design-states.md) → [トークン層](../design-token-layers.md) → [7部品](../components/README.md) → [共通ダイアログ](../screens/common-dialogs.md)／[患者選択](../screens/00-pet-selection.md)。その他sharedの見た目はcolocated stories/autodocsが正本で、全sharedを重複したMarkdownへ書き起こさない。MDは臨床sentinel・権限・router/query/print依存など非自明な契約に限定する。

印刷帳票、ManualContent本文、LIFF／line-reserve／shared-liffは既存design-system §2.6の画面意匠適用外。各領域の安全・認可・保存契約まで除外するものではない。本体88ルートと別アプリを混算しない。各select-petルートは利用領域へ配賦し、共通部品だけをC01で数える。

## 継承契約

<a id="requirement-fr-common-ui-state"></a>
### FR-common-ui-state — 状態と可読性

[状態正本](../design-states.md)のdefault／hover／active／focus-visible／disabled／readonly／loading／invalid／selectedを維持。キーボードフォーカスを不可視にせず、loadingはラベルを保持してスピナー・aria-busy・操作抑止を併用。readonlyはdisabledと意味が違い、読める値を維持する。sentinelを色だけで示さない。44pxタッチ目標とdense例外は正本の条件を使い、全画面を無条件に縮小しない。

色値はdesign-system、寸法・余白等はDESIGN、状態はdesign-states、配置方向はdesign-token-layers。PALETTE→C／STATE→STYLE／BADGE等の責務を維持し、Tailwindの静的文字列制約を理由にしたL2の値重複を欠陥としない。raw色を使うcanvas等の明示例外は保持する。

<a id="requirement-fr-common-ui-primitives"></a>
### FR-common-ui-primitives — 部品の操作境界

- [Badge](../components/badge.md)は識別用で、クリック操作はButton等へ。非色cueを維持。
- [Button](../components/button.md)は既定type=button、loading中に操作を起動せず、アイコン単独には名前を付ける。
- [Dialog](../components/dialog.md)はタイトル・説明・focus復元・ESC等の契約を保持。確認必須操作はAlertDialog。X占有領域を入力や操作ボタンへ重ねない。閉じる操作そのものに業務取消・保存・削除の意味を与えない。
- [Input](../components/input.md)／[Select](../components/select.md)等の必須入力には必須マークを示し、未入力・検証エラーは赤枠と文字メッセージで対象欄を識別できるようにする。Backlog Phase1 No.86 F87:G87、2026-10-15曽我回答が根拠。既存の任意項目を必須化せず、色だけに頼らない。invalidの文字説明とlabel接続を保持。Selectの選択肢取得中制御は親の責務。空文字value禁止。検索式選択は適切な検索部品へ。
- [Table](../components/table.md)の行全体onClickは禁止。明示した行内actionを使い、ヘッダ意味論・空／読込表示と既存例外を保つ。
- [Tabs](../components/tabs.md)は同一データの切替。別ページ遷移にはナビを使い、disabled理由を補う。

<a id="requirement-fr-common-ui-context"></a>
### FR-common-ui-context — 患者文脈・離脱・共有範囲

PatientInfoCard利用フォームはtrimming/checkups/examinations/hospitalization/vaccinations。カルテは専用MedicalRecordStickyHeader＋PatientContextHeaderで、同一部品利用済みとは書かない。欠損年齢・性別・去勢避妊を不明として示し、speciesがあれば実値を出す。現状onOwnerClickの利用画面なし、StaffSelectionModalはカルテ専用、PropertyFilterのdeferは利用側の責務という境界を保つ。

NavigationBlockerはdirty時のSPA内遷移、併用useUnsavedChangesはブラウザ離脱を担当する。ガードの存在は全writerの永続保存済みを意味しない。OwnerSearchModalの汎用変更確認と、OwnerForm側の会員区分／値引差の警告を混同しない。飼主の付替えを許可する対象・金銭影響・医院scopeはD03/D04/C04の正本へ渡す。

左メニューの各項目には画面の役割・説明を用意し、マウスホバーとキーボードフォーカスのどちらからも確認できる。[Backlog Phase1 F74:G74（No.73）](https://docs.google.com/spreadsheets/d/17jaFBv7NmbPzk88lZx94vFxa_G1shGNX2O_lFGt4QdQ/edit#gid=0&range=F74:G74)、2026-10-15曽我回答が根拠。説明を読むだけで画面遷移や保存を起動せず、既存のフォーカス可視性・メニュー表示権限を維持する。対応完了の原文報告を今回の実機合格にしない。

<a id="requirement-fr-common-ui-pet-selection"></a>
### FR-common-ui-pet-selection — 選択と作成の分離

予約作成・各カルテ作成時の患者選択は、ペット名・飼主名・よみ・電話番号を扱う単一の横断検索欄へ統一する。飼主の姓と名は空白の有無を吸収するが、ペット名と飼主名の2語は空白で区切り、両者の連結入力を対象にしない。予防接種の飼主検索でも姓名間の空白を吸収する。各一覧の検索述語の完全統一はPhase2で、既存の画面固有検索・認可・ページングを一律変更しない。根拠はBacklog Phase1 No.74・77・81（F75:G75、F78:G78、F82:G82）の2026-10-15回答。

対象はmedical-records/hospitalization/trimming/examinations/checkups/accounting/vaccinations。検索はsearch・owner_id=owners.id・動物種マスタID、20件サーバページング・300ms debounceという既存実装参照を保持し、住所検索は追加しない。通常は利用機能create、checkupsはmedical-recordsのcreateかつeditという既存例外を維持し、C04で再照合する。

死亡個体は死亡表示付きで選択できるが、新規未来予約・死亡後の新しい臨床実施は許可しない。生前診療の後日入力・誤記訂正・退院・精算は[C03の追加承認](clinical-safety.md#requirement-fr-common-safety-patient)に従い、死亡状態を維持した理由／履歴付きの専用経路として区別する。カルテselectorは当該ペットの一覧へ渡す。未知生死statusは選択fail-closed。死亡選択可を全フォームでの無制限な臨床操作可へ広げない。

特記はペットdanger_levelと飼主is_dangerousを別概念として扱う。ペット高／中は形状アイコンとPopover、メモ空は内容未登録。飼主は代替名「特記」の形状アイコンでPopoverなし。スタッフ向けで、分類が分かる文言を画面・aria-label・titleへ出さず、owner/LIFFへ出さない。一般の非色cue規則を理由に機微な分類語を追加しない。

## 受入条件（実行結果ではない）

| 条件ID | Given／When | Then／独立確認・成功条件 | 確認手順／根拠 |
|:---|:---|:---|:---|
| AC-common-ui-menu-purpose | 左メニューの各項目へホバー／キーボードフォーカス | 両経路で画面の役割を読め、説明表示だけで遷移・保存しない | Phase1 No.73、FR-common-ui-context。未実施 |
| AC-common-ui-patient-search | 予約／各カルテ作成の患者選択と予防接種の飼主検索で検索する | 患者選択は単一欄でペット名・飼主名・よみ・電話を扱い、飼主姓名の空白有無を吸収。ペット名＋飼主名は空白区切りを用い、各一覧の完全統一へ拡大しない | FR-common-ui-pet-selection、No.74・77・81。未実施 |
| AC-common-ui-required-error | 必須欄と任意欄のあるフォームで必須未入力／形式不正を検証する | 必須マーク・対象欄の赤枠・文字メッセージを示す。任意欄の新規必須化なし | FR-common-ui-primitives、No.86。未実施 |
| AC-common-ui-focus | 各interactive部品をキーボードで巡回 | focusが見え、disabled操作が発火しない | state＋各primitiveの状態表／UI確認は未実施 |
| AC-common-ui-pending | 保存ボタンのloading中に再押下 | ラベルとbusy表示を保ち同ボタン操作は抑止 | Button／サーバ冪等性は各D |
| AC-common-ui-readable | readonly／invalid／死亡を表示 | 読取り・拒否・sentinelを区別でき、色だけに依存しない | 状態正本／Input実装ギャップは残す |
| AC-common-ui-dialog | 開く→閉じる／ESC／右上操作 | focus復元、タイトル識別、X重なりなし。業務取消を勝手に実行しない | Dialogと利用Dの取消契約を両方確認 |
| AC-common-ui-selection | 死亡／未知statusを患者検索 | 死亡選択、通常の新規予約／実施拒否、理由／履歴付き生前分の専用処理、未知選択拒否を区別 | 患者選択の遷移先とC03死亡ゲートを照合 |
| AC-common-ui-special-mark | ペット高／中・飼主特記を表示 | 正本の形状・代替情報・Popover範囲を保持しownerへ漏らさない | 患者選択の特記契約 |
| AC-common-ui-dirty | dirtyでSPA遷移とリロード | それぞれのガード経路を確認。取消を保存成功としない | NavigationBlocker／useUnsavedChanges |
| AC-common-ui-scope | shared候補と別appを分類 | 実consumer・非自明契約・媒体境界を説明でき、全shared MD化・88二重計上なし | design-system／責務map |

## 既知差分・未確認

Input readonly視覚、Dialog mobile fullscreen、Select dense size、Table sticky/横scroll統一、Tabs mobile scrollableは既存部品の既知ギャップ。これは選定仕様がないことと同義ではなく、実装成立を未確認に保つ。stories全件保証の旧総括に対しtoken層§6はshared未カバー49件allowlistを明記している。allowlistを全件実装カバー済みに数えず、現在の実数は再実行未確認。原資料にある測定値は当時の履歴で、今回のAA実測PASSではない。

通信失敗・競合時の対象再取得、選択途中のscope変更や遅延応答の破棄、保存後の戻り先は利用D側の契約を必要とする。共通部品だけで一律の再送／上書き方式を作らない。各consumerの明示契約がない箇所は仕様不足として残り、C01全体完成判定へ転用しない。

## 全ownerへの適用照合

[全23領域の具体的な意味ID・保存・安全・保持・失敗条件との対応](owner-contract-coverage.md)を参照する。原則の再掲だけで個別操作の受入を満たしたとは扱わず、引用先のFR／AC・技術契約と照合する。

## 2026-10-15入力クリアと取消の区別

通常カルテの入力クリアは[D03](../domains/medical-records/lifecycle.md#medical-record-agreement-autosave-20261015)に従い現在タブを開いた時点へ、履歴複製の取消は複製直前へ戻す。自動保存が続くため「未保存入力だけ破棄」の共通閉じる操作と同一視しない。他タブ・他者版・金銭／在庫等の副作用を一括取消しない。自動保存・復元・診療終了の成功／失敗／結果不明を別表示し、記録保存成功を終了成功へ代用しない。

<a id="list-filter-contract-20261017"></a>
## 一覧初期値・復帰条件・ペット行検索（2026-10-17）

[決定根拠・適用範囲](../domains/SPECIFICATION_PROGRESS.md#confirmed-business-decisions-20261017)：ユーザーの2026-10-16〜17 JSTの直接決定、文書更新承認は10月17日03:20 JST。以下は目標仕様・未実行の受入条件であり、実装・実機合格を示さない。

<a id="requirement-fr-common-ui-list-filter"></a>
### FR-common-ui-list-filter — 開き方と一覧の責務

飼主・ペット一覧とカルテ履歴の医院初期値は閲覧権限内全医院。選択を次回保持し、権限喪失医院は保存済み選択からも除外する。

| 対象・操作 | 初期表示・保持 |
|:---|:---|
| 通常カルテ／トリミング／検査／予防接種／健診／通常会計一覧をメニューから開く | 当日。日付軸は順に診療日／施術日／検査実施日／接種日／健診実施日／会計予定日。一律作成日に置換しない |
| 一覧→詳細→戻る | 日付・ペット等の検索条件を保持。メニューから開き直す操作と区別 |
| 行の「このペットで絞り込む」 | 対象ペットIDで全期間（日付解除）。医院条件を維持し、ペット絞込は解除可能 |
| 飼主ペット・マスタ一覧 | 当日初期化の対象外 |
| 入院・ホテル一覧 | 当日初期化の対象外。現在入院中・滞在中を確認する既存表示保持 |
| 見積一覧 | 当日初期化の対象外。期限・承認状態の既存表示保持 |
| 受付ボード | 選択日の受付表示保持 |
| 未納一覧 | 繰越を含む既存未納確認表示保持 |

全期間化で状態・担当・動物種・期限等を無断解除せず、終了済み履歴条件・取消削除除外も維持する。同名別個体・同飼主別ペットを混ぜない。ページングを維持し、表示ページだけの絞込・固定取得上限で全期間0件や検索完了を断定しない。過去参照・既存期間指定は可。予約／シフト週月カレンダー・締め集計・個別レポート／カルテ内履歴の責務は維持する。

### 受入条件

| 条件ID | 前提／操作 | 期待結果 | 確認手順／証跡 |
|:---|:---|:---|:---|
| AC-common-ui-list-clinic-memory | 初回→医院選択→次回→権限喪失 | 初回全権限医院、選択保持、認可外除外 | 両対象一覧と履歴で確認、未実施 |
| AC-common-ui-list-entry-return | 対象6一覧をメニュー起動／詳細復帰 | 当日の日付軸／既検索条件保持を区別 | 各一覧で確認、未実施 |
| AC-common-ui-list-pet-all-time | 別ページの過去対象行もある | ペットID全期間・医院維持、ページ外対象も取得 | 同名・別頭・解除も確認、未実施 |
| AC-common-ui-list-exceptions | 入院ホテル・見積・受付・未納・台帳マスタ | 上表の例外・既存責務保持 | 各一覧で確認、未実施 |

## Backlog原条文の補足受入条件

保存済み原Q/Aと既存業務契約の照合による補足。対応済みという旧回答は実装・実機検証の合格証明に用いない。後続の本人合意、医院scope、権限、履歴・会計訂正の既存条件を維持する。

### FR-common-ui-pet-selection

| 条件ID | 受入条件 | 根拠 | 検証状態 |
|:---|:---|:---|:---|
| AC-common-ui-deceased-history-navigation | 死亡登録済みペットを死亡表示／グレーで区別し、通常カルテselectorで選ぶと該当ペットのカルテ履歴へ進む。閲覧だけで新規記録や生死状態変更を起動せず、生前分後日入力は既存の理由・履歴付き専用経路へ分ける。 | Phase1 F43:G43 | 未実施 |


<a id="enter-input-unintended-save-20261017"></a>
## 入力中のEnterによる誤保存を防ぐ（2026-10-17本人確認）

入力中のEnterを、本人が意図していないフォーム送信・保存確定へつなげない。全入力欄に一律「Enterを2回押す」操作を要求する方針にはしない。対象画面・入力欄、IMEが原因かどうか、既に実装修正済みかは未特定であり、特定画面で発生・再現・解消したという事実を補わない。

原要望は2026-09-15 16:56:45 JSTの城東PC（SLK-291、ts1789459005.162369）の「エンターを一回押すと入力確定してしまうのでエンター2回で確定にしてほしい」、19:30:30 JSTの曽我返信は「こちら修正対応を行います」。本人Sentinel_fd16cb27f1b4819192fee859e0ed3d7b「保存されてしまうということでは？」を受け、「入力中のEnterで誤保存しないこと」と整理し全入力欄に一律Enter2回を要求しない方針を確認、2026-10-17 21:52:59 UTCの本人Sentinel_245605c234d08191913aba96a2c85a4b「そうですね」で承認された。旧文言から全アプリのEnter回数仕様やIME原因を推測する解釈は、この本人確認に置き換える。

[通常カルテの既決自動保存](../domains/medical-records/lifecycle.md#medical-record-agreement-autosave-20261015)は維持する。入力変更に伴う正規の自動保存と、意図しないフォーム送信・明示保存の発火を区別し、保存通信があるという理由だけで誤保存と判定しない。複数行入力の改行、候補の選択・確定、既存のキーボード操作を一律禁止しない。保存ボタン等へフォーカスした明示的な実行まで禁止する意味にせず、入力部品内の操作とフォームの保存確定を区別する。

対象画面・入力欄と、変換中／変換確定後・候補選択中等の条件、保存操作の種類、イベントの伝達・制御、具体的な操作回数は実態調査・設計で照合する。今回の承認だけで「1回目は必ずIME変換、2回目は必ず保存」と確定しない。臨床記録・会計等の既存保存契約、認可・競合検査・状態保護は維持し、実装や実機合格とは分ける。

| 条件ID | 前提／操作 | 期待結果 | 検証状態 |
|:---|:---|:---|:---|
| AC-common-ui-enter-no-unintended-save | 対象画面・入力欄を調査で特定し、入力中のEnter操作とフォーム送信／保存確定の発火を確認する | 意図しないフォーム送信・保存確定を起こさず、一律Enter2回を強制しない。対象未特定の現在を再現確認済み・修正済みとはしない | 対象未特定・未実施 |
| AC-common-ui-enter-keeps-input-operations | 複数行入力、候補選択、保存操作をそれぞれ区別してEnterの役割を確認する | 改行や候補確定を一律禁止せず、入力操作を誤って保存確定へ流さない。IME原因や回数の具体動作は実態確認前に固定しない | 対象未特定・未実施 |
| AC-common-ui-enter-autosave-boundary | 通常カルテの既決自動保存がある入力変更と、Enterによるフォーム送信の経路を比較する | 正規の自動保存を維持し、意図しない送信／保存確定と区別する。自動保存停止や全保存通信禁止をこの承認の合格条件にしない | 対象未特定・未実施 |
