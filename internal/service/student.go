package service

import (
	"context"
	pb "student/api/student/v1"
	"student/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type StudentService struct {
	pb.UnimplementedStudentServer
	log *log.Helper
	stu *biz.StudentUsecase
}

func NewStudentService(stu *biz.StudentUsecase, logger log.Logger) *StudentService {
	return &StudentService{
		stu: stu,
		log: log.NewHelper(logger),
	}
}

func (s *StudentService) ListStudent(ctx context.Context, req *pb.ListStudentRequest) (*pb.ListStudentReply, error) {
	students, err := s.stu.List(ctx)
	if err != nil {
		return nil, err
	}
	var studentInfos []*pb.StudentInfo
	for _, stu := range students {
		studentInfo := &pb.StudentInfo{
			Name:   stu.Name,
			Id:     stu.ID,
			Status: stu.Status,
		}
		studentInfos = append(studentInfos, studentInfo)
	}
	return &pb.ListStudentReply{
		Student: studentInfos,
	}, nil
}

func (s *StudentService) GetStudent(ctx context.Context, req *pb.GetStudentRequest) (*pb.GetStudentReply, error) {
	stu, err := s.stu.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	studentInfo := &pb.StudentInfo{
		Name:   stu.Name,
		Id:     stu.ID,
		Status: stu.Status,
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
	err := s.stu.Create(ctx, &user)
	if err != nil {
		return nil, err
	}
	return &pb.CreateStudentReply{
		Message: "创建成功",
	}, nil
}

func (s *StudentService) UpdateStudent(ctx context.Context, req *pb.UpdateStudentRequest) (*pb.UpdateStudentReply, error) {
	err := s.stu.Update(ctx, req.Id, &biz.Student{
		Name:   req.Name,
		Info:   req.Info,
		Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &pb.UpdateStudentReply{
		Message: "更新成功",
	}, nil
}

func (s *StudentService) DeleteStudent(ctx context.Context, req *pb.DeleteStudentRequest) (*pb.DeleteStudentReply, error) {
	stu, err := s.stu.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	err = s.stu.Delete(ctx, stu.ID)
	if err != nil {
		return nil, err
	}
	return &pb.DeleteStudentReply{
		Message: "删除成功",
	}, nil
}
