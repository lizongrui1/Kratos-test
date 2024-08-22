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

func NewStudentRepo(data *Data, logger log.Logger, redisClient *RedisClient) biz.StudentRepo {
	return &StudentRepo{
		data: data,
		log:  log.NewHelper(logger),
		rdb:  redisClient,
	}
}

func (s *StudentRepo) ListStudent(ctx context.Context) ([]*biz.Student, error) {
	var students []*model.Student
	err := s.data.db.WithContext(ctx).Find(&students).Error
	if err != nil {
		return nil, err
	}
	var stus []*biz.Student
	for _, stu := range students {
		stus = append(stus, &biz.Student{
			ID:        stu.ID,
			Name:      stu.Name,
			Status:    stu.Status,
			Info:      stu.Info,
			UpdatedAt: stu.UpdatedAt,
			CreatedAt: stu.CreatedAt,
		})
	}
	for _, stu := range stus {
		stuJSON, err := json.Marshal(stu)
		if err != nil {
			return nil, err
		}
		err = s.rdb.RPush(ctx, "students:list", stuJSON).Err()
		if err != nil {
			return nil, err
		}
	}
	return stus, nil
}

func (s *StudentRepo) GetStudentById(ctx context.Context, id int32) (*biz.Student, error) {
	var stu biz.Student
	err := s.data.db.Where("id = ?", id).First(&stu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New(404, "USER_IS_NOT_EXIST", "用户不存在")
	}
	s.log.WithContext(ctx).Info("db: GetStudentById, id: ", id)
	return &biz.Student{
		ID:        stu.ID,
		Name:      stu.Name,
		Status:    stu.Status,
		Info:      stu.Info,
		UpdatedAt: stu.UpdatedAt,
		CreatedAt: stu.CreatedAt,
	}, nil
}

func (s *StudentRepo) GetStudentByName(ctx context.Context, name string) (*biz.Student, error) {
	var stu model.Student
	err := s.data.db.Where("name = ?", name).First(&stu).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New(404, "USER_IS_NOT_EXIST", "用户不存在")
	}
	s.log.WithContext(ctx).Info("db: GetStudentById, Name: ", name)
	return &biz.Student{
		ID:        stu.ID,
		Name:      stu.Name,
		Status:    stu.Status,
		Info:      stu.Info,
		UpdatedAt: stu.UpdatedAt,
		CreatedAt: stu.CreatedAt,
	}, nil
}

//func (s *StudentRepo) GetStuByRdb(ctx context.Context, id int32) (*biz.Student, error) {
//	var student biz.Student
//	s.log.WithContext(ctx).Infof("GetStuByRdb: %d", id)
//	cacheKey := "student:" + fmt.Sprint(id)
//	val, err := s.rdb.Get(ctx, cacheKey)
//	if errors.Is(err, redis.Nil) {
//		go func() {
//			fmt.Println("缓存未命中，异步加载数据并更新缓存")
//			studentPtr, err := s.GetStudentById(ctx, id)
//			if err != nil {
//				l.Printf("从数据库加载数据失败: %v", err)
//				return
//			}
//			data, err := json.Marshal(studentPtr)
//			if err != nil {
//				l.Printf("序列化 student 失败: %v", err)
//				return
//			}
//			err = s.rdb.Set(ctx, cacheKey, data, 3*time.Minute)
//			if err != nil {
//				l.Printf("异步更新缓存失败: %v", err)
//			} else {
//				fmt.Printf("缓存更新成功: %s = %s\n", cacheKey, data)
//			}
//		}()
//		return nil, fmt.Errorf("缓存未命中，数据正在异步加载")
//	}
//	if err := json.Unmarshal([]byte(val), &student); err != nil {
//		s.log.WithContext(ctx).Infof("biz.GetStuByRdb - Cache Hit: %v", student)
//		return nil, err
//	}
//	return &student, nil
//}

func (s *StudentRepo) GetStuByRdb(ctx context.Context, id int32) (*biz.Student, error) {
	var student biz.Student
	s.log.WithContext(ctx).Infof("biz.GetStuByRdb: %d", id)
	cacheKey := "student:" + fmt.Sprint(id)
	val, err := s.rdb.Get(ctx, cacheKey)
	if err == nil {
		if err := json.Unmarshal([]byte(val), &student); err == nil {
			s.log.WithContext(ctx).Infof("Cache Hit: %v", student)
			return &student, nil
		}
		s.log.WithContext(ctx).Errorf("failed to unmarshal student: %v", err)
	} else if !errors.Is(err, redis.Nil) {
		s.log.WithContext(ctx).Errorf("redis get error: %v", err)
	}
	// 缓存未命中或解析失败，从数据库获取数据
	studentPtr, err := s.GetStudentById(ctx, id)
	if err != nil {
		return nil, err
	}
	student = *studentPtr
	// 数据库查询成功，将结果缓存到 Redis
	data, err := json.Marshal(student)
	if err != nil {
		s.log.WithContext(ctx).Errorf("failed to marshal student to json: %v", err)
	} else {
		err = s.rdb.Set(ctx, cacheKey, data, 5*time.Minute)
		if err != nil {
			s.log.WithContext(ctx).Errorf("failed to set student data to redis: %v", err)
		}
	}
	return &student, nil
}

func (s *StudentRepo) CreateStudent(ctx context.Context, stu *biz.Student) error {
	_, err := s.GetStudentByName(ctx, stu.Name)
	if err == nil {
		return errors.New(409, "USER_IS_EXIST", "用户已存在，无法创建")
	} else {
		return s.data.db.Model(&model.Student{}).Create(&model.Student{
			Name:   stu.Name,
			Info:   stu.Info,
			Status: stu.Status,
		}).Error
	}
}

func (s *StudentRepo) UpdateStudent(ctx context.Context, id int32, stu *biz.Student) error {
	//return s.data.db.WithContext(ctx).Model(&model.Student{}).Where("id = ?", id).Updates(&model.Student{
	//	Name:   stu.Name,
	//	Info:   stu.Info,
	//	Status: stu.Status,
	//}).Error
	tx := s.data.db.WithContext(ctx).Begin()
	err := tx.Model(&model.Student{}).Where("id = ?", id).Updates(&model.Student{
		Name:   stu.Name,
		Info:   stu.Info,
		Status: stu.Status,
	}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	redisKey := fmt.Sprintf("student:%d", id)
	if err := s.rdb.Del(ctx, redisKey).Err(); err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (s *StudentRepo) DeleteStudent(ctx context.Context, id int32) error {
	//return s.data.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Student{}).Error
	tx := s.data.db.WithContext(ctx).Begin()
	if err := tx.Delete(&model.Student{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	redisKey := fmt.Sprintf("student:%d", id)
	if err := s.rdb.Del(ctx, redisKey).Err(); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
