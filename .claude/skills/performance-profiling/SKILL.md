---
name: performance-profiling
description: パフォーマンスプロファイリング・最適化（Go pprof、React DevTools、Lighthouse）
---

# Performance Profiling & Optimization

Backend (Go) と Frontend (React) のパフォーマンスを詳細に分析・最適化します。

## 実行スコープ

### 1. Go Backend Profiling (pprof)

**前提: backend には現在 net/http/pprof が未配線**。以下のコマンドは `import _ "net/http/pprof"` + :6060 リッスンを追加してから使う。未配線のまま実行しても接続失敗する。アプリ本体は backend :8080 / frontend :3003。pprof 未配線を「計測 PASS」と書かない。

#### CPU プロファイル
```bash
docker compose exec backend go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

**分析項目:**
- ホットパス特定
- 関数別CPU使用率
- Goroutine 生成位置

**チューニング例:**
```go
// 低速: 文字列連結（毎イテレーション）
var result string
for _, v := range items {
  result += v.String() + ","
}

// 高速: strings.Builder
var buf strings.Builder
for _, v := range items {
  buf.WriteString(v.String())
  buf.WriteByte(',')
}
return buf.String()
```

#### メモリプロファイル
```bash
docker compose exec backend go tool pprof http://localhost:6060/debug/pprof/heap
```

**分析項目:**
- メモリ割り当て箇所
- Goroutine リーク検出
- オブジェクト生存期間

**チューニング例:**
```go
// 低速: 毎回 slice 再割り当て
func process(items []Item) {
  result := make([]string, 0)
  for _, item := range items {
    result = append(result, item.Name)
  }
}

// 高速: 容量予約
func process(items []Item) {
  result := make([]string, 0, len(items))
  for _, item := range items {
    result = append(result, item.Name)
  }
}
```

#### Goroutine リーク検出
```bash
docker compose exec backend go tool pprof http://localhost:6060/debug/pprof/goroutine
```

**確認項目:**
- Goroutine 総数の増加傾向
- デッドロック検査
- context キャンセル伝播

#### Database クエリ分析
```bash
docker compose exec db psql -U "$DB_USER" -d "$DB_NAME" -c "
  SELECT query, calls, mean_time
  FROM pg_stat_statements
  ORDER BY mean_time DESC LIMIT 10;
"
```
⚠️ CLAUDE.md の自動実行禁止コマンド（psql 直叩き）。ユーザーに手動実行を依頼する

**N+1 クエリ検出:**
```go
// ❌ N+1: Owner ごとに Pet を取得
owners := []Owner{}
db.Find(&owners)
for _, owner := range owners {
  db.Where("owner_id = ?", owner.ID).Find(&owner.Pets) // N回実行
}

// ✅ 最適化: Preload
db.Preload("Pets").Find(&owners)

// または Joins
db.Joins("LEFT JOIN pets ON pets.owner_id = owners.id").
  Find(&owners)
```

### 2. React Frontend Profiling

#### React DevTools Profiler
```typescript
// コンポーネントレンダリング時間測定
import { Profiler } from 'react'

export function MyComponent() {
  const onRender = (
    id: string,
    phase: 'mount' | 'update',
    actualDuration: number,
    baseDuration: number,
    startTime: number,
    commitTime: number
  ) => {
    console.log(`${id} (${phase}): ${actualDuration}ms`)
  }

  return (
    <Profiler id="MyComponent" onRender={onRender}>
      <YourComponent />
    </Profiler>
  )
}
```

**測定対象:**
- 初回レンダリング（mount）
- 再レンダリング（update）
- 不要な再レンダリング検出

#### Chrome DevTools Performance タブ
```bash
# 手動で実施
1. Chrome DevTools開く
2. Performance タブ
3. Record 開始 → ユーザーアクション → Stop
4. 分析
```

**ボトルネック特定:**
- Long Task（50ms以上）
- Layout Thrashing
- 不要な DOM 操作

#### Lighthouse スコア
```bash
# 対象 URL を明示して実行（フロントのホストポートは 3003）
docker compose exec frontend pnpm exec lighthouse http://localhost:3003 --output html
```

**Core Web Vitals:**
- **LCP (Largest Contentful Paint)**: < 2.5s
- **INP (Interaction to Next Paint)**: < 200ms
- **CLS (Cumulative Layout Shift)**: < 0.1

### 3. バンドルサイズ分析

⚠️ `pnpm build` は CLAUDE.md の自動実行禁止コマンド。ユーザーに手動実行を依頼する

```bash
docker compose exec frontend pnpm build
docker compose exec frontend npx source-map-explorer 'dist/**/*.js'
# ※ source-map-explorer は依存未登録。使う場合は導入をユーザーに確認する
```

**最適化:**
```typescript
// ❌ 大きなライブラリ全体をインポート
import lodash from 'lodash'

// ✅ 必要な関数のみインポート
import { debounce } from 'lodash'

// または tree-shaking 対応
import { debounce } from 'lodash-es'
```

### 4. キャッシング戦略

#### HTTP キャッシング
```typescript
const api = axios.create({
  headers: {
    'Cache-Control': 'max-age=3600'
  }
})
```

#### React Query キャッシング
```typescript
const { data } = useQuery({
  queryKey: ['owners'],
  queryFn: getOwners,
  staleTime: 5 * 60 * 1000,      // 5分間は fresh
  gcTime: 10 * 60 * 1000,         // 10分間メモリ保持
})
```

## パフォーマンス目標

### Backend
```
API レスポンス:     < 100ms (p95)
メモリ使用量:       < 500MB
Goroutine 数:       < 1000
Database クエリ:    < 50ms (p95)
```

### Frontend
```
LCP:                < 2.5s
INP:                < 200ms
CLS:                < 0.1
バンドルサイズ:     < 200KB (gzip)
初回表示タイム:     < 3s
```

## チェックリスト

### Go Backend
- [ ] pprof で CPU ホットパス特定
- [ ] メモリプロファイルで alloc 最適化
- [ ] Goroutine リーク検査
- [ ] N+1 クエリ排除
- [ ] インデックス設定確認

### React Frontend
- [ ] React DevTools Profiler で不要再レンダリング確認
- [ ] Lighthouse スコア 90+
- [ ] バンドルサイズ 200KB 以下
- [ ] Core Web Vitals 達成
- [ ] 画像最適化（webp、lazy loading）

## 出力形式

```markdown
## Performance Analysis Report

### Backend Analysis
**CPU Hotspots:**
- handler.CreateOwner: 45% (1.2s/2.6s)
- repository.GetOwner: 30% (0.8s)
- database/sql: 15% (0.4s)

**Memory:**
- Total Alloc: 234MB
- Goroutines: 42
- Recommendation: Reduce allocations in GetOwner

### Frontend Analysis
**Lighthouse:**
- Performance: 92 ✅
- Accessibility: 95 ✅
- Best Practices: 88 ✅
- SEO: 100 ✅

**Core Web Vitals:**
- LCP: 1.8s ✅
- INP: 120ms ✅
- CLS: 0.05 ✅

**Bundle Size:**
- Main: 145KB (gzip)
- Vendor: 78KB
- Total: 223KB ❌ (200KB超過)
  → Code splitting recommended

### 推奨最適化
1. [High] N+1 クエリ除去 (2時間)
2. [High] バンドルサイズ削減 (1.5時間)
3. [Medium] 画像最適化 (1時間)
```

## 関連スキル

- `database-indexing` - クエリパフォーマンス
