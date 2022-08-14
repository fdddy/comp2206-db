FROM golang

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/runserver ./cmd/runserver/runserver.go

EXPOSE 15184
CMD ["runserver", "-c", "./exampleServerConfig.json"]
