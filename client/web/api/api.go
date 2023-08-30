package api

import (
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
		response.SuccessWithData(gin.H{"config": cfg}, ctx)
	}
}
func GetConfigFromFile(c *client.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var cfg client.Config
		cfg = *c.Config() //initialize cfg with running config, mainly for configPath
		cfg, err := service.GetConfigFormFile(c)
		if err != nil {
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
		cfg = *c.Config() //initialize cfg with running config, mainly for configPath
		c.Logger.Info().Msg("CONFIG:" + cfg.Config)
		c.Logger.Info().Msg("URL:" + cfg.Services[0].LocalURL.URL.String())
		response.SuccessWithDetailed(gin.H{"config": cfg}, "JSONBind Before", ctx)
		if err := ctx.ShouldBindJSON(&cfg); err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		response.SuccessWithDetailed(gin.H{"config": cfg}, "JSON", ctx)
		fullPath, err := service.SaveConfigToFile(&cfg)
		if err != nil {
			response.FailWithMessage(err.Error(), ctx)
			return
		}
		response.SuccessWithMessage("save config in"+fullPath, ctx)
	}
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
