package middleware

import (
	// "fmt"

	"log"
	"net/http"

	"tarkib.uz/config"
	"tarkib.uz/pkg/logger"
	jWT "tarkib.uz/pkg/token"

	"github.com/casbin/casbin/v2"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
	"github.com/spf13/cast"
)

type JWTRoleAuth struct {
	enforcer   *casbin.Enforcer
	cfg        *config.Config
	jwtHandler jWT.JWTHandler
}

func NewAuthorizer(e *casbin.Enforcer, jwtHandler jWT.JWTHandler, cfg *config.Config, l logger.Interface) gin.HandlerFunc {
	a := &JWTRoleAuth{
		enforcer:   e,
		cfg:        cfg,
		jwtHandler: jwtHandler,
	}

	return func(c *gin.Context) {
		allow, err := a.CheckPermission(c.Request, l)
		if err != nil {
			pp.Println("OUR ERROR IS: ", err)
			if err.Error() == "token is expired huh" {
				a.RequireRefresh(c)
				return
			}
			v, _ := err.(*jwt.ValidationError)
			if v.Errors == jwt.ValidationErrorExpired {
				a.RequireRefresh(c)
			} else {
				a.RequirePermission(c)
			}
		} else if !allow {
			a.RequirePermission(c)
		}
	}
}

func (a *JWTRoleAuth) CheckPermission(r *http.Request, l logger.Interface) (bool, error) {
	user, err := a.GetRole(r)
	if err != nil {

		log.Println("error get role", err)
		return false, err
	}

	method := r.Method
	if method == "OPTIONS" {
		return true, nil
	}
	path := r.URL.Path

	allowed, err := a.enforcer.Enforce(user, path, method)
	if err != nil {
		l.Error("enforce error", err)
		return false, err
	}

	return allowed, nil
}

func (a *JWTRoleAuth) GetRole(r *http.Request) (string, error) {
	var (
		role   string
		claims jwt.MapClaims
		err    error
	)

	jwtToken := r.Header.Get("Authorization")

	if jwtToken == "" {
		return "unauthorized", nil
	}

	a.jwtHandler.Token = jwtToken

	claims, err = a.jwtHandler.ExtractClaims()

	if err != nil {
		log.Println("error check token", err)
		return "", err
	}
	if cast.ToString(claims["role"]) == "admin" {
		role = "admin"
	} else if cast.ToString(claims["role"]) == "user" {
		role = "user"
	} else if cast.ToString(claims["role"]) == "unauthorized" {
		role = "unauthorized"
	} else if cast.ToString(claims["role"]) == "super-admin" {
		role = "super-admin"
	} else {
		role = "unknown"
	}

	return role, nil
}

func (a *JWTRoleAuth) RequireRefresh(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"error": "token is expired",
	})
	c.AbortWithStatus(401)
}

func (a *JWTRoleAuth) RequirePermission(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{
		"Error": "Sizga bu sahifaga kirishga ruxsat yo'q",
	})
	c.AbortWithStatus(403)
}
