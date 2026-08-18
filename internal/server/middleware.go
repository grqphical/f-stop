package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grqphical/f-stop/internal/auth"
)

func (s *Server) Authorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationCookie, err := c.Cookie("Authorization")
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims := auth.ValidateJWT(authorizationCookie)
		if claims == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// make sure user exists
		userId := claims["sub"].(float64)
		user, err := s.db.GetUserByID(int(userId))
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("user", user)

		c.Next()
	}
}
