# 検査結果・選択印刷・患者トップの用語と関係

2026-10-15追加採用の語彙。業務正本は[D07](../domains/examinations/lifecycle.md#examination-print-infection-agreement-20261015)、[技術契約](../../design/domains/examinations/technical-spec.md#examination-selected-print-contract)。物理schema・実装済み・臨床的意味の承認を表さない。

| 概念 | 意味・多重度 | 境界 |
|:---|:---|:---|
| 検査結果 | 一患者・医院に属するD07結果記録。複数項目、複数の保存版／訂正履歴を持ち得る | 原結果と保存版を保持。台帳表示用の別結果本文を作らない |
| 版内項目ref | 一つの固定結果版／snapshotに属する行の識別 | field名・表示順・別版のitem IDと同一視しない |
| 実施状況 | 実施済み・明示未実施・不明の区別 | 空値・0・陰性・欠測表現、検査全体statusだけでは決めない |
| 結果状態 | 原値あり・結果待ち・原値の意味不明等 | 実施済み結果待ちと未実施を区別。医学的な診断や正常保証ではない |
| 選択印刷 | 一検査の固定保存snapshotから版内項目集合を選ぶread | 未実施除外。臨床状態変更・自動保存・確定／解除・複数日比較帳票ではない |
| 感染症項目 | 猫FIV／FeLV、犬mf／HWの各記録上の項目 | 種と保存fieldの明示対応が必要。名称似・閾値による推測なし |
| 患者トップ要約 | D04/D03の認可済みスタッフが読むD07保存結果のprojection | 原値・検査日・出典・版／訂正状態を保持。danger_levelと別 |
| 院内／他院の出典 | 保存結果の由来。院内保存は自動投影、他院は手入力で記録 | 他院を院内機器受信に変換しない。出典詳細不明を許容して識別 |
| 検査日／日時精度 | 実際の検査／測定の日付又は日時と分かる精度 | 入力・受信・更新時刻とは別。不明を当日や架空の時刻で補わない |
| 直近候補と曖昧さ | 判明検査日の順に扱う候補。同日順序不明・複数出典・不一致は複数を保持 | 医学的統合や出典優先を作らず、日時不明を最新へ自動昇格しない |
| 訂正中／official | 同じ元結果の作業版と保存された公式版の状態 | 作業保存を公式確定へ変換せず、旧公式版を破壊しない |
| 記録なし／未検査／取得不能 | 読取0件、明示された実施情報、読取失敗は別の概念 | 記録なしや権限不足だけで未検査・陰性を断定しない |

患者トップの読取権限と結果のwrite権限を分ける。D20／LINEの公開健康情報、D21の印刷／共有、追加機器接続はこの採用の範囲外。通常カルテの構成は[D03の問診統合](../domains/medical-records/lifecycle.md#integrated-plan-20261017)に従う。旧9タブ/問診別タブを維持する根拠にしない。

<a id="clinical-results-and-certificates-boundary"></a>
## 計画・結果・自己申告・接種・健診の関係

検査の予定/注文とD07の保存結果、スタッフ聞取又はLINEの自己申告、D08の接種事実、D09の健診結果を同じ実体にしない。D07の固定結果版をD03/D04が読む場合も元結果・項目ref・版・訂正状態を保持し、表示先ごとに結果本文を複製して独立編集しない。空欄・未実施・結果待ち・陰性・0・取得失敗は異なる。

[検査印刷/感染症投影](../domains/examinations/lifecycle.md#examination-print-infection-agreement-20261015)と[患者台帳読取](../domains/owners-pets/lifecycle.md#requirement-fr-owners-pets-infection-summary)を参照。AC-owners-pets-infection-top-reference／AC-owners-pets-infection-ambiguity／AC-owners-pets-infection-write-boundary／AC-owners-pets-infection-privacyが出典・版・認可境界を追跡する。健診package/field/resultの既存概念は[D09](../domains/checkups/lifecycle.md)へ委ね、専用の作図・写真機能は新採用せず既存カルテ添付を維持する。接種証明の発行版は[証明書概念](medical-records.md#certificate-edition-concept)へ接続し、結果保存/確認と証明書発行/印刷を同一イベントにしない。

<a id="examination-state-and-receive-concepts"></a>
## 検査自身の版・機器受信の状態

| 概念 | 定義・関係 | 制約と追跡 |
|:---|:---|:---|
| 検査進捗／封印 | pending/in_progress/result_enteredは進捗、confirmedは検査自身の封印 | [FR-examinations-revision](../domains/examinations/lifecycle.md#fr-examinations-revision)、AC-examinations-confirm-lock。通常カルテ旧finalizedの撤回で検査封印を解除しない |
| official版／訂正作業版 | 元の公式結果を保持して専用解除で作る作業版。再確定と作業保存は別 | 同FR、AC-examinations-unconfirm／AC-examinations-unconfirm-denied／AC-examinations-reconfirm。理由/認可と元版・actor・時点を保持 |
| revisionなしcompleted／revisionありcompleted | 同じ旧status文字列でも前者は旧封印、後者は解除後の作業版となり得る | AC-examinations-legacy-seal／AC-examinations-finalized-parent。架空の版を作って解除しない |
| 受信フレーム／ジョブ | 1フレームを1測定/1ジョブとして同医院・source・fingerprintで再送同定する | [FR-examinations-mapping](../domains/examinations/lifecycle.md#fr-examinations-mapping)、[FR-examinations-idempotency](../domains/examinations/lifecycle.md#fr-examinations-idempotency)。対応種別NならN検査、ジョブ由来全件が補償単位 |
| 待機対象／未紐付け／要確認 | 医院ごとの有効待機は1件、期限切れ/待機なしは未紐付け、未知コードはneeds_reviewで推測補完しない | [FR-examinations-receive](../domains/examinations/lifecycle.md#fr-examinations-receive)、AC-examinations-wait／AC-examinations-wait-expired／AC-examinations-multiple-types。カルテ編集権の無操作維持とこの待機期限は別 |
| attach／device detach／fixture revert | 患者への明示紐付け、安全条件を満たす解除、fixture終端処理という別操作 | AC-examinations-attach／AC-examinations-detach／AC-examinations-detach-used／AC-examinations-receive-duplicate／AC-examinations-result-unknown。参照/確定/手修正等の使用中保護を迂回せず原指紋を保持 |

機器のorigin排他は受信ポートの保護であり、通常カルテの編集権ではない。受信日時は検査日時と別。既存NX600/AU10Vの日常経路の定義を、新機器/新接続方式の運用承認へ拡張しない。D09健診はpackage/typed field/result setと保存時snapshotを持つ別ownerで、package改定による過去結果/次回日自動再計算をしない。[FR-checkups-canonical](../domains/checkups/lifecycle.md#fr-checkups-canonical)、FR-checkups-results／FR-checkups-record、AC-checkups-value-types／AC-checkups-replace-missing／AC-checkups-replace-empty／AC-checkups-master-change／AC-checkups-next-history-preservedを参照し、0/falseと未入力、意図的空配列と未送信/nullを区別する。
