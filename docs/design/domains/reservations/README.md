# reservations/ — 予約ドメインの技術設計

予約区分、指名・対応可否、シフト、受付とカルテ、入院・ホテル、LINEの接続を設計する。読む順は[確定業務仕様](../../../spec/domains/reservations/README.md#approved-reservation-scope) → [語彙・概念関係](../../../spec/ontology/reservations.md) → [技術仕様](technical-spec.md) → 同書のコード・API・DB参照と検証方針。

技術仕様は実現方式・現行差分・判断待ちを記す。業務許可や受入結果を再定義しない。具体的な作業分割・依存関係・進捗はPlane、実施した調査・変更・検証の記録はdocs/workへ置く。今回tasks.md、docs/plan、Planeチケットを作らず、下位フォルダも新設しない。

現在コードを静的に照合した範囲と、提案方式・未確認を技術仕様で区別する。既存workの予約design/archive資料は当時の記録として保持し、今後の技術設計の正本を増やさない。[設計全体](../../README.md)
