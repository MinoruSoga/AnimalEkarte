# トリミングの用語と関係

> 最新保持方針: 保存済み原画像・添付、参照済みkey、業務原本／必要履歴か分類不明の画像はC05の原本保持対象とし、通常削除・clear・置換・取消でblobを物理削除しない。通常操作は論理削除／非表示とし元参照・必要監査を保持する。未保存かつ未送信の放棄作業copyと区別し、分類不明は保護する。物理削除は承認された別手続へ委ね、本整理から新しい例外や保存年限を推定しない。
文書確認日2026-10-15。既存仕様と2026-10-15追加承認の完了後訂正方針を整理する。実装済みとはしない。[業務正本](../domains/trimming/lifecycle.md)・[技術契約](../../design/domains/trimming/technical-spec.md)。

| 概念 | 意味・識別 | 関係・境界 |
|:---|:---|:---|
| トリミング予約 | appointmentsのトリミング分類 | appointment IDは予約writerが所有。detail未作成でも存在可 |
| 施術記録／detail | 当該予約の施術内容 | appointment_trimming_details。通常診療medical_recordとは別 |
| 施術担当 | 業務上のトリマー。途中保存では未定可、終了までに実担当を記録 | wireのdoctor_idや担当医ラベルは現行対応。職種が医師になる意味ではなく、登録actor・前回担当名とも別 |
| 登録actor | mutationを行う認証スタッフ | 前回担当名や今回担当者とは別 |
| コース種別 | コースのカテゴリ（大分類） | コースそのもの・料金・予約区分の業務分類ではない。予約区分マスタに近い一覧／編集UIへ整理する |
| コース | 対象サイズ・所要時間・価格等を持つ基本施術。途中保存では未定可 | 未定を既定コース・料金で補完しない。終了時もコース自体は必須でなく、オプション等の既存明細で実施内容を記録可能。種別は任意。犬種価格マトリックスなし |
| オプション | 追加施術項目、組合せ可否・所要時間・価格 | 複数選択。OptionIDsの明示配列は全置換 |
| 施術ステータス | 予約／進行中／完了という表示上の進捗 | 開始後in_consultationは受付の診療中列。専用画面の終了で会計待ちへ進むが、記録・請求・入金の確定とは別 |
| 希望画像／完成画像 | object keyで参照する施術画像 | preview／signed URLは永続keyとは別 |
| 削除予約tombstone（旧実装参照） | 既存のobject key新規参照を拒否する技術台帳を指す用語 | [技術契約の観察](../../design/domains/trimming/technical-spec.md)であり、永久台帳方式/物理cleanupの新しい業務採用ではない。現在の原本保持・通常論理削除の目標を優先し、storage削除完了・clinical auditの代用にしない |
| 履歴コピー | 所定の前回内容を今回のフォームへ反映 | 担当ID・バイタル・画像・日時・状態を引き継がない |
| 未請求施術明細 | 会計側が参照して取込むコース・オプションと保存済み料金調整の明細 | 完了時の会計pushではない。元値・調整値／scope・根拠、取込済snapshot、master現在値を分け、同じ値引きを二重適用しない |
| 施術側の料金手修正／値引き | 個別施術に保存する価格変更と値引き条件 | master変更ではない。コース／オプション各行をD10明細scopeへ対応させ、会計全体scopeと区別する。新しい初期値・計算順を作らず、他診療・他ペットへ勝手に拡張しない |
| record_shortcut | 予約に紐付かない直接記録の既存経路 | 固定時刻を使わずJST現在時刻。全予約制約免除という一般許可ではない |

同じ予約には重複detailを作らない。detailなしの正常NotFoundとDB取得失敗を分ける。医院／患者／予約を一緒に識別し、同じpetの通常診療と施術を同一recordへ統合しない。masterの現在価格だけで当時の施術単価を確定できると仮定しない。

## 追加承認の語彙

- 未定の途中保存: [最新第5項合意](../domains/trimming/lifecycle.md#trimming-partial-save-undecided)により担当だけ・コースだけ・両方未定でも記録を保存し、同じ記録で入力を続けること。取得失敗／未取得とは別。終了時は[追加承認](../domains/trimming/lifecycle.md#trimming-end-recorded-performance)の実担当・記録済み施術内容を別に確認する。
- 理由付き専用訂正: 完了後の原記録を保持し、理由と変更履歴を関連づける操作。自動返金・在庫巻戻しを含まない。

## 実施なし中止と請求判断

[実施なし中止](../domains/trimming/lifecycle.md#requirement-fr-trimming-zero-performance-abort)は、開始後に施術実績がないまま中止理由を残して閉じる業務結果。予定コースの有無と実施有無を分け、通常の施術完了・開始取消・記録削除と同一視しない。請求あり／なしは今回の中止料等についてのスタッフ判断であり、既存請求・預り金・他診療費の消去／免除を意味しない。中止料等は未実施コースを施術済みとして作る明細ではない。

<a id="trimming-course-option-concept"></a>
## 予約選択・施術実績・料金の意味境界

コースは任意の基本施術、追加optionは許可された組合せから選ぶ別施術。コースカテゴリはコースでも予約の業務分類でもない。予約でコース0..1、異種追加option0..複数、同一option1回・数量なしという[予約側の多重度](reservations.md#booking-course-option-concept)と、施術保存/会計数量の契約を同一にしない。コース内包施術と追加optionの重複を認めず、コースなし用候補を区別する。現在master/予約採用時の時間/施術当時内容/会計snapshotは異なる時点の値である。予約時間と料金を自動連動させない。

<a id="trimming-style-projection-concept"></a>
## 要望と当日受付への参照

| 用語 | 定義・区別 | 追跡先 |
|:---|:---|:---|
| 予約メモ／notes | 予約の一般連絡。仕上がり要望専用情報とは別 | [FR-reception-trimming-style-request](../domains/reception/pages/reception-board.md#requirement-fr-reception-trimming-style-request) |
| 仕上がり要望 | 飼主の希望する施術結果についての情報。過去の施術内容・完成写真と同一ではない | 同FR、AC-reception-trimming-style-display |
| 希望画像／完成画像 | 前者は希望を示す原資料、後者は結果を示す画像。保存keyと表示URLは別 | [画像保護](../domains/trimming/lifecycle.md#requirement-fr-trimming-image-lifecycle) |
| 当日受付の要望表示 | 同じ予約・ペット・対象日に対応する要望の読取projection | AC-reception-trimming-style-target／AC-reception-trimming-style-empty-failure。表示だけで記録作成・notes転記・原文更新を起こさず、未記入/失敗/別個体/別日/認可外を区別 |

希望画像を参照できることと紙の記録票への画像印刷採用は別。記録票の現採用範囲を専用仕様から参照し、希望/完成写真を必ず印刷する新条件を作らない。

<a id="trimming-start-end-technical-selection"></a>
## 2026-10-23開始・終了の技術接続

[開始・終了の限定契約](../../design/domains/trimming/trimming-start-end-contract.md#trimming-start-end-closed-wire)で468単一writer、T4 schema2の実担当者（actor/予約指名と別）、閉じたstart/end fingerprintと最小準備readproofを選定。準備はmetadataのみで記録作成/本文save/状態変更なし。現在認可/版検査を維持し、同key成功はproof再検証前に扱う。途中保存/開始に実担当者・料金の新gateを加えず、終了の既決9ACを保持。共有供給/実装/試験/全9AC受入/activationは未完（NOT_RUN）。
