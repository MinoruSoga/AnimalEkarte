---
description: Performance optimization standards (Go pprof, React DevTools, Lighthouse)
alwaysApply: false
globs: ["backend/**/*.go", "frontend/src/**/*.{ts,tsx}"]
---

# Performance Rules

Performance targets and optimization strategies.

## Core Rules

### 1. Go Performance Targets

```
API Response Time:   < 50ms (p95)
Memory Allocation:   < 100MB (baseline)
Goroutine Count:     < 100 (idle)
Database Query Time: < 50ms (p95)
```

### 2. CPU Profiling (pprof)

```bash
# Endpoint available at http://localhost:8080/debug/pprof

# CPU profile (30 second recording)
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# Interactive mode
(pprof) top10       # Top 10 functions
(pprof) list Main   # Function details
(pprof) graph       # Call graph
```

### 3. Memory Profiling

```bash
# Memory allocation
go tool pprof http://localhost:8080/debug/pprof/allocs

# Heap profile
go tool pprof http://localhost:8080/debug/pprof/heap

# Goroutine leak detection
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

### 4. Go Optimization Patterns

```go
// ✅ Buffer with capacity
buf := make([]Owner, 0, 100)  // Pre-allocate to avoid reallocation

// ✅ String concatenation with bytes.Buffer
var buf bytes.Buffer
for _, s := range strings {
  buf.WriteString(s)
}
result := buf.String()

// ❌ Inefficient
var result string
for _, s := range strings {
  result += s  // Creates new string each time
}

// ✅ GORM query optimization
db.Select("id", "name", "email").  // Only needed columns
  Preload("Pets", func(db *gorm.DB) *gorm.DB {
    return db.Select("id", "owner_id", "name")
  }).
  Where("clinic_id = ?", clinicID).
  Limit(100).
  Find(&owners)
```

### 5. React Performance Targets

```
FCP (First Contentful Paint):     < 1.8s
LCP (Largest Contentful Paint):  < 2.5s
CLS (Cumulative Layout Shift):    < 0.1
TTI (Time to Interactive):        < 3.8s
Bundle Size:                      < 200KB (JS)
```

### 6. React DevTools Profiler

```typescript
// React DevTools → Profiler tab to record

// memo() prevents component re-renders
export const OwnerCard = memo(function OwnerCard({ owner }: Props) {
  return <div>{owner.name}</div>;
});

// useCallback stabilizes handlers
const handleChange = useCallback((value) => {
  setData(value);
}, [setData]);

// useDeferredValue delays rendering
const deferredTerm = useDeferredValue(searchTerm);

// useMemo prevents component regeneration
const memoizedList = useMemo(() => (
  owners.map(o => <OwnerCard key={o.id} owner={o} />)
), [owners]);
```

### 7. Bundle Analysis

```bash
# Vite bundle size analysis
npm run build
npm install -g rollup-plugin-visualizer
# Check build output

# Critical JS < 200KB
# CSS < 50KB
# Image optimization (WebP)
```

### 8. Lighthouse Audit

```bash
# Check score with DevTools Lighthouse
# Target: Performance > 90

# Automated audit
npm run audit:lighthouse

# Check items:
- Unused JavaScript
- Unused CSS
- Image optimization
- Font optimization
- Minification
- Code splitting
```

### 9. Database Optimization

```go
// Detect and eliminate N+1 queries
// ✅ Preload to reduce queries
db.Preload("Pets").Where("clinic_id = ?", clinicID).Find(&owners)

// ✅ EXPLAIN ANALYZE to check execution plan
EXPLAIN ANALYZE SELECT * FROM owners WHERE clinic_id = 1 AND id = 100;
// → Index Scan (< 1ms)

// ❌ Avoid Seq Scan
EXPLAIN ANALYZE SELECT * FROM owners WHERE name LIKE '%太%';
// → Seq Scan (1000ms) → Consider index
```

## Checklist

- [ ] API Response Time < 50ms (p95)
- [ ] Regular pprof analysis (monthly)
- [ ] React memo() eliminates unnecessary re-renders
- [ ] useCallback stabilizes handlers
- [ ] useDeferredValue delays heavy computation
- [ ] Bundle size < 200KB (JS)
- [ ] Lighthouse Score > 90
- [ ] N+1 queries eliminated (Preload)
- [ ] EXPLAIN ANALYZE: no Seq Scan
- [ ] Memory allocation < 100MB (baseline)

## Performance Monitoring Commands

```bash
# Go
make test-backend  # go test ./... -v -cover

# React
make test-frontend # npm run test:run

# Production monitoring (dashboard)
make logs          # Docker Compose logs
```
