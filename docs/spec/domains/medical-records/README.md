# カルテ

**診断・バイタルの追加合意（2026-10-15）**：[第1・第2それぞれの複数診断、最新値＋測定時刻・同日を含む以前の測定履歴](pages/medical-record-form.md#medical-record-agreement-diagnosis-vitals-20261015) → [概念と多重度](../../ontology/medical-records.md#medical-record-diagnosis-vitals-ontology) → [技術契約／trace](../../../design/domains/medical-records/technical-spec.md#medical-record-diagnosis-vitals-contract)。2FR／12ACを追加し合意追跡は24FR／101AC（旧継承ID・C05追加1ACは別計数）。自動保存・終了後通常編集・復元保護は維持。初診判定に含める受診歴の対象は[予約／受付の別合意](../reservations/pages/reservation-management.md#visit-history-included-types)へ従う。実装／実機は未確認。

**最新の読む順**：[2026-10-15会議合意](lifecycle.md#medical-record-agreement-autosave-20261015)を先に確認する。自動保存維持・診療終了後の通常編集・履歴付きタブ開始時復元／複製直前取消・終了時警告／終了履歴を追加し、旧医師確定必須と本文ロックの目標を置換。以下の完了記録は当時の範囲で、今回の統合レビューは重大指摘なしで完了。保存済み版・会計確定歴・検査自体の確定等の独立保護は保持する。全過去の承認証跡網羅や実装完了は意味しない。

カルテ一覧・入力画面・自動保存・通常編集と履歴の契約を確認する入口です。2026-10-14初回5項目・入力4項目・訂正／例外4項目・追加3項目の合意範囲を反映。合意16FR／71ACと関連する通常カルテ・D03master・受け手技術契約の明示範囲は文書整理完了。実装・利用開始・全体横断最終確認・実機は未確認。

## 読む順・業務ルールの所在

まず [全体要件と臨床安全原則](../../specification.md) §2.1 を確認し、次に以下の既存本文を読みます。

[横断フロー](../../workflows/reservation-to-record-flow.md) §5 → [保存・確定契約](lifecycle.md#hybrid-sync)（特に §2）→ [一覧](pages/medical-record-list.md) → [詳細・入力](pages/medical-record-form.md)を参照します。合意の読む順は保存§2 → [初回5項目](lifecycle.md#medical-record-agreement-first-five) → [入力4項目](pages/medical-record-form.md#medical-record-agreement-input-four) → [訂正／例外4項目](lifecycle.md#medical-record-agreement-correction-four) → 用語／部分技術設計。入口に業務ルールの本文を重複掲載しません。

## 受入条件の所在

画面固有の挙動・バリデーション・保護・API 権限は上記本文、具体的な確認手順は [S06 記録ロックと監査](../../../ops/testing/scenarios/S06-record-lock-audit-trail.md)・[S17 処置数量の保存](../../../ops/testing/scenarios/S17-treatment-quantity-commit.md)・[V01 臨床フォーム](../../../ops/testing/scenarios/V01-clinical-forms.md)を参照します。これらは条件・手順の所在であり、今回の整理で実機受入済みになったことを示しません。実行状態と証跡の入口は [docs/work/](../../../work/README.md)です。

## 未決事項と実装状態

[初回5項目](lifecycle.md#medical-record-agreement-first-five)の業務判断と、医師による薬量値の別承認（#261）と横断フロー §7 の既知ギャップを保持します。コードのガードや Issue Closed と臨床承認を同一視しません。 新しい要望は決定前に既存の確定ルールへ混ぜず、決定状態と実装・検証状態を分けて扱います。

### 予約からの記録確認訂正の引継ぎ

[業務5点合意の記録確認訂正](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-record-reviewed-correction)では、受付担当が理由・変更案を入力し記録担当が確認して確定する。予約・記録・請求の整合確認と既存記録保持、以前の保存版・訂正履歴を保持し、通常本文編集は最新合意へ従う。患者／日付等の目的別確認と請求整合は維持する。具体的記録訂正項目・権限識別子・保存方法はカルテ側で詰める。本合意は臨床承認#261や保存ルーティング契約を変更しない。

## 共通事項

関連マスタは [問診](../../screens/settings/master-interview.md)・[処置](../../screens/settings/master-treatment.md)・[薬剤](../../screens/settings/master-medicine.md)を参照します。[共通用語・権限](../../shared/README.md)・[仕様全体の入口](../../README.md)。

## このフォルダの役割と使い方

カルテ領域の業務ルール・受入条件・未決事項を辿る入口。

pagesに画面詳細を置き、保存契約など領域固有本文は既存の配置を維持する。横断フロー・共通マスタ・語彙をここで再定義しない。

子フォルダ：[pages/](pages/README.md)

主要文書：主要文書は本書の既存索引・上記の読む順を参照します。

読む順：本書の「読む順・業務ルールの所在」 → [画面詳細](pages/README.md) → 関連する横断契約・マスタ。

正本・状態：[決定状態と検証状態の規約](../../authoring-guide.md#status)に従い、各節・FRの承認根拠と実装／実機状態を別に読む。READMEへの掲載・フォルダ配置は承認や実装済みを表さず、対象外の既存仕様を新たに確定しない。既存の別名入口・索引の責務は維持し、本書は案内を補う。

### 予約設計からの引継ぎ：記録participantとlock

docs限定承認 Sentinel_f19ab03e8c2c81919c6f589bcf563811に基づき、[受け手の返却契約](../../../design/domains/reservations/technical-spec.md#reservation-receiving-contracts)と[全writerの共通gate／下位取得順](../../../design/domains/reservations/technical-spec.md#reservation-global-lock-contract)を技術設計として選定した。本体と子入力を含むcontent_state／dependency_revision、確認済み訂正結果、以前の保存版・追記／訂正参照を返し、通常本文の現在版更新は最新合意へ接続する。作成・補完・finalize・削除・子入力・追記・関連masterを参加表へ含め、ambient participantは独立transactionを開始しない。保存契約§2・確定本文／元日付・同じIDを保持する。訂正内部の具体項目・専用記録本文の業務承認と実装受入は別条件。

## 初回合意と技術の入口

[合意5FR・受入16AC](lifecycle.md#medical-record-agreement-first-five) → [用語と関係](../../ontology/medical-records.md) → [部分技術設計](../../../design/domains/medical-records/README.md)。担当医以外の入力／編集者ログ・通常編集／必須追加なし／終了時の未完了警告・訂正／削除・過去入力・医師確認理由付き手入力の正本はlifecycle。医院の実grant／医師判定・専用操作の実データ対応と全owner実装／横断統合は未確認。全タブ文書契約は[最終照合](../../../design/domains/medical-records/technical-spec.md#medical-record-reconciled-contract)へ追従。


## 入力4項目の追加合意

この節から追加3項目までの進行中記載は各合意時点の履歴。現在の文書判定は末尾の完成記録と全体進捗表を参照。

2026-10-14、Sentinel_6287d7cbcf988191aac47fbe504cfad1 の回答を[画面正本の4FR／15AC](pages/medical-record-form.md#medical-record-agreement-input-four)へ反映。初回5FR／16ACと合わせた今回の追加合意は9FR／31ACで、既存継承FR／AC・予約69件と別に数える。選択複写・template手動挿入・診療master当時内容・項目状態とcheckbox／modalを整理。D03全体は進行中、実装／実機未確認。


## 訂正・例外4項目の追加合意

[保存契約の正本4FR／14AC](lifecycle.md#medical-record-agreement-correction-four)へ反映。初回・入力・今回の部分合意は計13FR／45AC。実施済み取消、必要内容の患者転記、当日カード／自動課金なしの過去入力、空誤作成の非表示と理由付き復元。D03全体は進行中、実装／実機未確認。


## 画像・印刷・後日結果・影響表示の追加合意

Sentinel_59d45d62afe08191bc0b289c42486962 により[3FR／12AC](pages/medical-record-form.md#medical-record-agreement-integration-three)を追加。[全タブ共通14AC](pages/medical-record-form.md#medical-record-shared-coverage-ac)も既決契約へ接続。今回合意計16FR／71AC。D&Dと選択は同じ検証、印刷は範囲選択と訂正識別、同じ結果／履歴と確定時版保持、会計／在庫影響表示とowner案内。自動公開・確定解除・自動返金／在庫処理なし。D03全体は進行中。

<a id="medical-record-next-business-questions"></a>
<a id="次の業務判断と設計受入残件"></a>
## 今回合意の判定範囲と次の技術整理

初回5・入力4・訂正例外4・追加3項目の計16FRと、横断14ACを含む71ACの範囲を整理した。既存継承FR／AC、予約69件と別に数える。今回までの合意から決まる通常カルテの業務に新しい質問は不要。採用済みD&Dや印刷範囲・後日結果・影響表示を再質問しない。将来の新形式・必須欄・飼主公開・自動課金等はこの承認に含めない。

| 区分 | 現在の状態・残件 | 次の扱い |
|:---|:---|:---|
| 今回業務・文書範囲 | 通常受付接続・入力／ログ・確定・コピー／template／master履歴・4状態・訂正／患者転記／過去入力／復元・画像／印刷／後日結果／影響表示は合意済み | [初回正本](lifecycle.md#medical-record-agreement-first-five) → [入力4](pages/medical-record-form.md#medical-record-agreement-input-four) → [訂正4](lifecycle.md#medical-record-agreement-correction-four) → [追加3](pages/medical-record-form.md#medical-record-agreement-integration-three) → [共通AC](pages/medical-record-form.md#medical-record-shared-coverage-ac) |
| 技術契約の整理 | 全9タブ＋バイタル・一覧／ヘッダー、型の意味・保存／取消／失敗／競合／結果不明・ownerごとの認可と参照を具体化 | [全タブ対応](../../../design/domains/medical-records/technical-spec.md#medical-record-full-coverage-contract)・[追加3契約](../../../design/domains/medical-records/technical-spec.md#medical-record-integration-contracts)。新API／DBの反映済みとはしない |
| D03文書の判定範囲 | 実wire／DTO差分、旧値・版・認可の未知への停止、全writer／consumer・受け手contractを最終照合A〜Fへ反映 | [完成条件と判定外](../../../design/domains/medical-records/technical-spec.md#medical-record-document-completion)。独立レビューと限定文書検査を通過し、この明示範囲を文書完了と判定。実装／実機・全体横断確認は別 |
| 他領域への引継ぎ | D07検査revision／機器受信・D08接種／D09健診、D10／11／12／13の未納・返金・締め後・金額／在庫増減、D05／06専用記録、D20／21公開 | [引継ぎ表](../../../design/domains/medical-records/technical-spec.md#文書判定範囲と関連領域への引継ぎ)。カルテ合意で他領域内部を確定しない |
| 別承認・実装受入 | 薬量の具体値・対象・単位・出典#261、C04実権限・C05保存期間／削除規約、実装・実データ／実機・並行性／性能 | 未承認／未検証のまま保持。Issue Closedや静的読取を医師の数値承認・実機PASSにしない |

最終静的照合の差分と受け手contractを[技術正本](../../../design/domains/medical-records/technical-spec.md#medical-record-reconciled-contract)・[16FR／71AC trace](../../../design/domains/medical-records/technical-spec.md#medical-record-final-trace)へ記録した。今回文書範囲は独立レビューと限定検査を通過。次はD10会計の内部文書整理を推奨する。追加業務が必要と分かった場合だけ、既決との差分をまとめて確認する。実装・API／DB変更・起票・公開は本整理の成果に含めない。

## 追加承認に基づく最終照合

Sentinel_7e6b0beed1b48191ae86bc2332800109 に基づくdocs限定更新。現API／DTOの実キー・型・未送信／null／clear、YAML差分、取得元訂正、GET初期化を含む全writer候補と監査／gate、医師と編集者、旧client／receipt／依存版、旧プラン／snapshot、空／専用intent、結果版／印刷projection、通常治療createの在庫副作用を整理。16FR／71ACを変えず、新しい業務判断を追加しない。文書の[完成判定範囲](../../../design/domains/medical-records/technical-spec.md#medical-record-document-completion)と実装／利用開始・全体統合を分離する。

## 文書完成記録

2026-10-14、追加docs承認に基づく最終照合A〜G、16FR／71AC trace、受け手contractを独立レビュー3名で確認し限定文書検査を通過。[完成範囲と判定外](../../../design/domains/medical-records/technical-spec.md#medical-record-document-completion)に限定してD03を文書完了とする。現API／DTOとの差、旧値不明停止、writer／監査／gateと医師確認、専用intent・在庫副作用を明示し、現コードの追従を承認根拠にしない。実設定・実旧データ・薬量具体値の医師承認・実装／利用開始／実機性能・他owner内部・全体横断最終確認は未確認のまま。新業務の質問は追加しない。次の推奨領域はD10会計。

## 最新差分の件数と実装検証の扱い

今回6FR／18ACを追加し、当時の16FR／71ACのIDを保持して対象期待値を置換。この6項目の時点では合意追跡22FR／89AC、診断・バイタルの追加後は冒頭の24FR／101AC。旧継承IDとC05追加1ACは別計数。最新技術は[自動保存・復元・終了](../../../design/domains/medical-records/technical-spec.md#medical-record-autosave-revert-contract)。S06等の既存ロック試験は現実装の確認先であり、旧finalized拒否を最新目標の合格条件として採用しない。実装時に最新版の通常編集・元版保持を受入へ対応付ける。
