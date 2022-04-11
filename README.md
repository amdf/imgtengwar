##Конвертер кириллицы в тенгвар
 
Build:
 `docker build -t amdf/imgtengwar -f build/package/Dockerfile .`
Run:
 `docker run --rm -d -p 3333:8081 -p 50051:50051 --name srv_imgtengwar amdf/imgtengwar`

 * :3333 - web gate
 * :50051 - gRPC