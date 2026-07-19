---
description: Go公式命名指針とDB/API contractの命名レビュー
---

# Naming Conventions

## Go official guidance

- package 名は短く、明確な、小文字の1単語を優先する。
- underscore、mixedCaps、曖昧な `util`、`common`、`misc` を避ける。
- package 名と exported name の stutter を避ける。
- acronym は `ID`、`HTTP`、`URL` のように一貫して扱う。
- getter に `Get` を機械的に付けない。
- receiver 名は短く一貫させ、`this` / `self` を使わない。
- interface は behavior を表す自然な名前にする。単一 method interface では `-er` が適切な場合があるが、機械的に強制しない。

Go は file 名、layer 名、DTO suffix、CRUD method 名を公式規約として定めない。

## API contract

- resource path は API 全体で一貫した名詞表現を使う。
- path/query/header/body の同じ概念には同じ名前を使う。
- JSON field は公開 contract に従い、一度公開した名前を内部 refactor で無断変更しない。
- status/error code は client が programmatically 扱える stable な名前にする。
- action endpoint が必要な場合は、resource state transition として意味が明確な名前を使う。

REST resource の複数形、kebab-case、envelope 等は Gin 公式要件ではない。project の OpenAPI contract で一貫性を決める。

## Database contract

- table/column/constraint/index 名は schema 全体で一貫させる。
- foreign key と timestamp は意味が分かる名前にし、略語を乱用しない。
- boolean 名は真偽の意味が一意になる肯定形を優先する。
- migration で既存名を変更する場合は application/API/analytics への互換性を評価する。

plural table、snake_case、suffix、最大長等は PostgreSQL/project の設計判断であり、Go/Gin公式規約ではない。既存 schema の naming contract を変更する場合は migration/ADR で扱う。

## Sources

- [Go package names](https://go.dev/blog/package-names)
- [Effective Go: Names](https://go.dev/doc/effective_go#names)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
