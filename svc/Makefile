.PHONY: all
all:svc.pb.go

svc.pb.go: svc.proto
	protoc --go_out=. --go-grpc_out=. svc.proto --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative

.PHONY: clean
clean: end
	$(RM) svc.pb.go svc_grpc.pb.go

.PHONY: end
end:
