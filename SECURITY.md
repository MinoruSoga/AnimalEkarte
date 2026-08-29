# Security Policy — AnimalEkarte

## 対応するバージョン

| バージョン | サポート状況 |
|-----------|------------|
| main (最新) | ✅ セキュリティパッチ適用中 |
| staging | ✅ セキュリティパッチ適用中 |
| その他 | ❌ 未サポート |

## 脆弱性の報告

**公開 Issue での報告は避けてください。**

脆弱性を発見した場合は、以下の連絡先に直接ご報告ください:

- **Email**: baritech.soga@gmail.com
- **件名**: `[SECURITY] AnimalEkarte - <概要>`
- **初期返答**: 48時間以内

### 報告に含める内容

1. 脆弱性の種類 (XSS, SQL Injection, 認証バイパス 等)
2. 再現手順 (できる限り詳細に)
3. 影響範囲 (どのデータ・機能が影響を受けるか)
4. 提案される修正方法 (任意)

## セキュリティ設計

> **正本**: 認証・認可の詳細は [`docs/architecture/auth.md`](docs/architecture/auth.md)。  
> データ経路・監査の path-dependent 方針は [`docs/architecture/data-flow.md`](docs/architecture/data-flow.md)。  
> 本ファイルは入口の要約であり、断定が食い違うときは上記を優先する。

### マルチテナント隔離

- すべての患者・飼い主・診療データは `clinic_id` で完全に隔離される
- clinic scope は **domain/capability 境界**で強制する（BE9 後。旧「repository 層で一律 `clinicScope`」という固定表現は使わない）
- 実装・境界の正本: [`docs/architecture/overview.md`](docs/architecture/overview.md) · ADR-006
- `clinic_id` なしの危険な SELECT/UPDATE/DELETE は規約・lint で抑止する（詳細は backend 規約）

### 認証・認可

- JWT (HS256) によるステートレス認証（dual-token）
  - **Access Token**: 約 **15 分**有効（httpOnly Cookie）
  - **Refresh Token**: 最大 **7 日** · family ID + JTI · ローテーションと再利用検知（詳細は auth.md §4.1）
  - 「リフレッシュが 15 分でローテートする」一括断定はしない
- Role-based access control: `system_admin` / `admin` / `staff` + リソース単位 RBAC（`AllResources` 37 種 · auth.md）
- スタッフ向け API は `RequirePermission` / `RequirePermissionAny` で権限チェックする
  - **例外**: LIFF 等の **公開ルート**はスタッフ RBAC 前提ではない（全エンドポイント RequirePermission ではない）
- リクエスト時の clinic 再解決・stale token fail-closed は auth.md §4.2

### 入力検証

- バックエンド: Gin の ShouldBindJSON + apperrors による型安全バリデーション
- フロントエンド: React 19 useActionState + TypeScript による型安全フォーム
- SQL インジェクション対策: **原則** GORM のパラメータバインディングを使う
  - 「生 SQL 全面禁止」ではない。限定された parameterized `Raw` / 運用クエリは domain 内に存在し得る（新規の文字列連結 SQL は禁止）

### シークレット管理

- 実運用シークレットは GitHub Actions Secrets / 環境変数のみ
- ローカルの `.env` / `.env.local` は Git 管理外（`.gitignore`）
- **例外**: `frontend/.env.production` は公開可能な `VITE_*` のみを意図的に track する（秘密を置かない）
- API キー・パスワードのハードコード禁止 (pre-commit フック + code review で検出)

### 医療情報保護

- 飼い主・ペット・診療記録は clinic_id スコープで隔離
- **監査 (audit) は path-dependent**: すべての CUD が機械的に `audit_logs` へ入るわけではない（例: LSTEP タグ同期は意図的に audit 対象外 · data-flow.md）
- 必須監査は業務 write と同一 transaction で fail-closed（auth.md §4.3）
- 削除は論理削除 (deleted_at) または変更追跡で履歴保持（経路による）

## 既知の制限事項

- 本システムは日本国内の動物病院向けです。人医療の患者記録は対象外です。
- 現在、二要素認証 (2FA) は未実装です (計画中)。

## セキュリティ更新の通知

GitHub の Watch → Custom → Security Advisories を有効にしてください。
