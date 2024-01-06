package api

import (
	"github.com/gin-gonic/gin"
	"github.com/isrc-cas/gt/predef"
	"github.com/isrc-cas/gt/server"
	"github.com/isrc-cas/gt/server/web/service"
	wServer "github.com/isrc-cas/gt/web/server"
	"github.com/isrc-cas/gt/web/server/model/request"
	"github.com/isrc-cas/gt/web/server/model/response"
	"github.com/isrc-cas/gt/web/server/util"
)

func HealthCheck(ctx *gin.Context) {
	response.Success(ctx)
}

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
		token, err := util.GenerateToken(s.Config().SigningKey, predef.DefaultTokenDuration, "gt-server", loginReq)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		response.SuccessWithData(gin.H{"token": token}, ctx)
	}
}

func VerifyTempKey(tokenManager *wServer.TokenManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tempKey := ctx.Query("key")
		if tempKey == "" {
			response.InvalidKey(ctx)
			return
		}

		actualToken, err := tokenManager.GetActualToken(tempKey)
		if err != nil {
			response.InvalidKey(ctx)
			return
		}
		response.SuccessWithData(gin.H{"token": actualToken}, ctx)
	}
}

func ChangeUserInfo(s *server.Server) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var userInfo = request.UserInfo{}
		if err := ctx.ShouldBindJSON(&userInfo); err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		if err := service.ChangeUserInfo(userInfo, s); err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		user := request.User{
			Username: userInfo.Username,
			Password: userInfo.Password,
		}
		token, err := util.GenerateToken(s.Config().SigningKey, predef.DefaultTokenDuration, "gt-server", user)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		response.SuccessWithData(gin.H{"token": token}, ctx)
	}
}

func GetUserInfo(s *server.Server) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var userInfo request.UserInfo
		cfg, err := service.GetConfigFromFile(s)
		if err != nil {
			userInfo = request.UserInfo{
				Username:    cfg.Admin,
				Password:    cfg.Password,
				EnablePprof: cfg.EnablePprof,
			}
		}
		// Get the user info from running config,if the config file is not specified
		userInfo = request.UserInfo{
			Username:    s.Config().Admin,
			Password:    s.Config().Password,
			EnablePprof: s.Config().EnablePprof,
		}
		response.SuccessWithData(userInfo, ctx)
	}
}

// GetMenu returns the permission menu based on the role of the user
func GetMenu(s *server.Server) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		lang := ctx.Query("lang")
		menu := service.GetMenu(s, lang)
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
		s.Logger.Info().Msg("get config from file")
		if err != nil {
			s.Logger.Info().Msg("get config from file failed, try to fetch running config")
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
		oldConfig, err := service.InheritConfig(s)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		var cfg server.Config
		if err := service.InheritImmutableConfigFields(&oldConfig, &cfg); err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		if err := ctx.ShouldBindJSON(&cfg); err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		fullPath, err := service.SaveConfigToFile(&cfg)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		s.Logger.Info().Str("config", fullPath).Msg("save config to file")
		response.SuccessWithMessage("save config to "+fullPath, ctx)
	}
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
