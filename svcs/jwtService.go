package svcs

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/erneap/go-models/logs"
	"github.com/erneap/go-models/users"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateToken(userid primitive.ObjectID, email string) (string, error) {
	key := []byte(os.Getenv("JWT_SECRET"))
	expireTime := time.Now().Add(6 * time.Hour)
	claims := &users.JWTClaim{
		UserID:       userid.Hex(),
		EmailAddress: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateToken(signedToken string) (*users.JWTClaim, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&users.JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*users.JWTClaim)
	if !ok {
		return nil, errors.New("couldn't parse claims")
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		return nil, errors.New("token expired")
	}
	return claims, nil
}

func GetRequestor(context *gin.Context) string {
	tokenString := context.GetHeader("Authorization")
	if tokenString == "" {
		return ""
	}
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return ""
	}
	return claims.UserID
}

func CheckJWT(app string) gin.HandlerFunc {
	return func(context *gin.Context) {
		tokenString := context.GetHeader("Authorization")
		if tokenString == "" {
			AddLogEntry(app, logs.Minimal,
				"CheckJWT: No Authentication Token passed")
			context.JSON(http.StatusUnauthorized, gin.H{"error": "request does not contain an access token"})
			context.Abort()
			return
		}
		claims, err := ValidateToken(tokenString)
		if err != nil {
			AddLogEntry(app, logs.Minimal, "CheckJWT: Validation Error: "+
				err.Error())
			context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			context.Abort()
			return
		}

		// replace token by passing a new token in the response header
		AddLogEntry(app, logs.Debug, "CheckJWT: Token Verified")
		id, _ := primitive.ObjectIDFromHex(claims.UserID)
		tokenString, _ = CreateToken(id, claims.EmailAddress)
		context.Writer.Header().Set("Token", tokenString)
		context.Next()
	}
}

func CheckRole(prog, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			AddLogEntry(prog, logs.Minimal,
				"CheckRole: No Authentication Token passed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "request does not contain an access token"})
			c.Abort()
			return
		}
		claims, err := ValidateToken(tokenString)
		if err != nil {
			AddLogEntry(prog, logs.Minimal, "CheckRole: Validation Error: "+
				err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		user, err := GetUserByID(claims.UserID)
		if err != nil {
			AddLogEntry(prog, logs.Minimal, "CheckRole: User Not Found: "+
				err.Error())
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found: " + err.Error()})
			c.Abort()
			return
		}
		if !user.IsInGroup(prog, role) {
			AddLogEntry(prog, logs.Minimal, "CheckRole: User Not in Group: "+
				user.LastName)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not in group"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func CheckRoles(prog string, roles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			AddLogEntry(prog, logs.Minimal,
				"CheckRoles: No Authentication Token passed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "request does not contain an access token"})
			c.Abort()
			return
		}
		claims, err := ValidateToken(tokenString)
		if err != nil {
			AddLogEntry(prog, logs.Minimal, "CheckRoles: Validation Error: "+
				err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		user, err := GetUserByID(claims.UserID)
		if err != nil {
			AddLogEntry(prog, logs.Minimal, "CheckRoles: User Not Found: "+
				err.Error())
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found: " + err.Error()})
			c.Abort()
			return
		}
		inRole := false
		for i := 0; i < len(roles) && !inRole; i++ {
			if user.IsInGroup(prog, roles[i]) {
				inRole = true
			}
		}
		if !inRole {
			AddLogEntry(prog, logs.Minimal, "CheckRoles: User Not In Group: "+
				user.LastName)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not in group"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func CheckRoleList(app string, roles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			AddLogEntry(app, logs.Minimal,
				"CheckRoleList: No Authentication Token passed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "request does not contain an access token"})
			c.Abort()
			return
		}
		claims, err := ValidateToken(tokenString)
		if err != nil {
			AddLogEntry(app, logs.Minimal,
				"CheckRoleList: Validation Error: "+err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		user, err := GetUserByID(claims.UserID)
		if err != nil {
			AddLogEntry(app, logs.Minimal, "CheckRoleList: User Not Found: "+
				err.Error())
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found: " + err.Error()})
			c.Abort()
			return
		}
		inRole := false
		for i := 0; i < len(roles) && !inRole; i++ {
			parts := strings.Split(roles[i], "-")
			if user.IsInGroup(parts[0], parts[1]) {
				inRole = true
			}
		}
		if !inRole {
			AddLogEntry(app, logs.Minimal,
				"CheckRoleList: User not in any of the roles provided: "+user.LastName)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not in group"})
			c.Abort()
			return
		}
		c.Next()
	}
}
