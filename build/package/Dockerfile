FROM golang:1.17-alpine as builder
ENV GOPATH=
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /main main.go

FROM alpine:3
COPY --from=builder main /bin/main
COPY *.ttf ./
ENTRYPOINT ["/bin/main"]
