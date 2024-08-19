package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"student/internal/biz"

	pb "student/api/student/v1"
)

type StudentService struct {
	pb.UnimplementedStudentServer
	stu *biz.StudentUsecase
	log *log.Helper
}

func NewStudentService(stu *biz.StudentUsecase, logger log.Logger) *StudentService {
	return &StudentService{
		stu: stu,
		log: log.NewHelper(logger),
	}
}

func (s *StudentService) GetStudent(ctx context.Context, req *pb.GetStudentRequest) (*pb.GetStudentReply, error) {
	stu, err := s.stu.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	studentInfo := &pb.StudentInfo{
		Name:   stu.Name,
		Status: stu.Status,
		Id:     stu.ID,
	}
	return &pb.GetStudentReply{
		Student: studentInfo,
	}, nil
}

func (s *StudentService) CreateStudent(ctx context.Context, req *pb.CreateStudentRequest) (*pb.CreateStudentReply, error) {
	s.log.WithContext(ctx).Infof("CreateUser Received: %v", req)
	user := biz.Student{
		ID:   0,
		Name: req.Name,
	}
	return &pb.CreateStudentReply{}, s.stu.Create(ctx, &user)
}

func (s *StudentService) UpdateStudent(ctx context.Context, req *pb.UpdateStudentRequest) (*pb.UpdateStudentReply, error) {
	err := s.stu.Update(ctx, req.Id, &biz.Student{})

	if err != nil {
		return nil, err
	}
	return &pb.UpdateStudentReply{}, nil
}

func (s *StudentService) DeleteStudent(ctx context.Context, req *pb.DeleteStudentRequest) (*pb.DeleteStudentReply, error) {
	err := s.stu.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.DeleteStudentReply{
		Message: "删除成功",
	}, nil
}
