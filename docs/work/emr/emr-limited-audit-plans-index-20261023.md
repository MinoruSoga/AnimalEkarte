# 限定監査と設計作業案の索引（2026-10-23）

**計画整備。製品実装Ready・公開契約採択・実行受入ではない。**

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.limited-audit-plans-index-20261023 |
| 種別 | 監査受領・有限な次作業の索引 |
| 決定状態 | 既存採択仕様を継承。後2資料の新技術案は未採択 |
| 実装状態 | 固定範囲の静的分類のみ。全票・全ACの実装監査完了ではない |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23 |

<a id="limited-audit-scope"></a>
## 対象と資料

基準mainは `171b56575e7ee028e54358dad22ae28574853125`（[採用PR #843](https://github.com/MinoruSoga/AnimalEkarte/pull/843)）。7監査資料はこの固定SHAでの限定静的報告を受領したもの。編集時に全AC・全callerを再監査した記録ではない。#844/#845は別担当の製品修正で、資料中の候補head・状態は監査時点の観測として保持する。main採用や製品受入へ読み替えない。

| 記録 | 対象・残条件 |
|---|---|
| [職員・入退院・配信](audit-next-493-499-504-20261023.md) | 493/499/504。採用型・限定reader/FEと未接続runtimeを区別し、既存ownerへ有限成果を対応付ける |
| [予約ID・パッケージ](audit-next-reservations-20261023.md) | 519/520/521/523/524。既修正の再計上を避け、配信段階依存・17メソッドの証跡・dirty保護・mock再表示を分離 |
| [証明書・院内マニュアル](audit-next-certificates-20261023.md) | 486/487/502。既存read/edit/mockを維持し、writer・artifact・描画・回復・公開/復元wireの残scopeを分離 |
| [診療マスタ・トリミングコース](audit-next-clinical-masters-20261023.md) | 454/455/457。採用CRUD/CAS/auditを維持し、receipt/adoption/固定snapshot、A06/reader引渡しの未接続を分離 |
| [停止支払方法・再診回答](audit-next-payment-methods-revisit-intake-20261023.md) | 459/460。459は465単一ownerの既知残差、460は再診回答の親でマスタ票ではない。段階供給とactivationを区別 |
| [入院プラン・診断マスタclear](audit-next-master-clears-20261023.md) | 528/529。旧clear/明示0は実装済み。新規欠陥0、528の既知未接続UI1件を維持 |
| [共通operation基盤](audit-next-operation-core-20261023.md) | 489。Runner/Gate/Q/ClaimMetadataWriterと未実装ContextGate拡張、consumer/001 fixture受領を分離 |
| [死亡後の生前事実入力の設計作業案](posthumous-contract-plan-20261023.md) | 331/495。既決の生前入力許可を再質問せず、life snapshot・writer lock順・vital専用command/履歴等の未選定成果を列挙 |
| [医師確認集合の設計作業案](doctor-review-contract-plan-20261023.md) | 462/481/484。非活性reader/projection・writer参加案。臨床snapshot供給、医師資格reader、confirm/return公開wire、準備/新規Complete互換、anchor/総lock宣言は未閉鎖 |

監査対象は計19票（493/499/504、519/520/521/523/524、486/487/502、454/455/457、459/460、528/529/489）＋448限定subsetで、計20票の限定確認。設計作業案の関連票をこの母数へ加算しない。

<a id="limited-audit-staff-copy-csv"></a>
## 448の限定補足

448のシフト登録関数ではcopy/CSV入口が未接続。専用画面導線は限定読取では未確認であり、他登録経路全体も未確認。493のcopy/CSV familyとの接点として保持する。全画面・全writerの不存在や未実装を断定せず、既存448/493の必要subsetへ引き渡す。専用シフト画面完成や全登録経路監査済みとはしない。

<a id="limited-audit-status"></a>
## 判定軸と次作業

- **implemented／実装済みsubset**：固定sourceで存在・接続を確認した限定成果。実DB適用・runtime公開・受入PASSを含まない。
- **missing**：各記載対象の未実装、未接続、又はfixture・原子性証跡の不足。本文の対象を読むこと。証跡不足だけをproduction機能不存在へ一般化しない。
- **dependency／依存待ち**：必要な供給subset、実資料、配信・環境receiptを未受領。全親票Done待ちへ置き換えない。
- **未選定**：具体wire・保存/reader対応・総取得順等の技術判断が残る。候補schemaを実在APIや採択済み契約と呼ばない。
- **unverified／未確認**：今回の読取範囲外又は静的追跡未完。不存在・不具合認定ではない。
- **NOT_RUN**：実行状態。未実装・未接続・未選定をこの一語へまとめず、未実行をPASSにしない。

共有RunnerのGate(ctx) errorとQ/ClaimMetadataWriterは実装済み。一方ContextGateの派生contextを返すruntime interface/hook自体は未実装で、consumerのDIだけの問題ではない。既存DB fixture helperの存在は将来の隔離実行必須条件ではなく、001適合と実行許可/環境指定を分離する。

次作業は各資料のAC・有限成果・既存owner対応から選び、[チケット規約](../../ops/plane-ticket-conventions.md)の必要subset/共有hunk/実行ID/claim/1PR完了条件を確認して引き渡す。この索引は実装セッションへの追加割当やReady昇格を行わない。仕様の再選定や共有technical-specの変更も行っていない。

<a id="limited-audit-plane-receipt"></a>
## 既存票への追補の受領範囲

親側で今回Planeへ追補済みなのは448/493/499/504/521/523/524/486/487/502と454/455/457/460の計14票。状態・担当・本文・relationは変更していない。本docs作業からの追加投稿はない。519/520は修正済みsubsetと既存配信条件への依存を確認し、重複追補していない。459は465で同じ既知残差を追跡済みのため重複投稿なし。528/529/489も既存票の具体的な既知残件と重複するため追加コメントなし。追補は計14票のまま。454/455/457は監査時のUATを維持し、状態の一般定義をmain採用へ書き換えていない。

過去153票は本文・コメントのtriageであり、全ACの実装監査ではない。今回19票の限定監査や448の補足と合算して全体完了・全件未着手を主張しない。

<a id="limited-audit-inputs"></a>
## 受領資料と公開時の整形

Library資料5件と追加受領の監査本文4件を読取り、投稿用HTMLの重複、内部の執筆経緯、ローカルscratchの絶対パスを除いた。根拠URL・AC対応・有限な次作業は保持する。後2資料は設計作業案であり、候補の型・総取得順を新たな公開契約として採択していない。設計採択時は[文書責務](../../spec/authoring-guide.md#ownership)に沿いdesign正本へ反映する。

差分・リンク・固定anchorの静的確認と、実装・実行受入は別。今回の成果は本記録10文書の公開であり、製品コード・API YAML・DDL・DB・生成物・設定・正式移行HOLDを変更しない。
