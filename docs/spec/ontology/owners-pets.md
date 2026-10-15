# 飼主・ペットの用語と関係

文書確認日2026-10-15。既存台帳仕様、追加合意4・5と[台帳・会員区分・LINE希望の共同確認](../domains/owners-pets/lifecycle.md#agreement-owner-membership-dm-20261015)の語彙整理であり、その他の関係の新しい承認や実装済み宣言ではない。[業務正本](../domains/owners-pets/lifecycle.md)・[技術契約](../../design/domains/owners-pets/technical-spec.md)。

| 概念 | 意味・識別 | 関係・境界 |
|:---|:---|:---|
| 飼主 | 医院の顧客台帳。owner ID、飼主Noはその表示別名 | 一人に複数pet。identity groupはD19の既存scope内で別リンク |
| ペット／患者 | 診療・施術等を受ける個体の台帳。pet IDとpet_numberを区別 | 現在主飼主一人、他の副飼主リンクは複数可。臨床記録IDとは別 |
| 主飼主 | pet.owner_idの現在の帰属 | 今後の予約・LINE通知先への影響一覧確認後に切替。過去請求・預り金は元ownerを保持 |
| 副飼主／共同飼育 | pet-ownerの追加リンクと続柄 | 主飼主との重複・同一owner重複なし。請求先・LINE通知先は自動変更せず、会計／LINE／割引の主飼主も変えない |
| identity group | 既存scope内の同一人物／個体の手動連携。医院間利用を本整理で拡大しない | 台帳移管・予約移管・自動名寄せとは別、D19正本 |
| 生死状態 | alive／deceased、死亡日・理由 | 専用死亡／復活writer。生前診療の後日入力・理由付き訂正・退院・精算はdeceasedを保つ専用経路で、復活や履歴消去ではない |
| 飼主特記 | owners.is_dangerous、人への対応上の注意 | ペットの特記レベルとは別概念。スタッフ限定 |
| ペット特記 | danger_levelとdanger_reason、取扱い上の注意 | 高／中／低・未設定。高は理由必須。スタッフ専用、限定DTO除外を維持 |
| 動物種 | 全医院共有の分類master。追加・名称変更・無効化はシステム管理者のみ | ワクチン／投薬量の固定enumとは別。名称犬／猫のサジェストとの現行関係を保持 |
| 会員区分 | 非会員／会員／退亡者／他診/準の既存ラベル | 退亡者は連絡なく来院しなくなった人で死亡ではない。他診/準の意味は未確定。区分変更だけで料金・予約受付・LINE可否を自動変更しない |
| DM希望 | LINEの未設定／必要／不要 | 未設定・不要は案内／販促を止める。予約確認・変更・予約リマインドは既存安全条件へ従う。独立停止やLINE宛先とは別 |
| 体重 | pet.weightは台帳値 | 個々の診察／施術時バイタルから自動引用する保証なし |
| サイグラム | 飼主生年月日からの派生表示 | 独立永続事実ではない、未設定／不正時分類なし |
| 未保存pet | temp-*等の画面draft | 永続IDでも深いリンク対象でもない |
| 重複連絡先候補 | 同医院の電話・メール重複時に示す既存飼主の参照候補 | 警告後も別owner登録可。人物同一性・自動統合・保存receiptの根拠ではない |
| 主飼主変更の影響一覧 | 保存前に確認する対象予約とLINE通知先の変更内容 | 確認対象と保存対象を一致させる。新しいLINE本人紐付けや過去資金移管ではない |

参照・多重度は業務正本のprimary-secondary／primary-transferへ。医院横断readは他医院writeの許可ではない。メール／電話／氏名の一致は同一IDの証明や保存receiptの代わりにならない。

## スタッフ患者トップの感染症結果

[D07の追加採用](../domains/examinations/lifecycle.md#examination-print-infection-agreement-20261015)による結果のread projection。猫FIV／FeLV・犬mf／HWの原値と検査日、院内／他院の出典、固定結果版・訂正状態を参照する。台帳の独立した検査結果本文、danger_level、現在の感染状態の診断とは異なる。日付不明・複数候補・記録なし・結果待ち・取得失敗を区別する。結果記録のwrite ownerはD07であり、他院手入力も台帳属性PATCHにはしない。[用語詳細](examinations.md)・[D04 read契約](../../design/domains/owners-pets/technical-spec.md#owner-infection-summary-read-contract)。
