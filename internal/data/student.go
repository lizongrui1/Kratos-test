package data

import (
	"context"
	"encoding/json"

	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"student/internal/biz"
	"student/internal/data/model"
	"time"
)

var _ biz.RedisClient = (*RedisClient)(nil)

type StudentRepo struct {
	data *Data
	log  *log.Helper
	rdb  *RedisClient
}

func NewStudentRepo(data *Data, logger log.Logger, redisClient *RedisClient) *StudentRepo {
	return &StudentRepo{
		data,
		log.NewHelper(logger),
		redisClient,
	}
}

func (s *StudentRepo) GetStudent(ctx context.Context, id int32) (*biz.Student, error) {
	var stu biz.Student
	err := s.data.gormDB.Where("id = ?", id).First(&stu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New(404, "用户不存在", "用户不存在")
	}
	s.log.WithContext(ctx).Info("gormDB: GetStudent, id: ", id)
	return &biz.Student{
		ID:        stu.ID,
		Name:      stu.Name,
		Status:    stu.Status,
		Info:      stu.Info,
		UpdatedAt: stu.UpdatedAt,
		CreatedAt: stu.CreatedAt,
	}, nil
}

func (s *StudentRepo) GetStuById(ctx context.Context, id int32) (*biz.Student, error) {
	var student biz.Student
	s.log.WithContext(ctx).Infof("biz.GetStuById: %d", id)
	cacheKey := "student:" + fmt.Sprint(id)
	val, err := s.rdb.Get(ctx, cacheKey)
	if err == nil {
		if err := json.Unmarshal([]byte(val), &student); err == nil {
			s.log.WithContext(ctx).Infof("biz.GetStuById - Cache Hit: %v", student)
			return &student, nil
		}
		s.log.WithContext(ctx).Errorf("failed to unmarshal student from redis: %v", err)
	} else if !errors.Is(err, redis.Nil) {
		s.log.WithContext(ctx).Errorf("redis get error: %v", err)
	}
	// 缓存未命中或解析失败，从数据库获取数据
	studentPtr, err := s.GetStudent(ctx, id)
	if err != nil {
		return nil, err
	}
	student = *studentPtr
	// 数据库查询成功，将结果缓存到 Redis
	data, err := json.Marshal(student)
	if err != nil {
		s.log.WithContext(ctx).Errorf("failed to marshal student to json: %v", err)
	} else {
		err = s.rdb.Set(ctx, cacheKey, data, time.Minute*5)
		if err != nil {
			s.log.WithContext(ctx).Errorf("failed to set student data to redis: %v", err)
		}
	}
	return &student, nil
}

func (s *StudentRepo) CreateStudent(ctx context.Context, stu *biz.Student) error {
	//_, err := r.GetStudent(ctx, stu.ID)
	//if err != nil {
	//	return errors.New(404, "用户名已存在", "用户注册失败")
	//
	//}
	return s.data.gormDB.Model(&model.Student{}).Create(&model.Student{
		Name:   stu.Name,
		Info:   stu.Info,
		Status: stu.Status,
	}).Error
}

func (s *StudentRepo) UpdateStudent(ctx context.Context, id int32, stu *biz.Student) error {
	return s.data.gormDB.WithContext(ctx).Model(&model.Student{}).Where("id = ?", id).Updates(&model.Student{
		Name:   stu.Name,
		Info:   stu.Info,
		Status: stu.Status,
	}).Error
}

func (s *StudentRepo) DeleteStudent(ctx context.Context, id int32) error {
	return s.data.gormDB.WithContext(ctx).Where("id = ?", id).Delete(&model.Student{}).Error
}
