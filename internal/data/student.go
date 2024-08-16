package data

import (
	"context"

	"student/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type StudentRepo struct {
	data *Data
	log  *log.Helper
}

func NewStudentRepo(data *Data, logger log.Logger) *StudentRepo {
	return &StudentRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (repo *StudentRepo) Save(ctx context.Context, stu *biz.Student) (*biz.Student, error) {
	return stu, nil
}

func (repo *StudentRepo) Get(ctx context.Context, stu *biz.Student) (*biz.Student, error) {
	return stu, nil
}
