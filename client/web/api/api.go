package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/client/web/model/request"
	"github.com/isrc-cas/gt/client/web/model/response"
	"github.com/isrc-cas/gt/client/web/service"
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

func GetServerInfo(ctx *gin.Context) {
	serverInfo, err := service.GetServerInfo()
	if err != nil {
		response.FailWithMessage(err.Error(), ctx)
		return
	}
	response.SuccessWithData(gin.H{"serverInfo": serverInfo}, ctx)
}
func GetRunningConfig(c *client.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var cfg = c.Config()
		c.Logger.Info().Msg("Running CONFIG:" + cfg.Config)
		response.SuccessWithData(gin.H{"config": cfg}, ctx)
	}
}
func GetConfigFromFile(c *client.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cfg, err := service.GetConfigFormFile(c)
		if err != nil {
			c.Logger.Error().Err(err).Msg("get config from file failed")
			// try to fetch running config
			GetRunningConfig(c)(ctx)
			return
		}
		response.SuccessWithData(gin.H{"config": cfg}, ctx)
	}
}

func SaveConfigToFile(c *client.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var cfg client.Config
		err := inheritImmutableConfigFields(c.Config(), &cfg)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		c.Logger.Info().Msg("SaveConfig in :" + cfg.Config)
		response.SuccessWithDetailed(gin.H{"config": cfg}, "JSONBind Before", ctx)
		if err := ctx.ShouldBindJSON(&cfg); err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		response.SuccessWithDetailed(gin.H{"config": cfg}, "JSON", ctx)
		response.SuccessWithDetailed(gin.H{"RunningConfig": *c.Config()}, "JSONBind After", ctx)
		fullPath, err := service.SaveConfigToFile(&cfg)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		response.SuccessWithMessage("save config in"+fullPath, ctx)
	}
}

// inheritImmutableConfigFields copy immutable fields from original to new
func inheritImmutableConfigFields(original *client.Config, new *client.Config) (err error) {
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

// ServerGroup api

func ReloadServices(ctx *gin.Context) {
	err := service.SendSignal("reload")
	if err != nil {
		response.FailWithMessage(err.Error(), ctx)
		return
	}
	response.SuccessWithMessage("reload services done", ctx)
}

// Restart for a brand-new config process,
// not only restart the services
func Restart(ctx *gin.Context) {
	err := service.SendSignal("restart")
	if err != nil {
		response.FailWithMessage(err.Error(), ctx)
		return
	}
	response.Success(ctx)
}
func Stop(ctx *gin.Context) {
	err := service.SendSignal("stop")
	if err != nil {
		response.FailWithMessage(err.Error(), ctx)
		return
	}
	response.Success(ctx)
}
func Kill(ctx *gin.Context) {
	err := service.SendSignal("kill")
	if err != nil {
		response.FailWithMessage(err.Error(), ctx)
		return
	}
	response.Success(ctx)
}

func GetConnectionInfo(c *client.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		poolStatus := service.GetConnectionPoolStatus(c)
		conn, err := service.GetConnectionInfo(c)
		if err != nil {
			c.Logger.Error().Msg(err.Error())
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		response.SuccessWithData(gin.H{"pool": poolStatus, "connection": conn}, ctx)
	}
}

func GetClientMenu(c *client.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		router := service.GetClientMenu(c)
		response.SuccessWithData(router, ctx)
	}
}
