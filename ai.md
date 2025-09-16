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

### 完整实现 fail2ban

```
#file:fail2ban.go 完整实现本文件，功能如下：
通过 logread -f 跟踪日志，可以得到日志的时间信息，以及文本信息。参考文件 #file:logread.go 
如果在一定时间内，出现 dropbear 或者 uhttpd 登录错误的一定的时间内，则封禁这个 IP 一定的时间。
这几个参数分为为 LoginErrorWindow LoginErrorThreshold BanDuration
Start 函数就是启动监控，并分析时间，以及文本，如果是 IsDropbearBadPasswordLog 或者IsUhttpdLoginErrorLog 则调用 #file:block.go 里面的相关函数封禁这个 IP，以及记录这个 IP 的封禁时间，等一定时间内，再不封禁这个 IP 以保证内存不会无限
```

### 实现 logNewIP 函数

```
#file:fail2ban.go 实现 logNewIP 函数。具体实现功能如下：
#file:fail2ban.go 完整实现本文件，功能如下：
如何在一定时间内，一个IP被记录多次，则封禁这个 IP 一定的时间后，再不封禁这个 IP 以保证内存不会过多。
这几个参数分为为 LoginErrorWindow LoginErrorThreshold BanDuration
```

### 实现 main 函数

```
#file:fail2ban_main.go 实现 main 函数，调用 #file:fail2ban.go 对应的对象的函数。并且 block 使用 #file:block.go 里面的封禁IP 以及解除封禁IP 的函数。
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

### SSH force login

* ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@192.168.30.244
