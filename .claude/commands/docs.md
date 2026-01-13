# ドキュメント生成

$ARGUMENTS のドキュメントを生成してください。

## 生成内容

1. 関数・クラスの概要
2. パラメータの説明
3. 戻り値の説明
4. 使用例
5. 注意事項

## 形式

JSDoc/TSDoc形式

## 例

```typescript
/**
 * ユーザー情報を取得する
 * @param userId - ユーザーID
 * @returns ユーザー情報オブジェクト
 * @throws {NotFoundError} ユーザーが存在しない場合
 * @example
 * const user = await getUser('123');
 */
```
