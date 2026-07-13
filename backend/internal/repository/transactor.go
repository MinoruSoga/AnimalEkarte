package repository

import (
	"context"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
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
	if err := t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	}); err != nil {
		return apperrors.Wrap(err, "transaction failed")
	}
	return nil
}

// txFromContext はコンテキストからトランザクションを取り出す。
// トランザクションがない場合は nil を返す。
func txFromContext(ctx context.Context) *gorm.DB {
	tx, _ := ctx.Value(txKey{}).(*gorm.DB)
	return tx
}

// DetachTx は ambient tx（txKey）を持ち越さない context.WithoutCancel(ctx) を返す。
// goroutine 境界を跨いで背景実行する際、親 ctx のキャンセルからは切り離したいが、
// context.WithoutCancel は ctx の値を全て保持するため ambient tx（コミット済みの可能性がある）
// も goroutine に持ち越してしまう（BE-refactor.md B-2）。
// context は値を削除できない（WithValue は既存値を shadow するのみ）ため、
// txFromContext が nil とみなす値で上書きして無効化する。
func DetachTx(ctx context.Context) context.Context {
	return context.WithValue(context.WithoutCancel(ctx), txKey{}, (*gorm.DB)(nil))
}
