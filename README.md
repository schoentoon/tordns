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
  tordns {
    controlsocket /var/lib/tor/control
  }
}
```

In this case it'll connect with tor through the control socket at `/var/lib/tor/control`. From there on it'll resolve the actual dns queries using the builtin tor resolver. Do keep in mind that due to the way this resolver works we can only do A and AAAA queries, so you may want to add another resolver to catch other types.