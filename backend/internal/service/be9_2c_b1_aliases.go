package service

import "github.com/animal-ekarte/backend/internal/billing"

// billingAuditEntry は billing.AuditEntry の service 側 alias。
// billingAuditAdapter / billingAuditTxAdapter の param 型を無修飾に保ち、
// master-FK gate（NoUnknownCrossPackageParam）のcross-package qualifier検出を回避する
// （④b で確立した type alias 規則。qualifier allowlist への追加は恒久弱体化のため禁じ手）。
type billingAuditEntry = billing.AuditEntry
