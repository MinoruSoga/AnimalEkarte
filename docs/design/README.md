# design/ — 技術設計

確定業務を実現するDB・API・画面連携・内部状態・整合性・移行・検証の技術設計を置く。業務上の許可・必須条件・受入結果は[spec](../spec/README.md)が正本、具体的作業・依存関係・進捗はPlane、実施記録は[work](../work/README.md)。設計提案は実装済みや業務承認済みを意味しない。

[domains/](domains/README.md)が領域別の入口。全体構造・write owner・認可・ADRは既存[architecture](../architecture/README.md)を参照し、同じ規則を再定義しない。API wire契約はbackend/docs/api.yaml、現行DB定義はbackend/migrationsを維持する。設計書には予定する変更と理由・未確定部分を記す。

読む順：業務正本 → 関連architecture → 領域の技術仕様 → 参照するAPI／DB／コードと検証方針。技術仕様を追加する場合はその領域READMEへ案内を置く。タスク台帳・tasks.md・新しい業務仕様のコピーは作らない。

今回の配置はユーザー Sentinel_4090fd2153c881918e1da545da9ab505 により承認。旧work配下の設計・計画資料は今回移動していない。文書分類とdocs-symbol-driftのroot allowlistにdesignを追加済み。検査側の追加はユーザー Sentinel_34b62bb163b881919e6f2a9fd05edfc3 の個別承認に基づき、未知カテゴリの拒否・シンボル／数値検査は維持する。Spec Kitの出力先・preset・生成skillsは未変更。
