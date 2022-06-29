set -e

if [ ! -d "./build" ]; then
    mkdir ./build
fi

go build -o ./build/run-server ./cmd/runserver/runserver.go
if [ ! -f "./build/config.json" ]; then
    cp ./exampleServerConfig.json ./build/config.json
fi