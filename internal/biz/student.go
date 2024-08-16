package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"

	v1 "student/api/helloworld/v1"
)

var (
	ErrUserNotFound = errors.NotFound(v1.ErrorReason_USER_NOT_FOUND.String(), "user not found")
)

type Student struct {
	ID      string
	Name    string
	Sayname string
}

type StudentRepo interface {
	Save(context.Context, *Student) (*Student, error)
	Get(context.Context, *Student) (*Student, error)
}

type StudentUsercase struct {
	repo StudentRepo
	log  *log.Helper
}

func NewStudentUsercase(repo StudentRepo, logger log.Logger) *StudentUsercase {
	return &StudentUsercase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *StudentUsercase) CreateStudent(ctx context.Context, stu *Student) (*Student, error) {
	uc.log.WithContext(ctx).Infof("CreateStudent: %v", stu.ID)
	return uc.repo.Save(ctx, stu)
}
