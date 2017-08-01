## wsps - websocket pub-sub
[![Build Status](https://travis-ci.org/cristaloleg/wsps.svg?branch=master)](https://travis-ci.org/cristaloleg/wsps)
[![License](https://img.shields.io/github/license/mashape/apistatus.svg)]()
[![Go Report Card](https://goreportcard.com/badge/github.com/cristaloleg/wsps?style=flat-square)](https://goreportcard.com/report/github.com/cristaloleg/wsps)
[![codecov](https://codecov.io/gh/cristaloleg/wsps/branch/master/graph/badge.svg)](https://codecov.io/gh/cristaloleg/wsps)


### How to start

Install RabbitMQ:
```
brew install rabbitmq
```

Update PATH after installation
```
PATH=$PATH:/usr/local/sbin
```

Start RabbitMQ:
```
rabbitmq-server
```

Web view:
```
http://localhost:15672/#/
```

To start a Pub:
```
cd pub/cmd
go build main.go
./main
```

To start a Sub:
```
cd sub/cmd
go build main.go
./main
```

To start a Demp:
```
cd demo
go build main.go
./main
```
