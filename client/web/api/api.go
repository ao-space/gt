package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/client/web/service"
	"github.com/isrc-cas/gt/web/server/model/request"
	"github.com/isrc-cas/gt/web/server/model/response"
	"github.com/isrc-cas/gt/web/server/util"
)

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
		token, err := service.GenerateToken(c.Config().SigningKey, loginReq)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		response.SuccessWithData(gin.H{"token": token}, ctx)
	}
}

// GetMenu returns the permission menu based on the role of the user
func GetMenu(c *client.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		menu := service.GetMenu(c)
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
			c.Logger.Error().Msg(err.Error())
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
		cfg, err := service.GetConfigFromFile(c)
		if err != nil {
			c.Logger.Error().Err(err).Msg("get config from file failed")
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
		var cfg client.Config
		err := inheritImmutableConfigFields(c.Config(), &cfg)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		if err := ctx.ShouldBindJSON(&cfg); err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		fullPath, err := service.SaveConfigToFile(&cfg)
		c.Logger.Info().Msg("save config in :" + cfg.Options.Config)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		response.SuccessWithMessage("save config to "+fullPath, ctx)
	}
}

// inheritImmutableConfigFields copy immutable fields from original to new
func inheritImmutableConfigFields(original *client.Config, new *client.Config) (err error) {
	if original == nil {
		err = errors.New("original config is nil")
		return
	}
	new.Options.Config = original.Options.Config
	new.EnableWebServer = original.EnableWebServer
	new.WebAddr = original.WebAddr
	new.WebPort = original.WebPort
	new.EnablePprof = original.EnablePprof
	new.SigningKey = original.SigningKey
	new.Admin = original.Admin
	new.Password = original.Password
	return
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
