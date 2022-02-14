.PHONY: all
all: svc.swagger.json svc_grpc.pb.go svc.pb.go

svc.swagger.json: svc.proto
	protoc -I . --openapiv2_out . --openapiv2_opt logtostderr=true svc.proto

svc_grpc.pb.go: svc.proto
	protoc --grpc-gateway_out . --grpc-gateway_opt=logtostderr=true --grpc-gateway_opt=paths=source_relative --grpc-gateway_opt=generate_unbound_methods=true svc.proto

svc.pb.go: svc.proto
	protoc --go_out=. --go-grpc_out=. svc.proto --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative

.PHONY: clean
clean: end
	$(RM) svc.pb.go svc_grpc.pb.go svc.pb.gw.go
	$(RM) svc.swagger.json

.PHONY: end
end:
