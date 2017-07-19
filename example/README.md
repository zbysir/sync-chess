# 四川麻将

> 源代码量900行左右. 主要代码量集中在玩家出牌逻辑处理, 其实这都是些低逻辑代码, 实际你要写的代码更少.


## 依赖

### 网络库
使用ws协议通讯, 一个请求一个协程对应一个handle, 清晰明了
github.com/bysir-zl/hubs

## 期望架构
```
// ->>> 代表可有多个服务
client -> agent -> db -(find chess server and conn)>>> chess server ->>> log server
```
