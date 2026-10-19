# 飼主・ペットの用語と関係

文書確認日2026-10-15。既存台帳仕様、追加合意4・5と[台帳・会員区分・LINE希望の共同確認](../domains/owners-pets/lifecycle.md#agreement-owner-membership-dm-20261015)の語彙整理であり、その他の関係の新しい承認や実装済み宣言ではない。[業務正本](../domains/owners-pets/lifecycle.md)・[技術契約](../../design/domains/owners-pets/technical-spec.md)。

| 概念 | 意味・識別 | 関係・境界 |
|:---|:---|:---|
| 飼主 | 共通基本情報を持つ顧客。owner IDと表示上の飼主Noを区別し、内部共有ID方式は未選定 | 一人に複数pet。共通患者の自院利用とD19の既存手動リンクを区別 |
| ペット／患者 | 診療・施術等を受ける個体の台帳。pet IDとpet_numberを区別 | 現在主飼主一人、他の副飼主リンクは複数可。臨床記録IDとは別 |
| 主飼主 | pet.owner_idの現在の帰属 | 今後の予約・LINE通知先への影響一覧確認後に切替。過去請求・預り金は元ownerを保持 |
| 副飼主／共同飼育 | pet-ownerの追加リンクと続柄 | 主飼主との重複・同一owner重複なし。請求先・LINE通知先は自動変更せず、会計／LINE／割引の主飼主も変えない |
| identity group | D19既存scope内の手動連携。共通患者の認可内自院利用の業務とは別で、内部共有方式を代用決定しない | 台帳移管・予約移管・自動名寄せとは別、D19正本 |
| 生死状態 | alive／deceased、死亡日・理由 | 専用死亡／復活writer。生前診療の後日入力・理由付き訂正・退院・精算はdeceasedを保つ専用経路で、復活や履歴消去ではない |
| 飼主特記 | owners.is_dangerous、人への対応上の注意 | ペットの特記レベルとは別概念。スタッフ限定 |
| ペット特記 | danger_levelとdanger_reason、取扱い上の注意 | 高／中／低・未設定。高は理由必須。スタッフ専用、限定DTO除外を維持 |
| 動物種 | 全医院共有の分類master。追加・名称変更・無効化はシステム管理者のみ | ワクチン／投薬量の固定enumとは別。名称犬／猫のサジェストとの現行関係を保持 |
| 会員区分 | 非会員／会員／退亡者／他診/準の既存ラベル | 退亡者は連絡なく来院しなくなった人で死亡ではない。他診/準の意味は未確定。区分変更だけで料金・予約受付・LINE可否を自動変更しない |
| DM希望 | LINEの未設定／必要／不要 | 未設定・不要は案内／販促を止める。予約確認・変更・予約リマインドは既存安全条件へ従う。独立停止やLINE宛先とは別 |
| 体重 | pet.weightは台帳値 | 個々の診察／施術時バイタルから自動引用する保証なし |
| サイグラム | 飼主生年月日から規則で導出しDBへ保存する結果 | DOBとルール版に従属し独立編集しない。導入時有効DOBも計算保存、不正/欠落は未設定 |
| 未保存pet | temp-*等の画面draft | 永続IDでも深いリンク対象でもない |
| 重複連絡先候補 | 同医院の電話・メール重複時に示す既存飼主の参照候補 | 電話/メール一致だけで同一と決めず、別人と確認した場合は別owner登録可。同一飼主と確認した場合は後段の新規禁止を適用。自動統合・保存receiptの根拠ではない |
| 主飼主変更の影響一覧 | 保存前に確認する対象予約とLINE通知先の変更内容 | 確認対象と保存対象を一致させる。新しいLINE本人紐付けや過去資金移管ではない |

参照・多重度は業務正本のprimary-secondary／primary-transferへ。医院横断readは他医院writeの許可ではない。メール／電話／氏名の一致は同一IDの証明や保存receiptの代わりにならない。

## スタッフ患者トップの感染症結果

[D07の追加採用](../domains/examinations/lifecycle.md#examination-print-infection-agreement-20261015)による結果のread projection。猫FIV／FeLV・犬mf／HWの原値と検査日、院内／他院の出典、固定結果版・訂正状態を参照する。台帳の独立した検査結果本文、danger_level、現在の感染状態の診断とは異なる。日付不明・複数候補・記録なし・結果待ち・取得失敗を区別する。結果記録のwrite ownerはD07であり、他院手入力も台帳属性PATCHにはしない。[用語詳細](examinations.md)・[D04 read契約](../../design/domains/owners-pets/technical-spec.md#owner-infection-summary-read-contract)。

<a id="common-patient-concept"></a>
## 共通基本情報と院別業務の識別

| 概念／関係 | 意味・方向・多重度 | 追跡先 |
|:---|:---|:---|
| 共通飼主・共通ペット | 飼主住所/電話とペット基本情報を利用医院から最新参照する対象。一飼主に複数ペット、現在主飼主一人という既存関係を維持 | [FR-owners-pets-common-use](../domains/owners-pets/lifecycle.md#requirement-fr-owners-pets-common-use)、AC-owners-pets-common-latest |
| 患者↔利用医院 | 一患者を権限内の複数医院で利用でき、他院既存患者を選び自院で受付・カルテを作れるという業務関係 | 同FR、AC-owners-pets-other-clinic-use／AC-owners-pets-common-denied。物理ID統合/共有ID/手動unlinkの再設計は未選定 |
| 院別診療・金銭帰属 | 各医院の診療記録・請求・預り等に元医院と元対象の参照がある | 共通基本情報更新や自院利用で他院記録/過去金銭を移管・公開しない。発行済版も不変 |
| 重複候補と同一確認 | 氏名、又は動物種＋ペット名から示す候補と、スタッフが本人/個体を確認する判断は別 | AC-owners-pets-duplicate-confirm。同一飼主なら警告し新規停止、別人なら可。文字一致を同一IDの証拠にしない |
| 同じ飼主の新ペット | 確認済み既存飼主へ追加する別個体 | AC-owners-pets-new-pet-existing-owner。新ペット登録を重複飼主作成へ変換しない |

<a id="common-subject-technical-relations"></a>
## 共通患者と本人対応の技術案への対応

[D04詳細設計案](../../design/domains/owners-pets/technical-spec.md#common-subject-link-proposal)の識別を既決業務へ対応付ける。共通UUID／mappingの採用・移行済み・本人の物理統合承認とはしない。

| 概念 | 関係・境界 |
|:---|:---|
| 共通subjectと医院利用 | 一subjectを複数医院で利用する既決。提案は各clinic/subjectにactive利用最大1と元local ID対応を持ち、元院の診療/金銭を移動しない |
| 共通基本情報版と院別業務版 | 住所電話・pet基本情報の現在値は共通、診療/請求・過去発行snapshotは元医院・元版を保持。認可のscopeを共通IDで拡大しない |
| identity groupとcommon subject | groupは既存の手動対応関係で、共通subjectの物理IDではない。group unlink・患者分割・自院利用解除を同じ操作へ潰さない |
| LINE binding版と本人確認 | 確認は当該clinic/ownerの当該bindingに対応。別LINE IDへ旧確認を継承せず、共通患者・主飼主・customerを別識別 |
| 移管previewと保存対象 | 対象集合・元新owner・LINE結果と版を確認する。保存時に対象追加/除外も再検査し、古い確認から自動実行しない |

Q-I1は具体的な矛盾データ移行／患者分割が必要になった時だけの条件付き検討で、現在の全体blockerではない。スタッフ同一確認は既決。Q-I2は[12:53 UTC本人第2項A](../line/lstep-integration.md#line-unlink-future-notification-20261018)により、予約と患者関連・参照履歴を保持しLINE通知を停止、再連携後は再開する予約をstaffが確認する。旧JSONのBと意味対応する。連携状態・残存参照と現在送信許可を同義にせず、自動再開や結果不明再送を追加しない。

<a id="saigram-derived-result-concept"></a>
## サイグラムの派生保存と時点

[FR-owners-pets-saigram-calculation](../domains/owners-pets/lifecycle.md#saigram-calculation-reference-20261018)が計算表と版の唯一正本。DOB→計算ルール版→派生保存結果の依存を保持する。計算を表示のたびに独立分類するだけにせず、結果とルール版をDBに保存する業務契約を採用するが、列構造/APIは選定しない。DOBの有効変更時は同基準で再計算、明示削除は結果なし、不正保存拒否で有効DOBと結果を片側だけ変更しない。

導入時の既存飼主も有効DOBなら計算保存、欠落/不正は未設定。氏名や別属性からDOBを推定しない。固定60対応表を語彙へ複製せず、指定7/16資料の基準/UTC暦日計算・版を参照する。AC-owners-pets-saigram-persist-consistency／AC-owners-pets-saigram-clear-invalid／AC-owners-pets-saigram-storage-scope／AC-owners-pets-saigram-initial-population。将来ルール改定・導入実行手順・臨床的又は性格診断の保証はこの整理で採用しない。
