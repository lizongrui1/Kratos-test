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
	"student/internal/service"
	"sync"

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

var client = &http.Client{}

func fetchData(url string, target interface{}) error {
	resp, err := client.Get(url)
	if err != nil {
		log.Errorf("Error making GET request to %s: %v", url, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("record not found")
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
		var bData []service.Message
		var cData service.Score

		//err := fetchData(serviceBUrl, &bData)
		//if err != nil {
		//	http.Error(w, "Failed to fetch data from service B", http.StatusInternalServerError)
		//	return
		//}
		//err = fetchData(serviceCUrl, &cData)
		//if err != nil {
		//	http.Error(w, "Failed to fetch data from service C", http.StatusInternalServerError)
		//	return
		//}

		var wg sync.WaitGroup
		errorChan := make(chan error, 2)
		wg.Add(2)
		go func() {
			defer wg.Done()
			errorChan <- fetchData(serviceBUrl, &bData)
		}()
		go func() {
			defer wg.Done()
			errorChan <- fetchData(serviceCUrl, &cData)
		}()

		//go func() {
		//	wg.Wait()
		//	close(errorChan)
		//}()
		wg.Wait()
		for err := range errorChan {
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		close(errorChan)
		if len(bData) == 0 {
			http.Error(w, "No data returned from service B", http.StatusNotFound)
			return
		}
		firstMessage := bData[0]

		combined := map[string]interface{}{
			"id":     firstMessage.ID,
			"name":   firstMessage.Name,
			"info":   firstMessage.Info,
			"status": firstMessage.Status,
			"score":  cData.Score,
		}

		log.NewHelper(logger).Info(fmt.Sprintf(
			"Query successful: ID: %d, Name: %s, Info: %s, Status: %s, Score: %d",
			firstMessage.ID,
			firstMessage.Name,
			firstMessage.Info,
			firstMessage.Status,
			cData.Score,
		))

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

	db, cleanup, err := data.NewData(logger, bc.Data)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	go service.ServiceB(db)
	go service.ServiceC(db)

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
