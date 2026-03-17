---
description: エラーハンドリング規約（Go Sentinel、HTTP Status、ユーザーメッセージ）
alwaysApply: true
globs: ["backend/**/*.go", "frontend/src/**/*.{ts,tsx}"]
---

# Error Handling Rules

エラーハンドリング標準規約。

## 核心ルール

### 1. Go エラーフロー（Sentinel → Wrap → Respond）

```go
// errors/errors.go - Sentinel 定義
var (
  ErrNotFound       = errors.New("not found")
  ErrConflict       = errors.New("conflict")
  ErrInvalidInput   = errors.New("invalid input")
  ErrUnauthorized   = errors.New("unauthorized")
  ErrForbidden      = errors.New("forbidden")
)

// service/owner_service.go - エラーラップ
func (s *OwnerService) Create(ctx context.Context, input CreateOwnerInput) (*Owner, error) {
  // バリデーション
  if err := input.Validate(); err != nil {
    return nil, fmt.Errorf("validation failed: %w", err)  // ErrInvalidInput に変換される
  }

  // repository 呼び出し
  owner, err := s.repo.Create(ctx, &Owner{
    ClinicID: input.ClinicID,
    Name: input.Name,
    Email: input.Email,
  })
  if err != nil {
    if errors.Is(err, ErrConflict) {
      return nil, fmt.Errorf("owner already exists with email %s: %w", input.Email, ErrConflict)
    }
    return nil, fmt.Errorf("failed to create owner: %w", err)
  }

  return owner, nil
}

// handler/owner_handler.go - HTTP レスポンス
func (h *OwnerHandler) Create(c *gin.Context) {
  var req CreateOwnerRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    RespondError(c, fmt.Errorf("invalid request: %w", ErrInvalidInput))
    return
  }

  owner, err := h.service.Create(c.Request.Context(), toCreateOwnerInput(req))
  if err != nil {
    RespondError(c, err)
    return
  }

  RespondSuccess(c, http.StatusCreated, toOwnerResponse(owner))
}
```

### 2. HTTP ステータスマッピング

```go
// handler/response.go
func RespondError(c *gin.Context, err error) {
  code := http.StatusInternalServerError
  message := "Internal server error"

  switch {
  case errors.Is(err, ErrNotFound):
    code = http.StatusNotFound
    message = "Resource not found"
  case errors.Is(err, ErrConflict):
    code = http.StatusConflict
    message = "Resource already exists"
  case errors.Is(err, ErrInvalidInput):
    code = http.StatusBadRequest
    message = "Invalid input"
  case errors.Is(err, ErrUnauthorized):
    code = http.StatusUnauthorized
    message = "Unauthorized"
  case errors.Is(err, ErrForbidden):
    code = http.StatusForbidden
    message = "Forbidden"
  }

  slog.ErrorContext(c.Request.Context(), "error", "message", message, "error", err)

  c.JSON(code, gin.H{
    "code": errorCode(err),
    "message": message,
    "timestamp": time.Now(),
  })
}

func errorCode(err error) string {
  switch {
  case errors.Is(err, ErrNotFound):
    return "NOT_FOUND"
  case errors.Is(err, ErrConflict):
    return "CONFLICT"
  case errors.Is(err, ErrInvalidInput):
    return "INVALID_INPUT"
  case errors.Is(err, ErrUnauthorized):
    return "UNAUTHORIZED"
  case errors.Is(err, ErrForbidden):
    return "FORBIDDEN"
  default:
    return "INTERNAL_ERROR"
  }
}
```

### 3. バリデーションエラー詳細化

```go
// service/validators.go
type ValidationError struct {
  Field   string `json:"field"`
  Message string `json:"message"`
}

type ValidationErrors struct {
  Errors []ValidationError `json:"errors"`
}

func (i *CreateOwnerInput) Validate() error {
  var errs []ValidationError

  if i.Name == "" {
    errs = append(errs, ValidationError{Field: "name", Message: "required"})
  }
  if i.Name != "" && len(i.Name) > 100 {
    errs = append(errs, ValidationError{Field: "name", Message: "max 100 characters"})
  }

  if i.Email == "" {
    errs = append(errs, ValidationError{Field: "email", Message: "required"})
  }
  if !isValidEmail(i.Email) {
    errs = append(errs, ValidationError{Field: "email", Message: "invalid format"})
  }

  if len(errs) > 0 {
    return fmt.Errorf("validation failed: %w", ErrInvalidInput)
    // errorsは応答に含める（RespondError内で拡張可能）
  }

  return nil
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
    // エラーロギングサービスに送信
  }

  render() {
    if (this.state.hasError) {
      return <ErrorFallback error={this.state.error} />;
    }
    return this.props.children;
  }
}

// 使用例
<ErrorBoundary>
  <OwnersList />
</ErrorBoundary>
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
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  });
}

// コンポーネント
export function OwnersList() {
  const { data, error, isLoading } = useGetOwners(clinicID);

  if (isLoading) return <div>Loading...</div>;

  if (error) {
    return (
      <div className="bg-red-100 p-4 rounded">
        <h3>エラーが発生しました</h3>
        <p>{error.message}</p>
        <button onClick={() => window.location.reload()}>再度読み込む</button>
      </div>
    );
  }

  return <div>{/* リスト表示 */}</div>;
}
```

### 6. ユーザーメッセージ（i18n）

```typescript
// lib/error-messages.ts
export const ERROR_MESSAGES: Record<string, string> = {
  NOT_FOUND: 'リソースが見つかりません',
  CONFLICT: 'このメールアドレスは既に登録されています',
  INVALID_INPUT: '入力内容に誤りがあります',
  UNAUTHORIZED: 'ログインしてください',
  FORBIDDEN: 'アクセス権限がありません',
  INTERNAL_ERROR: 'サーバーエラーが発生しました',
  NETWORK_ERROR: 'ネットワークエラーが発生しました',
};

// components/Toast.tsx
export function showErrorToast(error: unknown) {
  const message = error instanceof AxiosError
    ? ERROR_MESSAGES[error.response?.data?.code] || ERROR_MESSAGES.INTERNAL_ERROR
    : ERROR_MESSAGES.INTERNAL_ERROR;

  toast.error(message);
}
```

## チェックリスト

- [ ] Sentinel エラー定義（errors/errors.go）
- [ ] エラーラップ：fmt.Errorf("context: %w", err)
- [ ] HTTP Status マッピング（RespondError）
- [ ] バリデーション詳細化（Field + Message）
- [ ] ログ：slog.ErrorContext で構造化ログ
- [ ] React Error Boundary 実装
- [ ] React Query retry + retryDelay設定
- [ ] ユーザーメッセージ：ERROR_MESSAGES i18n
- [ ] Toast/Alert で エラー通知
- [ ] Console.error なし（本番環境）

## エラーレスポンス例

```json
{
  "code": "INVALID_INPUT",
  "message": "Validation failed",
  "errors": [
    { "field": "email", "message": "invalid format" },
    { "field": "phone", "message": "must be E.164 format" }
  ],
  "timestamp": "2026-03-17T10:30:00Z"
}
```
