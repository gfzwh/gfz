# gfz
一个以及protobuf的异步服务框架

# server的使用
```
svr := server.NewServer("./conf/gfz.xml")
defer svr.Release()

ctl := controller.Controller()
protocol.RegisterUserServiceHandler(svr, ctl)
svr.Run(server.Bind("127.0.0.1"), server.Port(8989))
```

# client的使用
```
// 先初始化客户端请求的信息
// 同时创建连接池
client.Pools().Registry(registry.NewRegistry(
		registry.Url("http://127.0.0.1:7171"),
		registry.Zone("guangzhou"),
		registry.Env("dev-0.0.1"),
		registry.Host("fgz-discovery")))


// 通过服务名调用对应服务的接口
userService := protocol.NewUserService("gfz-test")
resp, err := userService.CreateUser(context.TODO(), &protocol.CreateUserReq{
    Auth:      &protocol.Authorize{Appid: "", Appkey: ""},
    Username:  fmt.Sprintf("%s-%d", "shockerjue", i),
    Telephone: "1234567890",
    Email:     "12345@123.com",
}, client.Timeout(3))
if nil != err {
    zzlog.Errorf("CreateUser return error, err:%v\n", err)
} else {
    zzlog.Infof("Create return %v\n", resp)
}
```
