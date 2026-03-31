# BUG-059: クリニック切り替え後 JWT の clinic_id が古いままでデータが誤クリニックに保存される

**作成日**: 2026-03-29
**ステータス**: Closed（Superseded by TASK-049）
> 権限を company 単位にフラット化することで、`hasPermission()` が `currentClinicId` に依存しなくなる。権限判定の乖離問題は構造的に解消される。本バグは TASK-049 完了後に無効となる。
**優先度**: Critical
**領域**: Backend + Frontend
**関連**: BUG-055（Dual-Token 移行）, BE-078

---

## 背景・問題

JWT の `clinic_id` はログイン時に1つのクリニック ID で固定される。
フロントエンドの `switchClinic()` は `localStorage` の `currentClinicId` を更新するが、
**JWT は更新されない**。

結果として、クリニック切り替え後の全 API リクエストは
**JWT 由来の古い `clinic_id`** でバックエンドに到達する。

---

## 再現シナリオ

```
1. ユーザー（複数クリニック所属）が clinic-1 でログイン
   JWT: { user_id: 10, clinic_id: "1", user_type: "staff" }

2. UI で clinic-2 に切り替え（switchClinic("2")）
   localStorage: currentClinicId = "2"
   JWT: 変更なし（clinic_id = "1" のまま）

3. フロント UI では clinic-2 の画面を表示しながら、
   POST /api/v1/medical-records を送信

4. バックエンド extractClinicID(c) → JWT から clinic_id = "1" を読む

5. 結果: カルテが clinic-2 ではなく clinic-1 に作成される
```

**これはデータテナント混在を引き起こす CRITICAL バグ。**

---

## 影響範囲

- 複数クリニックに所属する `staff` / `clinic_admin` 全員
- 書き込み系 API（POST/PUT/PATCH/DELETE）すべて — 誤ったクリニックにデータが書き込まれる
- 読み取り系 API — 表示クリニックと異なるデータが返る可能性

---

## 解決方針

### 短期対応（即座）: `X-Clinic-ID` ヘッダーでクリニント切り替えを明示的に送信

JWT に `clinic_id` を含めたままにしつつ、フロントエンドが現在の `currentClinicId` を
リクエストヘッダーで明示的に送信する。バックエンドはヘッダーを優先する。

#### フロントエンド変更（`lib/axios.ts`）

```typescript
// lib/axios.ts
import { getCurrentClinicId } from "@/features/auth/hooks/use-auth";

api.interceptors.request.use((config) => {
  const clinicId = getCurrentClinicId();
  if (clinicId) {
    config.headers["X-Clinic-ID"] = clinicId;
  }
  return config;
});
```

`getCurrentClinicId()` は `useAuth()` hook から `currentClinicId` を Zustand または
localStorage 経由で読み出す純粋関数（React hook でないもの）。

#### バックエンド変更（`middleware/auth.go`）

```go
// ミドルウェアで X-Clinic-ID ヘッダーを優先し、
// JWT の clinic_id と整合性を検証する

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 既存の JWT 検証 ...
        claims := extractJWTClaims(c, jwtSecret)
        if claims == nil {
            return
        }

        // JWT の clinic_id をベースにセット
        effectiveClinicID := claims.ClinicID

        // X-Clinic-ID ヘッダーが存在する場合は上書き
        if headerClinicID := c.GetHeader("X-Clinic-ID"); headerClinicID != "" {
            // ユーザーが指定クリニックに所属しているか検証
            if !isUserMemberOfClinic(c.Request.Context(), claims.UserID, headerClinicID) {
                c.JSON(http.StatusForbidden, gin.H{"error": "not a member of the specified clinic"})
                c.Abort()
                return
            }
            effectiveClinicID = headerClinicID
        }

        c.Set("clinic_id", effectiveClinicID)
        c.Set("user_id", claims.UserID)
        c.Set("user_type", claims.UserType)
        c.Next()
    }
}
```

`isUserMemberOfClinic()` は `user_clinic_memberships`（または相当するテーブル）で
メンバーシップを確認する。

### 長期対応（BE-078 完了後）: Dual-Token でアクセストークンを再発行

BE-078 が完了しアクセストークンの有効期限が 15分になると、
`switchClinic()` 時に `POST /v1/auth/refresh?clinic_id=2` で新しいトークンを再発行することができる。

```
switchClinic(newClinicId)
  → POST /v1/auth/refresh { preferred_clinic_id: newClinicId }
  → 新しい JWT: { clinic_id: newClinicId } を発行
  → Cookie を更新
```

---

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| BE-080 | Backend | ~~`X-Clinic-ID` ヘッダーのミドルウェア検証~~ → TASK-049 で permission を company 単位にフラット化することで不要になった |
| ~~FE-137~~ | Frontend | ~~`axios.ts` に `X-Clinic-ID` ヘッダーインターセプター追加~~ → TASK-049 で不要（FE-137 は別目的で作成済み） |

> ⚠️ 本タスクは TASK-049（権限スコープ company 単位移行）により Superseded。上記の派生イシューは未作成のまま終了。

---

## 完了条件

- [ ] クリニック切り替え後の API リクエストが正しいクリニック ID で処理される
- [ ] 所属していないクリニック ID を `X-Clinic-ID` に指定したリクエストが `403` を返す
- [ ] 複数クリニック所属ユーザーのデータが正しいクリニックに保存される
- [ ] `docker compose exec backend go build ./...` 成功
- [ ] `docker compose exec frontend npm run build` 成功
