# 会計：請求確定・入金・訂正の契約

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | lifecycle.accounting／業務・保存境界 |
| 決定状態 | 初回5FR＋第二群3FR（計8FR）は合意済み。未決は末尾に分離 |
| 実装状態 | 未確認。現行との差分は技術仕様に明示 |
| 実機検証状態 | 未実施 |
| 文書確認日 | 2026-10-14 |

<a id="accounting-agreement-first-five"></a>
## 目的・対象範囲

同じ飼主の複数ペット・当日／過去の未請求分と、確定済み請求への回収を区別し、請求・入金・残額を追跡する。飼主による支払額の選択を全額決済の強制へ読み替えない。工程削減の対象は請求の作り直しと元明細の二重入力。所要時間・操作数の実測は未実施。

決定者は依頼ユーザー。根拠は Sentinel_51dfbdb2423c8191a01a32d1f934b91b の「1 はい はい（過去の未払い分などいくら払うかは自由に決められる）／2〜5 ok」、単価手修正・複数割引併用の追加確認 Sentinel_dd33527fe92c8191ab9b6c0f85b0e2d6 の「yes」。docs反映承認は Sentinel_c71afbfafbe0819198fb3f69dab9f53a。2026-10-14に反映。個人名の追記・優先度・数値目標は未確認で、コードや旧文書から補完しない。

本書が初回5点と第二群の業務正本。[語彙](../../ontology/accounting.md)は定義・関係、[技術仕様](../../../design/domains/accounting/technical-spec.md)は実現案と現行差分。[画面](pages/accounting-settlement.md)の継承本文・API表は旧挙動の参照で、今回の合意と異なる箇所を現行制約として区別する。保険オンライン請求や機器連携を追加しない。

第二群の根拠は Sentinel_e6565c89fd488191982ad524c997fe6c の質問へのユーザー Sentinel_6d647c8c719c81919daecc6db67d9455（1は前金を使う旨の修正、3 yes／4 ok／5 yes）と、Sentinel_b4d6e8ae144081919d9e19d797a2faba の訂正例・表示・前金説明へのユーザー Sentinel_67776f0e70c881919019002d47011b1e のok。前金を扱わない案は不採用。第2問の「どういうこと？」は承認とせず、後続説明へのokを根拠とする。下記の更新を第二群として2026-10-14に反映。

## 担当者・権限・前提

請求訂正・取消・返金は権限のあるスタッフが理由付きで行う。実際の職種／権限識別子、承認段階と金額別制限は未確認。現在のgrantを新しい全操作の認可へ無条件転用しない。医院・飼主・ペットの分離と原記録保護を維持する。

## 契機・入口・出口

[会計一覧](pages/accounting-list.md)・[精算詳細](pages/accounting-settlement.md)・[未納一覧](pages/unpaid-list.md)から請求と入金履歴へ接続。通常診療・専用記録・見積の明細を受けるが、臨床確認と会計確定は別。画面配置・新URL・API wireは今回確定しない。

<a id="accounting-requirements"></a>
## 状態遷移・引継ぎデータ

<a id="requirement-fr-accounting-owner-selected-charges"></a>
### FR-accounting-owner-selected-charges：飼主単位の選択と元請求への回収

同じ飼主の複数ペットをまとめて会計でき、当日分と過去の未請求分を選択して合算できる。元のペット・診療等の明細情報を保持し、まとめるために一つのペット／カルテへ付け替えない。

過去の確定済み未納は未請求候補とは別。払う額は自由に決められるが、元の確定済み請求を作り直さず、入金を元請求へ紐づけて残額を管理する。新しい請求と過去分回収を一度に扱う場合も、それぞれの請求帰属を保持する。第二群で[古い請求からの充当・スタッフ変更と釣銭](#requirement-fr-accounting-oldest-payment-allocation)、[予約への前金](#requirement-fr-accounting-reservation-advance)を合意した。自由な支払額の合意を全額回収の強制へ戻さない。予約なし預り金や自動繰越は未決。

<a id="requirement-fr-accounting-receivable-partial-zero"></a>
### FR-accounting-receivable-partial-zero：請求確定と入金状態を分ける

全額未納・一部入金でも請求を確定でき、残額は後日回収できる。請求額・入金履歴・残額を管理する。複数の支払方法を併用する決済と、後日回収を伴う部分入金は別の概念。請求確定を全額入金済みと表示しない。

0円会計は無料・割引等の根拠を残し、架空の入金を作らない。明細のない誤作成の確定可否・根拠入力の形式は未決。通常の入金だけの例では10,000円請求に3,000円入金なら残額7,000円、後日2,000円入金なら残額5,000円で、請求を再作成しない。

<a id="requirement-fr-accounting-reasoned-financial-correction"></a>
### FR-accounting-reasoned-financial-correction：訂正・取消・返金を区別する

権限のあるスタッフが理由付きで請求訂正・取消・返金を行い、元記録と履歴を残す。返金額は入金済み額から返金済み額を差し引いた範囲内。返金方法を指定する。実際に返金していない操作を返金済みとしない。

請求取消だけで返金や在庫返却を自動処理しない。第二群で請求減額と実際の返金を別操作とした。支払方法の誤記訂正は金額を動かさず履歴を残す。返金方法ごとの上限・元と異なる方法への返金、取消後の再請求手順は未決。返金の上限を請求総額だけから算定して未入金分まで返金可能としない。

訂正後の料金に対して未払額と返金必要額を表示し、実際に返金したときに方法と金額を記録する。通常請求への入金だけの例：10,000円全額入金後に9,000円へ減額すると、未払0円・返金必要1,000円。減額だけでは返金済みにならず、1,000円を実返金して初めて返金済みとなる。10,000円請求に3,000円入金後、9,000円へ減額なら未払6,000円で返金不要。料金訂正・返金必要・未実行返金・返金済みを区別する。入金・実返金・前金充当の帰属は別記録として保持し、未充当前金を請求の入金済み額へ混ぜない。

<a id="requirement-fr-accounting-post-close-separate-movement"></a>
### FR-accounting-post-close-separate-movement：締め後差分と金銭移動

締め後も元の締め記録を保存して訂正差分を追加する。金額訂正と実際の入金・返金を別々に記録し、金額だけの訂正を現金の移動にしない。[締め内部](../../cash-register.md)はD12の責務。元の請求日を変えず、今日の実際の入金・返金は今日の入出金として扱う。前金の実受領日／実返金日も保持し、後日の充当日を再受領日としない。期間・区分・差額照合の内部詳細を本書で決めず、元締めの取消／再開という新経路も追加しない。

<a id="requirement-fr-accounting-calculation-snapshot"></a>
### FR-accounting-calculation-snapshot：会計時点の根拠と手修正・併用

会計時点の価格・割引・保険等の計算根拠を保存し、マスタ変更だけで過去会計が変化しない。単価の手修正と複数割引の併用を認める。旧マスタ値／現在値／手修正値の区別を保持する。

第二群で割合割引→金額割引の順を合意。税・保険をまだ適用しない割引段階だけの例：10,000円に10%と500円を適用すると10,000→9,000→8,500円。保険の対象項目・割合・上限はスタッフが確認・修正でき、会計時点の確認／修正値を保持する。

複数割合の合算／順次、行／請求全体等の適用粒度、保険と割引の前後、税・端数処理の具体的な順序・丸め方、保険の具体的条件や修正権限mappingは未決。既存の最大割引選択・四捨五入・切捨てや旧割合の存在を今回の承認にしない。税務上の正しさ・法令適合は未確認で、医院の会計責任者等による根拠確認へ残す。

<a id="accounting-agreement-second"></a>
## 第二群：配分・前金・帳票の追加要件

<a id="requirement-fr-accounting-oldest-payment-allocation"></a>
### FR-accounting-oldest-payment-allocation：古い請求から充当・変更可能

過去未納への入金は古い請求から自動充当し、スタッフが配分を変更できる。A8,000円がB5,000円より古く今回5,000円を払うなら、初期配分はAへ5,000円（A残3,000円・B残5,000円）。スタッフがA2,000円・B3,000円に変更したらその配分を元請求へ記録する。合計入金額を配分変更だけで増減しない。

通常の請求回収で請求残額を超えて受け取った分はお釣りとして返し、請求へ過大充当しない。残額3,000円に5,000円を預かったら入金は3,000円・釣銭2,000円。明示的な予約前金の受領は別操作で、この釣銭規則を前金除外に使わない。同日請求の同順位処理、新規請求と過去回収を一度に扱う際の選択範囲、非現金の超過受領の訂正方法は未決。釣銭と後日の返金を混同しない。

<a id="requirement-fr-accounting-reservation-advance"></a>
### FR-accounting-reservation-advance：予約前金の受領・充当・返金

ホテルなどで前金を扱う。前金は飼主・対象予約に紐づけて受領し、最終会計で充当する。元予約・飼主と、受領額／実受領日、未充当残、充当先請求・額／充当日、実返金額／実返金日を追跡する。医院・飼主・予約の整合を維持し、未充当の前金を過去未納へ自動充当しない。

受領と充当は別。5,000円を前金として受領し、最終会計8,000円へ5,000円を充当したら、追加支払前の未払は3,000円、未充当前金は0円。充当時に5,000円をもう一度現金入金として二重計上しない。実際の受領日を充当日に書き換えない。

キャンセルや料金変更で余った分は返金できる。前金5,000円を受領し、確認された最終料金が3,000円なら3,000円を充当し未充当残2,000円を返金できる。料金変更だけで2,000円を実返金済みにしない。予約キャンセルだけで自動返金せず、権限・理由・方法と実際の返金記録を保持し、返金済み分を再充当／再返金しない。

ホテル固有の前金必要額・受領タイミング・キャンセル料はD05、別予約への振替・予約なし預り金・一予約の複数ペット会計への配分・支払方法別制約は未決。前金受領だけで請求確定・予約完了・通常診療終了Undoの停止を発生させない。最終会計を確定した場合は既決の確定歴境界へ従う。

<a id="requirement-fr-accounting-document-and-aged-debt"></a>
### FR-accounting-document-and-aged-debt：請求書・領収書と過去未納

請求書と実際に受け取った金額の領収書を分ける。再発行は再発行と表示し、訂正前の帳票も履歴に残す。請求10,000円・実入金3,000円の場合、請求と受領した3,000円を区別し、全額受領した領収書を出さない。

未納一覧は今月売上がなくても過去未払いのある飼主を表示する。今月の売上有無を理由に未納を除外しない。検索／期間filter・基準日の詳細は別途設計するが、従来の「期間内売上ありのみ」を目標仕様として残さない。

前金受領時／充当時の帳票、番号／宛名・訂正／返金表示・部分返金再発行・発行操作の権限／保存形式は未決。前金の充当を新たな実受領とする帳票は出さない。

## 異常系・保存・取消

[会計確定歴による通常Undo停止](../../workflows/reservation-to-record-flow.md#accounting-finalization-undo-boundary)を維持する。未納・部分入金・0円でも会計確定歴があれば通常Undo不可。返金・取消・削除後も確定歴を消さない。未保存入力は会計確定ではない。複数ペットをまとめても、正確な予約／記録との対応を失わない。

[カルテ保存正本§2](../medical-records/lifecycle.md#hybrid-sync)は変更しない。カルテ訂正・過去入力・復元から請求・返金・在庫を自動巻戻ししない。金銭処理の成否は会計側で確認し、カルテ保存成功や画面遷移で代用しない。

再送・競合・結果不明・監査の技術案は[保存契約](../../../design/domains/accounting/technical-spec.md#accounting-operation-contract)を参照。未決の配分・計算条件に依存する操作は推測で実行できる仕様にしない。

<a id="accounting-acceptance"></a>
## 受入条件

全条件は文書上の期待結果。実装・実機は未確認／未実施。確認手順は対象の請求・入金・履歴・元明細を照合する設計上のシナリオで、現行画面で実行可能との宣言ではない。

| 条件ID | 前提／操作 | 期待結果 | 要件／確認手順・証跡 |
|:---|:---|:---|:---|
| AC-accounting-owner-multi-pet | 同一飼主の2ペットの明細を選択し会計 | 合算可能。各行の元ペット・記録情報を保持 | [FR](#requirement-fr-accounting-owner-selected-charges)。合計と行の帰属を照合、実機未実施 |
| AC-accounting-selected-unbilled | 当日分と過去未請求分の一部を選択 | 選択分だけ合算し、未選択分を勝手に請求しない | [FR](#requirement-fr-accounting-owner-selected-charges)。選択前後と元明細を照合 |
| AC-accounting-old-bill-collection | 過去の確定請求10,000円から2,000円だけ回収 | 入金を元請求へ記録し残額8,000円。請求を作り直さない | [FR](#requirement-fr-accounting-owner-selected-charges)。請求ID・入金・残額を照合 |
| AC-accounting-no-invented-allocation | 第二群で配分・前金が承認された後 | 古い順・スタッフ変更・予約前金は追加FRへ。未決の別予約振替・予約なし預り金等を勝手に許可しない | [FR](#requirement-fr-accounting-owner-selected-charges)。仕様と技術案の未決境界を確認 |
| AC-accounting-unpaid-finalize | 入金なしで10,000円を請求確定 | 確定と未納を両立し残額10,000円を管理 | [FR](#requirement-fr-accounting-receivable-partial-zero)。請求状態と入金履歴を照合 |
| AC-accounting-partial-finalize | 10,000円を請求確定し3,000円入金 | 入金履歴と残額7,000円を管理、全額入金を強制しない | [FR](#requirement-fr-accounting-receivable-partial-zero)。確定・入金を別に照合 |
| AC-accounting-later-payment | 前条件へ後日2,000円入金 | 元請求の履歴へ追加し残額5,000円、二重請求なし | [FR](#requirement-fr-accounting-receivable-partial-zero)。初回／後日入金を照合 |
| AC-accounting-zero-no-fake-payment | 割引等により0円を確定 | 根拠を保持し架空の入金なし | [FR](#requirement-fr-accounting-receivable-partial-zero)。根拠と入金履歴を照合 |
| AC-accounting-finalization-undo-preserved | 未納・部分入金・0円を確定し、さらに取消／返金 | 通常診療終了Undoは再開せず確定歴を保持 | [横断正本](../../workflows/reservation-to-record-flow.md#accounting-finalization-undo-boundary)。履歴と判定を照合 |
| AC-accounting-correction-authorization | 無権限または理由なしで訂正・取消・返金を要求 | 成立させず元記録を保持 | [FR](#requirement-fr-accounting-reasoned-financial-correction)。各操作で拒否・履歴を照合 |
| AC-accounting-correction-history | 権限と理由付きで請求を訂正 | 元記録と変更履歴を保持 | [FR](#requirement-fr-accounting-reasoned-financial-correction)。変更前後・理由・actorを照合 |
| AC-accounting-refund-received-cap | 3,000円入金・1,000円返金済み | 残り返金可能額は2,000円以内。未入金分は返金不可 | [FR](#requirement-fr-accounting-reasoned-financial-correction)。入金・返金累計を照合 |
| AC-accounting-refund-method-required | 返金方法を指定しない | 返金を成立させない | [FR](#requirement-fr-accounting-reasoned-financial-correction)。入力と結果を照合 |
| AC-accounting-cancel-no-money-stock | 理由付きで請求取消 | 取消だけで実際の返金・在庫返却を自動処理しない | [FR](#requirement-fr-accounting-reasoned-financial-correction)。各owner結果を照合 |
| AC-accounting-post-close-preserve | 締め後に請求金額を訂正 | 元締めを保持し差分を追記 | [FR](#requirement-fr-accounting-post-close-separate-movement)。元締め・差分を照合 |
| AC-accounting-correction-no-cash-movement | 金銭授受なしで金額だけ訂正 | 訂正と実際の入金・返金を別記録、架空の移動なし | [FR](#requirement-fr-accounting-post-close-separate-movement)。差分と金銭履歴を照合 |
| AC-accounting-master-change-stable | 会計後に価格・割引・保険マスタ変更 | 保存済み過去会計と根拠は変化しない | [FR](#requirement-fr-accounting-calculation-snapshot)。変更前後の過去会計を照合 |
| AC-accounting-manual-unit-price | 単価を手修正して会計 | 手修正を許可し会計時点の根拠を保持 | [FR](#requirement-fr-accounting-calculation-snapshot)。元値・修正値を照合 |
| AC-accounting-discounts-combinable | 複数の割引を適用対象にする | 併用を許可し割合→金額の順。複数割合の合算／順次等の未決を勝手に確定しない | [FR](#requirement-fr-accounting-calculation-snapshot)。適用対象・根拠と未決表示を照合 |
| AC-accounting-clinical-no-auto-finance | カルテ訂正・過去入力・復元を行う | 請求・返金・在庫を自動巻戻しせず会計確認へ | [保存境界](#異常系保存取消)。会計側の結果を別に照合 |
| AC-accounting-oldest-default | 古いA8,000円・新しいB5,000円に計5,000円入金 | 初期充当A5,000円、A残3,000円・B残5,000円 | [FR](#requirement-fr-accounting-oldest-payment-allocation)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-allocation-staff-change | 前条件をA2,000円・B3,000円へ変更 | 元請求へ変更配分を記録。合計入金は5,000円のまま | [FR](#requirement-fr-accounting-oldest-payment-allocation)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-collection-change-return | 通常回収で残額3,000円に5,000円を預かる | 入金3,000円・釣銭2,000円。超過分を請求／前金へ無断充当しない | [FR](#requirement-fr-accounting-oldest-payment-allocation)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-advance-owner-reservation | 予約前金5,000円を受領 | 医院・飼主・対象予約・受領額／実日付を保持 | [FR](#requirement-fr-accounting-reservation-advance)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-advance-unallocated-balance | 前金5,000円を受領しまだ会計へ充当しない | 未充当残5,000円。請求入金／予約完了／請求確定歴と混同しない | [FR](#requirement-fr-accounting-reservation-advance)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-advance-apply-no-double-receipt | 前金5,000円を最終請求8,000円へ充当 | 未払3,000円・未充当残0円。充当で追加の現金受領なし | [FR](#requirement-fr-accounting-reservation-advance)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-advance-surplus-pending | 前金5,000円・確認された最終料金3,000円へ充当 | 未充当残2,000円を返金可能として表示し、実返金済みにしない | [FR](#requirement-fr-accounting-reservation-advance)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-advance-actual-refund | 前条件の2,000円を権限・理由・方法付きで実返金 | 実返金と日付を記録し残2,000円を再充当／再返金しない | [FR](#requirement-fr-accounting-reservation-advance)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-advance-duplicate-use | 同じ前金残へ二重充当／充当と返金が競合 | 受領済み残を超えて成立せず履歴・残高を整合させる | [FR](#requirement-fr-accounting-reservation-advance)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-advance-context-mismatch | 異医院・別飼主・未決の別予約へ前金を充当要求 | 紐づき不整合や未承認振替を成立させない | [FR](#requirement-fr-accounting-reservation-advance)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-paid-reduction-refund-pending | 10,000円全額入金後に請求9,000円へ訂正 | 未払0円・返金必要1,000円を表示、返金済みにはしない | [FR](#requirement-fr-accounting-reasoned-financial-correction)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-paid-reduction-actual-refund | 前条件で1,000円を実返金 | 方法・金額・実日付を記録し返金必要額を更新 | [FR](#requirement-fr-accounting-reasoned-financial-correction)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-partial-paid-reduction | 10,000円請求・3,000円入金後に9,000円へ訂正 | 未払6,000円・返金不要。減額だけで現金を動かさない | [FR](#requirement-fr-accounting-reasoned-financial-correction)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-payment-method-correction-no-money | 支払方法の誤記を訂正 | 入金額を動かさず元記録と訂正履歴を保持 | [FR](#requirement-fr-accounting-reasoned-financial-correction)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-percent-before-fixed | 割引段階で10,000円に10%＋500円 | 割合→金額で8,500円。税／保険を含む最終額と混同しない | [FR](#requirement-fr-accounting-calculation-snapshot)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-insurance-staff-review | スタッフが保険対象項目・割合・上限を確認修正 | 確認修正可能で会計時点の根拠を保存。未決の算定順を固定しない | [FR](#requirement-fr-accounting-calculation-snapshot)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-invoice-receipt-separated | 10,000円請求・実入金3,000円で帳票発行 | 請求書と受領3,000円の領収書を分け、全額受領としない | [FR](#requirement-fr-accounting-document-and-aged-debt)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-document-reissue-history | 帳票を訂正・再発行 | 再発行と表示し訂正前帳票を履歴に保持 | [FR](#requirement-fr-accounting-document-and-aged-debt)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-aged-debt-without-current-sales | 今月売上なし・過去未納ありの飼主 | 未納一覧に表示。今月売上なしを理由に除外しない | [FR](#requirement-fr-accounting-document-and-aged-debt)。対象・残高・履歴を照合、実機未実施 |
| AC-accounting-original-bill-and-movement-date | 過去請求へ今日実入金／実返金、後日前金充当 | 原請求日と実入出金日を保持、充当を再受領にしない。締め内部D12 | [FR](#requirement-fr-accounting-post-close-separate-movement)。対象・残高・履歴を照合、実機未実施 |

<a id="accounting-open-decisions"></a>
## 未決事項・関連仕様・根拠

初回20ACはIDを保持し、配分未承認／順序未決の条件を第二群の決定へ更新した。第二群20ACを追加して計8FR／40AC。未決は件数を減らすために推測で埋めない。D10全体は進行中。合意済みの許可・帳票分離・古い順・実日付を再質問しない。

1. 前金の専用業務：ホテルの必要額・受領時期・キャンセル料はD05。別予約振替・予約なし預り金・複数ペット会計への充当方法、前金の権限と帳票は未決。
2. 配分の例外：同日請求の同順位、新規請求＋過去回収の選択範囲、非現金の超過受領訂正方法、確定後の配分訂正・履歴。古い順／スタッフ変更と通常回収の釣銭は合意済み。
3. 計算の残件：複数割合の合算／順次、行／請求全体の粒度、保険と割引の前後、税・丸め段階、保険の具体値／条件と修正権限。割合→金額とスタッフ確認修正は合意済み。税務適合は未確認。
4. 訂正・帳票・権限：取消後の再請求、別方法返金／方法別制約、金額別承認、前金や部分返金時の帳票、番号／宛名・保存形式、督促／貸倒。減額と実返金の区別・訂正後表示・支払誤記の金額不変・再発行表示／旧帳票履歴は合意済み。
5. 受け手・集計：D05／06専用記録・D11見積の受渡し、D12締め区分／差分突合・D16集計、D13在庫、実支払機器の照合。原請求日保持と実入出金日の扱いは合意済み、締め内部は未完了。

[領域入口](README.md)・[予約／受付との横断正本](../../workflows/reservation-to-record-flow.md)・[カルテ保存正本](../medical-records/lifecycle.md#hybrid-sync)を維持。保険オンライン請求の着手しない判断と外部機器の未実装は[画面の既存記載](pages/accounting-settlement.md#decisions-evidence)を参照。
