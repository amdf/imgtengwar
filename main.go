package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/amdf/imgtengwar/internal/render"
	pb "github.com/amdf/imgtengwar/svc"
	"github.com/amdf/rustengwar"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/time/rate"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/tap"
)

const (
	SVCRPCADDR  = "0.0.0.0:50051"
	SVCHTTPADDR = "0.0.0.0:8081"
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

	err = render.ToPNG(ss, req.FontFile, float64(req.FontSize), &buf)
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

func (srv TengwarConverterServer) RunGateway() {
	//TODO: handle errors another way

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()

	conn, err := grpc.DialContext(
		ctx,
		SVCRPCADDR,
		//grpc.WithInsecure(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("fail to dial rpc", err)
	}
	err = pb.RegisterTengwarConverterHandler(ctx, mux, conn)
	if err != nil {
		log.Fatalln("fail to register http", err)
	}
	//TODO: move this to config
	allowedOrigins := []string{
		"http://localhost:8080",
		"http://127.0.0.1:8080"}

	log.Println("starting http server at ", SVCHTTPADDR)
	gatewayServer := &http.Server{
		Addr:    SVCHTTPADDR,
		Handler: cors(mux, allowedOrigins),
	}
	// Start HTTP server (and proxy calls to gRPC server endpoint)
	if err = gatewayServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalln("fail to serve http", err)
	}
}

func cors(h http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providedOrigin := r.Header.Get("Origin")
		matches := false
		for _, allowedOrigin := range allowedOrigins {
			if providedOrigin == allowedOrigin {
				matches = true
				break
			}
		}

		if matches {
			w.Header().Set("Access-Control-Allow-Origin", providedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, ResponseType, grpc-metadata-log-level")
		}
		if r.Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, r)
	})
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
	)
	pb.RegisterTengwarConverterServer(s, server)

	log.Println("starting server at ", lis.Addr())

	go server.RunGateway()

	err = s.Serve(lis)
	if err != nil {
		log.Fatalln("server failed", err)
	}

}
