package pkg

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-contrib/multitemplate"
	"github.com/unrolled/secure"

	"github.com/chenjiandongx/ginprom"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kr/pty"
	"github.com/prometheus/client_golang/prometheus/promhttp" // https://blog.csdn.net/u014029783/article/details/80001251 教程
	log "github.com/sirupsen/logrus"
)

var xterm *XtermJs

func init() {
	// 设置将日志输出到标准输出（默认的输出为stderr，标准错误）
	// 日志消息输出可以是任意的io.writer类型
	log.SetOutput(os.Stdout)

	log.SetReportCaller(false)
}

var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// func ServeGin(data.Host,data.Port, username, password, crtpath, keypath string, cmds []string, isdebug, isReconnect, isPermitWrite, isAudit, isXsrf, isProf, enabletls bool, MaxConnections int64) {
func ServeGin(data *Tty) {
	if data.IsDebug {
		// 设置日志级别为warn以上
		log.SetLevel(log.DebugLevel)
		gin.SetMode(gin.DebugMode)
	} else {
		// 设置日志级别为warn以上
		log.SetLevel(log.InfoLevel)
		gin.SetMode(gin.ReleaseMode)
	}

	if data.IsAudit {
		InitSqlite()
		Engine.Sync2(new(Aduit))
		Engine.Sync2(new(Whos))
	}

	router := gin.Default()

	// 使用 Recovery 中间件
	router.Use(gin.Recovery())

	if data.EnableTLS {
		router.Use(TlsHandler(data.Host, data.Port))
	}

	// 判断cmds输入，为空默认设置为bash
	if len(data.Cmds) == 0 {
		data.Cmds = append(data.Cmds, "bash")
	}

	// 初始化XtermJs全局属性配置
	connections := int64(0)
	xterm = &XtermJs{
		Options: Options{
			PermitWrite:    data.IsPermitWrite,
			CloseSignal:    9,
			MaxConnections: data.MaxConnections,
			Audit:          data.IsAudit,
			Xsrf:           data.IsXsrf,
			EnableTLS:      data.EnableTLS,
			CrtPath:        data.CrtPath,
			KeyPath:        data.KeyPath,
		},
		Title:       "Showme",
		Connections: &connections,
		Server:      NewServer(),
		XsrfToken:   sync.Map{},
	}

	// 静态二进制文件
	fs := assetfs.AssetFS{
		Asset:    Asset,
		AssetDir: AssetDir,
	}

	router.StaticFS("/static", &fs)
	// 静态文件
	// router.StaticFS("/static", http.Dir("./tty/static"))

	var apiGroup *gin.RouterGroup
	// 是否密码登录
	if data.Username == "" && data.Password == "" {
		apiGroup = router.Group("/")
	} else {
		apiGroup = router.Group("/", gin.BasicAuth(gin.Accounts{data.Username: data.Password}))
	}

	// 添加prometheus监控
	// use prometheus metrics exporter middleware.
	//
	// ginprom.PromMiddleware() expects a ginprom.PromOpts{} poniter.
	// It was used for filtering labels with regex. `nil` will pass every requests.
	//
	// ginprom promethues-labels:
	//   `status`, `endpoint`, `method`
	//
	// for example:
	// 1). I want not to record the 404 status request. That's easy for it.
	// ginprom.PromMiddleware(&ginprom.PromOpts{ExcludeRegexStatus: "404"})
	//
	// 2). And I wish to ignore endpoint start with `/prefix`.
	// ginprom.PromMiddleware(&ginprom.PromOpts{ExcludeRegexEndpoint: "^/prefix"})
	router.Use(ginprom.PromMiddleware(nil))
	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	doMetrics()
	apiGroup.GET("/metrics", ginprom.PromHandler(promhttp.Handler()))

	// 添加html template
	// 主页
	// 从内存取出然后渲染加载
	indexhtml := multitemplate.New()
	xterm3, err := Asset("xterm3.html")
	if err != nil {
		log.WithField("tty.go", "198").Error(err.Error())
		return
	}

	t, err := template.New("index").Parse(string(xterm3))
	if err != nil {
		log.Error(err.Error())
		return
	}

	admin, err := Asset("admin.html")
	if err != nil {
		log.WithField("tty.go", "198").Error(err.Error())
		return
	}

	ta, err := template.New("admin").Parse(string(admin))
	if err != nil {
		log.Error(err.Error())
		return
	}

	indexhtml.Add("index", t)
	indexhtml.Add("admin", ta)
	router.HTMLRender = indexhtml

	// 添加审计查询接口
	if data.IsAudit {
		apiGroup.GET("/check", func(c *gin.Context) {
			defer func() {
				who := &Whos{
					Remoteaddr: c.Request.RemoteAddr,
					Path:       "/check",
				}
				AddWhos(who)
			}()
			name := c.DefaultQuery("name", "")
			data, err := GetAduit(name)
			if err != nil {
				c.String(http.StatusOK, err.Error())
			} else {
				c.JSONP(http.StatusOK, data)
			}
		})

		apiGroup.GET("/who", func(c *gin.Context) {
			defer func() {
				who := &Whos{
					Remoteaddr: c.Request.RemoteAddr,
					Path:       "/who",
				}
				AddWhos(who)
			}()
			name := c.DefaultQuery("name", "")
			data, err := GetWhos(name)
			if err != nil {
				c.String(http.StatusOK, err.Error())
			} else {
				c.JSONP(http.StatusOK, data)
			}
		})

		apiGroup.GET("/admin", func(c *gin.Context) {
			defer func() {
				who := &Whos{
					Remoteaddr: c.Request.RemoteAddr,
					Path:       "/admin",
				}
				AddWhos(who)
			}()
			c.HTML(http.StatusOK, "admin", gin.H{})
		})
	}

	// 后端websocket服务
	apiGroup.GET("/ws", func(c *gin.Context) {
		defer func() {
			if data.IsAudit {
				who := &Whos{
					Remoteaddr: c.Request.RemoteAddr,
					Path:       "/ws",
				}
				AddWhos(who)
			}
		}()
		conns := atomic.AddInt64(xterm.Connections, 1)
		connects.Set(float64(conns))
		if xterm.Options.MaxConnections != 0 {
			if conns > xterm.Options.MaxConnections {
				log.WithFields(log.Fields{
					"tty.go": "147",
				}).Printf("Max Connected: %d", xterm.Options.MaxConnections)
				atomic.AddInt64(xterm.Connections, -1)
				return
			}
		}
		// 升级get请求为webSocket协议
		ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer ws.Close()

		cmd := exec.Command(data.Cmds[0], data.Cmds[1:]...)
		//这里得到标准输出和标准错误输出的两个管道，此处获取了错误处理
		ptmx, err := pty.Start(cmd)
		if err != nil {
			log.Errorf("ptmx[52] %s", err.Error())
			return
		}

		if xterm.Options.MaxConnections != 0 {
			log.WithField("tty.go", "169").Printf("Command is running for client %s with PID %d (args=%q), connections: %d/%d",
				c.Request.RemoteAddr, cmd.Process.Pid, data.Cmds, conns, xterm.Options.MaxConnections)
		} else {
			log.WithField("tty.go", "172").Printf("Command is running for client %s with PID %d (args=%q), connections: %d",
				c.Request.RemoteAddr, cmd.Process.Pid, data.Cmds, conns)
		}

		xterm.Server.StartGo()

		context := &ClientContext{
			Xtermjs: xterm,
			Request: c.Request,
			WsConn:  ws,
			Cmd:     cmd,
			Pty:     ptmx,
			// Cache:      bytes.NewBuffer([]byte("")),
			// CacheMutex: &sync.Mutex{},
			WriteMutex: &sync.Mutex{},
		}

		context.HandleClient()
		xterm.Server.WaitGo()
	})

	apiGroup.GET("/", func(c *gin.Context) {
		defer func() {
			if data.IsAudit {
				who := &Whos{
					Remoteaddr: c.Request.RemoteAddr,
					Path:       "/",
				}
				AddWhos(who)
			}
		}()
		var protocol, httproto string
		if xterm.Options.EnableTLS && IsPathExists(xterm.Options.CrtPath) && IsPathExists(xterm.Options.KeyPath) {
			protocol = "wss"
			httproto = "https"
		} else {
			protocol = "ws"
			httproto = "http"
		}
		newXsrf := GetRandomSalt()
		log.WithField("tty.go", "212").Debugf("%s xsrftoken %s", c.Request.RemoteAddr, newXsrf)
		if !xterm.Options.Xsrf {
			xterm.XsrfToken.Store(fmt.Sprintf("%s%s", newXsrf, strings.Split(c.Request.RemoteAddr, ":")[0]), time.Now().String())
		}
		c.HTML(http.StatusOK, "index", gin.H{
			"host":      c.Request.RemoteAddr,
			"Reconnect": data.IsReconnect,
			"Debug":     data.IsDebug,
			"Write":     data.IsPermitWrite,
			"MaxC":      data.MaxConnections,
			"Conn":      *xterm.Connections + 1,
			"Cmd":       strings.Join(data.Cmds, " "),
			"Xsrf":      newXsrf,
			"Protocol":  protocol,
			"Httproto":  httproto,
			"isAduit":   data.IsAudit,
		})
	})

	// automatically add routers for net/http/pprof
	// e.g. /debug/pprof, /debug/pprof/heap, etc.
	if data.IsProf {
		ginpprof.Wrapper(router)
	}

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", data.Host, data.Port),
		Handler: router,
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		// log.Println("receive interrupt signal")
		// if err := server.Close(); err != nil {
		// 	log.Fatal("Server Close:", err)
		// }

		log.WithField("tty.go", "249").Println("Shutdown Server ...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.WithField("tty.go", "254").Fatal("Server Shutdown:", err)
		}
		log.WithField("tty.go", "256").Println("Server exiting")
	}()

	if data.Host == "0.0.0.0" {
		ips := GetIPs()
		for _, ip := range ips {
			log.WithField("tty.go", "261").Infof("Listening and serving HTTPS on %s:%s", ip, data.Port)
		}
	} else {
		log.WithField("tty.go", "261").Infof("Listening and serving HTTPS on %s:%s", data.Host, data.Port)
	}

	if xterm.Options.EnableTLS {
		if IsPathExists(xterm.Options.CrtPath) && IsPathExists(xterm.Options.KeyPath) {
			if err := server.ListenAndServeTLS(xterm.Options.CrtPath, xterm.Options.KeyPath); err != nil {
				if err == http.ErrServerClosed {
					log.WithField("tty.go", "266").Println("Server closed under request")
				} else {
					log.WithField("tty.go", "268").Fatal("Server closed unexpect", err.Error())
				}
			}
		} else {
			log.WithField("tty.go", "277").Error("EnableTLS is true,but crt or key path is not exists")
		}
	} else {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.WithField("tty.go", "266").Println("Server closed under request")
			} else {
				log.WithField("tty.go", "268").Fatal("Server closed unexpect", err.Error())
			}
		}
	}

	log.WithField("tty.go", "272").Println("Server exiting")
}

func TlsHandler(host, port string) gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     fmt.Sprintf("%s:%s", host, port),
		})
		err := secureMiddleware.Process(c.Writer, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			return
		}

		c.Next()
	}
}