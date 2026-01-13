# テスト生成

$ARGUMENTS のテストを生成してください。

## テスト方針

- 正常系
- 境界値
- 異常系
- エッジケース

## フレームワーク

Jest / Vitest

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
