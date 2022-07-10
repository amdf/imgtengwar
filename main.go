package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/amdf/imgtengwar/internal/render"
	pb "github.com/amdf/imgtengwar/svc"
	"github.com/amdf/rustengwar"
	"golang.org/x/time/rate"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/tap"

	otgrpc "github.com/opentracing-contrib/go-grpc"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jcfg "github.com/uber/jaeger-client-go/config"
)

const (
	SVCRPCADDR = "0.0.0.0:50051"
)

type TengwarConverterServer struct {
	pb.UnimplementedTengwarConverterServer
	clientLimits map[string]*rate.Limiter
	conv         rustengwar.Converter
}

func (srv TengwarConverterServer) ConvertText(ctx context.Context, req *pb.SimpleConvertRequest) (resp *pb.ConvertResponse, err error) {
	resp = &pb.ConvertResponse{}
	t := time.Now()
	defer func() { fmt.Println("MakeImage ", time.Since(t)) }()
	resp.ConvertedText, err = srv.conv.Convert(req.InputText)
	if err != nil {
		err = status.Errorf(codes.Internal, err.Error())
	}
	return
}

func (srv TengwarConverterServer) MakeImage(ctx context.Context, req *pb.ConvertRequest) (body *httpbody.HttpBody, err error) {

	var buf bytes.Buffer
	t := time.Now()
	defer func() { fmt.Println("MakeImage", time.Since(t)) }()

	var convText string
	convText, err = srv.conv.Convert(req.InputText)
	if err != nil {
		err = status.Errorf(codes.Internal, err.Error())
	}

	ss := strings.Split(convText, "\n") //TODO: do smarter split

	time.Sleep(50 * time.Millisecond)
	parent := opentracing.SpanFromContext(ctx)
	sp := opentracing.StartSpan("render", opentracing.ChildOf(parent.Context()))

	err = render.ToPNG(ss, req.FontFile, float64(req.FontSize), &buf)
	sp.Finish()

	if err != nil {
		err = status.Errorf(codes.Internal, err.Error())
		return
	}
	body = &httpbody.HttpBody{
		ContentType: "image/png",
		Data:        buf.Bytes(),
	}
	return
}

func (srv TengwarConverterServer) GetClientAddr(ctx context.Context) string {
	ipaddr := "(unknown)"
	p, ok := peer.FromContext(ctx)
	if ok {
		ipaddr = p.Addr.String()
	}
	return ipaddr
}

func (srv TengwarConverterServer) RateLimiter(ctx context.Context, info *tap.Info) (context.Context, error) {
	user := srv.GetClientAddr(ctx)
	if srv.clientLimits[user] == nil {
		//TODO: store limits in config
		srv.clientLimits[user] = rate.NewLimiter(10, 50) // QPS, burst
	}

	if !srv.clientLimits[user].Allow() {
		return nil, status.Errorf(codes.ResourceExhausted,
			"client exceeded rate limit")
	}
	return ctx, nil
}

func MakeTengwarConverterServer() (srv *TengwarConverterServer, err error) {
	srv = &TengwarConverterServer{}
	srv.clientLimits = make(map[string]*rate.Limiter)
	err = srv.conv.InitDefault()
	return
}

func main() {
	err := render.Init()
	if err != nil {
		log.Fatalln("fail to create render", err)
	}

	newjcfg := jcfg.Configuration{
		Sampler: &jcfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jcfg.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}
	tracer, closer, jerr := newjcfg.New(
		"imgtengwar",
		jcfg.Logger(jaeger.StdLogger),
	)
	if jerr != nil {
		log.Fatalln("fail to create tracing", jerr)
	}
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()

	server, err := MakeTengwarConverterServer()
	if err != nil {
		log.Fatalln("fail to create server", err)
	}

	lis, err := net.Listen("tcp", SVCRPCADDR)
	if err != nil {
		log.Fatalln("fail to listen", err)
	}

	s := grpc.NewServer(
		grpc.InTapHandle(server.RateLimiter),
		grpc.UnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(tracer)),
	)
	pb.RegisterTengwarConverterServer(s, server)

	log.Println("starting server at ", lis.Addr())

	err = s.Serve(lis)
	if err != nil {
		log.Fatalln("server failed", err)
	}

}
