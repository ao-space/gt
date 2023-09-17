package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/isrc-cas/gt/server"
	"github.com/isrc-cas/gt/server/web/service"
	"github.com/isrc-cas/gt/web/server/model/request"
	"github.com/isrc-cas/gt/web/server/model/response"
	"github.com/isrc-cas/gt/web/server/util"
)

func Login(s *server.Server) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var loginReq = request.User{}
		if err := ctx.ShouldBindJSON(&loginReq); err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		if err := service.VerifyUser(loginReq, s); err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		token, err := service.GenerateToken(s.Config().SigningKey, loginReq)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		response.SuccessWithData(gin.H{"token": token}, ctx)
	}
}

// GetMenu returns the permission menu based on the role of the user
func GetMenu(s *server.Server) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		menu := service.GetMenu(s)
		response.SuccessWithData(menu, ctx)

	}
}

// GetServerInfo returns system info ( OS, CPU, Memory, Disk )
func GetServerInfo(ctx *gin.Context) {
	serverInfo, err := util.GetServerInfo()
	if err != nil {
		response.FailWithMessage(err.Error(), ctx)
		return
	}
	response.SuccessWithData(gin.H{"serverInfo": serverInfo}, ctx)
}

// GetConnectionInfo returns connection info ( client pool, external )
func GetConnectionInfo(s *server.Server) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serverPool, external, err := service.GetConnectionInfo(s)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		response.SuccessWithData(gin.H{"serverPool": serverPool, "external": external}, ctx)
	}
}

// GetRunningConfig returns the running config
func GetRunningConfig(s *server.Server) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var cfg = s.Config()
		s.Logger.Info().Msg("get running config")
		response.SuccessWithData(gin.H{"config": cfg}, ctx)
	}
}

// GetConfigFromFile returns the config from file,
// If failed, try to fetch running config
func GetConfigFromFile(s *server.Server) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cfg, err := service.GetConfigFromFile(s)
		if err != nil {
			s.Logger.Error().Err(err).Msg("get config from file failed")
			// try to fetch running config
			GetRunningConfig(s)(ctx)
			return
		}
		response.SuccessWithData(gin.H{"config": cfg}, ctx)
	}
}

// SaveConfigToFile saves the config to file,
// If the config file is not specified, save to default config file
func SaveConfigToFile(s *server.Server) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var cfg server.Config
		err := inheritImmutableConfigFields(s.Config(), &cfg)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		if err := ctx.ShouldBindJSON(&cfg); err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		fullPath, err := service.SaveConfigToFile(&cfg)
		s.Logger.Info().Msg("save config in:" + cfg.Options.Config)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		response.SuccessWithMessage("save config to "+fullPath, ctx)
	}
}

// inheritImmutableConfigFields copy immutable fields from original to new
func inheritImmutableConfigFields(original *server.Config, new *server.Config) (err error) {
	if original == nil {
		err = errors.New("original config is nil")
		return
	}
	new.Config = original.Config
	new.EnableWebServer = original.EnableWebServer
	new.WebAddr = original.WebAddr
	new.WebPort = original.WebPort
	new.EnablePprof = original.EnablePprof
	new.SigningKey = original.SigningKey
	new.Admin = original.Admin
	new.Password = original.Password
	return
}

func Restart(ctx *gin.Context) {
	err := util.SendSignal("restart")
	if err != nil {
		response.FailWithMessage(err.Error(), ctx)
		return
	}
	response.Success(ctx)
}

func Stop(ctx *gin.Context) {
	err := util.SendSignal("stop")
	if err != nil {
		response.FailWithMessage(err.Error(), ctx)
		return
	}
	response.Success(ctx)
}

func Kill(ctx *gin.Context) {
	err := util.SendSignal("kill")
	if err != nil {
		response.FailWithMessage(err.Error(), ctx)
		return
	}
	response.Success(ctx)
}
