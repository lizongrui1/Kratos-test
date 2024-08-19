package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"student/internal/biz"
	"student/internal/data/model"

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

func (r *StudentRepo) GetStudent(ctx context.Context, id int32) (*biz.Student, error) {
	var stu biz.Student
	r.data.gormDB.Where("id = ?", id).First(&stu)
	r.log.WithContext(ctx).Info("gormDB: GetStudent, id: ", id)
	return &biz.Student{
		ID:        stu.ID,
		Name:      stu.Name,
		Status:    stu.Status,
		Info:      stu.Info,
		UpdatedAt: stu.UpdatedAt,
		CreatedAt: stu.CreatedAt,
	}, nil
}

func (r *StudentRepo) CreateStudent(ctx context.Context, stu *biz.Student) error {
	// 判断名称是否存在,存在则返回错误
	_, err := r.GetStudent(ctx, stu.ID)
	if err != nil {
		return errors.New(400, "用户名已存在", "用户注册失败")

	}
	return r.data.gormDB.Model(&model.Student{}).Create(&model.Student{
		Name:   stu.Name,
		Info:   stu.Info,
		Status: stu.Status,
	}).Error
}

func (r *StudentRepo) UpdateStudent(ctx context.Context, id int32, stu *biz.Student) error {
	return r.data.gormDB.WithContext(ctx).Model(&model.Student{}).Where("id = ?", id).Updates(&model.Student{
		Name:   stu.Name,
		Info:   stu.Info,
		Status: stu.Status,
	}).Error
}

func (r *StudentRepo) DeleteStudent(ctx context.Context, id int32) error {
	return r.data.gormDB.WithContext(ctx).Where("id = ?", id).Delete(&model.Student{}).Error
}
