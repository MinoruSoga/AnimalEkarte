# 回帰テストシナリオ - Animal Ekarte

本ドキュメントは、Vitest + MSW を使用した自動化回帰テストシナリオを定義します。

---

## 1. Critical Paths（クリティカルパス）

### 1.1 受付 → 医療記録 → 会計 → 退院フロー

**目的:** コア業務フロー全体を検証（エンドツーエンド）

**テストシナリオ:**

```gherkin
Scenario: 新規ペット受付から会計確定まで
  Given: 新規ペット（シロ）が受付画面に表示される
  When:  「新規受付」ボタンをクリック
  Then:  受付済に移動 → 医療記録作成フォーム表示

  When:  医療記録で診断「食欲がない」を入力・保存
  Then:  医療記録一覧に「作成中」ステータスで表示

  When:  会計管理で当該記録を検索・選択
  Then:  請求金額計算・表示、支払方法選択可能

  When:  「確認」ボタンをクリック
  Then:  会計済ステータスに遷移、医療記録も「確定済」に

Expected Result: 医療記録 → 会計 の状態遷移一貫性 ✅
```

**実装場所:** `frontend/src/features/__tests__/critical-paths.test.ts`

---

### 1.2 予約 → 受付フロー

**目的:** 予約から当日受付への流れを検証

**テストシナリオ:**

```gherkin
Scenario: 予約済み患者の当日受付
  Given: 予約管理で「2026-04-02 10:00」にシロの予約が存在
  When:  当日の受付画面を開く
  Then:  予約済みリストに「シロ」が表示される（医師フィルタ機能動作確認）

  When:  受付予約から「シロ」をドラッグ→受付済へ移動
  Then:  リアルタイムで受付済セクションに移動、Kanban UI更新確認

Expected Result: 予約→受付の状態遷移 + UI同期 ✅
```

**実装場所:** `frontend/src/features/reception/routes/__tests__/Reception.test.ts`

---

### 1.3 医療記録 → 検査 → 結果確定フロー

**目的:** 複雑な医療ワークフロー（検査含む）を検証

**テストシナリオ:**

```gherkin
Scenario: 医療記録から検査予約→結果入力→確定
  Given: 医療記録ID=17（ルナの医療記録）が存在
  When:  「新規検査」ボタンをクリック
  Then:  検査フォーム表示（ペット・検査種別自動プリフィル）

  When:  検査種別「血液検査」を選択・検査技師を「田中太郎」に設定
  Then:  新規検査エントリが検査管理に作成

  When:  検査結果入力（status="result_entered"）
  Then:  結果確定ボタンが enabled に変更

  When:  結果確定クリック (status="result_confirmed")
  Then:  医療記録に検査結果リンク表示、タイムライン更新

Expected Result: 医療記録 ← → 検査 の双方向同期 ✅
```

**実装場所:** `frontend/src/features/examinations/routes/__tests__/ExaminationsList.test.ts`

---

## 2. Edge Cases（エッジケース）

### 2.1 エラーハンドリング

#### 2.1.1 ネットワークエラー

```typescript
Scenario: API 500エラー時のリトライと通知
  Given: MSW で POST /api/medical-records → 500 Internal Server Error
  When:  医療記録作成フォーム送信
  Then:
    - React Query retry 1回実行（設定に応じて）
    - エラートースト表示: 「医療記録の作成に失敗しました」
    - Submit ボタンは再度クリック可能（disabledではない）

Assertion:
  - Toast message presence ✅
  - Form remains fillable ✅
  - Retry logic works ✅
```

**実装位置:** `frontend/src/features/medical-records/__tests__/error-handling.test.ts`

---

#### 2.1.2 バリデーションエラー

```typescript
Scenario: フォーム入力バリデーション
  Given: 医療記録作成フォーム
  When:
    - 「診療日」未入力で送信
    - 「ペット」未選択で送信
  Then:
    - 各フィールドの下に赤いエラーメッセージ表示
    - Submit ボタン disabled（送信不可）

Assertion:
  - aria-describedby で error message linked ✅
  - Focus management (focus → error field) ✅
```

**実装位置:** `frontend/src/features/medical-records/__tests__/validation.test.ts`

---

### 2.2 同時並行操作

#### 2.2.1 複数タブでの同時編集

```typescript
Scenario: 医療記録を複数タブで同時編集
  Given: 医療記録ID=17 を2つのタブで開く（Tab A, Tab B）
  When:
    - Tab A: 「診断」フィールドを編集・保存
    - Tab B: 同時に「主訴」フィールドを編集・保存
  Then:
    - 最後の保存が反映される（Last-Write-Wins）
    - Tab A をリロード → Tab B の最新内容が表示
    - OR: 競合警告トースト表示（バージョン管理実装時）

Expected Behavior: Last-Write-Wins or Conflict Detection ✅
```

**実装位置:** `frontend/src/features/medical-records/__tests__/concurrent-editing.test.ts`

---

#### 2.2.2 フォーム送信中の複数クリック

```typescript
Scenario: Submit ボタンの二重送信防止
  Given: 医療記録作成フォーム（SubmitButton使用）
  When:
    - SubmitButton をダブルクリック
    - または Submit 中に別フォーム要素をクリック
  Then:
    - API は1回のみ呼び出し（isPending = true で button disabled）
    - ローディング UI が表示（spinner）
    - 完了後、button再度 enabled

Assertion:
  - API call count == 1 ✅
  - Button disabled during submission ✅
  - Spinner visible ✅
```

**実装位置:** `frontend/src/components/shared/__tests__/SubmitButton.test.ts`

---

### 2.3 データ整合性

#### 2.3.1 N+1クエリ防止

```typescript
Scenario: 医療記録一覧読み込み時のクエリ最適化
  Given: 医療記録20件を表示
  When:  医療記録一覧ページを開く
  Then:
    - GET /api/medical-records (1回)
    - Preload="pets" で飼主・ペット情報含む（追加クエリなし）
    - Network タブで XHR count == 1

Tool: DevTools Network Panel を自動検証
```

**実装位置:** DevTools integration test（手動確認 or Cypress）

---

#### 2.3.2 キャッシュ一貫性

```typescript
Scenario: React Query キャッシュの自動更新
  Given: 医療記録一覧キャッシュが存在（stale=false）
  When:
    - 新規医療記録作成 (POST成功)
    - 一覧ページに戻る
  Then:
    - キャッシュが自動更新される（新規レコード表示）
    - または invalidateQueries で手動更新

Tool: React Query DevTools で cache state 確認
```

---

### 2.4 アクセシビリティ（A11y）

#### 2.4.1 キーボード操作

```typescript
Scenario: フォーム全体をキーボード操作で完結
  Given: 医療記録作成フォーム
  When:
    - Tab キーで全フィールドをナビゲート
    - Enter/Space でドロップダウン・ボタン操作
    - Escape でモーダル・ポップアップ閉じる
  Then:
    - 全操作がマウスなしで実行可能
    - フォーカス指示（outline）が常に表示される

Tool: axe-core や手動キーボードテスト
```

---

#### 2.4.2 スクリーンリーダー対応

```typescript
Scenario: NVDA / JAWS で画面読み上げ
  Given: 医療記録フォーム
  When:  NVDA で読み上げ
  Then:
    - フォームラベルが正しく読み上げられる
    - エラーメッセージが aria-live で通知される
    - ボタン機能が明確（「保存ボタン」と読まれる）

Tool: NVDA/JAWS 実機テスト（月1回）
```

---

## 3. テスト実装フレームワーク

### 3.1 Vitest 設定

```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/testing/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      include: ['src/features/**'],
      exclude: ['src/**/*.test.ts', 'src/**/*.stories.ts'],
      lines: 80,
      functions: 80,
      branches: 75,
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
});
```

---

### 3.2 MSW (Mock Service Worker) 設定

```typescript
// src/testing/handlers.ts
import { http, HttpResponse } from 'msw';

export const handlers = [
  // Medical Records
  http.get('/api/medical-records', () => {
    return HttpResponse.json([
      { id: 1, petId: 1, visitDate: '2026-02-28', chiefComplaint: '食欲がない', status: 'draft' },
      { id: 2, petId: 2, visitDate: '2026-02-10', chiefComplaint: '嘔吐・下痢', status: 'draft' },
    ]);
  }),

  http.post('/api/medical-records', ({ request }) => {
    // バリデーション
    if (!request.body) {
      return HttpResponse.json({ error: 'Invalid request' }, { status: 400 });
    }
    return HttpResponse.json({ id: 21, status: 'draft' }, { status: 201 });
  }),

  http.patch('/api/medical-records/:id', ({ params }) => {
    return HttpResponse.json({ id: params.id, status: 'confirmed' });
  }),

  // Examinations
  http.get('/api/examinations', () => {
    return HttpResponse.json([]);
  }),

  // Accounting
  http.post('/api/accounting', ({ request }) => {
    return HttpResponse.json({ id: 100, status: 'paid' }, { status: 201 });
  }),
];

// src/testing/server.ts
import { setupServer } from 'msw/node';
import { handlers } from './handlers';

export const server = setupServer(...handlers);
```

---

### 3.3 テストベース構造

```typescript
// Example: medical-records.test.ts
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { server } from '@/testing/server';
import MedicalRecordForm from '@/features/medical-records/components/MedicalRecordForm';

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('MedicalRecordForm', () => {
  it('should create a medical record with valid input', async () => {
    const queryClient = new QueryClient();
    const user = userEvent.setup();

    render(
      <QueryClientProvider client={queryClient}>
        <MedicalRecordForm petId={1} />
      </QueryClientProvider>
    );

    // フォーム入力
    const visitDateInput = screen.getByLabelText(/診療日/i);
    await user.type(visitDateInput, '2026-04-02');

    const chiefComplaintInput = screen.getByLabelText(/主訴/i);
    await user.type(chiefComplaintInput, '食欲がない');

    // 送信
    const submitButton = screen.getByRole('button', { name: /保存/i });
    await user.click(submitButton);

    // 検証
    await waitFor(() => {
      expect(screen.getByText(/医療記録を作成しました/i)).toBeInTheDocument();
    });
  });

  it('should show validation error for empty chief complaint', async () => {
    const queryClient = new QueryClient();
    const user = userEvent.setup();

    render(
      <QueryClientProvider client={queryClient}>
        <MedicalRecordForm petId={1} />
      </QueryClientProvider>
    );

    const submitButton = screen.getByRole('button', { name: /保存/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/主訴は必須です/i)).toBeInTheDocument();
    });
  });
});
```

---

## 4. テスト実行スケジュール

| タイミング | テストスイート | 対象 | 時間 |
|-----------|--------------|------|------|
| **Pre-commit** | Lint + Unit | 変更ファイルのみ | < 10秒 |
| **Pre-push** | Unit + Integration | 全テスト | < 60秒 |
| **CI/CD (PR)** | All (+ Lighthouse + E2E) | 全セクション | < 5分 |
| **Nightly** | Smoke + Perf | 本番環境テスト | 毎晩 |
| **Weekly** | Full + Manual A11y | 全機能 + NVDA検証 | 金曜 |

---

## 5. コマンド

```bash
# ローカル開発
pnpm test:watch              # Watch mode
pnpm test:run                # Single run
pnpm test:coverage           # Coverage report

# CI/CD
pnpm test:ci                 # Strict mode (--no-coverage は除外)

# Lighthouse
pnpm audit:lighthouse        # Desktop audit
```

---

## 6. 既知のテスト実装状況

**実装済み:**
- ✅ `features/owners/` — useOwnerForm, getOwners hook テスト
- ✅ `features/medical-records/` — フォーム検証、API integration
- ✅ `components/shared/SubmitButton.tsx` — 二重送信防止テスト

**未実装（優先度順）:**
- ❌ `features/accounting/` — 会計確定フロー
- ❌ `features/hospitalization/` — cage board ドラッグドロップ
- ❌ `features/inventory/` — 在庫管理フロー
- ❌ Edge cases: N+1 クエリ検出、キャッシュ一貫性

---

## 7. 継続的改善

- **月次レビュー:** テストカバレッジ報告（目標 80%+）
- **四半期更新:** 新機能追加に合わせてシナリオ拡張
- **セキュリティテスト:** OWASP Top 10 対応テスト追加（Q2 2026）
