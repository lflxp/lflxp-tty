Browser sharing terminal

TTY function module provides terminal tool based on Web, and realizes full duplex communication mechanism of input and output of remote terminal through websocket.


![](https://github.com/lflxp/lflxp-tty/blob/master/asset/tty.png)
![](https://github.com/lflxp/lflxp-tty/blob/master/asset/ttyadmin.png)

# Requirements

* go get -u github.com/jteeuwen/go-bindata/...
* go get -u github.com/elazarl/go-bindata-assetfs/...

# Install

```
git clone https://github.com/lflxp/lflxp-tty
cd lflxp-tty
make install
lflxp-tty -h
```

`For Coder Demo`

> cmd/main.go

```go
...
import "github.com/lflxp/lflxp-tty/pkg"

func main() {
	tty := pkg.Tty{
		EnableTLS:      enableTLS,
		CrtPath:        crtPath,
		KeyPath:        keyPath,
		IsProf:         isProf,
		IsXsrf:         isXsrf,
		IsAudit:        isAudit,
		IsPermitWrite:  isPermitWrite,
		MaxConnections: MaxConnections,
		IsReconnect:    isReconnect,
		IsDebug:        isDebug,
		Username:       username,
		Password:       password,
		Port:           port,
		Host:           host,
		Cmds:           flag.Args(),
	}

	err := tty.Execute()
	if err != nil {
		panic(err)
	}
}
```

# Running

`Format`

```bash
➜  lflxp-tty git:(master) ✗ ./lflxp-tty -h                                                                                       
Usage of ./lflxp-tty:
  -H string
        http bind host (default "0.0.0.0")
  -P string
        http bind port (default "8080")
  -a    Open audit or not
  -c string
        *.crt file path (default "./server.crt")
  -d    debug log mode
  -f    Whether to enable pprof performance analysis
  -graceful
        listen on open fd (after forking)
  -k string
        *.key file path (default "./server.key")
  -m int
        maximum connection
  -p string
        password
  -r    auto reconnect
  -socketorder string
        previous initialization order - used when more than one listener was started
  -t    Open HTTPS or not
  -u string
        username
  -w    Whether to turn on write mode
  -x    Whether to turn on xsrf, which is on by default
```

* Read Only Command
    > ./lflxp-tty top
* ReadWrite Mode
    > ./lflxp-tty -w bash
* Full Options
    > ./lflxp-tty  -w -a -t -c ./server.crt -k ./server.key -m 10 -u admin -p admin -f -d 'docker run --rm -it -p 6379:7379 redis'

# Function point

* Independent binary operation, no other dependence
* Provide security audit
    * Access statistics
    * Command statistics
* Provide secure HTTPS connection
* Provide xsrf Security Assistance
* Provide performance analysis
    * pprof
* Provide interface monitoring
    * Prometheus
* Provide reconnection mechanism
* Provide current limiting mechanism
* Cross platform support
* Provide background web visual management
    * View access records
    * View operation record
    * View monitoring records (Prometheus metrics)
    * Full screen display
* Provide basicauth mode
* Provide host: port binding mode
* Provide debug log mode
* Support multiple remote modes
    * Read / write mode
    * Bash mode (default)
    * Custom mode
        * TMUX
        * docker run --rm -it -p 6379:6379 redis
        * Top

# Technology stack

* xterm.js
* pty/tty
* backend
    * golang
    * websocket
    * Prometheus
    * gin
    * restful
* front-end
    * HTML
    * JavaScript
    * Vue
    * element-ui

## Environment settings

When using 'systemctl' for service deployment, showme will report 'term environment variable not set'. You need to specify the environment variable 'term = xterm-256color' in the service file`

```
root@8.8.8.8:/etc/systemd/system# cat showme.service 
[Unit]
Description=showme
After=syslog.target
After=network.target

[Service]
# Modify these two values and uncomment them if you have
# repos with lots of files and get an HTTP error 500 because
# of that
###
#LimitMEMLOCK=infinity
#LimitNOFILE=65535
Type=simple
User=root
Group=root
WorkingDirectory=/tls
ExecStart=/usr/bin/showme tty -P 9999 -w -a -t -f -m 10 -u $user -p $pwd -c /tls/server.crt -k /tls/server.key
# ExecReload=/bin/kill -s HUP $MAINPID
Restart=always
Environment=USER=root HOME=/root TERM=xterm-256color

[Install]
WantedBy=multi-user.target
```

## Generate TLS file

```bash
cd lflxp-tty
➜  lflxp-tty git:(master) ✗ make crt  
rm -f cmd/cmd
rm -f lflxp-tty
rm -rf tls
mkdir -p tls
openssl genrsa -out tls/server.key 2048
Generating RSA private key, 2048 bit long modulus (2 primes)
....................................................+++++
.....+++++
e is 65537 (0x010001)
openssl req -nodes -new -key tls/server.key -subj "/CN=www.lflxp.cn" -out tls/server.csr
Can not load /home/xp/.rnd into RNG
140503399961024:error:2406F079:random number generator:RAND_load_file:Cannot open file:../crypto/rand/randfile.c:88:Filename=/home/xp/.rnd
openssl x509 -req -sha256 -days 3650 -in tls/server.csr -signkey tls/server.key -out tls/server.crt
Signature ok
subject=CN = www.lflxp.cn
Getting Private key
➜  lflxp-tty git:(master) ✗ ls -l tls/
total 12
-rw-rw-r-- 1 xp xp 1001 3月  23 17:11 server.crt
-rw-rw-r-- 1 xp xp  895 3月  23 17:11 server.csr
-rw------- 1 xp xp 1675 3月  23 17:11 server.key
```

# TODO

1. 自定义初始化命令通过http参数传递
2. 提供gin func插件