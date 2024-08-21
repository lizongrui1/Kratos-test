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
	// mysql
	ListStudent(ctx context.Context) ([]*Student, error)
	GetStudentById(ctx context.Context, id int32) (*Student, error)
	CreateStudent(ctx context.Context, stu *Student) error
	UpdateStudent(ctx context.Context, id int32, stu *Student) error
	DeleteStudent(ctx context.Context, id int32) error

	//redis
	GetStuByRdb(ctx context.Context, id int32) (*Student, error)
}

type StudentUsecase struct {
	repo StudentRepo
	log  *log.Helper
	rdb  RedisClient
}
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}

func NewStudentUsecase(repo StudentRepo, rdb RedisClient, logger log.Logger) *StudentUsecase {
	return &StudentUsecase{repo: repo, rdb: rdb, log: log.NewHelper(logger)}
}

func (s *StudentUsecase) List(ctx context.Context, id int32) ([]*Student, error) {
	return s.repo.ListStudent(ctx)
}

func (s *StudentUsecase) Get(ctx context.Context, id int32) (*Student, error) {
	return s.repo.GetStuByRdb(ctx, id)
}

func (s *StudentUsecase) Create(ctx context.Context, stu *Student) error {
	return s.repo.CreateStudent(ctx, stu)
}

func (s *StudentUsecase) Update(ctx context.Context, id int32, stu *Student) error {
	return s.repo.UpdateStudent(ctx, id, stu)
}

func (s *StudentUsecase) Delete(ctx context.Context, id int32) error {
	return s.repo.DeleteStudent(ctx, id)
}
