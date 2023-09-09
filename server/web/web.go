package web

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/isrc-cas/gt/server"
	"github.com/isrc-cas/gt/server/web/api"
	"github.com/isrc-cas/gt/web/server/middleware"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"
	"time"
)

var webserver *webServer

type webServer struct {
	http.Server
}

func NewWebServer(s *server.Server) {
	addr := s.Config().WebAddr + ":" + strconv.Itoa(int(s.Config().WebPort))

	s.Logger.Info().Msg("start web server on " + addr)
	f, _ := os.Create("Web_Server.log")
	gin.DefaultWriter = io.MultiWriter(f)

	r := gin.Default()
	setRoutes(s, r)

	srv := &webServer{
		Server: http.Server{
			Addr:    addr,
			Handler: r,
		},
	}
	webserver = srv
	startWebServer(s)
	return

}

// TODO : Connection and Config and Server
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
		//pprofGroup.Use(middleware.JWTAuthMiddleware(s.Config().SigningKey))
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

func startWebServer(s *server.Server) {
	go func() {
		defer func() {
			s.Logger.Info().Msg("web server stopped")
		}()
		err := webserver.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			s.Logger.Fatal().Msg("listen: " + err.Error())
			return
		}
		return
	}()
	return
}

func ShutdownWebServer() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := webserver.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
