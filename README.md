# rango
minimalist Go http server framework

> It's like [Echo](https://github.com/labstack/echo), but it's sweeter ~~(fake)~~

# Overview
总是会有些小想法想动手玩玩，别的库虽然是好又是高性能又是有社区，但是始终有点不适，于是写这个解决一系列的小问题

> 值得注意的是，在默认的函数处理中，将认为所有的body中都是json数据<br>
> 当然也对其他特殊格式进行支持，比如文件上传等操作

# Hello world
```golang
package main

import (
  "net/http"
  "github.com/zhzLuke96/Rango/rango"
)

func main() {
  sev := rango.New("hello")
  sev.Func("/", hello)
  sev.Start(":8080")
}

func hello(vars rango.ReqVars) interface{} {
  return "hello " + vars.GetDefault("name", "world") + " !"
}
```

GET:
```
$> curl 127.0.0.1:8080/?name=luke96
{...,data:"hello luke96 !",...}
```

POST:
```
$> curl -H "Content-Type:application/json" -X POST --data '{"name": "luke96"}' 127.0.0.1:8080/
{...,data:"hello luke96 !",...}
```

# Example
project in `example` folder list all `rango.functions` and common usage to help users get started quickly.

# Changelog
- 添加responser类，根据信息生成消息code
- 增加upload支持，通过sev.Upload创建上传服务器
- 修改ErrCatch中间件统一回复格式
- 修改Router中间件行为统一回复格式
- 修改rango.Func返回值为interface{}
- 增加非正则path matcher，自动识别，对于不符合的match设置将构建Regexp对象
- 修改HATEOAS为GitHub风格
- 修复static没有StripPrefix的bug

# Todo
- [x] updata .08h
- [ ] finish guide.md
- [ ] file upload handler
- [ ] BLOB stream
- [ ] RPC function
- [ ] add more test.go
- [ ] add more comment
- [ ] example on docker
- [ ] ...

# LICENSE
GPL-3.0