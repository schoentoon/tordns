# Make your dns requests over tor

In order to use this plugin you'll have to clone [coredns](https://github.com/coredns/coredns) and modify the plugin.cfg file to include the following line:
`remotehosts:github.com/schoentoon/tordns`
After this you can just build coredns the way you usually build it which is simply calling `make`. After this confirm that the plugin was build correctly into coredns using the following command.
```bash
$ ./coredns -plugins | grep tordns
  dns.tordns
```

# Configuration

Now to actually configure the plugin have a look at the following Corefile example
```
. {
  tordns dns4torpnlfs2ifuz2s2yf3fc7rdmsbhm6rw75euj35pac6ap25zgqad.onion:53 {
    proxy 127.0.0.1:9050
    retries 3
    poolsize 10
  }
}
```

In this case it'll use the dns server behind the hidden service on port 53. The `proxy` setting is used to specify where the tor socks5 proxy is located.
The `retries` option configures how often it'll retry a request (with a 5 second time window) whenever it detects a dead connection in the connection pool.
And lastly the `poolsize` configures how many connections it can have open at maximum.