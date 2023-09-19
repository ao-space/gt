package web

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/isrc-cas/gt/logger"
	"github.com/isrc-cas/gt/predef"
	"github.com/isrc-cas/gt/server"
	"github.com/isrc-cas/gt/server/web/api"
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

func NewWebServer(s *server.Server) (*Server, error) {

	err := checkConfig(s)
	if err != nil {
		return nil, err
	}

	addr := s.Config().WebAddr + ":" + strconv.Itoa(int(s.Config().WebPort))
	s.Logger.Info().Msg("start web server on " + addr)

	fullPath := filepath.Join(webUtil.GetAppDir(), "web_server.log")
	f, _ := os.Create(fullPath)
	gin.DefaultWriter = io.MultiWriter(f)

	r := gin.Default()
	setRoutes(s, r)

	ws := &Server{
		server: &http.Server{
			Addr:    addr,
			Handler: r,
		},
		logger: s.Logger,
	}

	go ws.start()
	return ws, nil
}

func checkConfig(s *server.Server) (err error) {
	if s.Config().WebAddr == "" {
		return errors.New("option webAddr must be set")
	}
	if s.Config().WebPort <= 0 {
		return errors.New("option webPort must be set")
	}
	if s.Config().Admin == "" {
		return errors.New("option admin must be set")
	}
	if s.Config().Password == "" {
		return errors.New("option password must be set")
	}
	if s.Config().SigningKey == "" {
		s.Config().SigningKey = util.RandomString(predef.DefaultSigningKeySize)
	}

	return
}

func setRoutes(s *server.Server, r *gin.Engine) {
	PublicGroup := r.Group("/")
	{
		PublicGroup.POST("/api/login", api.Login(s))
	}
	apiGroup := r.Group("/api")
	apiGroup.Use(middleware.JWTAuthMiddleware(s.Config().SigningKey))
	{
		configGroup := apiGroup.Group("/config")
		{
			configGroup.GET("/running", api.GetRunningConfig(s))
			configGroup.GET("/file", api.GetConfigFromFile(s))
			configGroup.POST("/save", api.SaveConfigToFile(s))
		}

		serverGroup := apiGroup.Group("/server")
		{
			serverGroup.GET("/info", api.GetServerInfo)
			serverGroup.PUT("/restart", api.Restart)
			serverGroup.PUT("/stop", api.Stop)
			serverGroup.PUT("/kill", api.Kill)
		}

		connectionGroup := apiGroup.Group("/connection")
		{
			connectionGroup.GET("/list", api.GetConnectionInfo(s))
		}

		permissionGroup := apiGroup.Group("/permission")
		{
			permissionGroup.GET("/menu", api.GetMenu(s))
		}
	}

	if s.Config().EnablePprof {
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
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error().Err(err).Msg("web server failed to serve")
	}
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}
