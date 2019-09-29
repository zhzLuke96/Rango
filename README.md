# rango
![LICENSE badge](https://img.shields.io/badge/license-GPL3.0-blue)
![size badge](https://img.shields.io/badge/line-1.5K-green)

minimalist Go http server framework

> It's like [Echo](https://github.com/labstack/echo), but it's sweeter ~~(fake)~~


# Index
- [Overview](#Overview)
- [Install](#Install)
- [Usage](#Usage)
  - [Hello world](#Hello-world)
- [Recommend](#Recommend)
  - [code style](#code-style)
  - [path cover](#path-cover)
- [Example](#Example)
- [Changelog](#Changelog)
- [Todo](#Todo)
- [LICENSE](#LICENSE)

# Overview
总是会有些小想法想动手玩玩，别的库虽然是好又是高性能又是有社区，但是始终有点不适，于是写这个解决一系列的小问题

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

# Example
project in `example` folder list all `rango.functions` and common usage to help users get started quickly.

# Changelog
- 修改route对象，增加了pathtpl变量，用于router排序
- 添加router排序功能，解决路由被覆盖问题
- 添加RangoSev.Sort，将遍历分组排序所有router
- 添加RangoSev.IsSorted，判断路由是否全部有序

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