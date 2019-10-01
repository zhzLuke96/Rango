# rango
![LICENSE badge](https://img.shields.io/badge/license-GPL3.0-blue)
![size badge](https://img.shields.io/badge/line-1.6K-green)

minimalist Go http server framework

> It's like [Echo](https://github.com/labstack/echo), but it's sweeter ~~(fake)~~


# Index
- [rango](#rango)
- [Index](#index)
- [Overview](#overview)
- [Install](#install)
- [Usage](#usage)
  - [Hello world](#hello-world)
- [Example](#example)
- [Middlewares](#middlewares)
  - [usage](#usage)
  - [list](#list)
- [Matchers](#matchers)
  - [usage](#usage-1)
  - [list](#list-1)
- [Compose](#compose)
- [Changelog](#changelog)
- [Todo](#todo)
- [LICENSE](#license)

# Overview
解决一些小问题的小玩具。

> 值得注意的是，在默认的函数处理中，将认为所有的body中都是json数据<br>
> 当然也对其他特殊格式进行支持，比如文件上传等操作

# Install
```
$ go get github.com/zhzLuke96/Rango
```

# Usage

## Hello world
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

# Middlewares
## usage
```golang
mid := NewMid(...)
sev.Use(mid)
```
## list
| name      | desc                 | example                         | effect                           |
| --------- | -------------------- | ------------------------------- | -------------------------------- |
| memCacher | Memory-based caching | `NewMemCacher(60).Mid` | 将body保存在缓存中 超时设置为60s |

# Matchers
## usage
```golang
mat := NewMat(...)
sev.Func("/", fn).AddMatcher(mat)
```
## list
| name     | desc   | example                              | effect                |
| -------- | ------ | ------------------------------------ | --------------------- |
| throttle | 限流器 | `newThrottle(500)` | 500ms内仅回复一个请求 |


# Compose

Matcher 常常是结合中间件使用的，可以搭配出更复杂的行为

> 例如 `throttle` 和 `cacher` 同时使用时，首先会判断是否被缓存，如果没缓存才调用接下来的serve，并穿过 `throttle` 决定是否响应。

```golang
func fn(vars rango.ReqVars)interface{}{...}

func main(){
  memCacher := middleware.NewMemCacher(10)
  throttle := matcher.newThrottle(500)

  sev := rango.NewSev()

  sev.Use(memCacher.Mid)
  sev.Func("/xxxapi", fn).AddMatcher()

  sev.Use(sev.Router.Mid)

  sev.Start(":8080")
}
```

# Changelog
- 修改middleware.cache，只缓存状态码小于400的情况
- 取消默认读取config行为，优化测试效率
- 添加简单的HTML功能
- 添加一些测试html，matcher，utils
- 增加了Bytes路由，可以将[]byte返回
- 修改了一些和测试有关的内容

# Todo
- [x] updata .08h
- [x] file upload handler
- [ ] finish guide.md
- [ ] BLOB stream
- [ ] RPC function
- [ ] add more test.go
- [ ] add more comment
- [ ] example on docker
- [ ] ...

# LICENSE
GPL-3.0