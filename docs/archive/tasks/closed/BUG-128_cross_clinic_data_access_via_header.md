# BUG-128: X-Clinic-ID ヘッダーでクロスクリニックデータアクセスが可能

## 概要
`X-Clinic-ID` HTTP ヘッダーを改ざんすることで、認証済みユーザーが**所属していないクリニックのデータを
閲覧・作成・更新・削除**できる。ヘッダーの値をバックエンドが無条件に受け入れ、
ユーザーの所属クリニックとの照合を行っていない。

**マルチテナントの根幹を破壊する致命的脆弱性。**

## 脆弱性分類
- **CWE-639**: Authorization Bypass Through User-Controlled Key (IDOR)
- **CWE-284**: Improper Access Control
- **OWASP A01:2021**: Broken Access Control
- **影響**: 任意のクリニックのデータを閲覧・改ざん・削除可能。データ漏洩・破壊のリスク。

## 再現手順

### 前提
- `admin@example.com`（八王子院、clinic_id=3）でログイン
- 城東医院（clinic_id=4）、敷島医院（clinic_id=5）は本来アクセス不可

### 再現1: 閲覧
```bash
# 城東医院の飼主データを取得
curl -b 'auth_token=<JWT>' \
  -H 'X-Clinic-ID: 4' \
  http://localhost:8080/api/v1/owners
# → 200 OK, 8件の城東医院データが返る
```

### 再現2: データ改ざん
```bash
# 城東医院のクリニック名を変更
curl -X PATCH -b 'auth_token=<JWT>' \
  -H 'X-Clinic-ID: 4' \
  -H 'Content-Type: application/json' \
  -d '{"name": "不正更新テスト"}' \
  http://localhost:8080/api/v1/clinics/4
# → 200 OK, クリニック名が変更される
```

### 再現3: データ削除
```bash
# 城東医院のスタッフを削除
curl -X DELETE -b 'auth_token=<JWT>' \
  -H 'X-Clinic-ID: 4' \
  http://localhost:8080/api/v1/masters/staffs/16
# → 204 No Content, スタッフが論理削除される
```

## ブラウザテスト結果

| テスト | ヘッダー | 期待 | 実際 | 判定 |
|--------|---------|------|------|------|
| GET /owners | X-Clinic-ID: 4 | 403 or 八王子データのみ | **200, 城東8件** | ❌ |
| GET /masters/staffs | X-Clinic-ID: 4 | 403 | **200, 城東4件** | ❌ |
| PATCH /clinics/4 | X-Clinic-ID: 4 | 403 | **200（更新成功）** | ❌ |
| DELETE /masters/staffs/16 | X-Clinic-ID: 4 | 403 | **204（削除成功）** | ❌ |
| GET /masters/permission-groups | X-Clinic-ID: 4 | 403 | **200, 城東2グループ** | ❌ |
| GET /owners | X-Clinic-ID: 5 | 403 | **200, 敷島8件** | ❌ |

## 現状コード

### `backend/internal/middleware/auth.go` — X-Clinic-ID の処理

```go
clinicID := claims.ClinicID
if headerClinicID := c.GetHeader("X-Clinic-ID"); headerClinicID != "" {
    clinicID = headerClinicID  // ❌ 無条件で上書き。所属チェックなし
}
c.Set("clinic_id", clinicID)
```

JWT の `clinic_id` を `X-Clinic-ID` ヘッダーで**無条件に上書き**している。
このヘッダーはクリニック切り替え機能（フロントエンドの Sidebar）のために導入されたが、
**ユーザーの所属クリニック（`staff_clinic_assignments`）との照合がない。**

## 修正方針

### `backend/internal/middleware/auth.go`

`X-Clinic-ID` ヘッダーを受け入れる前に、ユーザーがそのクリニックに所属しているか検証する:

```go
clinicID := claims.ClinicID
if headerClinicID := c.GetHeader("X-Clinic-ID"); headerClinicID != "" {
    // is_system_admin はすべてのクリニックにアクセス可能
    if claims.IsSystemAdmin {
        clinicID = headerClinicID
    } else {
        // ユーザーの所属クリニックに含まれているか検証
        headerID, err := strconv.ParseUint(headerClinicID, 10, 64)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid clinic id"})
            c.Abort()
            return
        }
        // staff_clinic_assignments から所属クリニック一覧を取得
        assignments, err := staffClinicRepo.FindByStaffID(c.Request.Context(), staffID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
            c.Abort()
            return
        }
        found := false
        for _, a := range assignments {
            if a.ClinicID == headerID {
                found = true
                break
            }
        }
        if !found {
            c.JSON(http.StatusForbidden, gin.H{"error": "not assigned to this clinic"})
            c.Abort()
            return
        }
        clinicID = headerClinicID
    }
}
c.Set("clinic_id", clinicID)
```

### パフォーマンス考慮
- 所属クリニック情報は JWT claims に含めるか、Redis キャッシュに保持して毎リクエスト DB 問い合わせを避ける
- JWT claims に `clinic_ids: [3, 4]` を含める案が最も軽量

### 代替案: JWT claims に所属クリニック一覧を含める

```go
type Claims struct {
    StaffID       uint64   `json:"staff_id"`
    ClinicID      string   `json:"clinic_id"`
    IsSystemAdmin bool     `json:"is_system_admin"`
    ClinicIDs     []uint64 `json:"clinic_ids"`  // 追加: 所属クリニック一覧
}
```

ログイン時に `staff_clinic_assignments` から全所属クリニックを取得し、JWT に含める。
ミドルウェアでは `claims.ClinicIDs` に含まれるかチェックするだけで DB 問い合わせ不要。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — バックエンド・アーキテクチャ規約
> handler → service → repository の軽量レイヤードを徹底

クリニック所属チェックは middleware 層の責務。全エンドポイントに自動的に適用される。

### `.claude/rules/database-design.md` — マルチテナント設計
> **全テーブルに `clinic_id` (マルチテナント)**
> **WHERE 句は `clinic_id` から開始**

マルチテナント設計の前提は「ユーザーが自分のクリニックのデータのみアクセスできる」こと。
`X-Clinic-ID` ヘッダーの無条件受け入れはこの前提を完全に破壊する。

### `.claude/rules/security.md` — Authentication
> "Use secure session management"

セッション（JWT）に含まれるクリニック ID をユーザー制御可能なヘッダーで上書きできるのは
セッション管理の根本的な欠陥。

### `.claude/rules/api.md` — Security
> "Validate all user input"

`X-Clinic-ID` ヘッダーはユーザー入力であり、バリデーション（所属チェック）が必要。

## 優先度
**Critical** — マルチテナントの根幹を破壊。任意のクリニックのデータを閲覧・改ざん・削除可能。
**本番環境デプロイ前に必ず修正が必要。**

## 関連チケット
- BUG-122（修正済み）: API 権限チェック
- BUG-125: CRUD 粒度

## 関連ファイル
- `backend/internal/middleware/auth.go` — X-Clinic-ID 処理（修正対象）
- `backend/internal/handler/auth_handler.go` — JWT 生成（claims に clinic_ids 追加の場合）
- `backend/migrations/001_init.sql` — `staff_clinic_assignments` テーブル定義
- `frontend/src/lib/axios.ts` — X-Clinic-ID ヘッダー送信
