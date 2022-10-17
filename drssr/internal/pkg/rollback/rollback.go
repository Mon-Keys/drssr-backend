package rollback

import (
	"context"
	"fmt"
)

type rollbackCtxKey struct{}

type funcs []func()

type Rollback struct {
	rollbackFuncs funcs
}

func NewCtxRollback(ctx context.Context) (context.Context, *Rollback) {
	rollback, err := GetFromCtx(ctx)
	if err != nil {
		newRollback := Rollback{}
		ctx = SetInCtx(ctx, &newRollback)
		return ctx, &newRollback
	}

	return ctx, rollback
}

func (r *Rollback) Add(newFunc func()) {
	r.rollbackFuncs = append(r.rollbackFuncs, newFunc)
}

func (r Rollback) Run() {
	for i := len(r.rollbackFuncs) - 1; i >= 0; i-- {
		r.rollbackFuncs[i]()
	}
}

func GetFromCtx(ctx context.Context) (*Rollback, error) {
	rollback, ok := ctx.Value(rollbackCtxKey{}).(*Rollback)
	if !ok {
		return nil, fmt.Errorf("failed to get rollback array from ctx")
	}

	return rollback, nil
}

func SetInCtx(ctx context.Context, rollback *Rollback) context.Context {
	return context.WithValue(ctx, rollbackCtxKey{}, rollback)
}
