# reception/ — 受付ドメインの技術設計

既合意の受付表示・操作・通信／競合・ブラウザ設定を実現する方式を置く。読む順は[業務正本](../../../spec/domains/reception/README.md) → [予約の語彙](../../../spec/ontology/reservations.md) → [横断フロー](../../../spec/workflows/reservation-to-record-flow.md) → [受付技術仕様](technical-spec.md) → 同書の参照API／コード・検証計画。

本書とtechnical-specがこのフォルダの全体。業務の許可・必須・受入結果はspecを正とし、操作版／再送・全writer／transactionは予約技術契約を参照する。別の技術正本やタスクファイルを増やさず、依存関係・担当・進捗はPlane、実施証跡はdocs/workへ分ける。今回は文書設計のみで実装・起票・公開は行わない。

作成根拠は2026-10-14の Sentinel_6bdf54904d3881918fb3e884820adfea。将来設計・現在実装の静的根拠・未実施検証・他owner内部を技術仕様で区別する。[領域設計入口](../README.md)
