# domains/ — 領域別の技術設計

業務領域ごとの実現方式と現行コードとの差分をまとめる。[予約](reservations/README.md)と[受付](reception/README.md)と[カルテの部分設計](medical-records/README.md)と[会計初回〜第六群の部分設計](accounting/README.md)が現在の入口。新しい領域を追加する際も、領域READMEに対象・読む順・正本を記す。

ここにはデータ・API・画面と処理・transaction・移行・検証の設計を置く。業務本文は[spec/domains](../../spec/domains/README.md)、全体アーキテクチャは[architecture](../../architecture/README.md)、実施記録は[work](../../work/README.md)を参照し、本文を二重管理しない。具体的タスク・依存関係・進捗はPlaneへ置く。

読む順：対象領域の業務正本 → 領域README → 技術仕様と参照コード。各記述の「現在実装／提案設計／業務判断待ち／未検証」を別軸で読む。本索引への掲載は実装・公開の許可ではない。
