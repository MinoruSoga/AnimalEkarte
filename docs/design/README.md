# design/ — 技術設計

確定業務を実現するDB・API・画面連携・内部状態・整合性・移行・検証の技術設計を置く。業務上の許可・必須条件・受入結果は[spec](../spec/README.md)が正本、具体的作業・依存関係・進捗はPlane、実施記録は[work](../work/README.md)。設計提案は実装済みや業務承認済みを意味しない。

[domains/](domains/README.md)が領域別の入口。全体構造・write owner・認可・ADRは既存[architecture](../architecture/README.md)を参照し、同じ規則を再定義しない。API wire契約はbackend/docs/api.yaml、現行DB定義はbackend/migrationsを維持する。設計書には予定する変更と理由・未確定部分を記す。

読む順：業務正本 → 関連architecture → 領域の技術仕様 → 参照するAPI／DB／コードと検証方針。技術仕様を追加する場合はその領域READMEへ案内を置く。タスク台帳・tasks.md・新しい業務仕様のコピーは作らない。

今回の配置はユーザー Sentinel_4090fd2153c881918e1da545da9ab505 により承認。旧work配下の設計・計画資料は今回移動していない。文書分類とdocs-symbol-driftのroot allowlistにdesignを追加済み。検査側の追加はユーザー Sentinel_34b62bb163b881919e6f2a9fd05edfc3 の個別承認に基づき、未知カテゴリの拒否・シンボル／数値検査は維持する。Spec Kitの出力先・preset・生成skillsは未変更。


## 2026-10-15 全領域仕様への案内

[全業務の技術契約](domains/README.md)を参照する。既存のルートprimary対応と過去の実行検証記録は保持し、今回の文書検査をruntime／実機合格へ転用しない。

## 2026-10-15会議4点の最新入口

自動保存維持・終了後通常編集・タブ開始時復元／複製直前取消・終了履歴、トリミング開始／終了・料金、勤務複製／CSVの合意を追加整理。[全体進捗](../spec/domains/SPECIFICATION_PROGRESS.md#meeting-four-decisions-20261015)と各owner本文を参照。途中の保存ボタン方式は撤回済み。今回差分は独立内容・構造レビューで重大指摘なしを確認し、文書再整合済み。全過去承認の監査完了・実装完了とは扱わない。

## 原資料追加採用6点の入口

[職員・月次未納・日次内訳・薬量の追加差分](../spec/domains/SPECIFICATION_PROGRESS.md#source-six-adoption-20261015)を参照する。採用範囲だけの文書再整合で、全原資料の読解完了ではない。職種／役職／雇用と認可、休業と退職、入金と充当／減額、税表示と元税額、自動薬量範囲と臨床数値承認を区別する。月次調整表示は回答待ち、内容レビュー・公開・実装は別段階。
