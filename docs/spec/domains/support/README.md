# D23 サポート

2026-10-17整理：日本語言い換え・ベトナム語対応はEMR-294として10/6 main統合済みで、採否未決へ戻さない。これはLibrary確定事項の引継ぎであり、今回のコード適合再監査・配備／設定／実サービス品質の合格を意味しない。

> 2026-10-15の仕様整理候補。PR #670の基準版から既存合意・既存仕様・未承認案を分離した。業務合意と実装／実機検証を別管理する。D22/D23は2026-10-15の一括承認を反映。技術設計案は業務承認と分ける。

## 目的・範囲

P2：不具合共有と支援履歴の参照。AI回答は参考で、臨床・金銭・権限変更を自動確定しない（安全側の範囲明示案）。

## 正本と優先順

- [docs/spec/screens/settings/README.md](../../screens/settings/README.md)
- [SECURITY.md](../../../../SECURITY.md)

最新の明示合意が適用範囲内の旧記述に優先する。コードは実装事実の根拠で、業務承認の代替ではない。既存FR／ACの本文は元正本に維持し、本書の接続要件で改番・複製しない。

## 業務要件

<a id="requirement-fr-support-shared-read"></a>
### FR-support-shared-read — 共有閲覧と個人履歴

認証済みスタッフは専用resource grantなしで、個人・臨床情報を含まないバグ報告と全医院・全スタッフのROOMボードを共有閲覧できる。患者・飼主を識別できる情報や臨床情報を含む未redactedの報告・chat・添付は報告元医院内に限定し、他医院へ返さない。医院横断で共有する必要がある場合は、共有前に明示的なredaction確認を行い、承認されたredacted内容だけを公開する。ROOMは同一医院×スタッフの30分超の空白による表示単位。widgetの個人history取得・削除は認証医院×本人の別経路であり、全chatを個人限定に変えない。

決定状態・根拠：既存製品決定・既存画面契約継承。上記の該当既存正本に基づく整理で、新たな承認を捏造しない。

<a id="requirement-fr-support-report-mutation"></a>
### FR-support-report-mutation — 報告元医院の操作

バグ報告作成はwidget、一覧は閲覧・状態変更・明示Plane起票・削除の入口。状態変更・起票・削除は報告元医院のみ。別医院からの操作を拒否し、削除は明示確認する。未redactedの患者・飼主・臨床情報は報告元医院だけに返し、医院横断の共有READには明示的なredaction確認済み内容だけを含める。この共有境界を臨床writeや他医院mutationへ拡張しない。

決定状態・根拠：既存製品決定継承。上記の該当既存正本に基づく整理で、新たな承認を捏造しない。

<a id="requirement-fr-support-external-send"></a>
### FR-support-external-send — 送信と失敗の境界

報告の送信前に、Discord・Planeへ渡す内容と送信先を表示して確認する。送信後は院内報告を残し、外部通知・起票が失敗しても報告自体を消さない。外部で作成済みか不明なら確認してから再送し、無条件に起票を重複させない。既存の共有内容・無期限保持とLLMのbest-effort screeningを継承する。秘密値を説明例や報告へ貼らない。本合意は製品仕様であり、実際の通知・起票を今回実行する許可ではない。

決定状態・根拠：合意済み：2026-10-15 ユーザー Sentinel_7a67fd30e47c819192cd36d416f0a2bf の「ok」。確認文 Sentinel_75a3a1bf81b481919ba4c34aa633feb4／Sentinel_9c0eb79acfd481919738187e9769f8cd と対象確認 Sentinel_8d0032ca45b881919b3d2ceb5339f792。上記の該当既存正本に基づく整理で、新たな承認を捏造しない。

## 受入条件

全条件は文書上の期待結果。実機未実施。技術方式の提案は実装設計として区別する。

| 条件ID | Given（前提） | When（操作） | Then（期待結果） | 独立確認 | 成功条件 | 確認手順／証跡 |
|:---|:---|:---|:---|:---|:---|:---|
| AC-support-01 | A院の報告とB院の認証スタッフ | 共有一覧・ROOMを見る | 個人・臨床情報を含まない内容だけを共有READできる。未redactedの患者・飼主・臨床情報はA院内に限定し、redaction確認済みの共有用内容だけを他院へ返す。個人historyとは別の範囲 | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-support-shared-read](#requirement-fr-support-shared-read)。実機未実施 |
| AC-support-02 | B院がA院報告を参照している | 状態変更・起票・削除APIを直接呼ぶ | 拒否し報告は不変 | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-support-report-mutation](#requirement-fr-support-report-mutation)。実機未実施 |
| AC-support-03 | 報告医院の利用者 | 削除確認を取りやめる | 報告を残す | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-support-report-mutation](#requirement-fr-support-report-mutation)。実機未実施 |
| AC-support-04 | 送信先と内容を確認できる報告フォーム | 送信を確定する | 承認された送信契約だけを実行し、確認前に外部へ送らない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-support-external-send](#requirement-fr-support-external-send)。実機未実施 |
| AC-support-05 | Plane未設定または外部失敗 | 明示起票する | 未設定／失敗を示し成功リンクを捏造しない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-support-external-send](#requirement-fr-support-external-send)。実機未実施 |
| AC-support-06 | A院本人履歴とB院別人履歴がある | widgetの本人履歴を削除 | A院×本人だけを削除しB院・他人を消さない。削除対象の共有ROOM元データにも反映する | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-support-shared-read](#requirement-fr-support-shared-read)。実機未実施 |
| AC-support-07 | 外部起票済みか応答だけでは不明 | 再送を試す | 作成済みの有無を照合し無条件に再起票しない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-support-external-send](#requirement-fr-support-external-send)。実機未実施 |
| AC-support-08 | 他院の報告を共有一覧で読む | 詳細と添付を開く | redaction確認済みの共有用内容だけを返し、未redactedの患者・飼主・臨床情報を含む原報告・添付は他院へ返さない。報告元以外のmutation権限は得ない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-support-report-mutation](#requirement-fr-support-report-mutation)。実機未実施 |

## 未決・責務境界

外部送信先／内容の確認、院内報告保持、外部結果不明時の照合は今回合意済み。現行のDiscord保存前送信と、未確認の自動起票は目標との差分として設計へ残す。個人history削除は医院×本人の元メッセージsoft deleteにより共有ROOMにも反映する。外部コピーの削除は約束せず、全chat個人限定化や全医院履歴削除へ拡張しない。

## レビュー状態

基準版の本文と関連する静的根拠を照合。独立レビュー・統合リンク検査は統合担当で実施する。仕様記述を実装完了、受入PASS、実環境変更承認へ変換しない。

## 操作・状態・例外の全域点検

| surface | read | write/delete・状態境界 |
|:---|:---|:---|
| bug report一覧・詳細／添付 | 非機微・redaction確認済みの共有用内容だけを全医院の認証スタッフへ共有。未redactedの患者・飼主・臨床情報を含む原内容は報告元医院内 | 作成はwidget。status／Plane起票／削除は報告医院のみ。削除確認取消は不変 |
| ROOMボード | 個人・臨床情報を含まない会話だけを全医院・全スタッフへ共有 | 閲覧専用。30分超空白で表示ROOM分割、永続session IDと混同しない |
| chat exchanges | 共有の質問回答ペア | ROOM画面の取得元とは別 |
| widget history | 現在医院×本人 | 取得／削除。元メッセージsoft deleteが共有READにも反映。全院削除不可 |
| widget chat質問 | 認証スタッフの質問・回答 | rate limitあり。回答失敗／制限を回答成功にしない。臨床・会計・権限writeの自動実行機能を新設しない |
| 外部Discord／Plane | 承認された送信先と内容のみ | 今回合意した送信前確認へ従う。ローカル保存と外部送信結果を別扱いし不明結果を照合 |

Loading／Error／Emptyを区別し、共有ROOMに選択IDがない場合の先頭選択と空一覧を区別する。削除・更新中の二重実行を抑止し、他医院mutationはUI非表示だけでなくサーバーで拒否する。

現行実装はPlane completedから院内resolvedへの一方向同期を持ち、Plane cancelledはresolvedに倒さず、院内open/resolved変更をPlaneへ反映しない。これは今回静的確認した実装事実で、新たに全外部同期を業務承認したという意味ではない。既存の一方向同期の実装参照として保持し、新たな双方向同期を追加しない。

添付削除の現行静的契約：スクリーンショットobject削除に失敗した場合は502で報告行を残す。storage削除が成功しDB側削除に失敗した場合は部分結果になり得るため、添付の有無と報告状態を再照会する。報告削除からPlane／Discordの外部コピー削除を推測しない。外部通知がDB保存前に行われる現実装は今回の院内保存保持方針と区別し、DB保存失敗時に外部だけ残るリスクを技術残差へ保持する。


## 保存・例外の追加受入条件

| 条件ID | Given（前提） | When（操作） | Then（期待結果） | 独立確認 | 成功条件 | 確認手順／証跡 |
|:---|:---|:---|:---|:---|:---|:---|
| AC-support-09 | 院内報告保存が失敗 | 報告を送信 | 外部送信を行わず院内保存成功を表示しない | 単独確認 | Thenの結果と失敗時の不変 | [選定配送・保持契約](../../../design/domains/support/technical-spec.md#support-selected-delivery-contract)。実機未実施 |
| AC-support-10 | 院内報告commit済み・外部通知失敗 | 配送結果を表示 | 報告・添付参照・監査を保持し外部失敗だけを表示 | 単独確認 | Thenの結果と失敗時の不変 | [選定配送・保持契約](../../../design/domains/support/technical-spec.md#support-selected-delivery-contract)。実機未実施 |
| AC-support-11 | 外部成功後にローカル結果／監査保存が失敗 | 再送要求を受ける | 元operationと外部結果を照合し新規送信しない。照合不能なら保留 | 単独確認 | Thenの結果と失敗時の不変 | [選定配送・保持契約](../../../design/domains/support/technical-spec.md#support-selected-delivery-contract)。実機未実施 |
| AC-support-12 | 同operationの並行要求／同ID別内容 | 送信または再送 | 同内容は同じ報告・receipt、別内容は拒否。宛先別claimで二重配送を抑止 | 単独確認 | Thenの結果と失敗時の不変 | [選定配送・保持契約](../../../design/domains/support/technical-spec.md#support-selected-delivery-contract)。実機未実施 |
| AC-support-13 | 保存済み添付・必要履歴または原本分類不明 | 報告を通常削除 | 論理削除／非表示のみ。blob・元参照・必要監査・再送対応を保持 | 単独確認 | Thenの結果と失敗時の不変 | [選定配送・保持契約](../../../design/domains/support/technical-spec.md#support-selected-delivery-contract)。実機未実施 |
| AC-support-14 | 旧報告に外部IDがなく送信有無不明 | 外部再送する | ID欠落を未送信とみなさず照合不能の間は停止 | 単独確認 | Thenの結果と失敗時の不変 | [選定配送・保持契約](../../../design/domains/support/technical-spec.md#support-selected-delivery-contract)。実機未実施 |

通常削除の目標は[共通C05](../../shared/retention-audit.md)と[選定配送・保持契約](../../../design/domains/support/technical-spec.md#support-selected-delivery-contract)に従う。保存済み添付・必要履歴・原本分類不明は通常操作でblobを物理削除せず、非表示／論理削除と元参照・必要監査保持とする。上記storage先行物理削除は現行事実として残し、最新目標への適合を示さない。院内保存前の外部送信は禁止し、外部成功後のclaim／log失敗は元operation照合・不明保留で扱い、新送信で救済しない。

## 保存済みSlack本文との補足受入条件

既存の支援widget・報告一覧を対象とする。SLK-340・344の使い方質問／バグ報告の入口、SLK-353の報告履歴と対応状態の確認を、既存画面と認可へ結ぶ。日本語言い換え・ベトナム語入力はSLK-342の要求とEMR-294の10/6 main統合済み範囲であり、採否未決へ戻さない。下表はその文書上の受入定義で、今回のコード適合・配備・応答品質の実測合格ではない。回答の言語や正答率等、未確認の数値条件を新設しない。

| 条件ID | 前提／操作 | 期待結果 | 根拠・検証状態 |
|:---|:---|:---|:---|
| AC-support-widget-purpose-tabs | 認証済みスタッフが既存支援widgetを開き、使い方質問とバグ報告の各入口を選ぶ | 質問と報告作成を区別して利用できる。質問だけで報告を起票せず、報告だけで臨床・金銭操作を実行しない | SLK-340・344、既存widget契約。未実施 |
| AC-support-japanese-paraphrase | マニュアルに該当記事がある同じ意図を、日本語の異なる言い回しで質問する | 近い内容の説明・関連項目から該当する使い方を確認できる。未記載や取得失敗を区別し架空の手順を確定情報として補わない | SLK-342、既決EMR-294。未実施 |
| AC-support-vietnamese-question | 認証済みスタッフがベトナム語で使い方を質問する | 入力を扱い既存の使い方回答・関連項目を確認できる。未記載・取得失敗を区別し、回答を臨床・会計の自動確定へ使わない | SLK-342、既決EMR-294。未実施 |
| AC-support-report-form-preview | 既存バグ報告タブで件名・詳細と現在画面のスクリーンショットを確認し送信する | 件名・詳細・自動取得した画面を報告内容として確認できる。患者・飼主・臨床情報が含まれる場合はredactionを明示確認し、未redactedの内容を他院へ共有しない。redactionできない場合は共有用公開を停止する。既決の送信前内容／送信先確認を経て送信し、取得・送信失敗を成功扱いせず院内報告と外部配送結果を分ける | SLK-344、既存送信前確認・保持契約。未実施 |
| AC-support-report-history-status | 許可された共有一覧で保存済みのopen／resolved報告とその詳細を読む | 元報告と対応状態を確認でき、再報告を要せず履歴を参照できる。報告元医院だけのmutation・個人chat履歴との別scopeを維持。取得失敗を解決済み／空一覧へ置換しない | SLK-353、既存一覧のstatus表示。未実施 |
