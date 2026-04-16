# UX-004: スタッフ編集フォームにログイン情報（メールアドレス）が表示されない

## 概要

スタッフマスタの編集画面でログイン情報（メールアドレス）が表示されていない。新規作成時のみ表示され、編集時は非表示になっている。

## 症状

- スタッフ編集フォームを開いても、メールアドレスが表示されない
- 新規作成時のみメールアドレス・パスワードが表示される
- ユーザーがスタッフの連絡先（メールアドレス）を確認できない

## 根本原因

### Backend
- `staff_response.go`: Account 情報（メールアドレス）が API レスポンスに含まれていない
- `staff_handler.go`: ListStaffs(), GetStaff() で Account を Preload していない

### Frontend
- `StaffSettings.tsx` (line 192-213): 新規作成時（isNew === true）のみメールアドレス・パスワード入力欄を表示
- 編集時（isNew === false）はフォーム自体に Account 情報のフィールドが存在しない

## 修正対象

| ファイル | 行番号 | 修正内容 |
|---------|--------|---------|
| `backend/internal/handler/staff_response.go` | 19-30 | Account 情報（Email）を staffResponse に追加 |
| `backend/internal/handler/staff_handler.go` | 30 | Account Preload を追加: `.Preload("Account")` |
| `frontend/src/features/master/routes/StaffSettings.tsx` | 94-101 | Account メールアドレス初期化を追加 |
| `frontend/src/features/master/routes/StaffSettings.tsx` | 192-213 | 編集時も Account メールアドレスを表示（読み取り専用） |
| `frontend/src/features/master/api/staffs.ts` | - | Staff 型に Account メールアドレスフィールドを追加 |

## 実装仕様

### Backend修正案

**staff_response.go:**
```go
type staffResponse struct {
    ID            uint64                   `json:"id"`
    Name          string                   `json:"name"`
    IsActive      bool                     `json:"is_active"`
    StaffRole     string                   `json:"staff_role"`
    JobTitleID    *uint64                  `json:"job_title_id,omitempty"`
    LicenseNumber string                   `json:"license_number"`
    SortOrder     int                      `json:"sort_order"`
    Email         string                   `json:"email"`  // ← 追加
    JobTitle      *jobTitleInStaffResponse `json:"job_title,omitempty"`
    CreatedAt     time.Time                `json:"created_at"`
    UpdatedAt     time.Time                `json:"updated_at"`
}

func toStaffResponse(s *model.Staff) staffResponse {
    return staffResponse{
        ID:            s.ID,
        Name:          s.Name,
        IsActive:      s.IsActive,
        StaffRole:     string(s.StaffRole),
        JobTitleID:    s.JobTitleID,
        LicenseNumber: s.LicenseNumber,
        SortOrder:     s.SortOrder,
        Email:         s.Account.Email,  // ← Account を Preload していることを前提
        JobTitle:      toJobTitleInStaffResponse(s.JobTitle),
        CreatedAt:     s.CreatedAt,
        UpdatedAt:     s.UpdatedAt,
    }
}
```

**staff_handler.go:**
```go
func (h *Handler) ListStaffs(c *gin.Context) {
    // ...
    staffs, _, err := h.svc.Staff.List(c.Request.Context(), clinicID, role, 1, 1000)
    // ... → service層で Preload("Account") を実装
}
```

### Frontend修正案

**StaffSettings.tsx:**
- 新規作成時: メールアドレス・パスワード入力欄を表示（既存）
- 編集時: メールアドレス読み取り専用で表示、パスワード変更は別機能に（ロードマップ）

```typescript
interface StaffFormData {
  name: string;
  staffRole: StaffRoleValue;
  licenseNumber: string;
  isActive: boolean;
  email: string;           // ← 常に存在（新規は入力、編集は読み取り専用）
  password: string;         // ← 新規作成時のみ
}

const StaffSidePanel = ({ item, ... }: StaffSidePanelProps) => {
  const [f, setF] = useState<StaffFormData>(() => ({
    name: item?.name ?? "",
    staffRole: item?.staffRole ?? "veterinarian",
    licenseNumber: item?.licenseNumber ?? "",
    isActive: item?.isActive ?? true,
    email: item?.email ?? "",  // ← 編集時も初期化
    password: "",
  }));

  return (
    <MasterSidePanel ...>
      {/* ... */}

      <PropertyRow label="メールアドレス">
        <input
          type="email"
          className={MASTER_INPUT_CLASS}
          value={f.email}
          onChange={isNew ? handleEmailChange : undefined}
          disabled={!isNew}  // ← 編集時は読み取り専用
          placeholder="例: staff@clinic.com"
        />
      </PropertyRow>

      {isNew && (
        <PropertyRow label="パスワード">
          <input
            type="password"
            className={MASTER_INPUT_CLASS}
            value={f.password}
            onChange={handlePasswordChange}
            placeholder="8文字以上"
          />
        </PropertyRow>
      )}
    </MasterSidePanel>
  );
};
```

## テスト仕様

### 新規作成時
- [ ] メールアドレス入力欄が表示される
- [ ] パスワード入力欄が表示される
- [ ] 両フィールド入力後、スタッフ作成に成功する

### 編集時
- [ ] メールアドレス（既登録の値）が読み取り専用で表示される
- [ ] パスワード入力欄が表示されない（編集不可）
- [ ] その他フィールド（氏名、職種等）は編集可能

## 優先度

**HIGH** — ユーザーが必要な情報を確認できないため、即座に修正が必要。

## 関連イシュー

- UX-001: マスタ設定スタッフキャッシュ問題（解決済み）
- UX-002: スタッフマスタ職種カラム表示（解決済み）
- UX-003: マスタ設定説明フィールド欠落（解決済み）
