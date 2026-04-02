---
description: パフォーマンス最適化規約（Go pprof、React DevTools、Lighthouse）
alwaysApply: false
globs: ["backend/**/*.go", "frontend/src/**/*.{ts,tsx}"]
---

# Performance Rules

パフォーマンス目標と最適化戦略。

## 核心ルール

### 1. Go パフォーマンス目標

```
API Response Time:   < 50ms (p95)
Memory Allocation:   < 100MB (baseline)
Goroutine Count:     < 100 (idle)
Database Query Time: < 50ms (p95)
```

### 2. CPU プロファイリング（pprof）

```bash
# http://localhost:8080/debug/pprof でエンドポイント公開

# CPU プロファイル（30秒間記録）
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# インタラクティブモード
(pprof) top10       # トップ 10 関数
(pprof) list Main   # 関数詳細
(pprof) graph       # 呼び出しグラフ
```

### 3. メモリプロファイリング

```bash
# メモリ割り当て
go tool pprof http://localhost:8080/debug/pprof/allocs

# ヒーププロファイル
go tool pprof http://localhost:8080/debug/pprof/heap

# Goroutine リーク検出
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

### 4. Go 最適化パターン

```go
// ✅ バッファ指定
buf := make([]Owner, 0, 100)  // capacity 確保で再割り当て回避

// ✅ 文字列連結は bytes.Buffer
var buf bytes.Buffer
for _, s := range strings {
  buf.WriteString(s)
}
result := buf.String()

// ❌ 非効率
var result string
for _, s := range strings {
  result += s  // 毎回新しい文字列生成
}

// ✅ GORM クエリ最適化
db.Select("id", "name", "email").  // 必要カラムのみ
  Preload("Pets", func(db *gorm.DB) *gorm.DB {
    return db.Select("id", "owner_id", "name")
  }).
  Where("clinic_id = ?", clinicID).
  Limit(100).
  Find(&owners)
```

### 5. React パフォーマンス目標

```
FCP (First Contentful Paint):     < 1.8s
LCP (Largest Contentful Paint):  < 2.5s
CLS (Cumulative Layout Shift):    < 0.1
TTI (Time to Interactive):        < 3.8s
Bundle Size:                      < 200KB (JS)
```

### 6. React DevTools Profiler

```typescript
// React DevTools → Profiler タブ で記録

// memo() でコンポーネント再レンダー防止
export const OwnerCard = memo(function OwnerCard({ owner }: Props) {
  return <div>{owner.name}</div>;
});

// useCallback でハンドラ安定化
const handleChange = useCallback((value) => {
  setData(value);
}, [setData]);

// useDeferredValue で遅延レンダー
const deferredTerm = useDeferredValue(searchTerm);

// useMemo でコンポーネント再生成防止
const memoizedList = useMemo(() => (
  owners.map(o => <OwnerCard key={o.id} owner={o} />)
), [owners]);
```

### 7. Bundle 分析

```bash
# Vite bundle サイズ分析
npm run build
npm install -g rollup-plugin-visualizer
# build output を확認

# Critical JS < 200KB
# CSS < 50KB
# 画像最適化（WebP）
```

### 8. Lighthouse 監査

```bash
# DevTools Lighthouse でスコア確認
# 目標: Performance > 90

# 自動監査
npm run audit:lighthouse

# チェック項目:
- Unused JavaScript
- Unused CSS
- Image optimization
- Font optimization
- Minification
- Code splitting
```

### 9. データベース最適化

```go
// N+1 クエリ検出・排除
// ✅ Preload でクエリ削減
db.Preload("Pets").Where("clinic_id = ?", clinicID).Find(&owners)

// ✅ EXPLAIN ANALYZE で実行計画確認
EXPLAIN ANALYZE SELECT * FROM owners WHERE clinic_id = 1 AND id = 100;
// → Index Scan (< 1ms)

// ❌ Seq Scan 回避
EXPLAIN ANALYZE SELECT * FROM owners WHERE name LIKE '%太%';
// → Seq Scan (1000ms) → インデックス追加検討
```

## チェックリスト

- [ ] API Response Time < 50ms (p95)
- [ ] 定期 pprof 分析（月 1 回）
- [ ] React memo() で不要再レンダー排除
- [ ] useCallback でハンドラ安定化
- [ ] useDeferredValue で重い計算遅延
- [ ] Bundle size < 200KB (JS)
- [ ] Lighthouse Score > 90
- [ ] N+1 クエリ排除（Preload）
- [ ] EXPLAIN ANALYZE で Seq Scan なし
- [ ] Memory allocation < 100MB (baseline)

## パフォーマンス監視コマンド

```bash
# Go
make test-backend  # go test ./... -v -cover

# React
make test-frontend # npm run test:run

# 本番監視（ダッシュボード）
make logs          # Docker Compose ログ監視
```
