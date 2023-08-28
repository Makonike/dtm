package main

import (
	"context"
	"github.com/asoul-fanclub/jaeger-middleware/middleware"
	"github.com/dtm-labs/dtm/busi"
	"github.com/dtm-labs/dtm/client/dtmgrpc"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/dtm-labs/dtmdriver"
	"github.com/dtm-labs/logger"
	"go.opentelemetry.io/otel"
	"testing"
	"time"

	"github.com/lithammer/shortuuid/v3"
)

func TestSagaGrpc(t *testing.T) {
	tp, _ := middleware.TracerProvider("http://localhost:14268/api/traces", false)
	otel.SetTracerProvider(tp)
	s := busi.GrpcNewServer()
	busi.GrpcStartup(s)
	logger.Infof("grpc simple transaction begin")
	gid := shortuuid.New()
	req := &busi.BusiReq{Amount: 30}
	dtmdriver.Middlewares.Grpc = append(dtmdriver.Middlewares.Grpc, middleware.NewJaegerClientMiddleware().UnaryClientInterceptor)
	ctx := context.WithValue(context.Background(), "trace-id", "867537eb51622b46d99652c618c1ed56")
	// req := &busi.BusiReq{Amount: 30, TransInResult: "FAILURE"}
	saga := dtmgrpc.NewSagaGrpcWithContext(ctx, dtmutil.DefaultGrpcServer, gid).
		Add(busi.BusiGrpc+"/busi.Busi/TransOut", busi.BusiGrpc+"/busi.Busi/TransOutRevert", req).
		Add(busi.BusiGrpc+"/busi.Busi/TransIn", busi.BusiGrpc+"/busi.Busi/TransInRevert", req)
	err := saga.Submit()
	if err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Second)
}
