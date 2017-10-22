ssh-proxy
=========

An ssh proxy dialer in Go.

Usage
-----

Use an ssh proxy to connect to a grpc service with something like this:

```
proxy := NewSSHProxy(host)
grpcOpt := []grpc.DialOption{
	grpc.WithInsecure(),
	grpc.WithDialer(proxy.Dial),
}
helloConn, err := grpc.Dial("10.132.93.129:8080", grpcOpt...)
if err != nil {
	panic(err)
}
defer helloConn.Close()

helloClient := hello.NewHelloClient(helloConn)

retort, err := helloClient.Greet(context.Background(), &hello.Greeting{Text: "What's up, doc?"})
if err != nil {
	panic(err)
}

fmt.Printf("Received this retort: %q\n", retort.GetText())
```
