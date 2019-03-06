# cqbot
[![Build Status](https://travis-ci.com/Bpazy/cqbot.svg?branch=master)](https://travis-ci.com/Bpazy/cqbot)
[![Docker pulls](https://img.shields.io/docker/pulls/hanziyuan08/cqbot.svg)](https://hub.docker.com/r/hanziyuan08/cqbot)

[coolq-http-api](https://github.com/richardchien/coolq-http-api) is necessary.

## Usage
```
$ cqbot --help
```

## Docker usage(recommend)
1. Get docker image
```shell
$ sudo docker pull coolq/wine-coolq
$ sudo docker pull hanziyuan08/cqbot
```
2. Create docker bridge network
```
$ sudo docker network create -d bridge coolq-net
```
3. Set up your configuration.  
Make sure you have installed `io.github.richardchien.coolqhttpapi.cpk` under `app`.  
Make sure you have added `coolqhttpapi general.json` under `data\app\io.github.richardchien.coolqhttpapi\config`.
```json
// general.json
{
    "log_level": "debug",
    "show_log_console": true,
    "use_http": true,
    "post_url": "http://cqbot:12345"
}
```
4. Start wine-coolq.
```shell
$ sudo docker run --name=coolq --rm -d -p 9000:9000 --network=coolq-net -v /home/han/coolq-data:/home/user/coolq -e VNC_PASSWD=your_vnc_password -e COOLQ_ACCOUNT=your_qq_number coolq/wine-coolq
```
5. Start Redis.
```shell
$ sudo docker run --name redis --rm -d --network coolq-net redis
```
6. Start cqbot.
```shell
$ sudo docker run --name cqbot --rm -d --network coolq-net hanziyuan08/cqbot -dns "user:password@tcp(localhost)/db_name" -redis "redis:6379"
```
