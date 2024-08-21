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
	GetStudent(ctx context.Context, id int32) (*Student, error)
	CreateStudent(ctx context.Context, stu *Student) error
	UpdateStudent(ctx context.Context, id int32, stu *Student) error
	DeleteStudent(ctx context.Context, id int32) error

	//redis
	GetStuById(ctx context.Context, id int32) (*Student, error)
}

type StudentUsecase struct {
	repo StudentRepo
	log  *log.Helper
	//rdb  RedisClient
}
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}

func NewStudentUsecase(repo StudentRepo, logger log.Logger) *StudentUsecase {
	return &StudentUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (s *StudentUsecase) Get(ctx context.Context, id int32) (*Student, error) {
	s.log.WithContext(ctx).Infof("biz.Get: %d", id)
	return s.repo.GetStudent(ctx, id)
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

//func (s *StudentUsecase) GetStuById(ctx context.Context, id int32) (*Student, error) {
//	var student *Student
//	s.log.WithContext(ctx).Infof("biz.GetStuById: %d", id)
//	cacheKey := "student:" + fmt.Sprint(id)
//	val, err := s.rdb.Get(ctx, cacheKey)
//	if err == nil && val != "" {
//		if err := json.Unmarshal([]byte(val), &student); err == nil {
//			s.log.WithContext(ctx).Infof("biz.GetStuById - Cache Hit: %v", student)
//			return student, nil
//		}
//		s.log.WithContext(ctx).Errorf("failed to unmarshal student from redis: %v", err)
//	}
//	// 缓存未命中或解析失败，从数据库获取数据
//	student, err = s.repo.GetStudent(ctx, id)
//	if err != nil {
//		return nil, err
//	}
//	// 数据库查询成功，将结果缓存到 Redis
//	data, err := json.Marshal(student)
//	if err != nil {
//		s.log.WithContext(ctx).Errorf("failed to marshal student to json: %v", err)
//	} else {
//		err = s.rdb.Set(ctx, cacheKey, data, time.Minute*5)
//		if err != nil {
//			s.log.WithContext(ctx).Errorf("failed to set student data to redis: %v", err)
//		}
//	}
//	return student, nil
//}
