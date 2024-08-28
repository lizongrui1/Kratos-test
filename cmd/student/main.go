package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"student/internal/biz"
	"student/internal/conf"
	"student/internal/data"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

//func fetchData(url string, target interface{}) error {
//	resp, err := http.Get(url)
//	if err != nil {
//		return err
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK {
//		return fmt.Errorf("failed to fetch data: status code %d", resp.StatusCode)
//	}
//
//	body, err := io.ReadAll(resp.Body)
//	if err != nil {
//		return err
//	}
//	return json.Unmarshal(body, target)
//}

func fetchData(url string, target interface{}) error {
	client := &http.Client{} // 创建一个独立的 HTTP 客户端
	resp, err := client.Get(url)
	if err != nil {
		log.Errorf("Error making GET request to %s: %v", url, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Errorf("Non-OK HTTP status: %d from %s", resp.StatusCode, url)
		return fmt.Errorf("failed to fetch data: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading response body from %s: %v", url, err)
		return err
	}
	log.Infof("Successfully fetched data from %s: %s", url, string(body))
	return json.Unmarshal(body, target)
}

func newApp(logger log.Logger, gs *grpc.Server, hs *kratosHttp.Server, studentRepo biz.StudentRepo) *kratos.App {
	app := kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)

	ctx := context.Background()
	go studentRepo.ConsumeStudentCreateMsg(ctx)
	go studentRepo.ConsumeStudentDeleteMsg(ctx)
	go studentRepo.ConsumeStudentUpdateMsg(ctx)
	hs.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Missing id parameter", http.StatusBadRequest)
			return
		}
		serviceBUrl := "http://localhost:8081/serviceB?id=" + id
		serviceCUrl := "http://localhost:8082/serviceC?id=" + id
		var bData Message
		var cData Score

		//err := fetchData(serviceBUrl, &bData)
		//if err != nil {
		//	http.Error(w, "Failed to fetch data from service B", http.StatusInternalServerError)
		//	return
		//}
		//
		//err = fetchData(serviceCUrl, &cData)
		//if err != nil {
		//	http.Error(w, "Failed to fetch data from service C", http.StatusInternalServerError)
		//	return
		//}

		// 使用通道（channels）来处理并发请求
		errChan := make(chan error, 2)

		// 并发请求 serviceB
		go func() {
			errChan <- fetchData(serviceBUrl, &bData)
		}()

		// 并发请求 serviceC
		go func() {
			errChan <- fetchData(serviceCUrl, &cData)
		}()

		// 等待两个请求完成
		for i := 0; i < 2; i++ {
			if err := <-errChan; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		combined := map[string]interface{}{
			"name":   bData.Name,
			"info":   bData.Info,
			"status": bData.Status,
			"score":  cData.Score,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(combined)
	})
	return app
}

func main() {

	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	// 初始化数据库连接
	db, cleanup, err := data.NewData(logger, bc.Data)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	go serviceB(db)
	go serviceC(db)

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
