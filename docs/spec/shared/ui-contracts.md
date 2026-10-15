# 共通UI・選択部品の契約（C01）

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | shared.ui-contracts／共通契約の整理候補 |
| 決定状態 | 既存記載の継承。新業務承認を含まない。根拠不明は未確認のまま。未回答案は別節 |
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
- [Input](../components/input.md)／[Select](../components/select.md)はinvalidの文字説明とlabel接続を保持。Selectの選択肢取得中制御は親の責務。空文字value禁止。検索式選択は適切な検索部品へ。
- [Table](../components/table.md)の行全体onClickは禁止。明示した行内actionを使い、ヘッダ意味論・空／読込表示と既存例外を保つ。
- [Tabs](../components/tabs.md)は同一データの切替。別ページ遷移にはナビを使い、disabled理由を補う。

<a id="requirement-fr-common-ui-context"></a>
### FR-common-ui-context — 患者文脈・離脱・共有範囲

PatientInfoCard利用フォームはtrimming/checkups/examinations/hospitalization/vaccinations。カルテは専用MedicalRecordStickyHeader＋PatientContextHeaderで、同一部品利用済みとは書かない。欠損年齢・性別・去勢避妊を不明として示し、speciesがあれば実値を出す。現状onOwnerClickの利用画面なし、StaffSelectionModalはカルテ専用、PropertyFilterのdeferは利用側の責務という境界を保つ。

NavigationBlockerはdirty時のSPA内遷移、併用useUnsavedChangesはブラウザ離脱を担当する。ガードの存在は全writerの永続保存済みを意味しない。OwnerSearchModalの汎用変更確認と、OwnerForm側の会員区分／値引差の警告を混同しない。飼主の付替えを許可する対象・金銭影響・医院scopeはD03/D04/C04の正本へ渡す。

<a id="requirement-fr-common-ui-pet-selection"></a>
### FR-common-ui-pet-selection — 選択と作成の分離

対象はmedical-records/hospitalization/trimming/examinations/checkups/accounting/vaccinations。検索はsearch・owner_id=owners.id・動物種マスタID、20件サーバページング・300ms debounceという既存実装参照を保持し、住所検索は追加しない。通常は利用機能create、checkupsはmedical-recordsのcreateかつeditという既存例外を維持し、C04で再照合する。

死亡個体は死亡表示付きで選択できるが新規作成を許可しない。カルテselectorは当該ペットの一覧へ渡す。未知生死statusは選択fail-closed。死亡選択可を全フォームでの臨床操作可へ広げない。

特記はペットdanger_levelと飼主is_dangerousを別概念として扱う。ペット高／中は形状アイコンとPopover、メモ空は内容未登録。飼主は代替名「特記」の形状アイコンでPopoverなし。スタッフ向けで、分類が分かる文言を画面・aria-label・titleへ出さず、owner/LIFFへ出さない。一般の非色cue規則を理由に機微な分類語を追加しない。

## 受入条件（実行結果ではない）

| 条件ID | Given／When | Then／独立確認・成功条件 | 確認手順／根拠 |
|:---|:---|:---|:---|
| AC-common-ui-focus | 各interactive部品をキーボードで巡回 | focusが見え、disabled操作が発火しない | state＋各primitiveの状態表／UI確認は未実施 |
| AC-common-ui-pending | 保存ボタンのloading中に再押下 | ラベルとbusy表示を保ち同ボタン操作は抑止 | Button／サーバ冪等性は各D |
| AC-common-ui-readable | readonly／invalid／死亡を表示 | 読取り・拒否・sentinelを区別でき、色だけに依存しない | 状態正本／Input実装ギャップは残す |
| AC-common-ui-dialog | 開く→閉じる／ESC／右上操作 | focus復元、タイトル識別、X重なりなし。業務取消を勝手に実行しない | Dialogと利用Dの取消契約を両方確認 |
| AC-common-ui-selection | 死亡／未知statusを患者検索 | 死亡選択と作成拒否、未知選択拒否を区別 | 患者選択の遷移先とC03死亡ゲートを照合 |
| AC-common-ui-special-mark | ペット高／中・飼主特記を表示 | 正本の形状・代替情報・Popover範囲を保持しownerへ漏らさない | 患者選択の特記契約 |
| AC-common-ui-dirty | dirtyでSPA遷移とリロード | それぞれのガード経路を確認。取消を保存成功としない | NavigationBlocker／useUnsavedChanges |
| AC-common-ui-scope | shared候補と別appを分類 | 実consumer・非自明契約・媒体境界を説明でき、全shared MD化・88二重計上なし | design-system／責務map |

## 既知差分・未確認

Input readonly視覚、Dialog mobile fullscreen、Select dense size、Table sticky/横scroll統一、Tabs mobile scrollableは既存部品の既知ギャップ。これは選定仕様がないことと同義ではなく、実装成立を未確認に保つ。stories全件保証の旧総括に対しtoken層§6はshared未カバー49件allowlistを明記している。allowlistを全件実装カバー済みに数えず、現在の実数は再実行未確認。原資料にある測定値は当時の履歴で、今回のAA実測PASSではない。

通信失敗・競合時の対象再取得、選択途中のscope変更や遅延応答の破棄、保存後の戻り先は利用D側の契約を必要とする。共通部品だけで一律の再送／上書き方式を作らない。各consumerの明示契約がない箇所は仕様不足として残り、C01全体完成判定へ転用しない。

## 全ownerへの適用照合

[全23領域の具体的な意味ID・保存・安全・保持・失敗条件との対応](owner-contract-coverage.md)を参照する。原則の再掲だけで個別操作の受入を満たしたとは扱わず、引用先のFR／AC・技術契約と照合する。
