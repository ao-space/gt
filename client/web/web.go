package web

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/client/web/api"
	"github.com/isrc-cas/gt/logger"
	"github.com/isrc-cas/gt/predef"
	"github.com/isrc-cas/gt/util"
	"github.com/isrc-cas/gt/web/server/middleware"
	webUtil "github.com/isrc-cas/gt/web/server/util"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Server struct {
	server *http.Server
	logger logger.Logger // have no right to close logger
}

func NewWebServer(c *client.Client) (*Server, error) {

	err := checkConfig(c)
	if err != nil {
		return nil, err
	}

	addr := c.Config().WebAddr + ":" + strconv.Itoa(int(c.Config().WebPort))
	c.Logger.Info().Msg("start web server on " + addr)

	fullPath := filepath.Join(webUtil.GetAppDir(), "web_client.log")
	f, _ := os.Create(fullPath)
	gin.DefaultWriter = io.MultiWriter(f)

	r := gin.Default()
	setRoutes(c, r)

	ws := &Server{
		server: &http.Server{
			Addr:    addr,
			Handler: r,
		},
		logger: c.Logger,
	}
	go ws.start()
	return ws, nil
}

func checkConfig(c *client.Client) (err error) {
	if c.Config().WebAddr == "" {
		return errors.New("option webAddr must be set")
	}
	if c.Config().WebPort <= 0 {
		return errors.New("option webPort must be set")
	}
	if c.Config().Admin == "" {
		return errors.New("option admin must be set")
	}
	if c.Config().Password == "" {
		return errors.New("option password must be set")
	}
	if c.Config().SigningKey == "" {
		c.Config().SigningKey = util.RandomString(predef.DefaultSigningKeySize)
	}

	return nil
}

func setRoutes(c *client.Client, r *gin.Engine) {
	PublicGroup := r.Group("/")
	{
		PublicGroup.POST("/api/login", api.Login(c))
	}

	apiGroup := r.Group("/api")
	apiGroup.Use(middleware.JWTAuthMiddleware(c.Config().SigningKey))
	{
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
}

func (s *Server) start() {
	defer s.logger.Info().Msg("web server stopped")
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error().Err(err).Msg("web server failed to serve")
		return
	}
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}
