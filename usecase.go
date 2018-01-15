package main

import "context"

type botUseCase interface {
	Init(ctx context.Context, repoInfo *repoInfo) error
	Execute(ctx context.Context, repoInfo *repoInfo, payload interface{}) error
}
