# 予防接種の業務仕様（D08）

| 項目 | 内容 |
|---|---|
| 文書ID | spec-D08-vaccinations-lifecycle |
| 種別 | 業務仕様・受入条件 |
| 決定状態 | 既存記載を継承（個別の決定根拠は本文参照。根拠未確認部分は承認済みとしない）＋D03合意目標への接続。新規の未決事項は未決のまま |
| 実装状態 | 本文の現コード観察と目標契約を分離。コードに存在するだけでは業務承認済み・目標実装済みとしない |
| 実機検証状態 | 今回未実施。既存証跡の引用は今回の実機受入を意味しない |
| 文書確認日 | 2026-10-15 |

既存仕様の継承とD03の合意済み目標を整理した文書候補。ソフトウェア仕様の整理、実装完了、実機受入、医師による具体的な臨床数値の承認は別の判定とする。

<a id="vaccination-document-scope"></a>
## 範囲・根拠
2026-10-15全ドメイン整理、基準commit `41fcf3eb0a8befb7c0b5f613d1df749a60e30be8`。[一覧](../../screens/14-vaccinations-list.md)・[入力](../../screens/15-vaccinations-form.md)・[カルテ保存§2](../medical-records/lifecycle.md#hybrid-sync)・[次回予定シナリオ](../../../ops/testing/scenarios/S03-vaccination-next-due-autocalc.md)・[同日複数の後続検証記録](../../../work/emr/emr-105-vaccine-multi-20261002.md)を整理する。ワクチン、狂犬病、フィラリア・ノミダニ等の予防実績を対象とする。接種判断・対象種・用量・間隔数値は医師の判断であり、本書は臨床的妥当性の承認ではない。新しい「臨床責任者」役割・追加の接種必須項目は作らない。

S29の古い「調査中」を現在の未解決状態とは扱わない。EMR-105は同日複数のカルテ内経路について当時の実DB／モックフォーム検証結果を記録する。今回再実行した証拠ではなく、製品全域／実ブラウザ全操作PASSへ拡張しない。[技術契約](../../../design/domains/vaccinations/technical-spec.md)に現状差を残す。

<a id="fr-vaccinations-record"></a>
## FR-vaccinations-record：一接種一記録を二つの入口で共有する
独立フォームとカルテ内フォームは同じ接種資源を扱う。独立新規は患者を選択し、カルテ内新規は保存済みカルテIDと患者を指定する。カルテ内に別系統の接種台帳や架空のnested vaccinations APIを設けない。カルテ外側の保存ボタンでは接種保存しない。

一回の送信は一接種。複数のワクチン／予防薬を同じ患者・同日に順次登録でき、名称・lot1〜4・次回予定・備考・担当・カルテ帰属を各行で保持する。バッチの部分成功をこの仕様から導かない。接種済み実績と将来の次回予定は別で、予定を登録して実施済みにしない。担当医属性と実編集actor、認められた医師の確認を分離する。

<a id="fr-vaccinations-schedule"></a>
## FR-vaccinations-schedule：次回予定を見える候補として計算し、医師の調整を保持する
接種日は必須、未来日は不可。有効なワクチンマスタを選ぶ。既存の標準候補は3weeks／4weeks／1year／other。マスタの間隔が設定されている場合は選択時に反映し、標準候補なら接種日変更に合わせて再計算、otherでは手動調整を保持する。次回予定は接種日より後、独立新規フォームでは既存の本日以降条件を維持する。過去記録専用入力を通常フォームの制約回避として新設しない。

独立フォームの既定1yearとカルテ内既定4weeksは現状差である。両方を医学的推奨値と記載せず、選択マスタ／計算結果／手動値が患者に適切かは医師が確認する。今回、既定の統一や新しい適用間隔を未承認で変更しない。種による候補絞込み・種不一致拒否が現行未実装であることも、異種の接種が安全／承認済みという意味ではない。

<a id="fr-vaccinations-correction"></a>
## FR-vaccinations-correction：訂正・削除・同時操作で別の接種を変えない
編集／削除は対象接種IDと現在の医院・患者・カルテ帰属・grantを確認する。死亡ペットの編集／削除を通常画面で行わない。未保存入力を保護し、保存失敗では入力を保持する。保存成功後だけフォームを初期化し、再取得で同日の複数行を独立表示する。単一行削除で同日の他の記録を消さない。

紐付くカルテfinalizedの通常追加／更新／削除を拒否する。患者取り違え、確定後訂正、過去専用入力、復元はD03の専用経路で原記録・変更前後・理由・actor／時刻を保持する。通常patient FK PATCHで付け替えない。専用経路未実装を通常APIで代用しない。競合時は古い値で上書きせず再取得する。送信結果不明は照会し、同日複数許可を無制限の再送重複許可にしない。

<a id="fr-vaccinations-history"></a>
## FR-vaccinations-history：履歴の範囲と公開先を明確にする
一覧は実施日・全文をAPIへ送り、担当医絞込みは取得済み最大100件内。終了日省略はJST当日。既定は次回日昇順、未設定末尾、ユーザー列ソートを優先。編集リンクはカルテ紐付きなら該当タブ／接種ID、未紐付きなら独立詳細。viewとedit/deleteを分け、閲覧リンクだけで編集権限を与えない。

同じ保存済み接種・次回予定をカルテ履歴、健康表示、レポートの認可済みprojectionへ渡す。取得不能・非開示・未設定を区別し、全件取得していないのに全履歴と断定しない。表示成功は飼主への公開／送信成功ではない。D20の2026-10-15承認（Sentinel_7a67fd30e47c819192cd36d416f0a2bf）は本人に紐付く生存ペットの保存済み接種実績を対象とし、親カルテfinalizedや新しい接種承認操作を要求しない。未保存・取消／削除は非公開、訂正保存後は更新する。検査・健診結果へ公開範囲を拡張しない。

<a id="fr-vaccinations-master-delivery"></a>
## FR-vaccinations-master-delivery：マスタ・期限通知・課金を別の責務へ接続する
名称・有効状態・間隔・価格等のワクチン項目管理は既存診療マスタに接続し、接種編集権限でマスタ全開放しない。過去の保存事実を現在マスタで推測置換しない。当時名称／間隔／価格のsnapshotがない旧値は未確認とし、D03／D10の保存時参照境界で扱う。

期限接近通知はD15のvaccine_deadline_60d／30d等の設定・配信判断が所有する。受付の自動アラートがあるとは記載しない。接種保存とタグ同期・配信成功を別に報告し、失敗した配信を直すために接種を再作成しない。死亡・退会／公開認可・現owner／pet scopeはD04/D15/D20側で再照合する。接種保存だけで課金済み／在庫減算済みとしない。

## 受入条件（GWT、実機未実施）
| AC ID | Given | When | Then | FR |
|:---|:---|:---|:---|:---|
| AC-vaccinations-standalone | 独立新規で患者選択済み | 保存 | 単一接種を対象患者へ、カルテIDを架空作成しない | record |
| AC-vaccinations-in-chart | 保存済みカルテと同じ患者 | タブ内保存 | medical_record_idを保持、外側保存／nested APIで二重保存しない | record |
| AC-vaccinations-same-day | 同患者・同日にA登録済み | B/Cを順次登録 | 各ID・名称・lot・次回日・備考が独立 | record |
| AC-vaccinations-reset | 1件保存成功 | 次記録を追加 | 前行の値を漏らさず再取得して別行表示 | record |
| AC-vaccinations-future | 未来接種日／接種日未入力 | 保存 | 拒否し実績化しない | schedule |
| AC-vaccinations-master-interval | 有効マスタに間隔設定あり | 選択・接種日変更 | 指定間隔で候補計算し表示、固定既定を強制しない | schedule |
| AC-vaccinations-manual | otherで手動日指定 | 再表示／保存 | 手動値を保持し標準候補で黙って上書きしない | schedule |
| AC-vaccinations-next-invalid | 次回が接種日以前 | 保存 | 拒否。独立新規の本日以降条件も維持 | schedule |
| AC-vaccinations-default-boundary | 2入口の既定が異なる／種混在 | 利用判断 | 現行差を明記、数値・異種適応の医師承認済みとしない | schedule |
| AC-vaccinations-lots | 1記録に複数lot | 保存・再表示 | 最大4欄を当該行で保持 | record |
| AC-vaccinations-finalized | 親がfinalized | 作成・更新・削除／関連変更 | 通常write拒否、D03専用訂正へ | correction |
| AC-vaccinations-scope | 他医院／患者不一致／grantなし | 操作 | 拒否し他患者情報を開示・更新しない | correction |
| AC-vaccinations-deceased | 死亡ペット | 通常編集・削除 | UIで許可せず専用取扱いを通常操作から推定しない | correction |
| AC-vaccinations-delete-one | 同日複数行 | 対象1行の確認後削除 | 対象だけ論理削除し他行を保持 | correction |
| AC-vaccinations-conflict | 他者変更／医院・権限切替 | 古い操作・応答 | 上書き・別scope反映をせず再取得 | correction |
| AC-vaccinations-unknown | 送信後結果不明 | 再送検討 | 対象保存結果を照会、無条件再createしない | correction |
| AC-vaccinations-edit-log | 必須履歴記録失敗 | 訂正write | 原記録と監査を分離成功にせず失敗扱い | correction |
| AC-vaccinations-filter-scope | 最大100件内の担当医絞込み | 一覧閲覧 | 現在取得範囲と全体を混同しない | history |
| AC-vaccinations-navigation | カルテ紐付き／未紐付き | 閲覧リンク | 正しい入口へ、viewだけでeditを許可しない | history |
| AC-vaccinations-projection | 保存済み接種と次回日 | 健康表示／レポート | 同じ患者・接種ID・取得状態。D20は本人紐付き生存ペットの保存済み実績のみ公開し、親finalizedを要求しない。未保存・取消／削除は非公開、訂正保存後更新。D21はスタッフ閲覧認可 | history |
| AC-vaccinations-master-history | マスタ変更／当時snapshot欠損 | 過去表示・訂正 | 当時値を現在値から架空生成しない | master-delivery |
| AC-vaccinations-notify-failure | 接種保存済み・タグ同期失敗 | 結果を表示 | 接種保存と配信成否を分離、接種再作成なし | master-delivery |
| AC-vaccinations-no-side-effect | 接種の保存／訂正 | 会計・在庫連携 | ownerへ影響を渡し、自動返金・新請求・在庫調整を推定しない | master-delivery |

## 引継ぎ・完成条件
業務の保存単位、独立／カルテ入口、同日複数、次回日、訂正保護、履歴範囲と受け手を整理済みと判定する範囲。種フィルタ追加・既定間隔統一・全履歴検索の改善は未承認の追加機能として今回の必須項目にしない。医師による製品／対象種／具体的接種間隔の確認、現権限・旧snapshot、全writer／版／監査／receipt適合、実機受入・通知運用は完了外。D03の完了済み保存§2と確定／訂正境界を変更しない。


D21 Owner Reportはスタッフの閲覧専用であり、同画面からの印刷・共有を設けない。D03や検査自ドメインに既存の印刷経路があることを、D21経由の印刷許可へ拡張しない。
