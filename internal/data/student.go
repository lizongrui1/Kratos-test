package data

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"student/internal/biz"
	"student/internal/data/model"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ biz.RedisClient = (*RedisClient)(nil)

type StudentRepo struct {
	data *Data
	log  *log.Helper
	rdb  *RedisClient
}

func NewStudentRepo(data *Data, logger log.Logger, rdb *RedisClient) *StudentRepo {
	return &StudentRepo{
		data: data,
		log:  log.NewHelper(logger),
		rdb:  rdb,
	}
}

// MySQL

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
	var stu model.Student
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

func (s *StudentRepo) CreateStudent(ctx context.Context, stu *biz.Student) error {
	_, err := s.GetStudentByName(ctx, stu.Name)
	if err == nil {
		return errors.New(409, "USER_IS_EXIST", "用户已存在，无法创建")
	} else {
		return s.SendCreateStudentMsg(ctx, stu)
	}
}

func (s *StudentRepo) UpdateStudent(ctx context.Context, id int32, stu *biz.Student) error {
	var existStu model.Student
	err := s.data.db.WithContext(ctx).Where("name = ?", stu.Name).First(&existStu).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existStu.Name == stu.Name && existStu.Info == stu.Info && existStu.Status == stu.Status {
		s.log.WithContext(ctx).Info("No changes detected")
		return errors.New(400, "USER_IS_UPDATED", "该学生信息不需要修改")
	}
	tx := s.data.db.WithContext(ctx).Begin()
	err = tx.Model(&model.Student{}).Where("id = ?", id).Updates(&model.Student{
		Name:   stu.Name,
		Info:   stu.Info,
		Status: stu.Status,
	}).Error
	if err != nil {
		tx.Rollback()
		s.log.WithContext(ctx).Errorf("Failed to update student in database: %v", err)
		return err
	}
	if err := tx.Commit().Error; err != nil {
		s.log.WithContext(ctx).Errorf("Failed to commit transaction: %v", err)
		return err
	}
	redisKey := fmt.Sprintf("student:%d", id)
	listKey := "students:list"
	if err := s.rdb.Del(ctx, redisKey, listKey).Err(); err != nil {
		s.log.WithContext(ctx).Errorf("Failed to delete Redis keys [%s, %s]: %v", redisKey, listKey, err)
		return err
	}
	if err := s.SendUpdateStudentMsg(ctx, stu); err != nil {
		s.log.WithContext(ctx).Errorf("发送创建学生消息失败: %v", err)
		return err
	}
	return nil
}

func (s *StudentRepo) DeleteStudent(ctx context.Context, id int32) error {
	tx := s.data.db.WithContext(ctx).Begin()
	if err := tx.Delete(&model.Student{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	redisKey := fmt.Sprintf("student:%d", id)
	listKey := "students:list"
	if err := s.rdb.Del(ctx, redisKey, listKey).Err(); err != nil {
		s.log.WithContext(ctx).Errorf("failed to delete Redis keys: %v", err)
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	if err := s.SendDeleteStudentMsg(ctx, id); err != nil {
		return err
	}
	return nil
}

// RedisClient

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

func (s *StudentRepo) SendGetStudentMsg(ctx context.Context, id int32) error {
	msg := &Msg{
		Topic:     "student_create",
		Body:      []byte(fmt.Sprintf("%d", id)),
		Partition: 0,
	}
	//topicPartition := fmt.Sprintf("%s:%d", msg.Topic, msg.Partition)
	return s.rdb.PushMsg(ctx, msg.Topic, msg.Body).Err()
}

func (s *StudentRepo) SendCreateStudentMsg(ctx context.Context, stu *biz.Student) error {
	data, err := json.Marshal(stu)
	if err != nil {
		s.log.WithContext(ctx).Errorf("failed to marshal student data: %v", err)
		return err
	}
	msg := &Msg{
		Topic:     "student_create",
		Body:      data,
		Partition: 0,
	}
	return s.rdb.PushMsg(ctx, msg.Topic, msg.Body).Err()
}

func (s *StudentRepo) SendDeleteStudentMsg(ctx context.Context, id int32) error {
	msg := &Msg{
		Topic:     "student_delete",
		Body:      []byte(fmt.Sprintf("%d", id)),
		Partition: 0,
	}
	if err := s.rdb.PushMsg(ctx, msg.Topic, msg.Body).Err(); err != nil {
		return err
	}
	return nil
}

func (s *StudentRepo) SendUpdateStudentMsg(ctx context.Context, stu *biz.Student) error {
	data, err := json.Marshal(stu)
	if err != nil {
		s.log.WithContext(ctx).Errorf("Failed to marshal student data: %v", err)
		return err
	}
	msg := &Msg{
		Topic:     "student_update",
		Body:      data,
		Partition: 0,
	}
	if err := s.rdb.PushMsg(ctx, msg.Topic, msg.Body).Err(); err != nil {
		s.log.WithContext(ctx).Errorf("Failed to push update student message: %v", err)
		return err
	}
	return nil
}

func (s *StudentRepo) ConsumeStudentCreateMsg(ctx context.Context) {
	topic := "student_create"
	for {
		cmd := s.rdb.PopMsg(ctx, 0, topic)
		messages, err := cmd.Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				s.log.WithContext(ctx).Info("No messages found")
				time.Sleep(5 * time.Second)
				continue
			}
			s.log.WithContext(ctx).Errorf("Error consuming messages: %v", err)
			continue
		}
		for _, message := range messages[1:] {
			s.HandleCreateStudentMsg(ctx, message)
		}
		time.Sleep(10 * time.Second)
	}
}

func (s *StudentRepo) ConsumeStudentDeleteMsg(ctx context.Context) {
	topic := "student_delete"
	for {
		cmd := s.rdb.PopMsg(ctx, 0, topic)
		messages, err := cmd.Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				s.log.WithContext(ctx).Info("No messages found")
				time.Sleep(5 * time.Second)
				continue
			}
			s.log.WithContext(ctx).Errorf("Error consuming messages: %v", err)
			continue
		}
		for _, message := range messages[1:] {
			id, err := strconv.ParseInt(message, 10, 32)
			if err != nil {
				s.log.WithContext(ctx).Errorf("Failed to parse student ID: %v", err)
				continue
			}
			s.HandleDeleteStudentMsg(ctx, int32(id))
		}
		time.Sleep(10 * time.Second)
	}
}

func (s *StudentRepo) ConsumeStudentUpdateMsg(ctx context.Context) error {
	topic := "student_update"
	for {
		cmd := s.rdb.PopMsg(ctx, 0, topic)
		messages, err := cmd.Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				s.log.WithContext(ctx).Info("No messages found")
				time.Sleep(5 * time.Second)
				continue
			}
			s.log.WithContext(ctx).Errorf("Error consuming messages: %v", err)
			continue
		}
		for _, message := range messages[1:] {
			s.HandleUpdateStudentMsg(ctx, message)
		}
		time.Sleep(10 * time.Second)
	}
}

func (s *StudentRepo) HandleCreateStudentMsg(ctx context.Context, message string) {
	var stu biz.Student
	if err := json.Unmarshal([]byte(message), &stu); err != nil {
		s.log.WithContext(ctx).Errorf("Failed to unmarshal student data: %v", err)
		return
	}
	modelStudent := &model.Student{
		Name:   stu.Name,
		Info:   stu.Info,
		Status: stu.Status,
	}
	if err := s.data.db.Model(&model.Student{}).Create(modelStudent).Error; err != nil {
		s.log.WithContext(ctx).Errorf("Failed to create student in database: %v", err)
		return
	}
	s.log.WithContext(ctx).Info("Successfully created student.")
}

func (s *StudentRepo) HandleDeleteStudentMsg(ctx context.Context, id int32) {
	var stu model.Student
	if err := s.data.db.Where("id = ?", id).Delete(&stu).Error; err != nil {
		s.log.WithContext(ctx).Errorf("Failed to delete student in database: %v", err)
		return
	}
	s.log.WithContext(ctx).Info("Successfully deleted student.")
}

func (s *StudentRepo) HandleUpdateStudentMsg(ctx context.Context, message string) {
	var stu biz.Student
	if err := json.Unmarshal([]byte(message), &stu); err != nil {
		s.log.WithContext(ctx).Errorf("Failed to unmarshal student data: %v", err)
		return
	}
	modelStudent := &model.Student{
		Name:   stu.Name,
		Info:   stu.Info,
		Status: stu.Status,
	}
	if err := s.data.db.Model(&model.Student{}).Where("id = ?", stu.ID).Updates(modelStudent).Error; err != nil {
		return
	}
	s.log.WithContext(ctx).Info("Successfully updated student.")
}

func (s *StudentRepo) Consume(ctx context.Context, topic string, partition int, h Handler) error {
	for {
		body, err := s.rdb.PopMsg(ctx, 2*time.Second, topic).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				s.log.WithContext(ctx).Info("no message found")
				time.Sleep(time.Second)
				continue
			}
			s.log.WithContext(ctx).Info("consuming message error")
			return err
		}
		for _, v := range body {
			if err := h(&Msg{Topic: topic, Body: []byte(v), Partition: partition}); err != nil {
				s.log.WithContext(ctx).Info("handle message error")
				continue
			}
			if err := s.rdb.DeleteMessage(ctx, topic, v).Err(); err != nil {
				s.log.WithContext(ctx).Info("delete message error")
				continue
			}
			s.log.WithContext(ctx).Info("message processed and deleted")
		}
	}
}
