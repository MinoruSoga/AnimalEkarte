# 検査・機器受信の業務仕様（D07）

| 項目 | 内容 |
|---|---|
| 文書ID | spec-D07-examinations-lifecycle |
| 種別 | 業務仕様・受入条件 |
| 決定状態 | 2026-10-17追加決定は末尾と進捗入口を参照。従来：既存記載を継承（個別の決定根拠は本文参照。根拠未確認部分は承認済みとしない）＋D03合意目標への接続。担当医未選択で保存可・入力者ログ保持の2026-10-15第3項合意を反映。検査項目選択印刷・感染症結果患者トップの追加採用は本文の直接回答に限定。対象外の未決は保持 |
| 実装状態 | 本文の現コード観察と目標契約を分離。コードに存在するだけでは業務承認済み・目標実装済みとしない |
| 実機検証状態 | 今回未実施。既存証跡の引用は今回の実機受入を意味しない |
| 文書確認日 | 2026-10-17追加差分（従来確認2026-10-15） |

既存仕様の継承とD03の合意済み目標を整理した文書候補。ソフトウェア仕様の整理、実装完了、実機受入、医師による具体的な臨床数値の承認は別の判定とする。

<a id="examination-document-scope"></a>
## 範囲・根拠・判定
2026-10-15の全ドメイン仕様整理。基準commitは `41fcf3eb0a8befb7c0b5f613d1df749a60e30be8`。既存[一覧](../../screens/12-examinations-list.md)、[入力](../../screens/13-examinations-form.md)、[検査マスタ](../../screens/settings/master-examinations.md)、[ADR007](../../../architecture/adr/007-lab-device-receive-and-commit.md)、[ADR008](../../../architecture/adr/008-local-lab-device-agent.md)、[現行接続正本](../../../ops/deploy/lab-devices/LAB_DEVICE_CONNECTIVITY.md)とD03の[後日結果契約](../medical-records/pages/medical-record-form.md#requirement-fr-medical-records-late-result-shared-history)を一つの業務境界へ整理する。新しい臨床数値・機器運用の承認ではない。

要件責任者は既存ADRのMinoruSoga。検査値／基準値の具体的内容・種・単位・出典の判断者は医師であり、新たな「臨床責任者」役割を要求しない。実装、医院実権限、基準値bundle承認、旧データ対応、実機受入／公開は今回の完了判定外。[技術契約](../../../design/domains/examinations/technical-spec.md)に現状差と停止条件を残す。

機器受信経路の優先順位は、2026-10-14 EMR-402を反映した現行接続正本／ADR008のcurrent daily pathが先。ADR007/008の古い「日常ポート選択ゼロ」「LaunchAgent常駐」を現在の日常受信要件として適用しない。その他の受信同一性・待機・attach/detach・隔離契約は維持する。

[旧Drワン解析資料の境界整理](../../lab-device-integration/README.md#analysis-contract-boundaries)は旧観測とコード参照であり、本書の既決条件を変更しない。大阪血清・PicJoint・医院差、旧ID/履歴/画像移行と初日最低互換は[未確認事項](../../lab-device-integration/README.md#migration-and-phase-unknowns)へ分け、新規採用・対象外・Phase2と推測しない。

## 業務状態
検査オーダー・結果入力・検査確定・カルテの保存／診療終了・機器受信を別の状態として扱う。旧カルテfinalizedは履歴上の値とし、医師確定操作を通常編集・終了の必要条件に戻さない。
- pending／in_progress／result_enteredは進捗。保存だけを全検査の臨床確定と扱わない。
- confirmedは検査全体を封印。revisionなしcompletedは結果編集・削除を封印する。
- revisionありcompletedは確定解除後の作業版となり得る。同じ文字列completedを理由に全件を同じロック状態へ変換しない。
- 検査確定解除は検査だけの専用手順。カルテの診療終了・旧確定履歴を解除／消去しない。
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

紐付くカルテが診療終了済み又は旧finalizedであることだけを理由に、カルテ本文の通常編集を一律拒否しない。一方、検査自身のconfirmed／旧封印・official版／専用unconfirm条件は維持し、D03の通常編集やタブクリアで検査原版を上書きしない。後日結果／訂正結果は[最新のD03接続](#examination-normal-edit-autosave)の変更履歴へ渡し、同じexamと固定結果版を参照する。原版と後日版を区別し、名前や行位置で跨版項目を同一視しない。

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
一覧と履歴は現在の取得範囲を明示する。一覧は期間サーバ抽出・20件ページ、進捗等は取得ページ内絞込み。履歴は現状先頭20件であり全履歴検索と表示しない。結果印刷は保存済みprint-snapshot。D03カルテ・D04スタッフ患者トップ・D21スタッフ専用レポートが使う院内scope・exam ID・固定版・版内item refと取得結果を渡す。表示／印刷は確定や飼主公開の代行ではない。読取権限不足／取得失敗を結果なしとして公開しない。

<a id="examination-print-infection-agreement-20261015"></a>
## 2026-10-15追加採用：検査項目の選択印刷と患者トップ表示

ユーザー回答 Sentinel_03d710da77408191ad91e022d77b6493 の「2 yes」「3 yes」により、(2)検査項目を選んで印刷し未実施項目を除外すること、(3)猫FIV／FeLV・犬mf／HWの保存結果と検査日をスタッフの患者トップへ表示し、院内保存結果は自動反映、他院結果は手入力で記録することを採用した。先行の担当医任意・原版／訂正履歴・閲覧認可・保存済みsnapshot契約は保持する。

参考原資料は[検歴印刷画面](https://drive.google.com/file/d/1_qkWPBHyC1xgwZDwwFB_T_Ec_t-1ni68/view)、[猫の結果表示](https://drive.google.com/file/d/1_aXNKWuNkZ0F_cktj7bY235K-FNyf0cl/view)、[犬の結果表示](https://drive.google.com/file/d/19R3gTbvlNFSQqigwvJhBvMFX_KoynttX/view)。旧資料の存在や更新日だけを採用根拠とせず、今回の直接回答を優先する。旧Drワンへの接続・データ移行・新機器の通常受信を承認した意味ではない。

以下は追加採用を満たす目標契約であり、既存API／DBでの実装済み・実機受入ではない。複数日比較帳票、全検歴検索の拡張、グラフ、医学的な期限／判定閾値・診断／再検査推奨は今回追加しない。問診は2026-10-17決定で治療プランへ統合する。検査結果の原本・認可・確定保護は維持する。D20／LINE健康表示・自動送付は拡張しない。語彙は[検査の概念](../../ontology/examinations.md)、具体的なread境界と適合差は[技術契約](../../../design/domains/examinations/technical-spec.md#examination-selected-print-contract)へ。

<a id="fr-examinations-selected-print"></a>
## FR-examinations-selected-print：保存済みの実施項目を選んで印刷する

1回の印刷対象は一つの保存済み検査・固定結果版／print-snapshotとし、その版内の項目を選択する。プレビューと出力は同じsnapshot・項目集合・順序を使う。保存済み実施項目を初期選択し、スタッフが項目単位で外す／選び直せる。明示的な未実施項目は印刷候補・出力から除外する。0・陰性・不等式・定性文字列を空／未実施と判定しない。

実施状況、結果の有無、結果の意味、取得状態を分ける。実施済みで結果待ちなら「結果待ち」と識別し、空欄を陰性に変換しない。空値・旧欠測記号・項目テンプレの存在だけで実施／未実施を推定せず、実施状況不明としてプレビューで区別する。不明行は自動選択・黙った未実施扱いをせず、元結果の照合へ案内する。不明行を解決できなくても、実施を確認できた選択行の印刷は妨げず、不明の除外があることをプレビューに示す。元結果の訂正が必要なら通常のD07訂正へ渡し、印刷選択で臨床状態を書き換えない。

項目名・原値・単位・保存当時の基準値・検査日、患者識別、元結果版と訂正／作業版の識別を保持する。項目を減らしても、その選択結果に関係する訂正識別を落とさない。現在マスタの項目や基準値で過去snapshotを再構成しない。選択0件は「印刷対象なし」とし、全項目印刷へ戻したり空帳票を成功扱いしたりしない。

未保存入力の自動保存・確定／解除・患者紐付け・送付は行わない。取消では入力と保存済み結果を保持する。医院・患者・認可・固定版が変われば旧選択を別結果へ適用せず、再取得して同じ対象の選択とプレビューを確認する。版内項目refの検証、旧snapshotの不明条件、再印刷は[選択印刷の技術契約](../../../design/domains/examinations/technical-spec.md#examination-selected-print-contract)へ。

<a id="fr-examinations-infection-summary"></a>
## FR-examinations-infection-summary：保存された感染症検査結果と日付を患者トップへ返す

スタッフ向け患者トップに猫FIV／FeLV、犬mf／HWを項目別に表示する。表示するのは保存された検査結果の記録であり、現在の感染状態の診断・陰性保証や有効期限ではない。結果原値、各結果自身の検査日、院内／他院の出典、保存結果／版の状態を分かるようにし、権限内の元結果・訂正履歴へ辿れる。日付・出典の異なる項目を一回の検査に見せず、患者の特記レベルや自由備考へ混ぜない。

院内は対象患者に保存されたD07結果を自動投影し、別の結果本文を台帳へ手で転記しない。未保存draft、未紐付け受信、保存失敗を保存済み結果として出さない。自動反映のために検査確定を新たな必須操作とせず、未確定／作業版はその状態を明示する。確定解除後の作業版をofficial版として表示せず、原official版への対応・訂正中の識別を保つ。

他院結果は対象患者・検査項目・原値・判明している検査日と出典を、D07が所有する結果記録へ手入力する。院内受信と偽装せず「他院・手入力」と識別し、入力者／入力時刻を実施日時と分ける。医院名・検査日が不明な旧資料を架空の値で埋めず「出典詳細不明」「検査日不明」とする。問診の自己申告による検査歴は未確認の申告として区別し、それだけでD07の保存結果／official版や患者トップの検査結果へ自動昇格させない。元結果を確認して他院結果として記録する場合は本D07入力・版履歴の経路を使う。証拠ファイルの添付、新しい医師確認役割、別の必須提出物はこの採用から追加しない。担当医任意と認証actorログは既存契約を継承する。

表示は検査日と出典に基づいて複数の記録を識別できるものとし、保存／入力の新しさだけで過去実施の結果を最新へ昇格しない。項目ごとに一意な直近検査日の記録を要約できる場合は、その日付と結果を表示する。同日で時刻の確度が足りない複数結果・内容の不一致・院内他院の競合は、医院側の固定優先順位や医学的統合を作らず、候補の結果と各日付／出典を併記して要確認とする。日付不明の結果は直近結果の代わりに自動採用せず、不明の別記録があることを示す。日付不明しかない場合はその原値と「検査日不明」を表示する。現在の履歴先頭20件だけから全患者の直近結果を断定しない。

記録なし、明示未実施、実施済み結果待ち、原値の意味不明、取得失敗／権限不足を区別する。未実施／結果待ちの新しいオーダーは結果値のある候補と分けて表示し、それだけで以前の保存結果を消したり日付を更新したりしない。結果なしを「未検査」と断定せず、陰性・0・陽性は保存された表現／明示対応がある範囲で表示する。名称・値の見かけだけでFIVとFeLV、mfとHWを相互変換せず、未対応は照合対象とする。猫／犬以外又は種の対応不明では別種の結果欄へ転用しない。

訂正保存・再確定・適法な無効化後は同じ元結果との関係を保って再投影し、古い陰性等を現在の結果としてキャッシュし続けない。再取得できなければ更新未確認とし、不変の過去記録を消さず照合する。手入力でも同じ認可・医院／患者・期待版・必要履歴・結果不明照合を適用する。D04/D03の表示権限だけでD07を編集できず、LINE／LIFF、公開健康情報、D21の印刷／共有へ流用しない。

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
| AC-examinations-finalized-parent | 親カルテは診療終了済み又は旧finalized | カルテ本文の通常編集／結果訂正／後日到着 | カルテの旧確定値だけで本文編集を拒否せず変更履歴を保持する。検査自身のconfirmed／official版は通常PATCHで迂回せず専用訂正し、同じ固定結果版参照を保つ | revision |
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
| AC-examinations-clear-copy-guard | カルテ検査タブに原結果・機器job・固定結果版の参照あり | タブクリア／履歴複製Undo | カルテ側対象入力の履歴付き復元と、機器受信・attach/detach・検査確定版を分離する。検査原版削除や実行副作用の逆実行をせず、競合／結果不明は停止照合 | revision |
| AC-examinations-print | 保存済み結果版あり | 印刷 | 同じ保存snapshot、未保存値を自動保存しない | read |
| AC-examinations-owner-handoff | D03カルテ／D21スタッフレポートが参照 | 受渡し | 同じ医院・患者・exam・固定版・取得状態を認可済みスタッフへ返す。D20への検査結果公開は行わない | read |
| AC-examinations-write-failure | 必須監査失敗／競合／認証医院変更 | write | 部分成功を避け原記録保持、古い応答を別scopeへ反映しない | revision |

### 選択印刷・感染症患者トップの追加受入条件

| AC ID | Given | When | Then | FR |
|:---|:---|:---|:---|:---|
| AC-examinations-print-item-selection | 実施済みの陰性・0・不等式と明示未実施が同じ保存版にある | 印刷項目を選ぶ | 実施済みを初期選択し項目単位で変更可。未実施は除外、陰性・0・原値は保持 | selected-print |
| AC-examinations-print-item-unknown | 実施済み結果待ち、空値だけの旧行、実施不明行がある | 候補とプレビューを表示 | 結果待ちと実施不明を区別。不明を自動選択／未実施断定せず、既知の選択行のみ印刷可。不明除外を示し元記録照合へ | selected-print |
| AC-examinations-print-item-snapshot | 訂正された結果と保存当時の単位／範囲がある | 項目を減らして印刷／再印刷 | 同じ固定snapshot・選択項目・順序・検査日・版／訂正識別を保持。現在マスタへ差替えない | selected-print |
| AC-examinations-print-item-empty | 未実施のみ又は選択0件 | 印刷する | 印刷対象なしを表示。全件fallback・空帳票の成功表示なし | selected-print |
| AC-examinations-print-item-version-change | プレビュー後に対象結果／版・scopeが変わる | 出力を要求 | 古い選択を新版・別患者へ移さず再取得・再プレビュー。item名／位置で補完しない | selected-print |
| AC-examinations-print-item-access | 他医院・他患者・別版のitem ref又は閲覧権限不足 | 印刷projectionへ要求 | 出力拒否。表示可能な項目への黙ったすり替えや権限拡張なし | selected-print |
| AC-examinations-print-item-cancel-draft | 未保存入力がある | 印刷準備又は取消 | 未保存非反映を表示、入力保持。保存・確定・解除・attach・送付なし | selected-print |
| AC-examinations-print-item-single-snapshot | 複数日・複数検査の履歴がある | 今回の選択印刷を使う | 一つの保存検査snapshot内の項目選択。新しい複数日比較帳票や全履歴検索を追加しない | selected-print |
| AC-examinations-infection-species | 猫・犬・その他種／種対応不明の患者 | 患者トップを表示 | 猫FIV／FeLV、犬mf／HWを項目別に表示。別種転用・危険度や備考への混同なし | infection-summary |
| AC-examinations-infection-auto-saved | 院内の対象結果を保存済み、未保存draft／未紐付けjobもある | 患者トップへ反映 | 保存済み対象結果だけ自動反映。同じ結果ID／版を参照し、確定を新しい必須条件にしない | infection-summary |
| AC-examinations-infection-external-source | 他院資料の結果を手入力する | 保存・再読込 | 他院／手入力・原値・判明日付／出典を保持。不明を補完せず、actor／入力日時を検査日と分離。問診の未確認自己申告だけでは結果／officialへ自動反映しない | infection-summary |
| AC-examinations-infection-date-order | 古い検査が後日入力され、項目間の日付も異なる | 要約表示 | 入力順で直近を置換せず各結果自身の検査日を併記。一つの検査日へまとめない | infection-summary |
| AC-examinations-infection-ambiguous | 同日競合・複数出典又は日付不明がある | 患者トップの結果を選ぶ | 固定出典優先や医学的統合なし。候補の日付・出典／不明を表示し、一意な最新と断定しない | infection-summary |
| AC-examinations-infection-revision-refresh | 元結果の訂正保存・確定解除・再確定又は適法な無効化がある | 再投影する | 固定版と原official履歴を保ち訂正中／訂正済みを識別。反映失敗は更新未確認、古い値を現在確定値にしない | infection-summary |
| AC-examinations-infection-result-states | 陰性・陽性・0・明示未実施・結果待ち・意味不明がある | 結果を読む | 保存原値と実施／結果状態を区別。空値を陰性にせず、未確定／作業版の状態を隠さない。新しい未実施／結果待ちだけで過去結果を消さない | infection-summary |
| AC-examinations-infection-no-record-failure | 記録0件・取得失敗・不完全な候補取得がある | 要約を取得 | 記録なしと失敗／不完全を区別。先頭20件で全履歴最新・未検査を断定しない | infection-summary |
| AC-examinations-infection-access-boundary | 台帳閲覧のみ／D07権限なし、又は医院切替・認可失効 | 患者トップ／LINE応答を読む | 現在の両owner閲覧認可へ従い、遅延漏洩なし。D20／LIFFへの結果公開・自動送付なし | infection-summary |
| AC-examinations-infection-mapping-unknown | 検査項目とFIV／FeLV／mf／HWの対応が未確認 | 院内結果を投影 | 名前の類似や任意IDで推測せず対応不明として照合。新機器受信や現在マスタで旧版書換えなし | infection-summary |
| AC-examinations-infection-write-safety | 他院手入力／訂正の版競合・必要監査失敗・応答不明 | 保存／再試行 | 原版保持・同一操作照合・二重作成防止。台帳PATCHや通常カルテ編集でD07保護を迂回しない | infection-summary |

## 担当境界と文書完了外
- D03はカルテ確定・追記・患者取り違え訂正・過去入力・復元。D07は検査revision／機器受信／補償。検査usage receiptをD03転記／復元receiptへ流用しない。
- D09が健診の唯一の正系統。検査マスタへ健診の同じ記録を再作成しない。
- D10は請求／訂正／返金、D13は在庫。結果保存・確定で勝手に課金／返金／在庫補正しない。
- D04患者トップはD07の認可済み結果projectionを読む。患者台帳は結果のwrite ownerにならず、他院入力もD07の結果履歴へ渡す。D03の患者ヘッダーも同じprojectionを参照する。
- D20は保存済み接種実績・最終来院日等の限定健康表示を所有し、今回D07の検査結果を公開しない。D21はスタッフ専用Owner Reportの閲覧認可を所有する。D07のviewだけでいずれの認可も代行しない。D21に飼主公開や新しい公開承認操作を追加しない。
- 未決の新規業務追加は本整理に含めない。具体的臨床数値、現場ポート／ケーブル確認、実データ旧版照合、現grant・writer適合は引継ぎ。これらを承認済み／実機PASSと記載しない。


D21の検査履歴sectionはスタッフの閲覧専用であり、同sectionからの印刷・共有を設けない。追加採用の予防薬専用2帳票とは別の範囲である。D03や検査自ドメインに既存の印刷経路があることを、D21検査履歴の印刷許可へ拡張しない。


<a id="examination-normal-edit-autosave"></a>
### カルテの通常編集・自動保存・復元との接続

会議追加第1項（Sentinel_e99d06bd6f0c8191b7bd18fd34d562f9）で診療終了後も通常編集できる目標へ変更した。保存ボタン方式は Sentinel_ffce3491042481918ffb701f9d84547b で撤回し、自動保存を維持する。クリアはタブを開いた時点、履歴複製Undoは複製直前の状態へ戻す回答 Sentinel_6583182163e881919433829835fa137e に従う。[D03正本](../medical-records/lifecycle.md#medical-record-agreement-autosave-20261015)の対象／版・変更履歴・現在権限・失敗／結果不明契約へ接続する。

D03の通常編集は、検査自身の確定解除や原結果変更の権限ではない。検査コメント／参照等のカルテ側入力と、機器の受信・患者紐付け・検査結果版のwrite ownerを区別する。受信済みjob、usage receipt、official版の保護は維持し、タブクリア／複製Undoをdevice Undoやfixture revertに変換しない。検査の原結果とカルテに残す表示／記載の変更を別履歴として追跡する。

タブクリア／履歴複製Undoは対象入力の復元であり、患者・親recordの付替え、保存済み原履歴の物理削除、別タブ／別患者／他者変更の消去、通知の回収、会計／入金・実在庫の逆実行へ拡張しない。現在版・対象ID・由来が復元基準から変わっていれば無警告上書きをせず、既存内容と入力を保持して再取得・確認へ進む。保存結果不明を復元成功や未保存扱いにせず、元操作結果を照合してから続ける。復元できた状態とその前後差分も履歴に残す。独立フォーム／機器実行の自動保存方式を、カルテタブの回答だけで新設・変更しない。

## Backlog原条文の補足受入条件

保存済み原Q/Aと既存業務契約の照合による補足。対応済みという旧回答は実装・実機検証の合格証明に用いない。後続の本人合意、医院scope、権限、履歴・会計訂正の既存条件を維持する。

### FR-examinations-receive／権限前提

| 条件ID | 受入条件 | 根拠 | 検証状態 |
|:---|:---|:---|:---|
| AC-examinations-receive-access-separation | 受信画面の閲覧権限のみ／受信作成権限あり／各紐付け等操作権限なしを比較し、現在の操作別grantに応じて閲覧とwriteを分ける。不足時は必要な操作権限を案内し、閲覧だけで受信・attach・detach・切戻しを許可しない。具体的既存grant名は既存正本へ従い、固定の新ロールを作らない。 | Phase1 F42:G42 | 未実施 |
