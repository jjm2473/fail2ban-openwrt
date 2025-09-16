# fail2ban-openwrt
* A better fail2ban specifically designed for OpenWrt/iStoreOS. 
* Now only works in OpenWrt/iStoreOS 24.10 that using nft
* 一个专门为 OpenWrt/iStoreOS 设计的更好的 fail2ban
* 当前只能运行在使用 nft 的 OpenWrt/iStoreOS 系统上

## Build ipk for openwrt

https://github.com/linkease/openwrt-app-actions/tree/main/applications/fail2banop

## 优化说明

一个专门为 OpenWRT 开发，性能更好的 fail2ban.

当用户用 ssh 或者 luci 登录 OpenWRT 时，如果在一定的时间内连续发生多次密码错误，则自动把登录 IP 封禁一段时间。

* 考虑到如果 IP 众多，直接封禁很多 IP 效率低，所以用了 ipset 来优化。
* 考虑到如果路由器运行半年甚至一年时间，会导致 IP 封禁越来越多占用内存，所以被封禁的 IP 会在一段时间后自动释放。
* 考虑到如果短时间内被太多请求 IP 攻击，则如果存储的 ipset 数量达到一定程度，则会自动封禁最新的 IP，而自动释放最老的 IP，保证内存占用恒定。

## Optimization Notes

A fail2ban specifically designed for OpenWRT with better performance.  

When a user attempts to log in to OpenWRT via SSH or LuCI and enters incorrect passwords multiple times within a certain period, the login IP will be automatically banned for a set duration.  

* To optimize efficiency when dealing with a large number of IPs, ipset is utilized for blocking.  
* To prevent excessive memory usage due to accumulated banned IPs over extended periods (e.g., six months to a year), banned IPs are automatically released after a certain time.  
* To handle situations where a sudden surge of IP attacks occurs, the system automatically bans the newest IPs while releasing the oldest ones once the stored ipset reaches a predefined limit, ensuring constant memory usage.

## Usage

```
fail2banop --help
NAME:
   fail2ban-openwrt - Fail2ban for OpenWrt

USAGE:
   fail2ban-openwrt [global options] command [command options]

COMMANDS:
   version       Show the current version
   show-ipset    Show the ipset used by fail2ban-openwrt
   remove-ipset  Remove the ipset used by fail2ban-openwrt
   help, h       Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --window value        Login error window (seconds) (default: 600)
   --threshold value     Login error threshold (default: 10)
   --ban-duration value  Ban duration (minutes) (default: 1440)
   --show-banned-ips     Show currently banned IPs, for debugging (default: false)
   --help, -h            show help
```
