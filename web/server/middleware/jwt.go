package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/isrc-cas/gt/web/server/model/response"
	"github.com/isrc-cas/gt/web/server/util"
	"time"
)

func JWTAuthMiddleware(signingKey string, tokenDuration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("x-token")
		if token == "" {
			response.InvalidToken(c)
			c.Abort()
			return
		}
		j := util.NewJWT(signingKey, tokenDuration)
		claims, err := j.ParseToken(token)
		if err != nil {
			response.InvalidToken(c)
			c.Abort()
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}
