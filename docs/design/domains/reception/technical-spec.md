# 受付ドメイン技術仕様

| 管理項目 | 内容 |
|:---|:---|
| 文書ID／役割 | design.reception.technical／確定業務の表示・通信・ブラウザ設定・画面連携を実現する技術設計 |
| 具体化の根拠 | 2026-10-14、Sentinel_6bdf54904d3881918fb3e884820adfea：docs限定の表示配置・異常時動作・受入条件の具体化と独立レビューを承認 |
| 業務の根拠 | 受付3点 Sentinel_ea8569c91d8c8191b70e7b2cf8df8db6／Sentinel_daeb79f859f48191a38d3a61f4f725d8、スタッフ別表示 Sentinel_67648c4b0030819193d97aacb34e9b16、ブラウザ保存 Sentinel_7fc2cd6f8c708191b6ab1d9160cb836b。業務本文は画面正本へ参照 |
| 基準・実装状態 | ローカルmain cea3f94ddeba10632ea372afcbe5d89efedfee36の限定静的読取。以下は将来設計、API／生成型・アプリ・DB・設定は未変更。実機／ブラウザ／並行試験未実施 |
| 管理先 | 本書は技術正本、業務／受入はspec、実施証跡はwork、作業・担当・進捗はPlane。タスクファイル・起票は作らない |

## 1. 正本・責務・保全

[受付画面](../../../spec/domains/reception/pages/reception-board.md)の業務契約・表示結果を実現する。用語・分類は[予約オントロジー](../../../spec/ontology/reservations.md)、状態・記録の接続は[横断フロー](../../../spec/workflows/reservation-to-record-flow.md)が正本。共通の操作版・再送・grant・transactionは[予約技術仕様§4.7](../reservations/technical-spec.md#reservation-api-closed-contract)・[参加規約](../reservations/technical-spec.md#reservation-global-lock-contract)を再定義しない。全体は[architecture](../../../architecture/README.md)、現API wireは[API YAML](../../../../backend/docs/api.yaml)、DBは[既存migrations](../../../../backend/migrations/)へ参照する。

受付はreadの投影と操作orchestratorであり、予約・記録・会計・宿泊の第二write ownerを作らない。表示列・ブラウザ非表示は業務statusの保存場所ではない。カルテ保存契約§2、88ルート、臨床安全、医院／患者／スタッフ分離を保持し、全ドメイン文書完了前に実装しない。保存期間・削除手続きは[共通の当面方針](../../../spec/shared/README.md#interim-retention-policy)へ分離する。

## 2. 現在コードの限定照合

| 参照箇所 | 確認した事実 | 設計との境界 |
|:---|:---|:---|
| [表示日モデル](../../../../frontend/src/features/reception/routes/reception-model.ts) | dateの実在検査、当日fallback、非本日walk-in抑制 | 同じ選択日契約を保持。現在時刻によるstatus変更に使わない |
| [投影](../../../../frontend/src/features/reception/api/transforms.ts) | cancelled/no_show/tentative除外。trimmingのin_consultationは受付済列という旧実装 | 会議追加2の診療中列投影へ追従が必要。4分類の明示接続は未適合／要実装照合。名称で分類を補完しない |
| [進行hook](../../../../frontend/src/features/reception/hooks/use-reception-kanban.ts) | 汎用状態更新、terminalのcompleted更新＋local除外 | 表示のみ片付け、目的別操作／版／再送と競合処理への追従が必要 |
| [詳細handler](../../../../frontend/src/features/reception/hooks/use-reception-modal-handlers.ts) | 独立予約編集と既存進行経路 | 予約管理編集への統一・患者入力保持・開始の準備／読込を実装済みとしない |
| [集計](../../../../frontend/src/features/reception/lib/reception-telemetry.ts) | フィルタ前columns計数、checked_in_at基準の待ち統計 | 選択日raw投影から集計し、表示filter／片付けで業務件数を変えない |
| [画面grant](../../../../frontend/src/features/reception/routes/Reception.tsx) | reservations create/edit/delete、記録・会計・入院等の個別権限 | reception:viewを全操作許可にせず、seedを実医院grantの証拠にしない |

これは全経路の適合・障害再現・性能のPASSではない。

<a id="reception-read-scope"></a>
## 3. 読取・投影と画面context

read contextは認証済み医院ID・スタッフID・解決済みJST表示日・認証世代・request generationから作る。GET reservationsの既存date入口にsnapshot=trueを指定して認証医院の対象日全件を取得する方式を選定する。既存snapshot応答のsnapshot=true、非負整数total、totalとdata件数の一致、医院・JST開始日・各IDを検査する。通常paginationの先頭100件だけを完全な受付母集団にせず、snapshot上限超過・取得不完了はerrorとして件数／待ちを確定しない。上限を新しい業務受付上限へ読み替えない。画面のID／versionは予約boundaryの正確な十進文字列へ揃え、JS Numberで丸めない。将来read adapterの必須項目は予約ID、医院ID、version、予定日時、予約status、明示business_class、既存の認可済みカード概要。専用関連はowner/kind/id/revisionの明示refで返し、専用本文は別view grantで取得する。API正本／生成型への反映は未実施。

分類・関連refを名称／職種・現在petの入院一覧から推測しない。医院外行・重複IDで版／関連が食い違う応答は正常snapshotにせず、再取得・対象操作停止へ。明示分類不明は既存予約を勝手に分類補完せず「分類未確認・操作条件要確認」を表示する。投影は予約ID単位で1回。診察後入院やunlinked入院管理行から新しい予約・カードを自動生成せず、既存診療カードの明示linked対象から同じ専用管理へ接続する。関連refが無ければ専用ownerの既存選択／登録導線で整合を確認し、日付／患者一致で既存滞在と自動統合しない。

表示の処理順は認可済み選択日snapshot → ID整合／既存status除外 → 全体件数・待ち統計 → 表示filter → 本人browser非表示overlay。全体件数は同じ予約IDを1件とし、同じpetの別予約をまとめない。前日データや全宿泊一覧を母集団に足さない。待ち統計は実statusがchecked_inかつchecked_in_at有効な対象のみ、診療中列の施術中in_consultationは加えない。非本日の待ち統計は非表示、対象0件は「—」。取得失敗を0件へ変換しない。

医院／スタッフ変更は以前の患者snapshot・開いているmodal・入力を新scopeへ流用せず、UIから取り除いて新readを完了する。表示日変更は新generationを発行し、遅い旧日応答を新日へmergeしない。modal／未保存案がある日付・scope変更は保持／破棄を確認してから移動し、患者draftはmemory内だけで保存する。認証喪失／grant失効時は旧患者表示と操作を閉じる。logoutでもbrowserの最小表示設定だけは残し、別スタッフではそのscopeを読まない。

30秒pollとfocus再取得の既存方式を保持。初回未取得はloading、初回取得失敗はerror＋再取得、成功0件は「対象日の受付はありません」、filter結果0件は「条件に一致するカードはありません」＋filter解除と区別する。同じcontextの更新失敗では直近正常snapshotを「更新できていません」・最終成功時刻付きで表示できるが、状態変更は再取得成功まで止める。scope変更後に旧医院／旧日をplaceholderとして表示しない。更新中・poll完了でmodalの理由／未保存入力を巻き戻さない。

<a id="reception-operation-contract"></a>
## 4. 操作・版・再送・戻り先

read-onlyのカード選択、filter、片付け／復元と、server業務writeを別dispatchへ分ける。表示列の隣接位置だけから任意status PATCHを作らない。

| UI操作 | owner契約への接続 | 受付側保持・停止条件 |
|:---|:---|:---|
| 受付／再受付 | checkIn／recheckInの版・キー。通常記録participant同時commitは診療のみ。他分類は専用owner境界へ | 同じ予約・記録ref。片方失敗は既存snapshot維持。再送は同じkey、未来院等の現在適格性はowner検査 |
| カード「カルテ作成」 | 準備／既存記録再利用 → 認可付き読込成功 → start intent | 同じボタン内の段階。読込失敗はstartを送らずprepared ref保持。通信結果不明は自動Undoや別記録作成へしない |
| 診療終了ボタン／確認DnD・各記録画面の終了 | 対象分類のownerを通るend intent、確認、期待版、同一操作key | 確認キャンセルはwriteなし。通常カルテ／専用トリミングの終了入口を追加し、終了だけで本文保存／記録確定・請求確定・入金を行わない。専用分類を通常カルテへ変換しない |
| 誤終了取消 | undo、理由・確認・期待版、会計確定歴のowner検査 | 理由draft保持。勤務外Undoと診療継続前の手動担当修正を分離。未収／分割／0円確定後、返金／取消後も普通Undoへ迂回しない |
| 取消／復元／手動未来院 | 各目的別intent、版、キー、状態履歴 | 自動no_show・自動カルテ削除なし。復元／再受付は同じIDと現在指名条件 |
| 予約変更／担当変更 | 予約管理の対象編集modalへ。同じ予約版・警告receipt／手動handoff | 受付独立フォームを第二write窓口にしない。メモだけの変更に日時／指名の未変更値を混ぜない |
| 会計待ちから記録を開く | 通常カルテ／トリミングの明示linked ref・同医院／予約・対象view grant | 既存記録読取のみ。新規作成・状態開始を送らず、ref欠落／取得失敗でnewへfallbackしない |
| 会計／専用管理 | 既存owner入口・明示対象ref・同医院のcontext | 関連ownerのgrant・版・結果を保持。予約statusの汎用進行で会計完了／入院開始／ホテル受入を代用しない |
| 会計済片付け／表示復元 | §5のbrowser overlayのみ | 予約write・会計write・業務監査writeを発火しない。結果照会の操作keyとは別 |

操作memoryは元の医院／actor／予約ID・intent・期待版・業務changes／理由・確認前提・UUID key・準備ref・stage・必要warning receiptを保持する。二重クリックは同じ対象intentの送信中buttonをdisableし同じ試行へ集約する。異なる端末間の競合はserver版／receiptで守り、UI禁止だけを根拠にしない。

状態writeはserver成功確認まで業務列を確定移動しない。DnDのpreviewはdrop/確認キャンセル・既知rollback失敗で元に戻す。通信断／timeout／abortは失敗確定とせず「結果を確認しています」とし、同じ医院actor intent keyの結果照会を行う。成功済みなら初回結果refを保持し最新状態を再取得、claimed_uncommitted／not_foundを未成功確定とせず別key作成へ落とさず元内容で同じkeyを再送する。result_unavailableは同じ成功IDの照合／再取得へ渡し、再作成しない。自動で業務内容を変えたwriteや無制限retryは行わない。

409 version conflictは入力・理由を保持し、現在版／変更点を取得して比較、利用者の再確認後だけ再適用する。warning_changedは根拠を再表示／再確認、禁止へ変われば止める。content_conflictは同keyの内容を変更して成功扱いにしない。transition／record_review／accounting_finalized／assignment等は正本の導線へ案内し、汎用PATCHで抜け道を作らない。同内容の確定拒否後は期待版・warning receiptだけを最新へ更新して同keyを再使用できる（両者は予約fingerprint対象外）。内容を変える次操作は旧結果を解消した後だけ別keyにする。400項目エラーは入力保持、401は認証回復、403はscope／grant再取得（receipt照会403も元操作失敗の証明ではない）、404は対象なし／認可範囲外を漏らさない表示。503依存取得失敗は判定不能として再取得、空き／不足等の確定に置換しない。

write送信後の日付／医院／actor変更やlogoutでは、abortが業務rollbackを保証しない。元contextのreceipt scopeを新contextへ付け替えず、遅い結果で新日カードを変更しない。同じ元医院actorへ認証復帰した後に結果を照合する。患者draft・送信payload・keyはlocalStorageへ永続化しないため、tab reloadで照会keyを失う場合は元医院の予約／記録を再取得する。確実に照合できる同一業務ID・履歴がある場合だけ既存対象へ接続する。0件・類似行・現在statusを元操作失敗の証拠にせず、照合不能なら結果不明のまま新keyによる自動再作成を停止する。先行処理の遅延commitがあり得るため、UUID消失を失敗証拠にしない。操作keyを自動的に全browserへ共有する案は採用しない。

戻り先は内部path検査済みの選択日受付URLとし、医院／対象IDのmemory contextを保持する。URL／browser storageへ患者名・理由・tokenを載せない。遷移先は対象clinicとpermissionを再確認し、医院を切り替えただけで旧対象へwriteしない。患者未linkedなら既存選択導線へ渡す。保存後／戻るでも選択日を復元し、失敗で新規作成へ送り直さない。患者draftを残すか捨てるかの確認は明示的な画面離脱時だけ、正常なpollごとには追加しない。

新resource／grantは作らず[予約の既存対応表](../reservations/technical-spec.md#reservation-api-closed-contract)を参照する。入口reception:viewと対象read reservations:view、予約create/edit/delete、記録・trimming・hospitalization・accounting・owners等の各owner grantを操作ごとに確認する。片付けは認可済みviewの本人表示設定で、reservations:editやaccounting:editを追加必須にしない。スタッフ名／勤務詳細は既存master-staff:view／shifts:view、詳細を取得できないときは医師不適格や勤務0人と推測せず表示を限定する。server側の最新認可・版・状態／関連検査をUI可否で免除しない。

### 会議追加2の接続差分

[業務正本](../../../spec/domains/reception/pages/reception-board.md#reception-meeting-start-end)に従い、トリミングのin_consultationは診療中列へ投影する。施術開始そのものは専用ownerの明示操作で予約writerへ参加させ、投影変更・途中保存・画面readからstatusを書かない。通常カルテとトリミング記録の終了入口も、同じ医院・予約／記録ref・現在版・既存認可・同一操作結果を使う。終了成功と本文保存成功は別に表示し、失敗時に入力を消さない。従来の臨床Undoの条件をトリミングへ自動拡張しない。

会計待ちの記録導線はread-onlyの既存対象解決とする。明示refがない場合や取得失敗にcreate/newを呼ばず、対象確認・再取得を可能にする。古いscopeの成功・失敗を新医院／新予約へ反映せず、既存戻り先の選択日を維持する。API/DTO・全入口・実機受入への追従は未確認。

### 終了対象と実効的な終了結果

終了対象は認証医院・予約ID／期待版と、通常カルテ／専用トリミング記録のowner/kind/id/revisionによる型付きrefで照合する。専用wireのIDが予約IDを表す場合とdetailの内部IDを区別し、数値一致だけで通常カルテIDへ変換しない。予約リンクなし／不明はownerによる対象確認まで終了を保留し、架空の予約を生成せず、同一患者・日付や最新記録から推測選択しない。

終了成功のoperation_resultは既存のoperation_id・intent・result_refs（予約と型付き記録）・committed_versions・committed_atを保持する。予約状態・終了結果／必要監査と成功receiptを同じcommitへ対応付け、記録保存や入金のreceiptと混用しない。初回成功結果とcurrent_state／現在版を分け、後の終了取消／再開後に古い成功receiptだけで「終了済み」と判定しない。D21等の受け手は実効的な終了結果と同じ記録への明示関連を読み、途中保存・本文のlegacy finalized表示・予定終了時刻を代用しない。本文view権限とscopeも別に照合する。具体的wire／DTO・全writer対応は実装前の適合確認で、既存APIが成立済みとはしない。

通常カルテはD03の最新自動保存を維持する。自動保存の失敗・競合・結果不明を終了成功で解消済みにせず、逆に自動保存成功だけで終了済みにしない。診療終了時の治療プラン未実施警告は確認すれば続行でき、自動保存では警告しない。警告根拠は対象記録・現在版へ結び付け、D06の新しい施術完了必須条件にしない。終了に「保存して終了／破棄して終了」を追加しない。

<a id="reception-visit-history-intent-handoff"></a>
### 初診／再診と次回予定の接続差分（2026-10-15）

[予約技術契約§4.8](../reservations/technical-spec.md#reservation-visit-history-intent-contract)へ接続する。バッジ／filterは保存済みvisit_typeを表示し、登録モードやowner単位の履歴から再判定しない。予約編集に渡す候補はpet単位・staff訂正可、unknownを受診歴なしへ変換しない。受診歴は業務正本どおり通常診療・ワクチン・健診を含み、トリミング・ホテルだけの利用を含めない。

snapshotから日付のみtentativeを除外し、日付到来やreadで実予約・受付カードを作らない。別件来院の受付は選択なしで元予定を消化しない。今回消化をスタッフが選ぶ場合だけ明示対象の予定ID／版・今回来院ref／版と選択を同じ操作として渡す。既存今回予約・記録を保持し、重複作成・無断統合なし。一般のcheckIn／recheckInと同じ医院・grant・receipt・record participant保護を維持する。

由来予約の取消は再設定案内を表示できるが、その表示・dismiss・再取得でsetを送信しない。スタッフが新日付を選び設定を確定する別操作で対象版・上書き権限を再検査する。競合・結果不明で消化済み／復活済みを先に表示せず、元contextで同じkeyの結果を照合する。FR／ACは予約正本と横断フローの4点にtraceし、D02へ二重計上しない。全UI／API経路の追従・実機確認は未実施。

<a id="reception-browser-storage-contract"></a>
## 5. ブラウザ非表示の具体設計

同じoriginのlocalStorageを初期方式に選定。namespaceはanimalekarte.reception.hidden.v1、医院・認証済みstaff・予約card IDを十進文字列でキーへ分離し、1カード1entryのhidden=trueを保存する。scope全体の配列をread-modify-writeして他tabの変更を上書きせず、片付けは当該entry設定、表示復元は当該entry削除とする。日付・患者情報・氏名・pet/owner ID・内容・金額・token・操作key・業務statusは保存しない。schema versionはnamespaceで表し、未知／不正key・値は適用せず通常表示へ。

scopeのスタッフIDは検証済み認証profile（現/meのidはstaff ID）から、医院IDは認可済み選択医院から決め、保存値から本人／所属を推測しない。両者が整合するまで設定を適用しない。storage変更は現在scopeと認可済み選択日raw rowsのcompleted card ID一致へだけ適用する。別staff／clinic entryを全消去しない。namespace以外のbrowser情報に触れない。日付をキーに複製せずcard IDを用い、対象日・statusは最新readから判定する。cardが他状態へ訂正されたときはhidden overlayを適用せず見える状態にし、古い設定で未完了を隠さない。これは業務状態の復元ではない。

「片付けたカードも表示」はmemory内の表示toggleで、ONでもhidden entryを一括解除しない。hidden cardには「片付け済み」と「表示を戻す」を置き、本人の1entryだけ解除する。toggleをOFFに戻せば未復元のカードは再び非表示。同ブラウザ・同originの別tabはstorage eventとfocus時再読込で表示設定を再照合する。同entryの片付け／復元が競合すれば最後に成功したstorage状態へ揃えるが、業務versionを増やさず患者dataをtab間送信しない。別browser／端末・別originへ同期しない。

save/read例外・quota・データ削除時は「表示設定を保存できません」等で保持失敗を知らせ、未保存の片付けを保持成功としない。初期read失敗は空のoverlayで通常の認可済みカードを表示する。片付け／復元write失敗は直前のoverlayを維持して再試行へ、全storage削除で救済しない。ブラウザによる保存消失では通常表示へ戻るだけで予約・会計を再実行しない。TTLによる自動expiryや業務情報の保管義務を新設しない。

<a id="reception-verification-contract"></a>
## 6. 検証計画・引継ぎ・完了境界

[画面の受入条件](../../../spec/domains/reception/pages/reception-board.md#acceptance)と、本書のcontext／operation／storageへtraceして実装前に検証する。今回は文書と静的読取だけで、下記は実行していない。

| まとまり | 実装前に必要な検証 | 引継ぎ先 |
|:---|:---|:---|
| read／分類投影 | 4分類、unknown、trim施術中の診療中投影、hotel不整合、専用ref、同じID、unlinked行の非自動生成、日付crossing | reservation readと各専用owner。API DTO／生成型の実適合 |
| UI／scope | 5列・group／filter・piece非表示、集計前後、0件／失敗／stale、日付／医院／staff変更の遅延応答、keyboard／touch／focus | frontend受付。現在のUIを適合済みにしない |
| state operations | check-in原子性、prepare/read/start、各記録画面のend／会計待ち既存ref、end確認／キャンセル、Undo会計境界、版409・key replay、通信断／abort／reload | 予約技術契約の全writer・関連owner。受付だけでDB安全性を宣言しない |
| browser | origin／scope、複数tab、再login、quota／不正値／削除、同entry競合、患者情報・token不保存 | frontend表示設定。サーバー同期を追加しない |
| 認可／戻り先 | 各grantとrevocation、閲覧限定、patient選択、内部URL検査、保存失敗時draft保持 | auth／共有navigation／各owner |
| 性能 | 実件数・全体query回数・再render・storage遅延・pollと同時操作の計測 | 将来の限定実測。SLO／実機PASSは未採用 |

受付の既合意・表示配置・通信／競合・ブラウザ設定の契約は本書と画面正本で具体化する。追加の受付固有業務判断は発見していない。臨床／施術の本文、hotel世話担当の職種・人数、入院専用記録内部、会計訂正／精算、共通保存期間／削除手続きは関連領域の文書完成条件へ渡す。全ドメイン確定後の連携再照合も残す。

文書整理の完了は、業務根拠・表示／異常時結果・技術責務・受入trace・参照と独立レビューが揃うこと。コード実装・API反映・実機受入・運用利用開始とは別で、レビュー結果と文書検査の証跡を完了報告に添える。既存の詳細に由来する未知を実データ確認済みにせず、実装前検証を省かない。

<a id="decision-alignment-20261017"></a>
## 2026-10-17確定仕様との接続

[決定根拠・適用範囲](../../../spec/domains/SPECIFICATION_PROGRESS.md#confirmed-business-decisions-20261017)：ユーザーの2026-10-16〜17 JSTの直接決定、文書更新承認は10月17日03:20 JST。以下は目標仕様・未実行の受入条件であり、実装・実機合格を示さない。

[業務正本・受入条件](../../../spec/domains/accounting/lifecycle.md#doctor-review-sync-billing-20261017)。会計ページは医師確認前から準備し待ち・金額未確定を表示する。確定精算は医師確認後。会計タブチェックが医師確認、診療終了で会計待ちへ進み、受付カードの確認状態は別軸とする。確認後の金額だけの変更は警告了承で可、薬剤処置追加削除・数量変更は再確認まで停止。未確定未手修正はプラン自動同期、手修正済みは値保護と差異判断、確定後は専用訂正。会計からプランへの自動逆書換えなし。通常請求対象は実施済み、持帰薬物販は内容数量確定・準備完了で、予定・実引渡し・在庫出庫と分離する。
