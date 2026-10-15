# 会計・未納

会計一覧・精算・売掛管理を確認する入口です。

## 読む順・業務ルールの所在

まず [全体要件と臨床安全原則](../../specification.md) §2.1 を確認し、次に[請求確定・入金・訂正の正本](lifecycle.md#accounting-agreement-first-five) → [会計語彙](../../ontology/accounting.md) → 以下の画面本文を読みます。

[会計一覧](pages/accounting-list.md) → [精算詳細](pages/accounting-settlement.md)（§2 業務・安全機能、§3.3 再試行）→ [未納一覧](pages/unpaid-list.md)。来院完了の transaction は [横断フロー §5](../../workflows/reservation-to-record-flow.md)、締めの業務は [cash-register.md](../../cash-register.md)を参照します。入口に業務ルールの本文を重複掲載しません。

## 受入条件の所在

画面固有の挙動・バリデーション・保護・API 権限は上記本文、今回合意の受入条件は[8FR／66AC](lifecycle.md#accounting-acceptance)、技術案と現行差分は[会計設計](../../../design/domains/accounting/README.md)。既存の具体的な確認手順は [S08 会計訂正](../../../ops/testing/scenarios/S08-accounting-corrections.md)・[S21 会計並行操作と冪等性](../../../ops/testing/scenarios/S21-accounting-concurrency-idempotency.md)・[S33 領収書](../../../ops/testing/scenarios/S33-receipt-print-pdf.md)を参照します。これらは条件・手順の所在であり、今回の整理で実機受入済みになったことを示しません。実行状態と証跡の入口は [docs/work/](../../../work/README.md)です。

## 未決事項と実装状態

精算詳細のレセプト送信（#240 の着手しない判断）と外部機器連携の未実装を保持します。レジ締め全体の改訂は今回の整理に含みません。 新しい要望は決定前に既存の確定ルールへ混ぜず、決定状態と実装・検証状態を分けて扱います。

<a id="reservation-correction-handoff"></a>
### 予約からの引継ぎ：支払い確定後の訂正

2026-10-14 12:07 JSTの予約対話では、診療終了取消は[支払い確定前・受付操作可能なスタッフ・日付制限なし・理由記録](../../workflows/reservation-to-record-flow.md#consultation-end-undo-before-payment)と合意した。支払い確定後の訂正手順は会計ドメインで決める。カルテ・会計データを保持する既存合意は維持する。

業務5点合意で[未収・分割・0円でも会計確定後の通常Undo停止、未保存入力の非確定、返金・取消後の自動再開なし](../../workflows/reservation-to-record-flow.md#accounting-finalization-undo-boundary)を確定した。受付担当の理由・変更案と記録担当の確認は[記録確認訂正](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-record-reviewed-correction)へ従う。権限識別子・返金／確定後訂正／締め後訂正の内部手順は会計領域の設計。勤務外医師がいる誤終了取消と診療継続前の手動担当修正は17:17合意へ従う。既存の精算・未納・レジ締め仕様をこの入口で改訂せず、今回の文書反映を会計の実装・実機受入の完了と扱わない。

## 共通事項

関連マスタは [支払方法](../../screens/settings/payment-methods.md)・[保険](../../screens/settings/master-insurance.md)を参照します。[共通用語・権限](../../shared/README.md)・[仕様全体の入口](../../README.md)。

## このフォルダの役割と使い方

会計・未納領域の業務ルール・受入条件・未決事項を辿る入口。

pagesに画面詳細を置き、保存契約など領域固有本文は既存の配置を維持する。横断フロー・共通マスタ・語彙をここで再定義しない。

子フォルダ：[pages/](pages/README.md)

主要文書：主要文書は本書の既存索引・上記の読む順を参照します。

読む順：本書の「読む順・業務ルールの所在」 → [画面詳細](pages/README.md) → 関連する横断契約・マスタ。

正本・状態：[決定状態と検証状態の規約](../../authoring-guide.md#status)に従い、各節・FRの承認根拠と実装／実機状態を別に読む。READMEへの掲載・フォルダ配置は承認や実装済みを表さず、対象外の既存仕様を新たに確定しない。既存の別名入口・索引の責務は維持し、本書は案内を補う。

### 予約設計からの引継ぎ：会計の同時更新参加

docs限定承認 Sentinel_f19ab03e8c2c81919c6f589bcf563811に基づき、[受け手の返却契約](../../../design/domains/reservations/technical-spec.md#reservation-receiving-contracts)と[全writerの共通gate／下位取得順](../../../design/domains/reservations/technical-spec.md#reservation-global-lock-contract)を技術設計として選定した。不変の会計確定イベントと正確な予約IDリンクを同時commitし、返金・取消・soft-delete後も確定歴を消さない。完了だけでなく返金・明細訂正・取消・締めのwriterも既存lock前からgate参加する。精算・未納・返金／締め後訂正の業務本文は維持し、実装・並行検証と会計内部の文書完成は別条件。

## カルテ訂正・過去入力からの引継ぎ

[カルテ追加合意](../medical-records/lifecycle.md#medical-record-agreement-correction-four)の実施取消・患者訂正で請求に影響するときは会計訂正へ接続する。過去専用入力には自動課金せず、必要な請求は会計で別途確認する。非表示／復元や転記だけで請求生成・返金・未納消去・精算確定歴解除をしない。具体的な請求／返金・締め後手順と権限はD10・X06／07の残件で、この案内を承認済み新手順としない。

## 訂正影響表示と会計側の確認

[追加3項目](../medical-records/pages/medical-record-form.md#requirement-fr-medical-records-correction-owner-impact)に従い、カルテ訂正の影響を表示して会計訂正入口へ案内する。元記録／項目と訂正履歴、関連請求／見積ID・現在版・理由を渡し、会計側で認可・精算確定歴・金額・締めを再確認する。カルテ側の訂正成功・画面遷移で請求訂正／返金成功と表示せずowner receiptを照合。確定済み請求・返金の自動変更なし。未納／分割／0円・取消／返金後・締め後の内部手順はD10／12／X06・07の文書判断へ残す。

## 最終照合：医師確認と会計訂正の返却

[現DTO／目標contract](../../../design/domains/medical-records/technical-spec.md#medical-record-reconciled-contract)へ接続する。confirmation ID＋record IDとaccounting IDは別。確認状態・確認／差戻しactor／時刻と、確認依存版の目標を分離する。現有効スタッフ所属検査を医院の医師確認済みとみなさず、実認可／確認医師対応は利用開始前に確認する。GET GetOrCreate初期化はwriter候補で、印刷用readは副作用なしprojectionへ。会計訂正の受渡しは対象ID／版・元訂正ref・理由・影響確認状態→owner側の現在検査／receipt。未納・返金・締め後の内部判断はD10／12へ残し、カルテの成功／案内で金銭処理完了としない。

## D10初回〜第四群の反映範囲

[初回](lifecycle.md#accounting-agreement-first-five)・[第二群](lifecycle.md#accounting-agreement-second)・[第三群](lifecycle.md#accounting-agreement-third)・[第四群](lifecycle.md#accounting-agreement-fourth)を反映し計8FR／66AC。既決の引継ぎ境界を維持し、内部判断の確定部分は正本、残件は[未決5群](lifecycle.md#accounting-open-decisions)へ参照する。

預り金は飼主単位で予約参照は任意。同飼主の別予約／別ペット会計へ、充当先と金額を確認して使える。元飼主・任意の元予約・受領／未充当残／利用可能額／充当／実返金を追跡し、二重入金・二重使用なし。別飼主への移管は未承認、全ホテル予約の前金必須を定義しない。

割合合算→金額→割引後の保険対象額、古い請求順／同日作成順・理由付き保存後再配分、非現金超過の返金必要、元カード→現金等の別方法返金／金額追加承認なし、受領証・最終会計充当／今回入金・返金元取引番号／再発行番号維持を合意。実決済サービスの自動連携／規約適合を承認済みにしない。

第四群で明細・会計全体の両割引、超過警告、会計取消時の残高復元と別返金を反映。D10は進行中。税・端数／両粒度の併用順序／配賦／警告後金額、取消以外の減額時の充当調整、具体権限・帳票詳細・各masterの全体、他owner内部・全体横断照合、実装・実機は残件。D05ホテル料金／キャンセル料・D12締め内部を分離し、通常Undo停止とカルテ保存§2を維持。
