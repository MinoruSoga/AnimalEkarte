package medicalrecord

// EMR-495 P1: 旧 finalize / Delete に紐づいていた来院完了タグ同期（SyncVisitCompletionTags）の呼び出しは、
// finalize の廃止（B の終了 route 未実装）に伴い発火点を失った。medicalRecordService.tagSyncSvc は
// B の終了 route が接続するまで読まれない（residual: owner B / LSTEP）。
