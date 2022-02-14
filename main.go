package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/amdf/imgtengwar/internal/render"
	pb "github.com/amdf/imgtengwar/svc"
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
	SVCRPCADDR  = "localhost:50051"
	SVCHTTPADDR = "localhost:8081" //"0.0.0.0:80"
)

type TengwarConverterServer struct {
	pb.UnimplementedTengwarConverterServer
	clientLimits map[string]*rate.Limiter
}

func (srv TengwarConverterServer) ConvertText(context.Context, *pb.SimpleConvertRequest) (*pb.ConvertResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConvertText not implemented")
}
func (srv TengwarConverterServer) MakeImage(context.Context, *pb.ConvertRequest) (*httpbody.HttpBody, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MakeImage not implemented")
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

func runGateway() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := pb.RegisterTengwarConverterHandlerFromEndpoint(ctx, mux, SVCRPCADDR, opts)
	if err != nil {
		log.Fatalln("fail to register http", err)
	}

	log.Println("starting http server at ", SVCHTTPADDR)

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	if err != http.ListenAndServe(SVCHTTPADDR, mux) {
		log.Fatalln("fail to serve http", err)
	}
}

func MakeTengwarConverterServer() (srv *TengwarConverterServer, err error) {
	srv = &TengwarConverterServer{}
	srv.clientLimits = make(map[string]*rate.Limiter)
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

	go runGateway()

	err = s.Serve(lis)
	if err != nil {
		log.Fatalln("fail to server", err)
	}

}
