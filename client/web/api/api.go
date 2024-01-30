package api

import (
	"github.com/gin-gonic/gin"
	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/client/web/service"
	"github.com/isrc-cas/gt/predef"
	"github.com/isrc-cas/gt/web/server"
	"github.com/isrc-cas/gt/web/server/model/request"
	"github.com/isrc-cas/gt/web/server/model/response"
	"github.com/isrc-cas/gt/web/server/util"
)

func HealthCheck(ctx *gin.Context) {
	response.Success(ctx)
}

func Login(c *client.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var loginReq = request.User{}
		if err := ctx.ShouldBindJSON(&loginReq); err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		if err := service.VerifyUser(loginReq, c); err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		token, err := util.GenerateToken(c.Config().SigningKey, predef.DefaultTokenDuration, "gt-client", loginReq)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		response.SuccessWithData(gin.H{"token": token}, ctx)
	}
}

func VerifyTempKey(tokenManager *server.TokenManager) gin.HandlerFunc {
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

func ChangeUserInfo(c *client.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var userInfo = request.UserInfo{}
		if err := ctx.ShouldBindJSON(&userInfo); err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		if err := service.ChangeUserInfo(userInfo, c); err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		user := request.User{
			Username: userInfo.Username,
			Password: userInfo.Password,
		}
		token, err := util.GenerateToken(c.Config().SigningKey, predef.DefaultTokenDuration, "gt-client", user)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		response.SuccessWithData(gin.H{"token": token}, ctx)
	}
}

func GetUserInfo(c *client.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var userInfo request.UserInfo
		cfg, err := service.GetMergedConfig(c)
		if err != nil {
			userInfo = request.UserInfo{
				Username:    cfg.Admin,
				Password:    cfg.Password,
				EnablePprof: cfg.EnablePprof,
			}
		}
		// Get the user info from running config,if the config file is not specified
		userInfo = request.UserInfo{
			Username:    c.Config().Admin,
			Password:    c.Config().Password,
			EnablePprof: c.Config().EnablePprof,
		}
		response.SuccessWithData(userInfo, ctx)
	}
}

// GetMenu returns the permission menu based on the role of the user
func GetMenu(c *client.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		lang := ctx.Query("lang")
		menu := service.GetMenu(c, lang)
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

// GetConnectionInfo returns connection info ( pool, external )
func GetConnectionInfo(c *client.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		poolStatus := service.GetConnectionPoolStatus(c)
		conn, err := service.GetConnectionInfo(c)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		response.SuccessWithData(gin.H{"clientPool": poolStatus, "external": conn}, ctx)
	}
}

// GetRunningConfig GetConfig returns the current config
func GetRunningConfig(c *client.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var cfg = c.Config()
		c.Logger.Info().Msg("get running config")
		response.SuccessWithData(gin.H{"config": cfg}, ctx)
	}
}

// GetConfigFromFile returns the config from file,
// if failed, try to fetch running config
func GetConfigFromFile(c *client.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cfg, err := service.GetMergedConfig(c)
		c.Logger.Info().Msg("get config from file")
		if err != nil {
			c.Logger.Info().Msg("get config from file failed, try to fetch running config")
			// try to fetch running config
			GetRunningConfig(c)(ctx)
			return
		}
		response.SuccessWithData(gin.H{"config": cfg}, ctx)
	}
}

// SaveConfigToFile save config to file,
// If the config file is not specified, save to the default path
func SaveConfigToFile(c *client.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		oldConfig, err := service.InheritConfig(c)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		var cfg client.Config
		if err := service.InheritImmutableConfigFields(&oldConfig, &cfg); err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		if err := ctx.ShouldBindJSON(&cfg); err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		service.SeparateConfig(&cfg)
		fullPath, err := service.SaveConfigToFile(&cfg)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		c.Logger.Info().Str("config", fullPath).Msg("save config to file")
		response.SuccessWithMessage("save config to "+fullPath, ctx)
	}
}

// ReloadServices reloads the services in the config file
// without restarting the current process
func ReloadServices(ctx *gin.Context) {
	err := util.SendSignal("reload")
	if err != nil {
		response.FailWithMessage(err.Error(), ctx)
		return
	}
	response.SuccessWithMessage("reload services done", ctx)
}

// Restart for a brand-new config process,
// not only reload the services,
// but also restart the process to load the brand-new config(mainly for options part)
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
