# Guide

- [Guide](#guide)
- [quick table](#quick-table)
      - [method](#method)
      - [outter Type](#outter-type)
      - [innner Type](#innner-type)
- [Static server](#static-server)
- [Middleware](#middleware)
      - [auth](#auth)
      - [session (NO RECOMMENDED)](#session-no-recommended)
- [Route.Matcher](#routematcher)
      - [throttle](#throttle)
      - [auth](#auth-1)
- [Router.Func](#routerfunc)
      - [HATEOAS](#hateoas)
- [Router.HandlerFunc](#routerhandlerfunc)
- [Confg](#confg)
- [Recommend](#recommend)
  - [code style](#code-style)
  - [path cover](#path-cover)
- [Work Flow (RECOMMENDED)](#work-flow-recommended)
      - [api server](#api-server)
      - [Api+中端](#api%e4%b8%ad%e7%ab%af)

# quick table
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


# Recommend

## code style
```golang
_, uploadHandler := sev.Upload("/upload", "./imgs", 10*1024, []string{"image"})
uploadHandler.Failed(func(code int, err error, msg string, w http.ResponseWriter) {
  rango.DefaultFailed(code, err, msg, w)
  fmt.Printf("[LOG] code, msg = %v, %v\n", code, msg)
}).After(func(fileBytes []byte, pth string) (error, interface{}) {
  err := rango.SaveFile(fileBytes, pth)
  _, filename := filepath.Split(pth)
  return err, map[string]string{
    "url": "/image/" + filename,
  }
})
```

上面这个代码来自`example/main.go`中，是可以正常运行，且通过`go-lint`的，但是显而易见的，看上去一团乱麻。所以虽然支持这么写，但是最好还是分开，golang不适合这种风格。

> 并且带来了一个最麻烦的问题，根本没有写注释的位置，这对于维护是很致命的

```golang
// upload handler setup
_, uploadHandler := sev.Upload("/upload", "./imgs", 10*1024, []string{"image"})
uploadHandler.Failed(imageUploadFailed)
uploadHandler.After(imageUploadAfter)
```

## path cover

snippet 1

```golang
sev.Static("/", "./www")
sev.Static("/image", "./imgs")
```

snippet 2

```golang
sev.Static("/image", "./imgs")
sev.Static("/", "./www")
```

按正常的逻辑，这乱段代码应该是相同的效果，然而，`snippet 1`其实是没法正常工作的。

> 原因其实在于其内部实现，router内部是以表的形式实现的，在做router匹配的时候将会根据定义顺序遍历，可想而知，`"/"`表示为一个文件夹则，其将匹配所有的URL

其实有很多方法可以避免这个问题，随意定义路由，比如pathMatcher定义在一个`jumpTable`或者`redBlackTree`上。更简单的，在路由映射之前进行简单的预处理，将routes根据path长短进行降序排序。

在rango的实现里选择了后者，提供了排序的方法，将`snippet 1`修改为如下则就可以正常使用。且会递归调用所有子组的Sort。

```golang
sev.Static("/", "./www")
sev.Static("/image", "./imgs")

sev.Sort()
```

> 如果你想实现动态修改路由的行为，则需要在每次modify之后都排序一次

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
