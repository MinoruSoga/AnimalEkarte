# LINE・Lステップ連携 統合ドキュメント (LINE Integration Hub)

> **目的**: LINE/Lステップ連携ドキュメントの索引を提供する。
> **読者**: 新規参加開発者・非エンジニア(院長等)。
> **タイミング**: LINE連携機能の全体像を把握したい時。

> **Animal Ekarte**: 飼い主向け予約体験と CRM 戦略の統合
> **最新更新**: 2026-08-31

---

## 1. ドキュメント構成

本ディレクトリには、LINE プラットフォーム（LIFF, Messaging API）および Lステップを活用した、Animal Ekarte の外部連携機能に関する全ての仕様が集約されています。

### 技術・アーキテクチャ
- **[architecture.md](./architecture.md)**: 予約システム(v1)と Lステップ(v2)の全体像、認証フロー、非同期同期ロジック。
- **[reservation-spec.md](./reservation-spec.md)**: 飼い主向け予約アプリ（LIFF）の機能要件、画面遷移、空き枠計算エンジン。

### 導入・設定
- **[setup.md](./setup.md)**: LINE Developers Console、Messaging API、および Lステップ管理画面の初期設定手順。
- **[lstep-integration.md](./lstep-integration.md)**: **【重要】** マーケティング戦略、CPM 判定ロジック、15 種の自動配信トリガー詳細。
- **[cost-analysis.md](./cost-analysis.md)**: Messaging / Lステップ課金と配信ボリューム試算（docs-only）。

### 原本
- クライアント受領原本の外部 restricted evidence location は CorpVault `evidence/2026-08-20-docs-cleanup/client/`（repository 外であり、この commit の検証対象・開発時の必須依存ではない）。製品側の repository SoT は [lstep-integration.md](./lstep-integration.md)。

---

## 2. システム概要

本機能は、電子カルテ本体（Go API）を共通の脳とし、飼い主が直接操作する **LIFF App** と、病院側が運用する **Lステップ管理基盤** を高度に連携させます。

### 主要な価値
- **オペレーションの自動化**: LINE 予約がカルテ受付（カンバン）へ即座に反映され、スタッフの手入力コストを削減。
- **臨床データに基づく CRM**: 診察結果や最終来院日に基づき、Lステップが「忘れられない病院」として自動で飼い主をフォロー。
- **競合を拒否する空き枠管理**: 作成transaction内で予約種別規則とappointment conflictを再検証する。明示staffはclinic所属・capability・LIFF公開/activeを検査するが、**選択時刻のshift再検証は未実装のsource gap**であり、frontendのshift絞り込みだけを安全根拠にしない。

---

## 3. クイックリンク

| 項目 | 管理画面 URL |
|:---|:---|
| **連携設定** | `/settings/integrations/lstep` |
| **文言編集** | `/line-reservation/page-editor` |
| **タグ管理** | `/settings/lstep/tags` |
| **分析レポート** | `/lstep/analytics` |
| **対象者抽出** | `/lstep/checkup-sync` |
| **配信監視** | `/lstep/delivery-monitor` |

---
