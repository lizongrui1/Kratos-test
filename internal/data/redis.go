package data

import (
	"context"
	"encoding/json"
	"fmt"
	"student/internal/biz"
	"time"
)

func (r *StudentRepo) GetStuById(ctx context.Context, id int32) (*biz.Student, error) {
	var student *biz.Student
	r.log.WithContext(ctx).Infof("biz.GetStuById: %d", id)
	cacheKey := "student:" + fmt.Sprint(id)
	val, err := r.rdb.Get(ctx, cacheKey)
	if err == nil && val != "" {
		if err := json.Unmarshal([]byte(val), &student); err == nil {
			r.log.WithContext(ctx).Infof("biz.GetStuById - Cache Hit: %v", student)
			return student, nil
		}
		r.log.WithContext(ctx).Errorf("failed to unmarshal student from redis: %v", err)
	}
	// 缓存未命中或解析失败，从数据库获取数据
	student, err = biz.StudentRepo.GetStudent(ctx, id)
	if err != nil {
		return nil, err
	}
	// 数据库查询成功，将结果缓存到 Redis
	data, err := json.Marshal(student)
	if err != nil {
		r.log.WithContext(ctx).Errorf("failed to marshal student to json: %v", err)
	} else {
		err = r.rdb.Set(ctx, cacheKey, data, time.Minute*5)
		if err != nil {
			r.log.WithContext(ctx).Errorf("failed to set student data to redis: %v", err)
		}
	}
	return student, nil
}
