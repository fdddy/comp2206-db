#!/usr/bin/env julia
const build_dir = "./build"
const bin_name = "run-server"
const src_path = "./cmd/runserver/runserver.go"
const example_config_path = "./exampleServerConfig.json"
const config_name = "config.json"
const config_path = joinpath(build_dir, config_name)

if isfile(build_dir)
    println("$build_dir is a file, not a dir")
    exit(1)
end
if isdir(config_path)
    println("$config_path is a dir not a file")
    exit(1)
end

if !isdir(build_dir)
    mkdir(build_dir)
end

if !success(`go build -o $build_dir/$bin_name $src_path`)
    println("go build failed")
    exit(1)
end

if !isfile(config_path)
    cp(example_config_path, config_path)
end
