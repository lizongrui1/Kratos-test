package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"time"

	v1 "student/api/student/v1"
)

var (
	ErrUserNotFound = errors.NotFound(v1.ErrorReason_USER_NOT_FOUND.String(), "user not found")
)

type Student struct {
	ID        int32
	Name      string
	Info      string
	Status    int32
	UpdatedAt time.Time
	CreatedAt time.Time
}

type StudentRepo interface {
	GetStudent(ctx context.Context, id int32) (*Student, error)
	CreateStudent(ctx context.Context, stu *Student) error
	UpdateStudent(ctx context.Context, id int32, stu *Student) error
	DeleteStudent(ctx context.Context, id int32) error
}

type StudentUsecase struct {
	repo StudentRepo
	log  *log.Helper
}

func NewStudentUsecase(repo StudentRepo, logger log.Logger) *StudentUsecase {
	return &StudentUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *StudentUsecase) Get(ctx context.Context, id int32) (*Student, error) {
	uc.log.WithContext(ctx).Infof("biz.Get: %d", id)
	return uc.repo.GetStudent(ctx, id)
}

func (uc *StudentUsecase) Create(ctx context.Context, stu *Student) error {
	return uc.repo.CreateStudent(ctx, stu)
}

func (uc *StudentUsecase) Update(ctx context.Context, id int32, stu *Student) error {
	return uc.repo.UpdateStudent(ctx, id, stu)
}

func (uc *StudentUsecase) Delete(ctx context.Context, id int32) error {
	return uc.repo.DeleteStudent(ctx, id)
}
