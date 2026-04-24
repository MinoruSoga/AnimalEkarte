---
description: 指定パスのテスト自動生成（Vitest / go test）
argument-hint: "<path> (e.g. frontend/src/features/owners)"
---

# テスト生成

$ARGUMENTS のテストを生成してください。

## テスト方針

- 正常系
- 境界値
- 異常系
- エッジケース

## フレームワーク

- Frontend: **Vitest** + Testing Library + MSW
- Backend: **go test** + testify

## パターン

```typescript
describe('関数名', () => {
  describe('正常系', () => {
    it('should [期待される動作]', () => {
      // Arrange
      // Act
      // Assert
    });
  });

  describe('異常系', () => {
    it('should throw error when [条件]', () => {
      // ...
    });
  });
});
```

## 出力

- テストファイルの内容
- 必要なモックのセットアップ
