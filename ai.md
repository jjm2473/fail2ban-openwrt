### 原始日志

```
root@iStoreOS:~# logread -e dropbear                                          
Mon Sep 15 14:54:36 2025 authpriv.info dropbear[30519]: Child connection from 192.168.9.204:55200                                                           
Mon Sep 15 14:54:39 2025 authpriv.notice dropbear[30519]: Password auth succee
ded for 'root' from 192.168.9.204:55200 
root@iStoreOS:~# 

root@iStoreOS:~# logread -e dropbear
Mon Sep 15 14:54:36 2025 authpriv.info dropbear[30519]: Child connection from 192.168.9.204:55200                                                           
Mon Sep 15 14:54:39 2025 authpriv.notice dropbear[30519]: Password auth succeeded for 'root' from 192.168.9.204:55200                                       
Mon Sep 15 14:55:07 2025 authpriv.info dropbear[30601]: Child connection from 
192.168.9.204:55374                                                           
Mon Sep 15 14:55:11 2025 authpriv.warn dropbear[30601]: Bad password attempt f
or 'root' from 192.168.9.204:55374
```

### 解析日志

```
实现日志的跟踪，使用 logread -f 来持续跟踪日志。得到日志产生的时间，以及文本。为之后的 dropbear 以及 uhttpd 的日志分析提供基础。
命令在 OpenWRT 下面的日志效果如下：

logread -f
Mon Sep 15 15:57:09 2025 authpriv.info dropbear[6862]: Child connection from 192.168.9.204:59200
Mon Sep 15 15:57:10 2025 authpriv.warn dropbear[6862]: Bad password attempt for 'root' from 192.168.9.204:59200
Mon Sep 15 15:57:14 2025 authpriv.warn dropbear[6862]: Bad password attempt for 'root' from 192.168.9.204:59200
Mon Sep 15 15:57:15 2025 authpriv.warn dropbear[6862]: Bad password attempt for 'root' from 192.168.9.204:59200
Mon Sep 15 15:57:16 2025 authpriv.info dropbear[6862]: Exit before auth from <192.168.9.204:59200>: (user 'root', 3 fails): Max auth tries reached - user 'root'
```


### 判断 dropbear 日志

```
实现一个函数，判断从 #file:logread.go 读到的一行一行的日志，是否为 dropbear 的 SSH 密码错误日志。具体的登录错误日志如下：

authpriv.warn dropbear[8843]: Bad password attempt for 'root' from 192.168.9.204:63779

如果是 dropbear 的登录错误日志，则拿到登录错误的时间，以及对应 IP。为下一步是否封禁这个 IP 做准备。

authpriv.info dropbear[10089]: Exit before auth from <192.168.9.204:50339>: (user 'root', 3 fails): Max auth tries reached - user 'root'
这种情况下也要匹配到

做两个正则匹配，不够优雅，是否能跟规律合并为一个。

```

### 判断 uhttpd 日志

```
实现一个函数，判断从 #file:logread.go 读到的一行一行的日志，是否为 openwrt uhttpd 的 HTTP 密码错误日志。具体的登录错误日志如下：

daemon.err uhttpd[11561]: [info] luci: failed login on / for root from 192.168.9.204

如果是 uhttpd 的登录错误日志，则拿到登录错误的时间，以及对应 IP。为下一步是否封禁这个 IP 做准备。

正则表达式里面也可以去掉时间的正则

```

### 高性能禁用IP访问

```
我们的这个程序运行在 Linux 当中，我们把 ip 添加到某个 ipset 的规则，而这个 ipset 的规则，会加入到防火墙的禁止访问列表中，从而达到无法访问 SSH 或者 uhttpd 的目的。
请帮实现一个函数，可以把一个 IP 加入禁用的 ipset 集合中。
```


### 一些 nft 命令参考

```

nft list table inet fw4|grep fail2ban
nft add set inet fw4 fail2banop { type ipv4_addr \; }
nft add element inet fw4 fail2banop { 192.0.2.1, 192.0.2.3 }
nft delete element inet fw4 fail2banop { 192.0.2.1 }
nft flush set inet fw4 fail2banop
nft list set inet fw4 fail2banop

```