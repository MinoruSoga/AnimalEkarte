// Package sharedkernel は domain package 間で共有される純粋（stateless・I/O なし）な
// 不変条件ヘルパーの単一実装を提供する（BE9-2D ④後の共有カーネル昇格batch・DDD Shared Kernel）。
//
// 昇格根拠は複製数（rule-of-three）ではなく「呼び出し側が恒久ドメイン境界を跨ぐ」こと:
// 例えば LockDraftMedicalRecord は medicalrecord 系 service と billing 系 service
// （billing_confirmation/estimate）の両方が呼ぶが、ADR-006 は billing→medicalrecord 依存を
// 禁じるため、どちらの domain package にも置けない。安全ガード・許可集合の複製は
// ドリフト＝安全性劣化源のため、本 package が唯一の実装となる。
//
// import 面は {apperrors, model, stdlib} のみ（ADR-006 acyclic 検証済み）。repository/
// service/handler/domain package を import してはならない。stateful な cross-cutting 依存
// （Audit/Permission 等）は本 package の対象外 — それらは consumer-side interface +
// composition root adapter で解決する（BE9-2B パターン）。
//
// 移行中の internal/service / internal/medicalrecord には既存呼び出し面互換の 1行 delegate が
// 残る（各 domain の BE9-2C/2D 移行時に呼び出し側を本 package 直参照へ切替えて解消）。
package sharedkernel
