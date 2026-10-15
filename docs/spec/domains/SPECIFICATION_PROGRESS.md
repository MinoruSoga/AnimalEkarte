# 仕様整理の全体進捗

更新日: 2026-10-14。対象はAnimalEkarteの仕様文書整理。ユーザー Sentinel_d23d519d0a608191b52b444b5bcd8bce／Sentinel_26c69b86a0d88191bfefc3fd5809bf6e の依頼により、この1ファイルを全体管理の入口とする。

実装、移行、API／DB反映、実機受入、運用開始は今回の対象に含めない。実装タスクと担当・依存管理は引き続きPlaneで扱い、本書から実装完了を数えない。業務本文・FR／AC・技術契約はリンク先が正本。本書の分類・掲載・進捗は新しい業務ルールや未承認案の承認を表さない。

## 管理単位と数え方

現在のdomains子フォルダは4個（予約・受付・カルテ・会計）だが、全業務領域の数ではない。以下は既存仕様を見落とさないための「仕様整理の管理単位」で、実装のパッケージ境界を再定義しない。各画面・マスタ・画面外契約には主責務を一つ割り当て、利用側との接続は横断表へ分ける。共有画面や複数ルートを持つ同じ仕様を重複計数しない。

根拠は[全体要件・臨床安全](../specification.md)、[全画面索引](../screens/README.md)、[88製品ルートの正本対応](../ui-design-compliance.md)、[横断フロー](../workflows/README.md)、[語彙・関係](../ontology/README.md)、[共通仕様](../shared/README.md)、[FE／BE責務対応](../../architecture/fe-feature-be-domain-map.md)、[write-owner](../../architecture/model-write-owner-catalog.md)と[横断処理カタログ](../../architecture/cross-domain-orchestration-catalog.md)。画面を持たない保存・配信・集計や別アプリも含む。

## 状態と文書整理の完了条件

| 状態 | 意味 |
|:---|:---|
| 完了 | 表の明示範囲について業務・受入・技術文書を整理し、独立レビューと限定文書検査を通過。領域外内部や全体統合の完了へ広げない |
| 進行中 | 今回の領域全体整理を開始し、文書・レビューに残件がある |
| 未着手 | 既存本文はあるが、今回の領域全体見直しを開始していない。入口整形・他領域からの接続整理だけでは開始／完了と数えない |
| 確認待ち | 範囲は整理済みだが、判断・受け手の確認・全体再照合が残る |
| 未確認 | 範囲・正本・状態を特定できていない。既存実装やClosedだけで補完しない |

完了には、対象・用語・業務ルール・例外／訂正・権限・状態／IDの境界、正本と読む順、受入条件、通信失敗／競合・保存／再送・連携の技術契約が必要。承認根拠・未決・実装状態は[記述規約](../authoring-guide.md#status)で分け、独立レビュー指摘とリンク・アンカー・本文／対応表欠落を解消する。関連領域へ残した内部条件は渡し先を明示する。

個別領域の完了と全体の連携確認完了は別。全体完了には未着手領域の整理、共通判断の扱いの確定、全ownerの文書が揃った後の接続再照合が必要。今回の静的検査はruntimeや性能の合格を示さない。

## 集計

| 集計対象 | 総数 | 完了 | 進行中 | 未着手 | 確認待ち | 未確認 |
|:---|---:|---:|---:|---:|---:|---:|
| 業務領域（D01〜D23） | 23 | 3 | 1 | 19 | 0 | 0 |
| 共通領域（C01〜C06） | 6 | 0 | 3 | 3 | 0 | 0 |
| 横断接続（X01〜X08） | 8 | 0 | 0 | 0 | 8 | 0 |
| 管理行合計 | 37 | 3 | 4 | 22 | 8 | 0 |

完了3件は表に限定した予約・受付・カルテの文書範囲。全機能・全チャネル・実装の3領域完了とは数えない。現在の在庫の主責務は配賦できたが、未着手領域の内部判断・未文書化要件まで網羅／確定済みとは判断できない。「未確認0」は管理行の分類が未特定の行が0という意味で、未着手内部・未文書化要件・実装状態に未確認事項がないことを示さない。新たに発見した範囲は主責務を決め、表と集計・ルート対応を同時更新する。

## 業務領域

技術の共通参照は[責務対応](../../architecture/fe-feature-be-domain-map.md)、[ownerカタログ](../../architecture/model-write-owner-catalog.md)、[API設計](../../architecture/README.md)、[ERD](../../architecture/erd.md)。下表の専用技術入口が「今回未整理」の領域では、これら既存資料の掲載をレビュー完了と数えない。担当は責務上の主領域であり、人員担当の新任命ではない。

| ID・業務領域 | 主責務・境界／業務正本 | 技術設計の入口 | 状態・根拠／残る範囲 |
|:---|:---|:---|:---|
| D01 予約 | 登録・変更・取消、区分、指名・対応能力・勤務・容量の予約判定、院内／LINEの合意済み接続。[予約入口](reservations/README.md)・[語彙正本](../ontology/reservations.md)。LINE管理全域はD18、職員管理内部はD14 | [予約技術仕様](../../design/domains/reservations/technical-spec.md) | 完了。69固有FR＋継承参照2件と技術接続の明示範囲。独立レビュー・索引／trace照合済み。[完了境界](reservations/README.md#reservation-document-completion-current)を維持。他owner内部・全体再照合はX表へ |
| D02 受付 | 選択日の4分類・5列／例外区画、通常診療開始／終了への接続、表示片付け／復元、filter／集計。[受付入口](reception/README.md)・[画面とAC](reception/pages/reception-board.md)。施術・滞在・精算内部は各owner | [受付技術仕様](../../design/domains/reception/technical-spec.md) | 完了。承認済み表示・通信／競合・ブラウザ設定の文書範囲。独立レビュー2指摘修正／再確認済み、AC21件。専用内部・実機・全体再照合を含めない |
| D03 カルテ | 通常診療記録・一覧／入力・保存／確定・訂正、診療項目／診断／薬剤／問診／主訴マスタ。[カルテ入口](medical-records/README.md)、[保存正本§2](medical-records/lifecycle.md)。検査・予防・健診はD07〜09 | [カルテ技術契約](../../design/domains/medical-records/README.md)。予約participant・既存owner／保存正本へ接続 | 完了。追加docs承認 Sentinel_7e6b0beed1b48191ae86bc2332800109 の最終照合を反映。16FR／71AC・全9タブ＋バイタル／共通操作・D03master取扱い・専用訂正／過去／復元・型／互換／旧値停止・writer／監査／gate／receiptと受け手の明示docs範囲。独立レビュー3名と限定文書検査を通過。[完成条件と判定外](../../design/domains/medical-records/technical-spec.md#medical-record-document-completion)。実権限／実旧値・数値医師承認・実装／利用開始／実機性能・他owner内部・全体横断最終確認は未確認／引継ぎ |
| D04 飼主・ペット | 飼主登録／変更とペットの基本台帳、動物種。[一覧](../screens/03-owners-list.md)・[フォーム](../screens/04-owners-form.md)。同一個体連携D19・健康表示D20・レポートD21を別管理 | 今回専用技術整理未着手。既存owner／pet write-owner・ERD | 未着手。台帳登録・変更・削除／参照保護・複数医院の内部契約と受入を確認する |
| D05 入院・ホテル | 臨床入院／宿泊の2範囲、専用記録・実滞在／ケージ・世話・退院／退室、入院plan／cageマスタ。[一覧](../screens/07-hospitalization-list.md)・[詳細](../screens/08-hospitalization-detail.md)・[フォーム](../screens/09-hospitalization-form.md)。予約受入境界はD01 | 今回専用技術整理未着手。既存medicalrecord owner・予約受け手契約 | 未着手。予約に関する頭数・延泊・担当条件は先行整理済み。臨床入院とホテルの記録本文／内部状態、世話担当職種／人数を別scopeで確認する |
| D06 トリミング | 施術記録・担当・コース／選択肢・完了と会計への接続、trimming／course-typeマスタ。[一覧](../screens/16-trimming-list.md)・[フォーム](../screens/17-trimming-form.md) | 今回専用技術整理未着手。既存trimming owner・予約intent契約 | 未着手。予約／受付の分類接続だけ先行整理済み。施術内部・訂正・課金の受入を確認する |
| D07 検査・機器受信 | 検査結果・確定／訂正・機器受信／対応付け、検査マスタ／機器項目。[一覧](../screens/12-examinations-list.md)・[フォーム](../screens/13-examinations-form.md)・[検査マスタ](../screens/settings/master-examinations.md) | 既存[ADR007](../../architecture/adr/007-lab-device-receive-and-commit.md)・[ADR008](../../architecture/adr/008-local-lab-device-agent.md)。今回全域レビュー未着手 | 未着手。機器受信画面だけで結果確定／訂正・重複・失敗時契約全体の完了としない。健診正系統はD09 |
| D08 予防接種 | 接種記録・次回予定・担当帰属、ワクチン項目。[一覧](../screens/14-vaccinations-list.md)・[フォーム](../screens/15-vaccinations-form.md) | 今回専用技術整理未着手。既存medicalrecord owner・ERD | 未着手。独立フォーム／カルテ内入力の関係、履歴・訂正・健康表示の受渡しを確認する |
| D09 健診 | 健診パッケージ・項目／結果・担当、健診記録。[健診仕様](../screens/25-checkups-list.md)。配信対象同期はD15 | 既存[ADR004 正系統](../../architecture/adr/004-checkup-canonical-system.md)。今回全域レビュー未着手 | 未着手。旧検査系との二重管理禁止を保持。現行Acceptedと今回整理完了を別に読む |
| D10 会計・未納 | 請求／入金・未納／部分入金／0円・飼主預り金／任意予約前金・返金／訂正、支払／保険／割引／物販マスタ。[会計入口](accounting/README.md)・[初回〜第八群9FR／116AC](accounting/lifecycle.md#accounting-agreement-eighth)。締め内部D12、見積D11 | [会計部分設計](../../design/domains/accounting/README.md)・既存[ADR003](../../architecture/adr/003-payment-method-identity-and-consistency.md)／billing owner | 進行中。第八群 Sentinel_d6659ec5387c81918f87d61d112dc855で減額時の預り金解除・元飼主残高復元／併用内訳確認、自動失効なし・同医院同飼主限定、元明細選択確認による新番号再請求／元取消請求リンク、新規会計と過去回収の同時扱い／元請求別指定、共通料金別枠を合意。医院間移行・飼主間譲渡、督促／貸倒は今回対象外。第七群 Sentinel_1bf8d148d1488191b94bd7126faa5e83でスタッフが外部画面等により資格／残限度を確認し結果・保険負担額を記録、窓口会計限定・ペットごと1利用契約を合意。保険会社への請求・入金照合・査定差額管理、同一診療費の複数契約併用は対象外。自動給付算式は将来の別スコープで、現在の記録運用の前提にしない。第六群 Sentinel_f677bbe0853081918d52ae34f4be0f35で複数保険会社（アニコムは例）、品目マスタの全体割引除外、飼主宛名／ペット別明細と元取引／今回変更額、管理者と権限付きスタッフのマスタ変更を反映。現grantの読取mappingは技術仕様へ、権限設定は変更なし。第五群 Sentinel_d0dce252895c819191e2b73dd7e47e8d／Sentinel_a32cd855d7f48191bcb975c93c76b832で明細→全体・比例配賦、税抜計算／税込表示、割合額四捨五入・税率別税切捨て、実値引き上限／他行転用なしを反映。第四群 Sentinel_e7c15be2b17081918c35bdc4fc8f937bで両粒度割引・超過警告・会計取消時の預り金残高復元／別返金を追加。追加回答 Sentinel_2fd30155c5a88191956dc21f6f7a7740で警告を承認して進めた請求額0円下限を確定し、有効充当額に取消の解除・復元を反映。第三群 Sentinel_2925c6a799a08191b2b6d43efcdc66dc／Sentinel_ea9c25aa96208191bd320999a81a97d6を追加。飼主単位預り金／予約任意・同飼主別予約／ペット確認付き利用、割合合算→金額→割引後保険、作成順／理由付き再配分、非現金超過、別方法返金／金額追加承認なし、用途別帳票・番号維持を反映。[未決5群](accounting/lifecycle.md#accounting-open-decisions)・保険確認記録の設計／除外の初期値・移行／小数数量等の元明細額確定・再請求の価格時点／元入金／保険結果再利用・出所不明の既実返金の解消・全master／具体権限／帳票詳細・他owner整合・実装／実機は残件。通常Undo停止・カルテ保存§2を維持 |
| D11 見積 | 見積作成／変更／詳細と会計受渡し。[一覧](../screens/22-estimate-list.md)・[フォーム](../screens/23-estimate-form.md)・[詳細](../screens/26-estimate-detail.md) | 今回専用技術整理未着手。既存billing owner・ERD | 未着手。見積状態と会計確定を同義にせず、転記・訂正・金額契約を確認する |
| D12 レジ締め | 締め・実査・過不足・AM/PM/EMG境界、締め時間設定。[締め仕様](../cash-register.md)・[画面](../screens/29-closing-aggregation.md) | 今回専用技術整理未着手。既存billing owner・ERD | 未着手。今回ここまで改訂範囲を広げていない。会計確定D10・帳票読取D16との境界を確認する |
| D13 在庫 | 在庫記録／残数・登録／変更。[一覧](../screens/18-inventory-list.md)・[フォーム](../screens/27-inventory-form.md) | 今回専用技術整理未着手。既存inventory owner・ERD | 未着手。物販マスタD10との実際の関係を確認し、在庫連動を推測しない |
| D14 職員・シフト | 職員／職種／勤務・休憩・シフトtemplate。[シフト](../screens/24-shift-calendar.md)・[職員](../screens/settings/master-staff.md)・[職種](../screens/settings/master-occupation.md) | 今回専用技術整理未着手。既存staff owner・予約lock契約 | 未着手。予約への保護・対応能力は先行整理済み。職員管理内部・全勤務運用まで完了としない |
| D15 CRM・Lステップ | タグ／Webhook／配信・対象同期・停止／再開、CPM配信policy。[LINEハブ](../line/README.md)・[業務本文](../line/lstep-integration.md)・[管理画面](../screens/31-lstep-integration.md)・[配信監視](../screens/34-lstep-delivery-monitor.md) | 既存[LINE構成](../line/architecture.md)、[scheduler接続](../../architecture/cross-domain-orchestration-catalog.md)。今回全域レビュー未着手 | 未着手。17配信トリガー／背景同期も対象。先方enable・実送信等の既存停止／再開条件を未承認のまま解除しない。実行状態は元資料 |
| D16 経営・顧客集計 | 月次／売上レポート・ダッシュボード・顧客分析。[月次](../screens/32-accounting-reports.md)・[集計](../screens/36-aggregation-dashboard.md)・[CPM／LTV定義](../customer-aggregation.md)・Lステップ分析 | 今回専用技術整理未着手。既存billing／lstep読取責務・ERD | 未着手。分析の計数／権限と、D15の配信判定・D12の締めを区別して確認する |
| D17 医院・設定入口 | 医院設定とマスタの閲覧入口。[医院](../screens/19-clinic-settings.md)・[設定トップ](../screens/20-master-settings.md)。個別マスタ本文は各D／Cへ配賦 | 今回専用技術整理未着手。既存clinic owner・[医院分離ADR](../../architecture/adr/002-multitenancy-clinic-id-isolation.md) | 未着手。入口があるだけで個別マスタ完了と数えない。医院設定全域の状態・変更影響を確認する |
| D18 LINE予約チャネル | 院内LINE設定／エディタ／枠と飼主予約app。[設定](../screens/28-line-reservation.md)・[別app37](../screens/37-line-reserve-owner-flow.md)・[予約正本](../line/reservation-spec.md)。空き判定の合意済み条件はD01参照 | 既存[LINE構成](../line/architecture.md)・[導入](../line/setup.md)・予約技術接続。全チャネル整理未着手 | 未着手。予定枠基準等の予約合意は完了済み。設定／導線／公開・変更／取消・通知などチャネル全域の完了へ広げない |
| D19 同一飼主・ペット連携 | 手動identity group link／unlink・連携履歴。[identity仕様](../screens/40-identity-links.md)。台帳D04・認可C04を別責務に保つ | 今回専用技術整理未着手。既存identitylink owner・ERD | 未着手。所属医院内の既存scopeを保持し、氏名類似や予約接続から自動統合を決めない |
| D20 LIFF診察券・健康情報 | 別appの健康手帳・LINEアカウント紐付け。[LIFF38](../screens/38-liff-pet-health.md)。診療内容の正本はD03／08、基本台帳D04、identity groupはD19 | 既存LIFF認証／LINE構成・ERD。今回全域レビュー未着手 | 未着手。main88へ加算しない。未紐付け・token失効・公開範囲・自己認証とスタッフ認可の違いを確認する |
| D21 飼主カルテレポート | 診療前情報・履歴の複合読取／別ウィンドウ。[レポート39](../screens/39-owner-report.md)。各記録writeを持たない | 既存各section読取grant・FE責務map。今回全域レビュー未着手 | 未着手。ページ認可とsection別縮退・患者context・遅延応答を確認。台帳／記録の完了を転用しない |
| D22 院内マニュアル | 記事閲覧／編集・印刷・院内説明。[manual35](../screens/35-internal-manual.md) | 今回専用技術整理未着手。既存manualarticle owner・ERD | 未着手。製品内manual仕様と、この仕様整理の管理文書を別責務にする |
| D23 サポート | バグ報告・ヘルプチャット／履歴の2範囲。[settings README本文](../screens/settings/README.md)。共通UIの索引だけではない | 既存support owner・FE責務map。今回全域レビュー未着手 | 未着手。共有READ・報告医院mutation・個人履歴scopeを区別し、全chatを個人限定に変更しない |

## 共通領域

| ID・共通領域 | 正本・主責務／技術入口 | 状態・残件 |
|:---|:---|:---|
| C01 共通UI・選択部品 | [UI体系](../design-system.md)・[状態](../design-states.md)・[token層](../design-token-layers.md)・[components](../components/README.md)・[共通dialog](../screens/common-dialogs.md)・[ペット選択00](../screens/00-pet-selection.md)。各業務の結果を定義せず再利用契約を持つ | 未着手。今回の全体見直し未開始。選択routeは利用D側へ一度配賦、部品本文はC01。空template・旧番号案内は別業務に数えない |
| C02 用語・文書状態・オントロジー | [書き方](../authoring-guide.md)・[ontology入口](../ontology/README.md)・[予約語彙](../ontology/reservations.md)。概念／画面名／実装モデル、決定／実装状態を分離 | 進行中。予約・受付・カルテ・会計の明示範囲の語彙と記述規約は先行整理済み。全領域の概念・関係／制約・名称を未照合 |
| C03 全体要件・臨床安全 | [全体要件§2.1](../specification.md)・[product-philosophy](../../product-philosophy.md)。臨床安全原則の優先を維持 | 未着手。原則本文は保持済み。全ownerの個別受入が安全原則を満たすか今回横断レビューは未実施 |
| C04 認証・権限・医院scope | [ログイン21](../screens/21-login.md)・[権限group](../screens/settings/master-permission-group.md)・[auth](../../architecture/auth.md)・[医院分離ADR](../../architecture/adr/002-multitenancy-clinic-id-isolation.md)。各業務に固有の操作条件は各D | 未着手。予約／受付の既存grant接続は先行整理済み。認証回復・全grant・例外scopeを全域で未照合 |
| C05 保存・廃棄・監査 | [shared当面方針](../shared/README.md#interim-retention-policy)・[delivery U11の入口](../../delivery/DELIVERY_PACKAGE.md)、owner別履歴／監査・ERD | 進行中。当面自動削除なしは承認済み。具体的保存期間・削除手続きは全ドメイン共通で確定待ち。法的無期限・法令適合としない |
| C06 技術接続基盤・文書検査 | [architecture](../../architecture/README.md)・[API／DB](../../architecture/overview.md)・[責務map](../../architecture/fe-feature-be-domain-map.md)・[owner](../../architecture/model-write-owner-catalog.md)・[処理カタログ](../../architecture/cross-domain-orchestration-catalog.md)・[testing](../../ops/testing/README.md)。個別技術方式は各Dへ | 進行中。予約／受付のdocs方式は具体化済み。全領域の型・ID／版・transaction／lock・通知・失敗回復・互換と文書参照を未再照合。既存ADR Acceptedを今回整理済みと数えない |

共通マスタは[screens/settings](../screens/settings/README.md)に保持。21本文の主責務は、動物種D04、cage／hospitalization-plan D05、trimming／course-type D06、examinations D07、payment／insurance／campaigns／merchandise D10、closing-time D12、staff／occupation／shift-template D14、reservation-type D01、treatment／diagnosis／medicine／interview／chief-complaint D03、permission-group C04。設定入口はD17。利用する別領域は参照だけで二重計数しない。

## 横断接続・全体再照合

全8行は確認待ち。予約・受付側の先行整理済み契約を消さず、全ownerの内部仕様が揃ってから同じID／状態・承認境界・保存／再送・認可・訂正・通知を再照合する。連携の正本は[横断業務](../workflows/reservation-to-record-flow.md)、保存は[カルテlifecycle§2](medical-records/lifecycle.md)、技術の現行実装境界は[処理カタログ](../../architecture/cross-domain-orchestration-catalog.md)。現行separate-txと将来選定same-txを同じ状態として扱わない。

| ID・接続 | 主責務と参加領域／先行済み・参照 | 状態・残る照合 |
|:---|:---|:---|
| X01 予約 → 受付 → 通常カルテ | D01予約intent、D02画面、D03記録participant。受付／再受付、準備→読込→開始、終了、版／再送、通常記録同時commitは診療のみ。[予約技術](../../design/domains/reservations/technical-spec.md)・[受付技術](../../design/domains/reception/technical-spec.md) | 確認待ち。予約／受付側は完了。D03全域確定後、現行auto-create・backfill／finalizeと選定契約の差分を再照合 |
| X02 取消・未来院・再受付・記録訂正 | D01状態、D03記録保護、D02導線。手動未来院／同ID再受付・確定後追記等の境界。カタログcancel-cleanup／batch-noshow | 確認待ち。D03の特別訂正内部と記録保存・取消を確認。自動削除・自動持越し・無条件逆遷移の新許可を作らない |
| X03 予約／受付 ↔ 専用施術・滞在 | D01予約、D05入院／ホテル、D06施術、D02画面。分類・typed ref、頭数／延泊・実開始担当、trim appointment create／update／delete | 確認待ち。D05／06内部記録・担当人数／職種・精算との受渡しを照合。通常カルテへ強制変換しない |
| X04 患者台帳・同一個体・マスタ／勤務 | D04台帳、D19identity、D14勤務、D17医院、各master owner、C04scope。owner-pet-registerと予約同医院／個人対応の接続 | 確認待ち。台帳同時登録／変更保護、LINE紐付けD20、scope・master履歴の全owner整合を確認。類似名でID推測しない |
| X05 臨床記録 ↔ 検査・予防・健診・読取公開 | D03記録、D07機器／検査、D08接種、D09健診、D20健康情報、D21レポート。ADR004／007／008 | 確認待ち。記録確定／訂正・結果対応・読取権限／token・履歴公開の境界を再照合 |
| X06 会計確定・訂正 ↔ 診療／見積／退院 | D10決済、D03記録、D01予約complete、D05退院、D06施術、D11見積。通常Undo禁止・確定歴の先行合意、billing-res-complete／hospitalization-discharge-bill | 確認待ち。D10初回5点を部分反映。D10第二群で古い順／変更・予約前金・訂正後表示を追加。第三群で預り金の予約任意・同飼主別予約／ペット利用を追加。別飼主移管／医院間利用は未承認。会計取消時の残高復元を第四群で追加、減額時の充当調整・取消後再請求・権限／具体手順、専用記録／見積内部と全体再照合は残件 |
| X07 会計 ↔ 締め・集計・在庫 | D10金額／支払、D12締め境界、D16分析読取、D13在庫。cash-register／customer-aggregation／ADR003 | 確認待ち。金額・期間・取消／訂正・履歴の集計根拠、物販／在庫の実際の接続を確認。予約整理から締め全文改訂へ広げない |
| X08 予約チャネル・臨床／会計 ↔ CRM背景処理 | D18チャネル、D15配信／同期、D16分析、D01予約条件、各イベントowner。LINE予定枠正本・scheduler／delivery／dormant／tag-health／billing-lstep-CPM | 確認待ち。通知commit後・重複／失敗／停止／再開・Webhook／17トリガーと分析意味を照合。稼働・外部送信の既存承認境界を保持 |

## 対応表の範囲・読む順

main appの88製品leafは業務84＋C04認証／権限4へ一度ずつ配賦。画面の正本対応は[ui-design-compliance§2](../ui-design-compliance.md)を維持し、以下は進捗の責務割当だけ。redirect12件とwildcard404は業務領域や製品画面へ加算しない。別app37はD18、別app38はD20として管理し、main88へ加算しない。

画面なしの通常カルテ保存はD03、予約接続X01／02、レジ集計境界D12／X07、CPM／LTV D16、LINEタグ／Webhook／配信／scheduler D15／X08。技術・運用の既存境界はC06経由で参照し、インフラ変更／デプロイ作業は対象外。master／README本文のサポート契約を対象外にしない。未文書化要件・全APIごとの受入網羅は未確認で、今回のroute配賦だけから完了としない。

読む順は本書の集計・未着手行 → [予約](reservations/README.md) → [受付](reception/README.md) → [カルテ](medical-records/README.md) → [会計・未納](accounting/README.md) → 関連業務D05〜23／共通C01〜06 → 横断8接続の再照合。次の領域へ進む前に、その主責務・未決・技術入口と全体表の状態を更新する。

## main88ルートの主責務配賦

| 管理ID | main製品leaf数 | 対象パス（テンプレート） |
|:---|---:|:---|
| D01 | 2 | `/reservations`、`/settings/reservation-type` |
| D02 | 1 | `/` |
| D03 | 10 | `/medical-records`、`/medical-records/select-pet`、`/medical-records/new`、`/medical-records/:id`、`/settings/treatment-items`、`/settings/diagnosis`、`/settings/medicine`、`/settings/inquiry-templates`、`/settings/interview/chief-complaint`、`/settings/interview/templates` |
| D04 | 4 | `/owners`、`/owners/new`、`/owners/:id`、`/settings/animal-species` |
| D05 | 7 | `/hospitalization`、`/hospitalization/select-pet`、`/hospitalization/new`、`/hospitalization/:id`、`/hospitalization/:id/edit`、`/settings/hospitalization`、`/settings/cage` |
| D06 | 6 | `/trimming`、`/trimming/select-pet`、`/trimming/new`、`/trimming/:id`、`/settings/trimming`、`/settings/trimming-course-type` |
| D07 | 6 | `/examinations`、`/examinations/select-pet`、`/examinations/new`、`/examinations/:id`、`/settings/lab-device-item-masters`、`/lab-device` |
| D08 | 4 | `/vaccinations`、`/vaccinations/select-pet`、`/vaccinations/new`、`/vaccinations/:id` |
| D09 | 3 | `/checkups`、`/checkups/select-pet`、`/checkups/new` |
| D10 | 8 | `/accounting`、`/accounting/select-pet`、`/accounting/new`、`/accounting/:id`、`/settings/merchandise-items`、`/settings/insurance`、`/settings/payment-methods`、`/settings/campaigns` |
| D11 | 4 | `/estimates`、`/estimates/new`、`/estimates/:id`、`/estimates/:id/edit` |
| D12 | 3 | `/accounting/close`、`/accounting/close/history`、`/settings/closing-time` |
| D13 | 3 | `/inventory`、`/inventory/new`、`/inventory/:id` |
| D14 | 4 | `/shifts`、`/settings/staff`、`/settings/occupations`、`/settings/shift-templates` |
| D15 | 4 | `/lstep/checkup-sync`、`/lstep/delivery-monitor`、`/settings/integrations/lstep`、`/settings/lstep/tags` |
| D16 | 3 | `/accounting/reports`、`/aggregation`、`/lstep/analytics` |
| D17 | 2 | `/settings/clinic`、`/settings` |
| D18 | 4 | `/line-reservation`、`/line-reservation/settings`、`/line-reservation/page-editor`、`/line-reservation/slots` |
| D19 | 1 | `/identity-links` |
| D20 | 0 | なし。別app38はD20で管理 |
| D21 | 1 | `/owners/:id/report` |
| D22 | 2 | `/manual`、`/manual/:category/:slug` |
| D23 | 2 | `/settings/bug-reports`、`/settings/chat-history` |
| C04 | 4 | `/login`、`/forgot-password`、`/reset-password`、`/settings/permission-groups` |

主責務未配賦0・重複0。88パスの対応元は上記UI88正本対応表。配賦によって元の仕様正本や画面の実装状態を変更しない。

## 今回の確認記録

2026-10-14、受付の通常記録同時commitを診療に限定し、reloadによる操作key消失・read0件でも先行処理の遅延commitを失敗と判定して再作成しないよう技術契約と既存ACを補完した。独立レビューで2指摘とも解消を確認。受付FRを追加せず、受付AC21件を維持した。

全体管理表は別の読み取り担当2名が分類・集計・状態・主責務・88ルート／21マスタ・別app・画面なし契約・サポート本文を確認し、阻害指摘なし。予約69件の正本索引と技術traceが一致し、仕様全体82FR／214ACの定義に重複なし。これら全体件数を予約の承認69件や領域数に換算しない。

限定文書検査は138文書・2,393相対リンク・1,315アンカー参照、欠落0。全表の列数、37管理行の状態集計、88パスの一意配賦、git diff --check、docs-symbol-drift（634 token）を確認。実装・API・DB・runtime／性能検査、フルmake ciは実施しない。

作業した既存checkoutはmain、HEAD cea3f94ddeba10632ea372afcbe5d89efedfee36。既存dirty変更を保持し、今回の選択7文書以外6,442ファイルの内容hash一致を確認。コード・API YAML・DB・検査script・UI88正本対応・カルテ保存正本・予約本文等へ今回追加編集はしていない。commit／push／PR・Plane操作・公開・デプロイは行わない。

## カルテ初回5項目の部分反映

2026-10-14、初回回答に基づきD03を未着手から進行中へ更新。業務23領域は完了2・進行中1・未着手20。全37管理行は完了2・進行中4・未着手23・確認待ち8。初回5項目をカルテ全体完了にしない。[合意と受入](medical-records/lifecycle.md#medical-record-agreement-first-five)、[次の質問案](medical-records/README.md#medical-record-next-business-questions)を参照。上記の検査数・hash件数は全体管理ファイル作成時の履歴で、今回更新の検証は別に報告する。

カルテ初回反映の検証：独立レビュー2名で旧状態表／図の実装参照明示と医師承認の再質問を修正し再確認、追加阻害指摘なし。141文書・2,466相対リンク・1,364アンカー参照に欠落なし、5新FR／16新AC・仕様全体87FR／230ACは一意。予約69件の索引／技術trace・88ルート配賦・進捗表集計・13文書の表幅、git diff --check、docs-symbol-drift（634 token）を確認。保存§2と確定済み本文保護は保持。既存main cea3f94dを用い、今回13文書以外6,439ファイルの内容hash一致を確認。実装・API YAML・DB・設定・script／test・Plane・commit／push／公開は変更していない。実機／性能・フルmake ciは未実施。

## カルテ入力4項目の追加反映

2026-10-14、Sentinel_6287d7cbcf988191aac47fbe504cfad1 の回答を4FR／15ACへ反映。初回5項目と合わせた追加合意9FR／31ACの文書範囲で、D03全体は進行中のまま。状態数は変更しない。[業務正本](medical-records/pages/medical-record-form.md#medical-record-agreement-input-four)・[次の専用業務質問](medical-records/README.md#medical-record-next-business-questions)。複写の部分成功・保存§2・初回合意を保持、実装／実機は未確認。旧検査数は各時点の履歴で、今回の検証は別に報告する。

入力4項目の検証記録：独立レビュー2名で、無効化と有効master改定のAC条件・技術入口の対象表記を修正し再確認、追加阻害指摘なし。新4FR／15ACと初回5FR／16ACを保持、仕様全体の定義91FR／245ACは一意。141文書・2,500相対リンク・1,395アンカー参照に欠落なし。12文書の表幅、予約69件索引／trace、88ルート配賦、進捗集計、git diff --check、docs-symbol-drift（634 token）を確認。保存§2・確定保護・逐次複写本文と旧アンカーを保持。main cea3f94dと今回12文書以外6,440ファイルの内容hash一致を確認。実装・DB・API YAML・設定・script／test変更・起票・commit／push・公開なし。実機／性能とフルmake ciは未実施。

## カルテ訂正・例外4項目の追加反映

2026-10-14、Sentinel_d24ee1395f8c819196572baa20a16cd5 に基づき[追加正本4FR／14AC](medical-records/lifecycle.md#medical-record-agreement-correction-four)を反映。初回・入力・今回の部分合意計13FR／45AC。D03全体は進行中、業務23領域・全37行の集計は変更しない。治療項目取消と診療終了Undo、必要内容転記と患者ID付替え、空誤作成復元と診療開始を区別。保存§2・確定原記録・会計境界を維持。全タブ・全owner・臨床数値別承認・実装受入は残件。[次の業務質問と設計・受入残件](medical-records/README.md#medical-record-next-business-questions)を参照。

訂正／例外4項目の限定検証：独立レビュー2名で、項目取消のACが会計影響なしまで会計訂正必須に読める条件を修正し再確認、阻害指摘なし。今回4FR／14ACと累計部分合意13FR／45AC、仕様全体の定義95FR／259ACは一意。140文書・2,507相対リンク・1,402アンカー参照、対象11文書の表幅、予約69件索引／技術trace、88ルート、23業務／6共通／8接続と進捗集計を確認。git diff --check、docs-symbol-drift（634 token）はPASS。保存§2はHEADと一致、旧複写本文・旧アンカーも独立レビューで保持確認。main cea3f94d・今回11文書以外6,441ファイルの内容hash一致。実装コード・API YAML・DB・設定・script／test変更、起票、commit／push、公開なし。実機／性能・フルmake ciは未実施。全9タブ・共通操作・他owner接続の残件を部分技術設計に配賦、D03は進行中。

## カルテ追加3項目と全タブ契約の反映

2026-10-14、質問 Sentinel_f99722e55be0819189e5d50080928914 →回答 Sentinel_59d45d62afe08191bc0b289c42486962「ok」に基づき[追加3FR／12AC](medical-records/pages/medical-record-form.md#medical-record-agreement-integration-three)を反映。[全タブ共通14AC](medical-records/pages/medical-record-form.md#medical-record-shared-coverage-ac)も既決保存／認可／確定契約へ接続し、今回合意計16FR／71ACを整理。D&Dと選択の同検証、範囲選択印刷と訂正追記識別、同じ検査結果／履歴と確定時版保持、会計在庫の影響表示とowner案内。自動飼主公開・確定解除・確定請求／返金／在庫の自動訂正を追加しない。

[全9タブ＋バイタルの型／操作契約](../../design/domains/medical-records/technical-spec.md#medical-record-full-coverage-contract)と共通状態／異常系、専用訂正・複写・template・master履歴・owner受渡しを具体化した。今回通常業務の追加質問は不要。D03全体は進行中：実wire／全DTO・旧値・実grant・全writer／consumerとowner契約の最終照合を残す。文書の整理範囲と実装／実機受入、医師による薬量具体値承認#261を分離。[判定範囲と引継ぎ](medical-records/README.md#medical-record-next-business-questions)。業務23／共通6／接続8の集計は据置。

追加3項目・全タブの限定検証記録：独立レビュー3名（根拠／状態／型・API）で、予約Undoの条件限定、AC現累計71件、検査結果全体と版内項目ID、医師確認IDと会計ID、治療master4参照、検査確定契約の所在を修正し再確認。未解消の阻害指摘なし。新3FR／12ACと共通14AC、合意計16FR／71ACが一意。仕様全体の定義98FR／285ACも一意。140文書の2,583相対リンク／1,458アンカー参照と旧FR／AC／アンカー保持、対象14文書の表幅、予約69件索引と技術trace、88ルート、23業務／6共通／8接続と集計を確認。保存§2は作業前・HEADと一致。git diff --check、docs-symbol-drift（634 token）はPASS。今回変更13文書、対象14文書外6,438ファイルの内容hash一致と対象内一覧本文の無変更を確認。main cea3f94dを保持。既存script／test差分を含めdocs外を変更せず、実装・API YAML・DB・設定・起票・commit／push・公開なし。実機／性能・フルmake ci未実施。今回業務の追加質問は不要だが、D03全体は技術最終照合・関連owner確認を残すため進行中、薬量値の医師承認も別工程。

## D03最終照合の追加docs反映

2026-10-14、追加承認 Sentinel_7e6b0beed1b48191ae86bc2332800109 により読み取り監査の差分をdocsへ反映。API／DTO基幹48request・response（補助の体重／summary型は別）の実キー・型・未送信／null／clear、YAML差分、取得元、GET初期化含むwriter・監査／gate、確認医師／編集者・依存版／receipt／旧client、プラン旧状態／snapshot、空判定／専用intent、結果／印刷版と在庫副作用の誤流用防止を具体化。[最終照合](../../design/domains/medical-records/technical-spec.md#medical-record-reconciled-contract)・[16FR／71AC trace](../../design/domains/medical-records/technical-spec.md#medical-record-final-trace)・[文書判定範囲](../../design/domains/medical-records/technical-spec.md#medical-record-document-completion)。現コードを新仕様の承認根拠にせず、数値承認・実旧値／権限・実装／実機／性能／全体統合は未確認のまま。独立レビューと限定検査後にD03の限定範囲文書完成を判定する。

## D03限定文書完成の判定・検証記録

独立レビュー3名で、未送信のPATCH保持／PUT置換定義、会計return主体の医師限定誤読、バイタルcreate／delete監査actorとPATCHの差、健診field PUTの未送信／null／[]、検査PATCHと旧PUT全消去の互換差、在庫連動create整数数量・DTO集計範囲を修正し再確認。未解消阻害なし。明示した16FR／固有57AC＋共通14AC＝71ACの文書traceに欠落／重複なし。D03はGの通常カルテ・関連master・受け手技術契約の文書範囲で完了と判定し、業務23行は完了3・進行中0・未着手20、管理37行は完了3・進行中3・未着手23・確認待ち8へ更新。共通6行・横断8行の状態は維持。

140文書・2,665相対リンク・1,506アンカー参照と旧FR／AC／アンカー保持、仕様全体の定義98FR／285ACの一意性、対象13文書の表幅、予約69件索引／技術trace、88ルートと進捗実行分類を確認。git diff --check、docs-symbol-drift（634 token）はPASS。保存§2は作業前・HEADと一致。main cea3f94d、今回13文書以外6,439ファイルの内容hash一致を確認し、backend/docs/api.yaml・既存script／test差分を保持。docs外・実装・DB・設定・起票・commit／push・公開を変更していない。実権限・実旧値／DB制約・薬量具体値の医師承認、全writer／consumer実装適合・旧client切替・利用開始、他owner内部・全体横断最終確認、実機・並行性／性能は未確認／未実施。文書完了を実装／運用完了へ変換しない。次の推奨はD10会計。

## 会計初回5点の部分反映

2026-10-14、[合意根拠と5FR／20AC](accounting/lifecycle.md#accounting-agreement-first-five)を業務・語彙・画面参照・技術案へ反映。D10は進行中で全体完了にしない。業務23領域は完了3・進行中1・未着手19、全37管理行は完了3・進行中4・未着手22・確認待ち8。単価手修正と割引併用は許可済み、割引順・税／端数／保険詳細、配分優先順位・前受け／過入金等は未決。既存検査記録は各時点の履歴で、今回の実装・実機合格ではない。保存§2・通常Undo停止・カルテ訂正からの自動金銭／在庫巻戻し禁止を保持。起票・実装・API／DB・docs外変更・commit／push／公開は行わない。

会計初回部分反映の確認：読み取り専用の独立業務レビュー・独立技術レビューで修正を要する指摘なし。5FR／20ACと語彙・技術traceを照合し、仕様全体の103FRアンカー／305AC定義に重複なし。spec／designの138文書・2,545相対リンク・1,553アンカー参照に欠落なし、新規本文の表幅と37管理行の集計を確認。git diff --checkとdocs-symbol-drift（634 token）はPASS。対象12文書以外6,444ファイルの内容hash、main／HEAD cea3f94d、カルテlifecycle全体とUI88対応表は開始時と一致。保存§2はレビューでもHEADとの本文一致を確認。実装／API／DB／設定／script変更・起票・commit／push／公開はなし。runtime・実機・移行・並行性／性能・税務適合は未確認／未実施で、フルmake ciは行わない。

## 会計第二群の追加反映

2026-10-14、[第二群の決定根拠](accounting/lifecycle.md#accounting-agreement-first-five)と[追加3FR](accounting/lifecycle.md#accounting-agreement-second)を反映。初回5FR／20ACのIDを保持し、旧未決の配分・割引順条件は追加合意へ更新。20AC追加で計8FR／40AC。前金を扱わない案は不採用。前金の受領／未充当残／充当／実返金と元予約・飼主を追跡し、充当の二重入金・取消だけで返金済みになる扱いを禁止。訂正後未払／返金必要額・未実行／実返金、支払方法の誤記訂正、割合→金額・保険確認修正、帳票分離／再発行・旧履歴・売上なし過去未納表示、原請求日／実入出金日を整合。D10は進行中、管理37行の集計は前回と同じ。ホテル固有必要額／キャンセル料D05、別予約振替／予約なし預り金・複数割合・保険割引前後／粒度・税端数・具体認可／帳票詳細等は未決。以前の確認記録は各時点の履歴、今回の検査は別記する。

会計第二群の確認：読み取り専用の独立業務レビューに修正必須指摘なし。独立技術レビューのP3（画面3文書の初回のみmetadata表記）を初回・第二群へ修正し、同レビュー担当の再確認で解消。8FR／40ACの定義と技術trace40件に重複・欠落・余分なし。仕様全体106FRアンカー／325ACは一意。spec／designの138文書・2,582相対リンク・1,590アンカー参照に欠落なし、表幅・37管理行の集計を確認。git diff --check、docs-symbol-drift（634 token）はPASS。本作業の書込は対象12文書のみ。対象外6,444ファイルの全体hashは最終確認で開始時と不一致になり、対象外ontology/reservationsの更新時刻が新しくなっているため他の編集として保持し、全ファイル不変とは報告しない。差分の全件特定と編集者は未確認。カルテlifecycle全体／UI88対応・mainのHEAD cea3f94dは開始時と一致。独立レビューで保存§2とHEADの本文一致も確認。実装・docs外変更・起票・commit／push／公開なし。runtime・実機・全writer／権限適合・移行／並行性／性能・税務適合は未確認／未実施。フルmake ciは行わない。

## 会計第三群の追加反映

2026-10-14、[第三群の合意根拠](accounting/lifecycle.md#accounting-agreement-third)に基づき既存8FRを更新し20AC追加、計8FR／60AC。予約必須から飼主単位預り金・予約任意へ更新。同飼主の別予約／別ペットへ確認付きで利用し、元飼主・任意元予約・受領／充当／実返金と利用可能額、二重使用防止・原子的保存／監査を整合。割合合算→金額→割引後保険、同日作成順・理由付き保存後再配分、非現金超過の返金必要、別方法返金・金額追加承認なし、用途別帳票と再発行時の元番号維持を反映。D10は進行中、管理37行の集計は変更しない。税端数／粒度／100%超・値引き超過、充当解除／残高復元、ホテルD05条件・D12締め内部・全master／owner照合などは残件。既決の通常Undo・カルテ保存§2・自動金銭／在庫巻戻し禁止を維持し、実装や外部決済自動連携／規約適合を承認しない。過去の検証記録は各時点の履歴。今回の基準は実mainのHEAD 09e28ddb、開始時clean。

会計第三群の確認：独立業務／技術レビューで共通権限前提の旧表記（返金追加承認まで未決と読めるP2）を指摘。訂正・取消の確認段階等／実grant mappingと、返金の金額追加承認なしという既決を分離して修正し、両担当の再確認で解消。8FR／60ACと技術trace60件に重複・欠落・余分なし。仕様全体106FRアンカー／345ACは一意。spec／designの138文書・2,611相対リンク・1,619アンカー参照に欠落なし、新規表幅と管理37行の状態集計を確認。git diff --check・docs-symbol-drift（634 token）はPASS。対象12文書以外6,475ファイルの内容hash、カルテlifecycle全体・UI88対応・予約語彙とmainのHEAD 09e28ddbは開始時と一致。両レビューでも保存§2のHEAD本文一致を確認。実装・docs外変更・起票・commit／push／公開なし。実grant／全writer実装適合、API／DB反映・移行／並行性／性能・実機・税務／決済規約適合は未確認／未実施。フルmake ciは行わない。

## 会計第四群の部分反映

2026-10-15、[第四群](accounting/lifecycle.md#accounting-agreement-fourth)のユーザー Sentinel_e7c15be2b17081918c35bdc4fc8f937bに基づき、明細と会計全体の両割引、100%超／商品額超の値引きは入力エラーでなく警告、会計取消で充当済み預り金を飼主残高へ戻し実返金は別操作を反映。既存8FRを更新し6AC追加、計8FR／66AC。警告後の0円下限／負額は回答待ちで確定せず、税端数・両粒度併用順序／配賦も未決。取消／充当解除／残高復元の一度処理・原子的保存／監査を扱い、新受領／実返金／通常Undo再開を副作用にしない。D10は進行中、管理37行の集計は変更しない。実mainのHEAD 09e28ddbを基準に既存12docsの第三群変更を保持し、実装・docs外・設定・起票・commit／push／公開は行わない。過去の記録は当時の履歴、今回検査は別記する。

## 会計第四群の追加回答と残高式の訂正

2026-10-15、[第四群の追加回答](accounting/lifecycle.md#accounting-agreement-fourth) Sentinel_2fd30155c5a88191956dc21f6f7a7740に基づき、超過値引きは警告のまま承認して進められ、その請求額はマイナスにせず0円下限と確定した。上の部分反映時点の「回答待ち」は履歴として保持し、現在の決定状態は本項と正本を参照する。既存AC識別子を保持して期待結果・語彙・画面参照・技術traceを更新し、計8FR／66ACは変わらない。

預り金の残高式は受領累計−有効充当額−未充当分の実返金累計へ訂正。有効充当額は会計取消で解除・残高復元済みの額を充当累計から差し引き、元飼主へ戻す残高を反映する。新受領・実返金と復元を区別し、実返金／復元済み分の二重計上を防ぐ。税・端数、両粒度割引の併用順序／配賦、取消以外の減額や既実返金等が混在する訂正手順は未決。D10は進行中、予約領域の完了判定と管理37行の集計は変更しない。

今回の基準はリモートmain 7309b3a048cb781298c5ed315f6a653d0dfdda1b。GitHub上のdocs限定ブランチからmain向けDraft PRとする。実装・API／DB・設定・移行・実機・税務適合の検証は行わない。過去のローカル検査記録を今回の合格証拠へ転用せず、今回のリモート文書検査・独立レビューの結果と実行できない検査はPRに分けて記録する。

## 会計第五群の計算順序・税・端数の反映

2026-10-15、[第五群の承認根拠](accounting/lifecycle.md#accounting-agreement-fifth)に基づき既存8FRを更新し12ACを追加、計8FR／78AC。明細値引き→会計全体値引き、明細後の対象額比の比例配賦、同scope割合合算・割合額四捨五入→定額、配賦切捨て後の剰余降順調整、税抜計算／飼主向け税込表示、請求書・税率単位の税切捨て、超過警告後の実額上限／他行転用なしを反映。過去確認記録の税・順序・配賦「未決」は当時の履歴で、現在は正本の第五群へ従う。

業務正本の計算例・snapshot、語彙、技術trace、画面と横断フローの参照要約を追従。同率配賦の凍結明細順は技術提案で、業務承認とは区別する。税区分・税率はマスタ由来の会計時点の値を保持し、全品目一律税率を前提にしない。国税庁の適格請求書の税率別1回端数処理と、製品が選んだ切捨て方針を区別する。保険の具体式／税基準、対象外項目の定義、小数数量等の元明細額確定、その他の未決5群と税務適合全体の確認は残す。

PR #670の既存head 59888f82f3dff8476d80ed3146257be6eae07c0eから同じdocs限定ブランチへ追記し、預り金残高式の訂正を保持する。D10は進行中、予約領域の完了・管理37行の状態集計・通常Undo停止・カルテ保存§2は不変。実装・API／DB・設定・移行・実機は変更／検証せず、GitHubでの本文限定静的検査と独立レビューの結果・未実行検査をPRへ分けて記録する。CIの完了は待たない。

## 会計第六群の帳票・保険・マスタ反映

2026-10-15、[第六群の承認根拠](accounting/lifecycle.md#accounting-agreement-sixth)に基づき、複数保険会社を扱う前提（アニコムは例）、品目マスタでの会計全体割引除外、飼主宛名・ペット別明細、訂正／部分返金の元取引と今回変更額の区別、管理者と権限付きスタッフによる価格・割引・保険・支払方法マスタ変更を反映。既存8FR／78ACを保持して1FR・14ACを追加し計9FR／92AC。保険会社・商品一覧と具体式は未決。旧マスタ記載の自動設定未実装・固定4種や品目の税率例を新しい方針の承認にしない。

PR #670の開始head 2fb7cdc92249c5f8a3266a0aa184c3f8a8bba52aから同じdocs限定ブランチを更新。現行permissionの対応と不足を読取調査し、実際のstaff割当・runtime設定は変更／検証しない。D10進行中、他管理行の集計・予約の完了判定・通常Undo停止・カルテ保存§2、預り金の有効充当額による残高式・0円下限・第五群の税抜計算／端数を保持。限定静的検査・独立レビュー・リモート最終本文再取得をPRに記録し、CI完了は待たない。実装・API／DB・移行・公開は対象外。

## 会計第七群の窓口保険範囲反映

2026-10-15 03:49 UTC（12:49 JST）、[第七群の承認根拠](accounting/lifecycle.md#accounting-agreement-seventh)の3点を反映。スタッフが保険会社の画面等で資格・残限度を確認し結果と保険負担額をAnimalEkarteへ記録する。今回の範囲は窓口会計まで。保険会社への請求送信・入金照合・査定差額管理と、同じ診療費への複数契約併用／給付調整は対象外で、1会計につき各ペットの利用契約は1つ。未確認／不明と有効、窓口利用不可と補償対象外を混同せず、ローカル訂正から外部保険状態を自動変更しない。会社／商品別の自動給付算式は将来自動化時の別スコープで、現在のスタッフ記録運用の前提にしない。過去記録の保険具体式「未決」は当時の履歴とし、現在の範囲は第七群を正とする。

既存9FR／92ACの識別子を保持し7ACを追加、計9FR／99AC。PR #670の開始head caf1a622c40dd14782e30eef08c3cbaa547eb21dから同じdocs限定ブランチを更新。追加の必須入力欄・API連携を確定しない。D10進行中、他管理行と予約の完了判定・通常Undo停止・カルテ保存§2、預り金残高式・税抜計算／端数・値引き上限・帳票／マスタ権限は保持。独立レビュー・静的照合・最終本文再取得の結果と未実行検査をPRに記録する。ユーザーのローカルファイル・runtime・設定・API／DB・移行・merge／deploy・Cloudflare Artifactsは対象外、CI完了は待たない。

## 会計第八群の預り金・訂正・回収反映

2026-10-15 04:03 UTC（13:03 JST）、[第八群の承認根拠](accounting/lifecycle.md#accounting-agreement-eighth)に従い5点を反映。既存9FR／99ACの識別子を保持して17ACを追加し9FR／116AC。減額解除分の元飼主残高への復元と併用内訳確認、自動失効なし・同医院同飼主限定、選択確認した取消元明細の新番号再請求と関連付け、新規会計と過去未納回収の同時扱い／元請求別配分、共通料金別枠を同期した。別医院移行・別飼主譲渡、督促／貸倒は今回対象外。

PR #670の開始head 2bcc49b5bdc3cddb3f72e1cef28d1456dc68b792からdocs限定で更新。D10は進行中。再請求時の価格／税／割引の採用時点・元請求に残る入金／保険確認結果の扱い、出所不明の既実返金の解消手順と既存設計残件を未決として保持し、全体完了としない。他管理行・予約の完了判定／通常Undo境界・カルテ保存§2は不変。独立レビュー、AC追跡・リンク・表構造の限定静的照合、最終本文の再取得結果をPRに記録する。実装・ローカルファイル・runtime・設定・API／DB・移行・merge／deploy・Cloudflare Artifactsは対象外、CI完了は待たない。

過去の会計追記にある減額・失効・再請求・共通料金等の「未決」は当時の履歴。現在の合意範囲は第八群とD10行を正とし、解決済み項目を未決へ戻さない。
