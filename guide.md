# Guide

#### method
|func name|param|desc|
|---|---|---|
|Func|string rangoFunc|Api Func|

#### outter Type
|type name|struct|desc|
|---|---|---|
|router|[]matchers|router subitem|

#### innner Type
|type name|struct|desc|
|---|---|---|
|router|[]matchers|router subitem|

# Static server

# Middleware
#### auth

#### session (NO RECOMMENDED)

# Route.Matcher
#### throttle

#### auth

# Router.Func
#### HATEOAS

# Router.HandlerFunc

# Confg

# Work Flow (RECOMMENDED)
> 对于www文件夹其实是非常没必要的<br>
> 在这里rango的定位是作为一个api服务器开发工具，理论上不应该传递过多的静态文件

> 一个比较推崇的方法是增加一个`nodejs`中端服务器，处理web静态文件访问流量

dev directory
```
~/app
  ├── api
  │   └── v1
  │       ├── db
  │       |   ├── main.go
  │       |   ├── model0.go
  │       |   └── model1.go
  │       ├── handlers.go
  |       ├── funcs.go
  │       ├── work.go
  │       └── main.go
# ├── www
# |   ├── static
# |   |    ├── css
# |   |    ├── img
# |   |    └── js
# |   └── index.html
  |── storage
  |   ├── cache
  |   └── logs
  ├── apis.go
  ├── main.go
  ├── build.go
  ├── entrypoint.sh
  └── config.json
```

- main.go 启动文件，配置基本信息如运行环境编译环境静态文件映射等
- apis.go 定义二级路由如将`/v1`映射到`~/app/api/v1`中
- /confg.json 配置文件
- /entrypoint.sh docker入口
- /build.sh 编译部署

> 文件夹

- /www 静态文件
- /storage 生成文件 若使用sqlit db文件也应该放在此处
- /api api文件，内部根据版本号进行划分，也可无版本号

> 显而易见，除了api目录下的代码，其余的部分应该在初始化之后就固定

> 并且很不同的一点是`db`代码将与`api`版本所对应，不同的版本目录放置不同的`db`代码，即使两个版本只是没有改变`db模型`

prod directory
#### api server
```
~/sev
├── docker-compose.yml
|── storage
|   ├── cache
|   └── logs
|── config.json
└── app
    ├── Dockerfile
    ├── sev
    └── entrypoint.sh
```

docker-compose
```yml
version: '3'

services:
    mongodb:
      image: mongodb:latest
      volumes:
        - home/volumes/mongo:/data/db
      networks:
        - apiSev
    
    rango:
      build: ./app
      volumes:
        - ~/storage:./storage
        - ~/www:./www
        - ~/conf.json:./conf.json
      networks:
        - apiSev
    
    networks:
      apiSev:
```

#### Api+中端
```
~/sev
├── docker-compose.yml
├── web   
|   ├── Dockerfile
|   ├── entrypoint.sh
|   ├── app.js
|   └── www
|       ├── vendors
|       ├── css
|       ├── img
|       └── js
|       └── index.html
└── app1
    ├── Dockerfile
    |── storage
    |   ├── cache
    |   └── logs
    ├── sev
    ├── entrypoint.sh
    └── config.json
```
