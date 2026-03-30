---
description: エラーハンドリング規約（Go Sentinel、HTTP Status、ユーザーメッセージ）
alwaysApply: true
globs: ["backend/**/*.go", "frontend/src/**/*.{ts,tsx}"]
---

# Error Handling Rules

エラーハンドリング標準規約。

## 核心ルール

### 1. Go エラーフロー（Repository → Service → Handler）

```go
// errors/errors.go - Sentinel 定義
var (
  ErrNotFound       = errors.New("not found")
  ErrConflict       = errors.New("conflict")
  ErrInvalidInput   = errors.New("invalid input")
)

// repository/owner_repository.go - GORM エラー変換
func (r *OwnerRepository) GetByID(ctx context.Context, id uint) (*model.Owner, error) {
  var owner model.Owner
  if err := r.db.WithContext(ctx).First(&owner, id).Error; err != nil {
    // ✅ MANDATE: Repository では FromGORM を使用
    return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", id))
  }
  return &owner, nil
}

// service/owner_service.go - エラーラップ
func (s *OwnerService) GetOwner(ctx context.Context, id uint) (*model.Owner, error) {
  owner, err := s.repo.GetByID(ctx, id)
  if err != nil {
    // ✅ MANDATE: Service では Wrap を使用
    return nil, apperrors.Wrap(err, "failed to get owner")
  }
  return owner, nil
}
```

### 2. Frontend エラーハンドリング（handleApiError）

すべての `catch` ブロックで `handleApiError` を呼び出す。

```typescript
// ✅ MANDATE: すべての catch ブロックで handleApiError を使用
try {
  await api.updateOwner(id, data);
} catch (error) {
  handleApiError(error, "オーナーの更新");
}
```

### 3. HTTP ステータスマッピング

```go
// handler/response.go
func RespondError(c *gin.Context, err error) {
  code := http.StatusInternalServerError
  message := "Internal server error"

  switch {
  case errors.Is(err, apperrors.ErrNotFound):
    code = http.StatusNotFound
    message = "Resource not found"
  case errors.Is(err, apperrors.ErrConflict):
    code = http.StatusConflict
    message = "Resource already exists"
  case errors.Is(err, apperrors.ErrInvalidInput):
    code = http.StatusBadRequest
    message = "Invalid input"
  }

  slog.ErrorContext(c.Request.Context(), "error", "message", message, "error", err)

  c.JSON(code, gin.H{
    "code": errorCode(err),
    "message": message,
    "timestamp": time.Now(),
  })
}
```

### 4. React エラーバウンダリー

```typescript
// components/errors/ErrorBoundary.tsx
export class ErrorBoundary extends React.Component<Props, State> {
  state: State = { hasError: false, error: null };

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('Error:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return <ErrorFallback error={this.state.error} />;
    }
    return this.props.children;
  }
}
```

### 5. React Query エラーハンドリング

```typescript
// hooks/use-owners.ts
export function useGetOwners(clinicID: number) {
  return useQuery({
    queryKey: ['owners', clinicID],
    queryFn: async () => {
      const response = await axios.get(`/api/owners?clinic_id=${clinicID}`);
      return response.data;
    },
    retry: 1,
  });
}
```

## チェックリスト

- [ ] Repository: `apperrors.FromGORM(err, "resource", id)` 使用
- [ ] Service: `apperrors.Wrap(err, "context")` 使用
- [ ] Frontend: すべての `catch` ブロックで `handleApiError(error, "context")` 使用
- [ ] HTTP Status マッピング（RespondError）
- [ ] ログ：slog.ErrorContext で構造化ログ
- [ ] React Error Boundary 実装
- [ ] React Query retry 設定
- [ ] Console.error なし（本番環境）
