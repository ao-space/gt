package web

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/client/web/api"
	"github.com/isrc-cas/gt/client/web/middleware"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"
	"time"
)

var server *webServer

type webServer struct {
	http.Server
}

func NewWebServer(c *client.Client) {
	addr := c.Config().WebAddr + ":" + strconv.Itoa(int(c.Config().WebPort))

	c.Logger.Info().Msg("start web server on " + addr)
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)

	r := gin.Default()
	setRoutes(c, r)

	srv := &webServer{
		Server: http.Server{
			Addr:    addr,
			Handler: r,
		},
	}
	server = srv
	startWebServer(c)
	return
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
			permissionGroup.GET("/menu", api.GetClientMenu(c))
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

func startWebServer(c *client.Client) {
	go func() {
		defer func() {
			c.Logger.Info().Msg("web server stopped")
		}()
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			c.Logger.Error().Err(err).Msg("web server failed to serve")
			return
		}
		return
	}()
	return
}

// TODO: shutdown for restart?

// ShutdownWebServer used to shut down web server before the next restart,
// to avoid the port being occupied.
func ShutdownWebServer() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
