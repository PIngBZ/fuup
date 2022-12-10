# fuup

home: 有公网IP

office：内网

home tcp(socks5) <-----kcp on udp----> office tcp(socks5)

两端均通过socks5协议访问对方(可以通过配置文件禁用某一个方向)

可以通过redsocks和iptable规则实现把office所在网络和home所在网络连到一起，透明互相访问跟访问本地网络一样


可以配置一个ip更新服务器，home的公网IP变化时候注册过去，office端会定期请求，发现IP改变重新连接

