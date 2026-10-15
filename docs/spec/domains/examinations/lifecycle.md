# 検査・機器受信の業務仕様（D07）

| 項目 | 内容 |
|---|---|
| 文書ID | spec-D07-examinations-lifecycle |
| 種別 | 業務仕様・受入条件 |
| 決定状態 | 既存記載を継承（個別の決定根拠は本文参照。根拠未確認部分は承認済みとしない）＋D03合意目標への接続。担当医未選択で保存可・入力者ログ保持の2026-10-15第3項合意を反映。対象外の未決は保持 |
| 実装状態 | 本文の現コード観察と目標契約を分離。コードに存在するだけでは業務承認済み・目標実装済みとしない |
| 実機検証状態 | 今回未実施。既存証跡の引用は今回の実機受入を意味しない |
| 文書確認日 | 2026-10-15 |

既存仕様の継承とD03の合意済み目標を整理した文書候補。ソフトウェア仕様の整理、実装完了、実機受入、医師による具体的な臨床数値の承認は別の判定とする。

<a id="examination-document-scope"></a>
## 範囲・根拠・判定
2026-10-15の全ドメイン仕様整理。基準commitは `41fcf3eb0a8befb7c0b5f613d1df749a60e30be8`。既存[一覧](../../screens/12-examinations-list.md)、[入力](../../screens/13-examinations-form.md)、[検査マスタ](../../screens/settings/master-examinations.md)、[ADR007](../../../architecture/adr/007-lab-device-receive-and-commit.md)、[ADR008](../../../architecture/adr/008-local-lab-device-agent.md)、[現行接続正本](../../../ops/deploy/lab-devices/LAB_DEVICE_CONNECTIVITY.md)とD03の[後日結果契約](../medical-records/pages/medical-record-form.md#requirement-fr-medical-records-late-result-shared-history)を一つの業務境界へ整理する。新しい臨床数値・機器運用の承認ではない。

要件責任者は既存ADRのMinoruSoga。検査値／基準値の具体的内容・種・単位・出典の判断者は医師であり、新たな「臨床責任者」役割を要求しない。実装、医院実権限、基準値bundle承認、旧データ対応、実機受入／公開は今回の完了判定外。[技術契約](../../../design/domains/examinations/technical-spec.md)に現状差と停止条件を残す。

機器受信経路の優先順位は、2026-10-14 EMR-402を反映した現行接続正本／ADR008のcurrent daily pathが先。ADR007/008の古い「日常ポート選択ゼロ」「LaunchAgent常駐」を現在の日常受信要件として適用しない。その他の受信同一性・待機・attach/detach・隔離契約は維持する。

## 業務状態
検査オーダー・結果入力・検査確定・カルテ確定・機器受信を別の状態として扱う。
- pending／in_progress／result_enteredは進捗。保存だけを全検査の臨床確定と扱わない。
- confirmedは検査全体を封印。revisionなしcompletedは結果編集・削除を封印する。
- revisionありcompletedは確定解除後の作業版となり得る。同じ文字列completedを理由に全件を同じロック状態へ変換しない。
- 検査確定解除は検査だけの専用手順。カルテfinalizedを解除しない。
- 受信済みreceived、患者紐付け／persisted、要確認needs_review、fixtureのrevertedを混同しない。

<a id="fr-examinations-record"></a>
## FR-examinations-record：患者に帰属した検査結果を一度だけ記録する
独立新規はペット選択→検査入力、既存カルテ紐付きはカルテ検査タブ、未紐付きは独立詳細へ遷移する。検査種別・担当医・日付・結果項目を対象医院と患者に結び付ける。画面の担当医候補はactiveのdoctor。2026-10-15の6点確認（Sentinel_4230d91e7dd88191b67f30714ed073ba）に対するユーザー「ok」（Sentinel_ae272463dc248191b51fbef540e04109）の第3項により、担当医は未選択でも保存できる。検査種別の既存必須条件は維持し、担当医の未選択だけで保存を止めない。独立入力・カルテ内の関連入力でも同じ条件とし、担当医と実際の入力／編集actorを分離する。認証された入力者のログを保存結果と結び付け、未選択の担当医を入力者・カルテ担当医から自動補完しない。現行独立フォームの担当医必須は目標との差分である。

結果は文字列の原値（数値、定性、不等式、欠測表現）と単位を保持する。適用する基準範囲を保存結果／版と結び付け、HIGH/LOWはサーバ導出の表示補助とする。未取得・未評価・基準値なしを「正常が確認された」としない。新しい閾値・診断・投薬推奨を本整理で追加しない。

<a id="examination-deceased-retrospective"></a>
### 死亡登録後の生前検査記録

2026-10-15の死亡後取扱い確認（Sentinel_2b63d8f5aad88191aa2b4dff79fb5ac2）に対するユーザー「ok」（Sentinel_45be2cab0fe4819188c7c0f9a211358c）により、生前に行われた検査の後日入力・誤記訂正を死亡状態を戻さず理由と履歴付きで行える。D03の既決の専用過去入力／確定後追記、D07の原結果・official版保護と照合する。実施時点・元記録・現在の入力日時・actor／確認者を区別し、生前分か不明なら確認する。通常の機器待機／新たな検査実施を死亡ペットへ開放せず、専用adapterを通常create／attachや死亡状態取消で代用しない。自動返金・在庫戻しは行わない。現行死亡gateとの差と実装適合は技術契約へ残す。

<a id="fr-examinations-revision"></a>
## FR-examinations-revision：確定・訂正で原結果と版履歴を守る
confirmedは普通のPATCH／項目置換／削除では変更しない。専用unconfirm権限と理由を用いて作業版を作り、修正保存後に再確定する。解除理由は空白のみ不可、既存上限500文字を維持する。revisionのない旧confirmedを架空の版で解除しない。元official版を保持し、変更前後・理由・認証actor・時刻を辿れるようにする。作業変更と再確定を一度の曖昧な保存に混ぜない。

紐付くカルテがfinalizedなら通常の結果更新、unconfirm、再確定で原記載を変更しない。後日結果／訂正結果はD03の専用追記境界へ渡し、同じexamと固定結果版を参照する。原版と後日版を区別し、名前や行位置で跨版項目を同一視しない。

<a id="fr-examinations-receive"></a>
## FR-examinations-receive：転記せず受信し、患者不明は未紐付けに保持する
通常受信は検査用MacのHTTPS Chromeで明示選択したNX600/AU10VのWeb Serialをページ表示中に受信する。origin-wide排他が取れなければ開かない。日常経路は受信専用。PU-4010のdecoder、IDEXXのdecoder/persist実装の存在を日常運用承認へ読み替えない。COM4/7000V・Drワンデータ・検査器本体への直結は対象外。

本日診療中カルテカードの選択を日常の患者指定とし、電文の検体IDで患者を自動決定しない。医院ごとの有効待機は1件、別患者選択で置換、解除／期限切れまで同じ患者に持続する。期限切れ到着や待機なしは未紐付けへ保存する。診察端末から遠隔待機起動しない。接続・最終受信・未紐付け・保存済み・要確認を区別して表示する。

<a id="fr-examinations-mapping"></a>
## FR-examinations-mapping：機器の原値と対応付けを保存する
1フレームを1測定／1ジョブとして記録し、measured_atを検査日時、received_atを受信監査時刻とする。原コードのsuffix、不等式、定性値、単位を削らない。生バイト・接続識別子・検体IDをログやGitへ出さない。生バイトはサーバでデコード後に保存しない。

医院＋source_type＋device_item_codeのマッピングで結果項目を解決する。未知コードを推測して自動登録せずneeds_reviewとしてジョブ項目に残す。対応した検査種別がNならN検査に分割し、未対応項目を黙って落とさない。後日のマスタ対応後はそのジョブだけを再評価する。保存済み臨床版を現在マスタで自動的に書き換えない。

<a id="fr-examinations-idempotency"></a>
## FR-examinations-idempotency：再送・後付け・Undoの結果を一意にする
機器の同じ医院/source/fingerprintは同じジョブを返す。fixture側は患者・日・種別だけの重複判定を使わず、保存対象内容が完全同一の場合のみskipする。同日同種別でも内容違いは別結果である。

未紐付け受信へのattachは対象ジョブ＋患者を使い、クライアントから受信値を改変しない。別患者への付替えをattachで強制しない。device Undoは安全条件を満たしたdetachでpersisted→received・患者NULLとし、指紋を残して再attach可能にする。fixture revertの終端revertedとは別。由来検査が複数なら全件を一つの補償単位として扱う。

利用receipt、手動変更、検査確定、確定カルテなどの使用・保護状態によりdetachできない場合は拒否し、検査訂正へ案内する。確認ダイアログを追加してこの制約を解除しない。

<a id="fr-examinations-recovery"></a>
## FR-examinations-recovery：失敗・結果不明を成功にしない
未送信フレームはブラウザメモリ上だけで、ページ終了・停止・医院／権限変更で消失し得る。再送は機器側で行い同一指紋で照合する。受信成功と結果persist成功を別に示し、解析失敗・未対応・保存失敗を隠さない。通信断後は既存job／結果を照会し、成功不明を新規作成で補わない。fixture auto_commitは医院/sourceごとの初期falseを維持し、既存のauthorized enable・停止・失敗通知・監査の条件を充たすまで解禁しない。

<a id="fr-examinations-master"></a>
## FR-examinations-master：基準値マスタと臨床利用を分離する
検査プラン本体と測定項目・単位・基準値は既存診療項目の検査タブで管理する。プラン名は医院内一意。操作grantはmaster-medical、機器項目対応はlab-importの該当grantに従う。検査編集権限だけで全マスタを編集可能にしない。過去結果は保存時の範囲／単位／版を優先し、欠損した当時値を現在値から作らない。マスタ登録できることと、医師がその具体的範囲を臨床承認したことは別である。

<a id="fr-examinations-read"></a>
## FR-examinations-read：読取・印刷・他領域への受渡しを同じ結果にそろえる
一覧と履歴は現在の取得範囲を明示する。一覧は期間サーバ抽出・20件ページ、進捗等は取得ページ内絞込み。履歴は現状先頭20件であり全履歴検索と表示しない。結果印刷は保存済みprint-snapshot。D03カルテとD21スタッフ専用レポートが使う院内scope・exam ID・固定版・版内item refと取得結果を渡す。表示／印刷は確定や飼主公開の代行ではない。読取権限不足／取得失敗を結果なしとして公開しない。

## 受入条件（仕様のGWT。今回の実機PASSではない）
| AC ID | Given | When | Then | FR |
|:---|:---|:---|:---|:---|
| AC-examinations-entry | 同じ患者の検査 | 一覧から開く | カルテ紐付きは検査タブへ、未紐付きは独立詳細へ着地 | record |
| AC-examinations-scope | 他医院／異なる患者・カルテの参照 | 保存する | 拒否し他患者へ混入しない | record |
| AC-examinations-required | 検査種別未選択 | 独立フォーム保存 | 種別の既存必須エラーを表示し保存成功にしない。担当医未選択を必須エラーへ追加しない | record |
| AC-examinations-doctor-optional | 種別等の既存条件を満たし担当医未選択 | 独立入力／カルテ内関連入力から保存・再表示 | 保存でき、担当医未選択を保持する。入力者・カルテ担当医から勝手に担当を補完しない | record |
| AC-examinations-editor-log | 担当医未選択または入力者と異なる担当医 | 検査を入力・編集して保存 | 認証された入力／編集actor・対象・時刻をログに残し、担当医属性と分けて追跡できる。必要ログ失敗を保存成功にしない | record |
| AC-examinations-value | 不等式／定性値／欠測記号 | 入力・受信を保存 | 原値・単位を保持し数値への強制変換で失わない | record |
| AC-examinations-unassessed | 範囲なし／取得失敗／未評価 | 結果表示 | 正常確定と区別し取得失敗は再取得へ | record |
| AC-examinations-deceased-history | 死亡登録済みで生前検査の対象・実施時点を確認済み | 専用後日入力／誤記訂正 | 死亡状態を戻さず理由・履歴・原結果／公式版を保持して記録可。入力時刻・actorを実施時点と分離 | record |
| AC-examinations-deceased-boundary | 死亡後の新たな実施・未来予定、または生前分か不明 | 検査保存を要求 | 新たな実施は拒否、不明は確認まで停止。自動返金・在庫戻しなし | record |
| AC-examinations-confirm-lock | confirmed | 普通更新・項目置換・削除 | 全ロックで拒否 | revision |
| AC-examinations-legacy-seal | revisionなしcompleted | 結果更新／削除 | 封印を維持 | revision |
| AC-examinations-unconfirm | revisionありconfirmed・専用grantあり | 理由付き解除 | 原officialを保持し作業版、actor／理由／時刻を記録 | revision |
| AC-examinations-unconfirm-denied | grantなし／理由空白／旧confirmed版なし | 解除 | 拒否し架空revisionを作らない | revision |
| AC-examinations-reconfirm | 作業版に未保存修正あり | 再確定 | 修正保存と再確定を分離し固定版を生成 | revision |
| AC-examinations-finalized-parent | 親カルテfinalized | 結果訂正／後日到着 | 原版を保持、通常PATCHやunfinalizeで迂回せず専用追記へ | revision |
| AC-examinations-cross-version | item対応不明の旧版／新版 | 結果参照 | 名前／位置で補完せず照合停止 | revision |
| AC-examinations-browser | NX600/AU10V・許可したポート | 日常受信開始 | ページ表示中・受信専用・単一consumer | receive |
| AC-examinations-lock-denied | 別タブが受信占有 | 開始 | 排他取得前にポートを開かない | receive |
| AC-examinations-unsupported | PU-4010／IDEXX／旧agent | 日常経路利用 | コード存在を通常運用承認にせず対象外／別gate表示 | receive |
| AC-examinations-wait | 有効待機中 | 別患者選択／解除 | 1件へ置換／解除、遠隔待機開始なし | receive |
| AC-examinations-wait-expired | 待機なし／期限切れ | フレーム到着 | 未紐付け保存、以前の患者へ付けない | receive |
| AC-examinations-specimen | 電文に検体IDあり | 受信 | 表示用に限定し患者自動照合に使わない | receive |
| AC-examinations-clock | 電文日時と受信日時が異なる | 結果保存 | 検査日と監査時刻を分離し黙って置換しない | mapping |
| AC-examinations-unknown-code | 未対応コードあり | persist | 対応行のみ結果、未対応は原値付き要確認として残す | mapping |
| AC-examinations-multiple-types | 1フレームにN種別 | persist／detach | N検査に分割しjob由来全件を補償単位とする | mapping |
| AC-examinations-receive-duplicate | 同医院/source/fingerprint再送 | 受信 | 同じjob、二重結果なし | idempotency |
| AC-examinations-fixture-difference | 同日同種別で内容違い | fixture取込 | 日だけでskipせず別結果 | idempotency |
| AC-examinations-attach | 未紐付け・同医院生存患者 | attach | 保存済み原値で紐付け、別患者既存紐付けは拒否 | idempotency |
| AC-examinations-detach | 未使用で補償条件充足 | device Undo | 同jobをreceived／患者NULLへ、指紋保持で再attach可 | idempotency |
| AC-examinations-detach-used | usage receipt／手動変更／確定保護あり | Undo | 拒否し検査訂正へ、fixture revertで迂回しない | idempotency |
| AC-examinations-frame-loss | 未送信状態でページ終了 | 復帰 | 保存済みと断定せず機器再送・指紋照合へ | recovery |
| AC-examinations-result-unknown | 送信後応答不明 | 再操作 | job／保存結果を照会、無条件新規作成なし | recovery |
| AC-examinations-autocommit | 有効化条件未確認 | fixture自動取込 | 初期falseを保持 | recovery |
| AC-examinations-master-grant | 結果編集のみのactor | マスタ変更 | master／lab-importの別grantで検査 | master |
| AC-examinations-master-history | 基準値改定／当時値欠損 | 過去結果表示 | 保存版を保持、不明値を現在値で補完しない | master |
| AC-examinations-clinical-approval | マスタ投入／コード完了 | 利用開始判断 | 医師の具体値確認・実機受入済みとしない | master |
| AC-examinations-list-scope | ページ内絞込み／先頭履歴のみ | 検索 | 全件検索・全履歴と誤表示しない | read |
| AC-examinations-print | 保存済み結果版あり | 印刷 | 同じ保存snapshot、未保存値を自動保存しない | read |
| AC-examinations-owner-handoff | D03カルテ／D21スタッフレポートが参照 | 受渡し | 同じ医院・患者・exam・固定版・取得状態を認可済みスタッフへ返す。D20への検査結果公開は行わない | read |
| AC-examinations-write-failure | 必須監査失敗／競合／認証医院変更 | write | 部分成功を避け原記録保持、古い応答を別scopeへ反映しない | revision |

## 担当境界と文書完了外
- D03はカルテ確定・追記・患者取り違え訂正・過去入力・復元。D07は検査revision／機器受信／補償。検査usage receiptをD03転記／復元receiptへ流用しない。
- D09が健診の唯一の正系統。検査マスタへ健診の同じ記録を再作成しない。
- D10は請求／訂正／返金、D13は在庫。結果保存・確定で勝手に課金／返金／在庫補正しない。
- D20は保存済み接種実績・最終来院日等の限定健康表示を所有し、今回D07の検査結果を公開しない。D21はスタッフ専用Owner Reportの閲覧認可を所有する。D07のviewだけでいずれの認可も代行しない。D21に飼主公開や新しい公開承認操作を追加しない。
- 未決の新規業務追加は本整理に含めない。具体的臨床数値、現場ポート／ケーブル確認、実データ旧版照合、現grant・writer適合は引継ぎ。これらを承認済み／実機PASSと記載しない。


D21 Owner Reportはスタッフの閲覧専用であり、同画面からの印刷・共有を設けない。D03や検査自ドメインに既存の印刷経路があることを、D21経由の印刷許可へ拡張しない。
