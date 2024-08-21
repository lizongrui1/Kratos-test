package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"student/internal/biz"
	"student/internal/data/model"
)

var _ biz.RedisClient = (*RedisClient)(nil)

type StudentRepo struct {
	data *Data
	log  *log.Helper
	rdb  RedisClient
}

func NewStudentRepo(data *Data, logger log.Logger) *StudentRepo {
	return &StudentRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *StudentRepo) GetStudent(ctx context.Context, id int32) (*biz.Student, error) {
	var stu biz.Student
	err := r.data.gormDB.Where("id = ?", id).First(&stu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New(404, "用户不存在", "用户不存在")
	}
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
	//_, err := r.GetStudent(ctx, stu.ID)
	//if err != nil {
	//	return errors.New(404, "用户名已存在", "用户注册失败")
	//
	//}
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
