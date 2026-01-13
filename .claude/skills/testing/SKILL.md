---
name: Testing
description: ユニット、統合、E2Eテストの作成と実行
---

# テストスキル

## このスキルを使用するタイミング

- 新機能のテスト作成
- 既存テストの修正
- テストカバレッジの向上

## テスト種別

| 種別 | ツール | コマンド |
|------|--------|---------|
| ユニット | Jest | `npm test` |
| 統合 | Jest | `npm run test:integration` |
| E2E | Playwright | `npm run test:e2e` |

## テスト作成パターン

### ユニットテスト

```typescript
describe('関数名', () => {
  it('should [期待される動作]', () => {
    // Arrange
    const input = ...;

    // Act
    const result = functionName(input);

    // Assert
    expect(result).toBe(...);
  });
});
```

### 統合テスト

```typescript
describe('API: /api/endpoint', () => {
  it('should return 200 on valid request', async () => {
    const response = await request(app)
      .get('/api/endpoint')
      .expect(200);

    expect(response.body).toMatchObject({...});
  });
});
```

## 重要な注意事項

- モックは最小限に
- テストデータはファクトリパターンを使用
- 非同期テストは適切にawait

## テンプレート

- ユニット: [templates/unit.test.ts](./templates/unit.test.ts)
- 統合: [templates/integration.test.ts](./templates/integration.test.ts)
