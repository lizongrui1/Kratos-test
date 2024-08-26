// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"student/internal/biz"
	"student/internal/conf"
	"student/internal/data"
	"student/internal/server"
	"student/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
)

// Injectors from wire.go:

// wireApp init kratos application.
//func wireApp(confServer *conf.Server, confData *conf.Data, logger log.Logger, repo data.StudentRepo) (*kratos.App, func(), error) {
//	db, err := data.NewGormDB(confData)
//	rdb := data.NewRdb(confData, logger)
//	redisClient := data.NewRedisClient(rdb, logger)
//	if err != nil {
//		return nil, nil, err
//	}
//	dataData, cleanup, err := data.NewData(logger, db, rdb)
//	if err != nil {
//		return nil, nil, err
//	}
//	studentRepo := data.NewStudentRepo(dataData, logger, redisClient)
//	studentUsecase := biz.NewStudentUsecase(studentRepo, redisClient, logger)
//	studentService := service.NewStudentService(studentUsecase, logger)
//	grpcServer := server.NewGRPCServer(confServer, studentService, logger)
//	httpServer := server.NewHTTPServer(confServer, studentService, logger)
//	app := newApp(logger, grpcServer, httpServer, studentRepo)
//	return app, func() {
//		cleanup()
//	}, nil
//}

func wireApp(confServer *conf.Server, confData *conf.Data, logger log.Logger) (*kratos.App, func(), *data.StudentRepo, error) {
	db, err := data.NewGormDB(confData)
	if err != nil {
		return nil, nil, nil, err
	}
	rdb := data.NewRdb(confData, logger)
	redisClient := data.NewRedisClient(rdb, logger)
	dataData, cleanup, err := data.NewData(logger, db, rdb)
	if err != nil {
		return nil, nil, nil, err
	}
	studentRepo := data.NewStudentRepo(dataData, logger, redisClient)
	studentUsecase := biz.NewStudentUsecase(studentRepo, redisClient, logger)
	studentService := service.NewStudentService(studentUsecase, logger)
	grpcServer := server.NewGRPCServer(confServer, studentService, logger)
	httpServer := server.NewHTTPServer(confServer, studentService, logger)
	app := newApp(logger, grpcServer, httpServer, studentRepo)
	return app, cleanup, studentRepo, nil
}


