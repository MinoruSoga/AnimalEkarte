# トリミングの業務仕様

| 項目 | 内容 |
|:---|:---|
| 文書ID・種別 | D06 / 業務ライフサイクル・受入条件 |
| 決定状態 | 既存記載を継承（個別決定根拠は参照先）。完了後の専用訂正は2026-10-15追加回答で承認（実装適合は別） |
| 実装状態 | 限定ソース静的照合。既存記載の継承であり全機能再監査なし |
| 実機検証状態 | 未実施 |
| 文書確認日 | 2026-10-15 |

## 目的・対象範囲

予約済み／飛込の施術を専用記録へつなぎ、担当・コース／オプション・身体状態・希望／完成画像・施術結果を記録する。一覧・患者選択・新規・編集・トリミングmaster・コース種別masterの6ルートが対象。詳細は[一覧](../../screens/16-trimming-list.md)・[登録／編集](../../screens/17-trimming-form.md)、masterは[コース／オプション](../../screens/settings/master-trimming.md)・[コース種別](../../screens/settings/master-trimming-course-type.md)を正本として保持する。

語彙とID関係は[オントロジー](../../ontology/trimming.md)、transaction・画像・再送・現在差分は[技術契約](../../../design/domains/trimming/technical-spec.md)。診察カルテと入力画面を一体化する#242は着手しない判断が既存本文に記録されており、本整理で再開しない。

## 担当者・権限・前提

trimming:viewで一覧・詳細、createで新規、editで変更、deleteで削除。画像upload／未保存画像cleanupはcreateまたはeditの個別経路であり、記録delete権限と同一視しない。master-trimmingのview/create/edit/deleteを各操作で確認する。コース種別の旧概要「スタッフ管理者以上」と具体的resource/action表の差は実grant確認を要し、職位だけで許可しない。

担当は業務上トリマー。既存wireのdoctor_idや患者ヘッダーの担当医という文言から「医師だけが担当する」と解釈しない。D01の医院所属・有効性・個人対応可能区分・勤務等の予約適格性を保持し、過去担当者名のコピーだけで今回担当者を確定しない。

## 契機・入口・出口

予約カレンダーでトリミング分類の予約が作成されると専用一覧の予約対象として扱う。予約が存在することと施術detailの存在は別。既存予約へ新しいdetailを付ける経路と、直接新規で予約＋detailを同時作成する経路を区別し、同じ予約を二重作成しない。予約／受付カードの分類接続はD01/D02、施術本文はD06、請求の確定・入金はD10が所有する。

<a id="trimming-rules"></a>
## 状態遷移・引継ぎデータ

<a id="requirement-fr-trimming-reservation-and-detail"></a>
### FR-trimming-reservation-and-detail

トリミング分類の予約と専用記録を同じ医院・飼主・ペット・予約IDで結ぶ。予約行のwriterはreservation、detail/optionsはtrimming。通常カルテを無条件に生成する接続へ置き換えない。既存予約にdetailがなければ新規作成、既にあれば更新し、重複detail作成は409で拒否する既存契約を維持する。

新規画面の登録時ステータスは進行中が既定／予約も選べるが、既存予約紐付きの場合は新規ステータス選択を出さない。直接登録の現在時刻はJSTの実時刻を用いる現行仕様であり、固定10:00で同一担当同日の複数ペットを衝突させない。具体的予約wire状態と施術表示の対応はD01の技術正本を参照し、一覧の「完了」を入金済とみなさない。

<a id="requirement-fr-trimming-input-master-history"></a>
### FR-trimming-input-master-history

新規保存には患者・担当・コースを要し、トリミング予約区分を解決できなければ送信しない。既存更新には新規用validationをそのまま再適用しないが、現在scope・状態・安全制約を免除しない。

コース・オプションは有効なマスタから選び、既存記録に結び付く無効項目は「無効」として保持可能。未使用の無効項目を新規追加しない。コースは対象サイズ区分の料金であり、犬種別価格マトリックスを作ったことにしない。コース種別は分類であり料金／所要時間を持つコースと別。オプションの組合せ可否はマスタ定義を維持し、表示だけから全組合せを許可しない。

履歴コピーはコース・オプション・希望・使用品・備考・前回担当名の表示を反映。担当ID・体重・体温・画像・日時・ステータスはコピーしない。前回値が未取得／無効なら現在値を推測して埋めない。保存前に今回の担当と選択内容を確認する。

<a id="requirement-fr-trimming-safety-and-status"></a>
### FR-trimming-safety-and-status

死亡ペットは保存・削除不可。患者情報読込中は保存／削除を無効化し、取得失敗を生存として扱わない。安全特記はD04の内部向け表示を継承する。施術中に見つかった異常を備考へ記録できるが、それだけで通常診療記録や診断・投薬指示を生成しない。

既存予約状態の変更可能条件はreservation側を通す。施術の進行・完了と会計の確定は別であり、会計確定歴の通常Undo停止を施術側のPATCH／DELETEで迂回しない。完了後の専用訂正方針は末尾の追加合意に従う。D03の追記操作をそのまま使えると推定しない。

<a id="requirement-fr-trimming-image-lifecycle"></a>
### FR-trimming-image-lifecycle

希望・完成画像は選択／撮影直後にuploadし、成功object keyを記録の保存時に渡す。プレビュー表示と記録への永続化は別。画像upload中は記録保存不可。保存送信中は画像も含め操作を止め、送信済payloadと表示の乖離を防ぐ。

未変更keyは省略して保持、明示削除は空文字。署名付きURLは保存しない。upload失敗時は直前のkey／previewを保持し失敗を通知する。画像だけの選択／削除も未保存変更として離脱確認する。

未保存の置換／放棄画像をbest-effort cleanupするが、保存に送信済みのkeyはクライアントから削除しない。参照のあるkeyを消さず、削除予約済keyは再保存不可で再uploadを必要とする。旧HTTP(S)値は未変更表示互換に限り、内部objectとして署名／削除しない。他医院key・正規化別名を拒否する。具体的lock／tombstone／再試行は技術契約と画像画面正本。

<a id="requirement-fr-trimming-billing-handoff"></a>
### FR-trimming-billing-handoff

コース・オプション料金は会計側が同一患者の未請求明細APIから取り込む。施術完了時の会計draft push生成ではない。通常診療との同時施術でも入力記録は別に保ち、会計が未請求明細を集約する。

請求元は同じ予約・detail／コース・オプションの参照を保持し、会計の取込済・確定済の範囲と整合させる。原価／単価の再評価・税・割引・保険・取消／再請求はD10正本へ。予約時価格、施術保存時価格、会計取込時価格が常に同じという未確認の保証はしない。変更済マスタと請求済snapshotの意味を分け、確定会計をマスタ変更で再計算しない。

<a id="requirement-fr-trimming-delete-and-atomicity"></a>
### FR-trimming-delete-and-atomicity

トリミング保存の予約行・detail・options・必要監査は同一transaction。失敗時に予約のみ／一部optionsのみを残さない。削除もreservationの専用intentを通し、関連カルテ等の既存依存ガードを保つ。recordの削除と画像objectのcleanupは別境界で、storage削除失敗だけで臨床record削除のcommitを否定しない。画像台帳を監査ログの代替にしない。

<a id="trimming-pending-decisions"></a>
## 追加合意: 完了後訂正

2026-10-15追加合意: 提示文「完了後の誤りは原記録を残して理由付き訂正とし、返金や在庫を自動で戻しません。宿泊料金・必要前金・取消料はスタッフが個別確認して入力し、日数による自動課金や取消時の自動請求・返金は設けません。」を残り方針として再確認し、ユーザーが「ok」と回答（Sentinel_7a67fd30e47c819192cd36d416f0a2bf、確認文 Sentinel_8d0032ca45b881919b3d2ceb5339f792）。業務方針の承認であり、実装・実機PASSではない。

<a id="requirement-fr-trimming-special-correction"></a>
### FR-trimming-special-correction

完了後の施術記録・完了誤りは原記録を残して理由付き専用訂正とする。原記録との関連、変更内容、理由、操作者、時刻を識別できる履歴を保持する。自動返金・在庫巻戻しを行わず、確定済会計の訂正はD10へ引き継ぐ。通常PATCH／DELETEや予約Undoで原記録保護を迂回しない。誤記訂正を新しい再施術記録と混同しない。API・権限・版制御の適合は未確認。

## 異常系・保存・取消

| 状況 | 検知／表示と次の操作 | 保持するもの |
|:---|:---|:---|
| master無効化／組合せ不可 | 新規選択・保存を拒否。既存選択保持の例外と区別 | 保存済記録／当時の参照 |
| 既存予約detail作成競合 | 409後に同じ予約のdetailを取得 | 先行成功を別記録で複製しない |
| 保存／削除応答不明 | 元操作・同じ対象の状態照合まで自動再作成しない | 入力と送信済画像key |
| 患者／予約取得失敗 | 操作を止め再読込 | 生存・空detail・権限ありと推定しない |
| 画像upload失敗 | toast、直前画像保持、再選択／再試行 | 直前の成功key |
| storage cleanup失敗 | 削除台帳を保持して同keyの再削除を可能にする | 臨床保存結果と削除予約 |
| 参照中画像の削除 | 409で拒否 | 他記録の画像参照 |
| 未保存で離脱 | 確認取消なら編集維持、破棄後は未保存画像のみcleanup | 保存済みrecordを巻戻さない |

<a id="trimming-acceptance"></a>
## 受入条件

| 条件ID | 前提／操作 | 期待結果 | 確認手順／証跡 |
|:---|:---|:---|:---|
| AC-trimming-existing-reservation | トリミング予約はあるがdetailなしでフォーム保存 | 同じ予約IDへdetail作成、通常カルテ／別予約を作らない | newと既存紐付きを別確認。未実施 |
| AC-trimming-duplicate-detail | 同じ予約へdetail作成を競合させる | 先行1件だけ成功、後行409で再取得 | 孤立予約／二重detailなし。未実施 |
| AC-trimming-direct-time | 同一担当が同日に2患者を直接登録 | 固定10:00を使わず、現行適用対象の制約を満たして保存 | JST時刻とrecord_shortcut境界。未実施 |
| AC-trimming-new-required | 患者・担当・コース・区分を順に欠落させる | 新規保存しない、対象欄エラー | 既存更新は別validation。未実施 |
| AC-trimming-inactive-master | 既存選択のみ無効化した記録を開き別の無効項目を追加する | 既存選択保持、新規追加拒否 | 有効→無効の競合も確認。未実施 |
| AC-trimming-history-copy | 前回記録をクリックして今回へコピーする | 所定の内容のみ反映。担当ID／バイタル／画像／日時／状態は不変 | 前回担当名を選択確定としない。未実施 |
| AC-trimming-deceased-loading | 患者読込中／失敗／死亡で保存・削除する | 操作不可／拒否、既存記録保持 | 3状態を個別確認。未実施 |
| AC-trimming-image-inflight | 画像upload中または保存中に保存／画像変更を試す | upload中は保存不可、保存中はフォーム操作不可 | 遅延応答・連続選択。未実施 |
| AC-trimming-image-failure | 新画像uploadを失敗させる | 直前keyとpreviewを維持、失敗通知 | 初回失敗と置換失敗。未実施 |
| AC-trimming-image-clear | 画像未変更で再保存／明示削除する | 省略は保持、空文字はclear、署名URLを保存しない | payloadとDBの比較。未実施 |
| AC-trimming-image-race | 保存とcleanupを同じkeyで競合させる | 参照済は削除拒否、削除予約先行は保存409 | どちらの順序も確認。未実施 |
| AC-trimming-image-cross-clinic | 他医院key・パス別名・旧URLを渡す | 新保存は正規key検査、旧URL未変更だけ表示互換 | 旧URLをstorage削除しない。未実施 |
| AC-trimming-atomic-write | detail/options／必要監査を失敗させる | 予約と全detail変更rollback | create/update/delete別確認。未実施 |
| AC-trimming-delete-dependency | カルテ依存がある記録を削除する | reservation intentで拒否、関連record保持 | FE非表示だけに依存しない。未実施 |
| AC-trimming-billing-pull | 施術を完了し、会計で未請求明細を取得する | 完了時push生成なし、会計側で対象明細を取得 | 二重取込と確定済再計算なし。未実施 |
| AC-trimming-size-pricing | 同サイズの異なる犬種と異なるサイズを選ぶ | 犬種価格表を推定せずmasterのサイズ区分を使用 | course/type/optionを別確認。未実施 |
| AC-trimming-image-only-dirty | 画像だけ変更して離脱する | 未保存確認が出る、取消で保持 | 破棄cleanupと送信済key保護。未実施 |
| AC-trimming-context-result | 医院／患者切替後に古い保存結果が戻る | 新contextに混入せず元対象で照合 | 遅延成功・失敗・不明。未実施 |
| AC-trimming-correction-history | Given 完了済み施術、When 理由付き訂正を保存、Then | 原記録と変更内容・理由・actor・時刻・関連を保持。再施術とは区別 | 原記録／履歴比較。未実施 |
| AC-trimming-correction-no-side-effects | Given 確定会計・在庫に関連する施術、When 訂正、Then | 自動返金・在庫巻戻しなし。必要会計訂正はD10へ引継ぐ | 金銭・在庫前後比較。未実施 |
| AC-trimming-stale-update | Given 表示後に他操作が同じ記録を更新、When 元表示版で保存、Then | 現在の原記録は不変、入力保持、競合を通知し再取得・差分確認する。新版への無警告再送不可 | 結果不明は同一操作照合、別create不可。未実施 |
| AC-trimming-image-original-retention | Given 保存済み／参照済み／分類不明の画像、When clear・置換・通常削除、Then | 原blobと元参照・必要履歴を保持し論理削除／非表示に限定。未保存未送信の作業copyと区別 | 現行物理cleanupを許可根拠としない。未実施 |

## 未決事項・関連仕様・根拠

D01の予定枠・勤務適格性、D02受付表示、D04患者／特記、D10未請求取込・確定履歴、D14staff、C04認可、C05監査／保存を再照合する。施術完了後訂正は追加合意を反映。master価格のsnapshot時点・取込後変更・全APIのreceipt／版対応は技術契約の現行差分であり、文書だけで実装済みとはしない。

参照ref: PR670 / 41fcf3eb0a8befb7c0b5f613d1df749a60e30be8。独立レビュー・文書検査は統合時、実機・本番稼働は別工程。


## 画像の原本保持に関する優先契約

保存済み原画像・添付、参照済みkey、業務原本／必要履歴か分類不明の画像はC05の原本保持対象とし、通常削除・clear・置換・取消でblobを物理削除しない。通常操作は論理削除／非表示とし元参照・必要監査を保持する。未保存かつ未送信の放棄作業copyと区別し、分類不明は保護する。物理削除は承認された別手続へ委ね、本整理から新しい例外や保存年限を推定しない。
