package repository

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

// Transactor はトランザクション境界を管理するインターフェース。
// サービス層は *gorm.DB を直接保持せず、Transactor を介してトランザクションを開始する。
// Repository メソッドは dbOrTx(ctx, r.db) を使うことで、WithTx 内から呼ばれた場合に
// 自動的に同一トランザクションを使用する。
type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type gormTransactor struct{ db *gorm.DB }

// NewTransactor は Transactor を初期化して返す。
func NewTransactor(db *gorm.DB) Transactor {
	return &gormTransactor{db: db}
}

func (t *gormTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// txFromContext はコンテキストからトランザクションを取り出す。
// トランザクションがない場合は nil を返す。
func txFromContext(ctx context.Context) *gorm.DB {
	tx, _ := ctx.Value(txKey{}).(*gorm.DB)
	return tx
}
