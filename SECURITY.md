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

### マルチテナント隔離

- すべての患者・飼い主・診療データは `clinic_id` で完全に隔離される
- GORM の `clinicScope` を repository 層で強制適用し、クロステナントアクセスを構造的に防止
- `clinic_id` なしの SELECT/UPDATE/DELETE は P4 規約違反として CI で検出

### 認証・認可

- JWT (HS256) によるステートレス認証
- Role-based access control: `system_admin` / `admin` / `staff`
- `RequirePermission` ミドルウェアにより全エンドポイントで権限チェック
- リフレッシュトークン 15分ローテート

### 入力検証

- バックエンド: Gin の ShouldBindJSON + apperrors による型安全バリデーション
- フロントエンド: React 19 useActionState + TypeScript による型安全フォーム
- SQL インジェクション対策: GORM パラメータバインディングを強制、生 SQL 禁止

### シークレット管理

- 実運用シークレットは GitHub Actions Secrets / 環境変数のみ
- `.env` ファイルは `.gitignore` で除外済み
- API キー・パスワードのハードコード禁止 (pre-commit フック + code review で検出)

### 医療情報保護

- 飼い主・ペット・診療記録は clinic_id スコープで隔離
- audit_log テーブルにより全データ変更を記録
- 削除は論理削除 (deleted_at) または変更追跡で履歴保持

## 既知の制限事項

- 本システムは日本国内の動物病院向けです。人医療の患者記録は対象外です。
- 現在、二要素認証 (2FA) は未実装です (計画中)。

## セキュリティ更新の通知

GitHub の Watch → Custom → Security Advisories を有効にしてください。
