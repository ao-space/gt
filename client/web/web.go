package web

import (
	"context"
	"crypto/tls"
	"embed"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/client/web/api"
	"github.com/isrc-cas/gt/logger"
	"github.com/isrc-cas/gt/predef"
	"github.com/isrc-cas/gt/util"
	"github.com/isrc-cas/gt/web/server"
	"github.com/isrc-cas/gt/web/server/middleware"
	"github.com/isrc-cas/gt/web/server/model/request"
	webUtil "github.com/isrc-cas/gt/web/server/util"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var FS embed.FS

type Server struct {
	server       *http.Server
	logger       logger.Logger // have no right to close logger
	tokenManager *server.TokenManager
	enableTLS    bool
}

func NewWebServer(c *client.Client) (*Server, error) {

	err := checkConfig(c)
	if err != nil {
		return nil, err
	}

	tm := server.NewTokenManager(server.DefaultTokenManagerConfig())

	fullPath := filepath.Join(util.GetAppDir(), "web_client.log")
	f, _ := os.Create(fullPath)
	gin.DefaultWriter = io.MultiWriter(f)

	r := gin.Default()
	err = setRoutes(c, tm, r)
	if err != nil {
		return nil, err
	}

	ws, err := getServer(c, tm, r)
	if err != nil {
		return nil, err
	}

	serverErr := make(chan error, 1)
	go ws.start(serverErr)
	timeout := time.After(2 * time.Second)

	select {
	case err = <-serverErr:
	case <-timeout:
		enableHTTPS := ws.enableTLS
		validURL, _ := webUtil.SwitchToValidWebURL(ws.server.Addr, enableHTTPS)
		if isFirstStart(c) {
			tempKey, err := getTempKey(ws, c)
			if err != nil {
				return nil, err
			}
			//add tempKey to url
			validURL = webUtil.CreateUrlWithTempKey(validURL, tempKey)
			c.Logger.Info().Str("url", validURL).Msg("first start, browser url")
			c.Logger.Info().Msg("You have 3 chances to use this within the next 3 minutes. Please use it as soon as possible.")
		}
		if err := webUtil.OpenBrowser(validURL); err != nil {
			c.Logger.Info().Msg("failed to open browser, please open it manually")
		}
	}
	return ws, err
}

func checkConfig(c *client.Client) error {
	if len(c.Config().WebAddr) == 0 {
		return errors.New("option webAddr must be set")
	}
	if strings.IndexByte(c.Config().WebAddr, ':') == -1 {
		c.Config().WebAddr = ":" + c.Config().WebAddr
	}
	if len(c.Config().Config) == 0 {
		c.Config().Config = predef.GetDefaultClientConfigPath()
		c.Logger.Info().Str("config", c.Config().Config).Msg("use default config path")
	}
	if len(c.Config().SigningKey) == 0 {
		c.Config().SigningKey = util.RandomString(predef.DefaultSigningKeySize)
	}

	if len(c.Config().WebCertFile) > 0 && len(c.Config().WebKeyFile) > 0 {
		if c.Config().WebCertFile != "auto" && c.Config().WebKeyFile != "auto" {
			_, certErr := os.Stat(c.Config().WebCertFile)
			_, keyErr := os.Stat(c.Config().WebKeyFile)
			if os.IsNotExist(certErr) || os.IsNotExist(keyErr) {
				return errors.New("provided webCertFile or webKeyFile does not exist")
			}
		}
	}
	return nil
}

func getServer(c *client.Client, tokenManager *server.TokenManager, r *gin.Engine) (*Server, error) {
	webServer := &Server{
		server: &http.Server{
			Addr:    c.Config().WebAddr,
			Handler: r,
		},
		logger:       c.Logger,
		tokenManager: tokenManager,
		enableTLS:    false,
	}

	certFile := c.Config().WebCertFile
	keyFile := c.Config().WebKeyFile

	if len(certFile) == 0 && len(keyFile) == 0 {
		// if both certFile and keyFile are empty, use HTTP, no need to do anything else
		return webServer, nil
	} else if certFile == "auto" && keyFile == "auto" {
		// if both certFile and keyFile are "auto", generate a self-signed certificate and enable HTTPS
		webServer.server.TLSConfig = webUtil.GenerateCertification()
		webServer.enableTLS = true
	} else {
		// if user provides certFile and keyFile, load them and enable HTTPS
		tlsCert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			err = errors.New("failed to load certFile or keyFile for web server")
			return nil, err
		}
		webServer.server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
		}
		webServer.enableTLS = true
	}

	return webServer, nil
}

func setRoutes(c *client.Client, tm *server.TokenManager, r *gin.Engine) error {
	staticFiles, err := fs.Sub(FS, "dist")
	if err != nil {
		return err
	}
	r.StaticFS("/static", http.FS(staticFiles))

	PublicGroup := r.Group("/")
	{
		PublicGroup.POST("/api/login", api.Login(c))
		PublicGroup.GET("/api/health", api.HealthCheck)
		PublicGroup.GET("/api/verify", api.VerifyTempKey(tm))
	}

	apiGroup := r.Group("/api")
	apiGroup.Use(middleware.JWTAuthMiddleware(c.Config().SigningKey, predef.DefaultTokenDuration))
	{
		userGroup := apiGroup.Group("/user")
		{
			userGroup.POST("/change", api.ChangeUserInfo(c))
			userGroup.GET("/info", api.GetUserInfo(c))
		}
		configGroup := apiGroup.Group("/config")
		{
			configGroup.GET("/running", api.GetRunningConfig(c))
			configGroup.GET("/file", api.GetConfigFromFile(c))
			configGroup.POST("/save", api.SaveConfigToFile(c))
		}

		serverGroup := apiGroup.Group("/server")
		{
			serverGroup.GET("/info", api.GetServerInfo)
			serverGroup.PUT("/reload", api.ReloadServices)
			serverGroup.PUT("/restart", api.Restart)
			serverGroup.PUT("/stop", api.Stop)
			serverGroup.PUT("/kill", api.Kill)
		}

		connectionGroup := apiGroup.Group("/connection")
		{
			connectionGroup.GET("/list", api.GetConnectionInfo(c))
		}

		permissionGroup := apiGroup.Group("/permission")
		{
			permissionGroup.GET("/menu", api.GetMenu(c))
		}
	}

	if c.Config().EnablePprof {
		pprofGroup := r.Group("/debug/pprof")
		{
			pprofGroup.GET("/", gin.WrapF(pprof.Index))
			pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
			pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
			pprofGroup.POST("/symbol", gin.WrapF(pprof.Symbol))
			pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
			pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
			pprofGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
			pprofGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
			pprofGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
			pprofGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
			pprofGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
			pprofGroup.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
		}
	}

	r.NoRoute(func(ctx *gin.Context) {
		data, err := fs.ReadFile(FS, "dist/index.html")
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})
	return nil
}

func (s *Server) start(serverErr chan<- error) {
	defer s.logger.Info().Msg("web server stopped")
	if predef.IsNoArgs() {
		startTime := time.Now()
		for {
			s.logger.Info().Str("addr", s.server.Addr).Msg("web server started")
			err := s.server.ListenAndServe()
			if err == nil || errors.Is(err, http.ErrServerClosed) {
				return
			}
			if opErr, ok := err.(*net.OpError); ok {
				if sysErr, ok := opErr.Err.(*os.SyscallError); ok {
					if sysErr.Err == syscall.EADDRINUSE {
						if time.Since(startTime) > 1*time.Second {
							err = errors.New("retry timeout, web server stopped")
							serverErr <- err
							return
						}
						s.logger.Error().Err(err).Msg("web server failed to serve")
						s.logger.Warn().Msg("web server address already in use, retrying...")
						//Try to use a random port
						s.server.Addr = ":" + strconv.Itoa(util.RandomPort())
						continue
					}
				}
			}
			serverErr <- err
			return
		}
	} else {
		s.logger.Info().Str("addr", s.server.Addr).Msg("web server started")
		var err error
		if s.enableTLS {
			s.logger.Info().Msg("start web server with TLS")
			err = s.server.ListenAndServeTLS("", "")
		} else {
			s.logger.Info().Msg("start web server without TLS")
			err = s.server.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error().Err(err).Msg("web server failed to serve")
			serverErr <- err
		}
	}
}

// isFirstStart checks if the server is started for the first time, and set a random user
func isFirstStart(c *client.Client) bool {
	var isFirst bool
	// if the Admin and Password is not set,
	// treat it as the first start
	if len(c.Config().Admin) == 0 {
		c.Config().Admin = util.RandomString(predef.DefaultAdminSize)
		isFirst = true
	}
	if len(c.Config().Password) == 0 {
		c.Config().Password = util.RandomString(predef.DefaultPasswordSize)
		isFirst = true
	}
	return isFirst
}

func getTempKey(ws *Server, c *client.Client) (string, error) {
	tempUser := request.User{
		Username: c.Config().Admin,
		Password: c.Config().Password,
	}
	token, err := webUtil.GenerateToken(c.Config().SigningKey, predef.DefaultTokenDuration, "gt-client", tempUser)
	if err != nil {
		return "", err
	}
	tempKey := ws.tokenManager.GenerateTempKey(token)

	return tempKey, nil
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}
