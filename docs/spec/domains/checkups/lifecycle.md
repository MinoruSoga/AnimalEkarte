# 健診の業務仕様（D09）

| 項目 | 内容 |
|---|---|
| 文書ID | spec-D09-checkups-lifecycle |
| 種別 | 業務仕様・受入条件 |
| 決定状態 | 既存記載を継承（個別の決定根拠は本文参照。根拠未確認部分は承認済みとしない）＋D03合意目標への接続。死亡後の生前分後日入力・誤記訂正の最新承認を反映。対象外の未決は保持 |
| 実装状態 | 本文の現コード観察と目標契約を分離。コードに存在するだけでは業務承認済み・目標実装済みとしない |
| 実機検証状態 | 今回未実施。既存証跡の引用は今回の実機受入を意味しない |
| 文書確認日 | 2026-10-15 |

既存仕様の継承とD03の合意済み目標を整理した文書候補。ソフトウェア仕様の整理、実装完了、実機受入、医師による具体的な臨床数値の承認は別の判定とする。

<a id="checkup-document-scope"></a>
## 範囲・根拠
2026-10-15全ドメイン整理、基準commit `41fcf3eb0a8befb7c0b5f613d1df749a60e30be8`。[ADR004](../../../architecture/adr/004-checkup-canonical-system.md)、[一覧・登録](../../screens/25-checkups-list.md)、[D03保存§2](../medical-records/lifecycle.md#hybrid-sync)が既存根拠。[技術契約](../../../design/domains/checkups/technical-spec.md)は現コードの保存経路・型・不足を分離する。健診パッケージ、項目・結果、担当、次回予定、カルテ帰属を対象とし、配信はD15、結果読取はD03カルテとD21スタッフ専用Owner Reportへ渡す。D20への健診結果公開は今回の対象に含めない。具体的閾値・対象種・臨床妥当性を承認するのは医師であり、別の「臨床責任者」を設けない。

<a id="fr-checkups-canonical"></a>
## FR-checkups-canonical：健診をCheckup系に一本化する
Checkupパッケージ（checkup_types／checkup_type_fields／checkups／checkup_field_results）だけを健診の正系統とする。Examinationへ同じ健診を重ねて保存・集計しない。血液検査等の臨床検査はD07に残し、必要なら参照として結ぶ。退役したexam_types 12000〜12003／exam_type_fields 45〜58は再利用禁止。既適用環境に残る旧値を名称一致で自動移行／削除しない。既存ADRのAcceptedは今回の実環境照合完了ではない。

<a id="fr-checkups-record"></a>
## FR-checkups-record：どの入口でも同じカルテ帰属を保つ
Path Aは保存済みカルテの健診タブで親に紐付ける。Path Bは一覧→患者選択→独立フォームで実施日を診療日とするカルテを作成し健診を紐付ける既存経路。Path Cは一覧から対象カルテの健診タブ＋checkupIdを開いて編集し、別の独立編集系を作らない。通常カルテ全タブ一括保存を追加しない。

健診種別・実施日を既存フォーム必須とし、担当医・次回日・所見・型付き項目結果は保存対象。担当医未選択を新たな必須にしない。担当医の解除と未送信を区別する。独立フォームの表示既定は1年候補で、手動次回日を保持する。臨床的な間隔の適否は医師判断であり、全パッケージ一律の医学的推奨としない。通常Path BをD03の過去専用入力や患者訂正の代用にしない。

<a id="fr-checkups-results"></a>
## FR-checkups-results：パッケージの型に合う結果を保持する
対象医院・パッケージの項目定義を参照し、number／text／boolean／single_select／multi_select／checklistに対応する値だけを保存する。false、0、未入力を区別する。選択肢にない値、別医院／別パッケージのfieldは拒否する。項目名・型・単位・数値範囲はサーバで定義から解決し、保存時内容を結果へ保持する。HIGH/LOWはサーバの範囲比較で導出する表示補助で、欠損／未評価を臨床的正常と断定しない。

パッケージ本体は名称・有効状態・親区分・価格・間隔・対象年齢・説明・表示順を扱い、名称必須／医院内一意・非負価格・同医院の親参照という既存条件を保つ。子種別や使用中の健診記録がある本体の削除は拒否する。項目定義の作成・更新・削除・並替と結果値の保存を別操作にし、既存結果を削除／再計算する権限をマスタ編集から推定しない。現行マスタの操作grantはcheckupsであり、診療項目画面を共有していてもmaster-medicalへ一律変換しない。

結果配列は一括置換が保存単位。未送信／nullは全消去に変換せず拒否、意図的全消去は明示[]。途中の項目だけ消えて成功扱いにならないよう、置換と必要な監査を一つの保存結果にする。パッケージ変更時は旧項目値を新しいfieldへ名前／並び順で移さない。マスタ改定で過去結果を自動再判定・上書きしない。当時snapshotが欠損する場合は現在マスタから補完しない。

<a id="fr-checkups-protection"></a>
## FR-checkups-protection：確定・失敗・競合でも既存記録を守る

<a id="checkup-deceased-retrospective"></a>
2026-10-15の死亡後取扱い確認（Sentinel_2b63d8f5aad88191aa2b4dff79fb5ac2）に対するユーザー「ok」（Sentinel_45be2cab0fe4819188c7c0f9a211358c）により、生前に行った健診の後日入力・誤記訂正は死亡状態を戻さず理由と履歴を残す専用経路で可能にする。実施日・元記録・現在の入力時刻・actor／確認者を照合し、生前分か不明なら確認する。D03の既決の過去専用入力／確定後追記と連携し、通常Path Bの親自動作成や死亡状態取消で代用しない。死亡後の新しい健診・未来予約／実施は許可せず、原記録保持・自動返金／在庫戻しなしを維持する。現行通常死亡gateの適合差は技術契約へ残す。
finalized親への通常の健診追加・編集・削除・field置換を拒否する。確定後訂正／取り違え／過去記録はD03の専用手順へ渡し、元記載を保持する。全writeで現在の医院・患者・親record・checkup・master関係とgrantを照合し、医師確認者と実編集actorを区別して必要な前後・理由・時刻を残す。旧版・scope変更・取得不明で上書きしない。

Path Bは現状、カルテ作成→健診作成→値があればfield置換の連続要求であり、全体atomicな保存ではない。後段失敗・権限切替・死亡状態変化時に前段が残る可能性を明示し、「全部保存できなかったから存在しない」と判定しない。作成済みID／結果を照会して再開対象を特定し、無条件にもう一つのカルテを作らない。作成済み記録を黙って削除する補償も新設しない。結果不明・ID対応不明なら停止して照合する。

独立CheckupFormには現状離脱保護がない。カルテ本体のNavigationBlockerが独立フォームも守るとは記載しない。離脱保護改善を既存実装済み／新規承認済みとしない。

<a id="fr-checkups-followup"></a>
## FR-checkups-followup：結果・次回予定を必要な受け手へ渡す
一覧は1ページ20件で、全文は取得ページ内絞込み。次回日を表示し、期限切れ／30日以内の期限間近を既存アラートとして区別する。期限未設定は期限切れへ捏造しない。患者と現在owner／医院の整合を維持し、件数／検索範囲を全件結果と混同しない。

健診保存とタグ同期／フォローアップ配信成功は別。2026-10-15の確認案 Sentinel_a408975ce6408191be7a25734aae0aab の7に対するユーザー承認 Sentinel_4f050b1c0ed08191b5ca67deeaeec843「ok」により、記録作成だけでフォロー通知を送らず、スタッフが実施済み・送信対象を確認した後に扱う。お礼等の一般案内と再検査／受診を促す医療的案内を分け、後者は医師の要フォロー指定がある場合だけとする。担当医欄・異常値・次回日から指定を推定しない。[D15確認契約](../../line/lstep-integration.md#crm-followup-confirmation)の認可・必要監査・現在対象再照合・結果不明保留／重複防止へ接続し、Path B途中保存を実施済み確認や送信許可に変換しない。D15へ対象clinic・owner・pet・健診／次回日を渡し、送信条件・死亡等の対象外・停止・再試行・監査はD15が管理する。配信失敗を理由に健診を再作成しない。D03とD21は各スタッフ閲覧認可の下で保存済み結果projectionを参照し、読取表示は原結果を書換えない。D20は保存済み接種実績・最終来院日等の限定表示であり、今回健診結果を公開しない。会計／在庫の訂正を健診保存から自動実行しない。

## 受入条件（GWT、今回実機未実施）
| AC ID | Given | When | Then | FR |
|:---|:---|:---|:---|:---|
| AC-checkups-canonical | 健診を新規記録 | 保存・集計 | Checkup系一つ、Examinationへ同じ記録を重複生成しない | canonical |
| AC-checkups-retired-id | 旧健診seedの予約ID範囲 | 新割当／移行 | 再利用せず、実旧値確認前に削除・推測移行しない | canonical |
| AC-checkups-chart-entry | 保存済みカルテあり | タブから追加 | 同親へ紐付けて即時個別保存 | record |
| AC-checkups-quick-entry | 独立入口で患者・種別・実施日選択 | 保存 | 実施日を診療日とする親と健診を対応付け | record |
| AC-checkups-edit-entry | 一覧の既存行 | 編集 | 同recordの健診タブ＋対象checkupを開く | record |
| AC-checkups-required | 種別／実施日欠損 | 保存 | 既存必須エラー、担当医等を新規必須にしない | record |
| AC-checkups-doctor-clear | 既存担当医あり | 明示解除／未送信 | 解除操作と保持を区別、nullを無条件解除にしない | record |
| AC-checkups-next-manual | 手動指定した次回日 | 日付／候補を操作 | 手動意図を保持し一律医学的推奨としない | record |
| AC-checkups-value-types | 0／false／空欄を含む項目 | 保存・再読込 | 型どおり保持し0・falseを未入力にしない | results |
| AC-checkups-field-scope | 別医院／別packageのfield | PUT | 拒否し異なる基準値を適用しない | results |
| AC-checkups-invalid-choice | optionsにない値 | PUT | 拒否し任意文字列へ黙って変換しない | results |
| AC-checkups-replace-missing | results未送信／null | PUT | 拒否し既存結果を消さない | results |
| AC-checkups-replace-empty | 編集可能な対象で意図的[] | PUT | 対象結果を明示全消去し必要監査を同時保持 | results |
| AC-checkups-replace-failure | 置換中／監査で失敗 | 保存 | 部分削除・部分挿入を成功扱いしない | results |
| AC-checkups-master-change | 項目定義改定／package変更 | 新旧結果表示 | 保存時snapshotを保持、名前・順序で旧値を移さない | results |
| AC-checkups-unassessed | 値／基準値欠損 | 表示 | 判定未成立を正常確認済みとしない | results |
| AC-checkups-deceased-history | 死亡登録済みで生前健診の対象・実施日を確認済み | 専用後日入力／誤記訂正 | 死亡状態を戻さず理由・履歴・元記録を保持して記録可。実施日と入力時刻を分ける | protection |
| AC-checkups-deceased-boundary | 死亡後の新たな健診・未来予約／実施、または生前分か不明 | 保存を要求 | 新たな実施は拒否、不明は確認まで停止。通常Path Bで専用処理を代用せず、自動返金・在庫戻しなし | protection |
| AC-checkups-finalized | finalized親 | CRUD／field置換 | 通常write拒否、専用訂正へ | protection |
| AC-checkups-scope-grant | 他医院／患者・親不一致／権限失効 | write | 拒否して原記録を保持 | protection |
| AC-checkups-partial-save | Path Bの親作成後に健診／fieldが失敗 | 再開 | 前段の存在を照会、無条件親再create・自動削除なし | protection |
| AC-checkups-unknown-result | 応答断で保存結果不明 | 再試行 | 作成済みID・関係を照合、不明なら停止 | protection |
| AC-checkups-concurrent | 他者変更／親確定と競合 | 保存 | 最新依存で再照合、古い値の上書きなし | protection |
| AC-checkups-leave-limit | 独立フォーム未保存 | 離脱 | 現行保護なしの制約を明記しカルテguardで保護済みとしない | protection |
| AC-checkups-alert | 次回日過去／30日内／未設定 | 一覧 | 期限状態を区別、未設定を期限切れにしない | followup |
| AC-checkups-list-scope | 20件ページの全文絞込み | 検索 | 全件検索として件数を誤表示しない | followup |
| AC-checkups-followup-confirmation | 健診記録作成済み・スタッフ未確認 | 通知判定 | 作成だけで送信せず実施済み・送信対象確認を待つ | followup |
| AC-checkups-followup-purpose | 一般案内／医療的案内の候補 | スタッフ確認後に判定 | 一般と医療を分離し、後者は同一対象への医師要フォロー指定必須。異常値・担当医・次回日で代用しない | followup |
| AC-checkups-followup-recovery | 保存段階不明・確認保存失敗・送信結果不明／同一操作反復 | 再開 | 元IDと確認／送信結果を照合し、不明時保留・重複配信なし・健診再作成なし | followup |
| AC-checkups-notify-failure | 健診保存済み・同期／配信失敗 | 結果通知 | 保存と送信を分離、健診再作成なし | followup |
| AC-checkups-publication | D03/D21が結果を参照しD20の公開範囲を確認 | 読取 | 現scopeとD03/D21のスタッフ閲覧認可で同じ保存結果・型・版／取得状態を返す。D20へ健診結果を公開しない | followup |

## 文書完成条件・完了外
正系統、3入口、型・置換単位、確定保護、失敗／部分成功、マスタ履歴、期限／受け手を整理する範囲。具体的臨床閾値・旧データの移行実態・全writerの監査／版／再送receipt・Path B全体の再開実装・医院実権限・実機受入を完了に含めない。独立フォーム離脱保護、全体atomic化、新しい日付制限等は今回新たに承認された業務としない。


D21 Owner Reportはスタッフの閲覧専用であり、同画面からの印刷・共有を設けない。D03や検査自ドメインに既存の印刷経路があることを、D21経由の印刷許可へ拡張しない。
