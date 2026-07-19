---
description: Go公式の言語・API設計レビュー指針の要約
---

# Go Language Review Guide

backend 全体の正本は [Go/Gin Backend Guidelines](../rules/go-gin-backend-guidelines.md)。ここでは Go 言語固有の確認事項だけを要約する。

## Formatting and source

- `gofmt` を適用する。
- import は formatter/tool に任せ、dot import は特殊な test 等を除いて避ける。
- package comment と exported identifier の comment は GoDoc で意味が通る文にする。
- acronym を識別子内で一貫して扱う（`ID`、`HTTP`、`URL`）。
- receiver 名は短く一貫させ、`this` や `self` を使わない。

## Package API

- package 名は短く明確な小文字の1単語を優先する。
- package 名と exported name の stutter を避ける。
- export を最小化し、zero value を可能なら有用にする。
- interface は一般に利用側で、必要な最小メソッドだけを定義する。
- 実装 package は concrete type を返すことを基本にする。
- constructor は依存や不変条件を確立する場合に使う。

## Errors

- error は戻り値で明示的に扱い、無視しない。
- 呼び出し側が判定する必要がある場合だけ stable な sentinel/type を公開する。
- 文脈追加は `%w` を使い、`errors.Is` / `errors.As` を可能にする。
- error string は通常、小文字で始め、句点や改行で終えない。
- panic は通常の失敗経路に使わない。

## Context and concurrency

- request-scoped な関数は `context.Context` を第1引数で受け取る。
- Context を struct に保存しない。
- cancel function を必ず呼び、deadline/cancellation を下流へ伝播する。
- goroutine の ownership、終了条件、error/panic 経路を明確にする。
- channel は通信と ownership の表現に使い、不要な async 化を避ける。
- shared mutable state は synchronization または ownership のどちらで守るか明確にする。

## Types and control flow

- 型変換を明示し、情報が失われる numeric conversion を検証する。
- `any` や reflection は型で表現できない境界に限定する。
- early return で正常系を左に保ち、深い nesting を避ける。
- defer の評価時点と loop 内での resource lifetime を意識する。
- map/slice を共有する場合は caller/callee の mutation ownership を明確にする。

## Tooling

- 対象変更に応じて `go test`、`go vet`、race detector、project lint を使う。
- AnimalEkarte では Go command を host で直接実行せず、project の Docker 実行規則に従う。

## Primary sources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Package names](https://go.dev/blog/package-names)
- [`context` package](https://pkg.go.dev/context)
