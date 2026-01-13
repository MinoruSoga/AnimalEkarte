# Animal Ekarte

動物病院向け電子カルテシステム

## ⚠️ 重要: Docker環境

**このプロジェクトはDocker環境で動作します。npm/goコマンドはローカル実行禁止。**

```bash
# ✅ 正しい実行方法
docker compose exec frontend npm run <command>
docker compose exec backend go <command>
```

## 詳細ドキュメント

- [プロジェクト詳細](.claude/CLAUDE.md)
- [Backend](.claude/docs/backend.md)
- [Frontend](.claude/docs/frontend.md)

## 技術ドキュメント

- [ドキュメント目次](docs/README.md)
- [ERD（データベース設計）](docs/ERD.md)
- [API仕様（Swagger）](http://localhost:8080/swagger/index.html)
- [API設計ロードマップ](docs/API-ROADMAP.md)
- [仕様定義書](spec.md)

## クイックリファレンス

| タスク | コマンド |
|--------|---------|
| 起動 | `make up` |
| 停止 | `make down` |
| ログ | `make logs` |
| DB接続 | `make db` |
