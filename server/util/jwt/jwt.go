package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"kuukaa.fun/leaf/cache"
)

type Claims struct {
	UserId    uint
	TokenType uint // 0:accessToken,1:refreshtToken
	jwt.StandardClaims
}

/**
 * 生成验证用的token
 * param: id 用户id
 * return: token字符串、错误信息
 */
func GenerateAccessToken(id uint) (string, error) {
	accessJwtKey := []byte(viper.GetString("security.access_jwt_secret"))
	// token过期时间
	expirationTime := time.Now().Add(cache.ACCESS_TOKEN_EXPRIRATION_TIME * time.Minute) // 5分钟有效
	accessClaims := &Claims{
		UserId:    id,
		TokenType: 0,
		StandardClaims: jwt.StandardClaims{
			//发放时间等
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "leaf",
		},
	}
	return generateToken(accessJwtKey, accessClaims)
}

/**
 * 生成刷新用的token
 * param: id 用户id
 * return: token字符串、错误信息
 */
func GenerateRefreshToken(id uint) (string, error) {
	refreshJwtKey := []byte(viper.GetString("security.refresh_jwt_secret"))
	// token过期时间
	expirationTime := time.Now().Add(cache.REFRESH_TOKEN_EXPRIRATION_TIME * time.Hour) // 14天有效

	refreshClaims := &Claims{
		UserId:    id,
		TokenType: 1,
		StandardClaims: jwt.StandardClaims{
			//发放时间等
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "leaf",
		},
	}
	return generateToken(refreshJwtKey, refreshClaims)
}

/**
 * 获取token的荷载数据
 * param: tokenString token字符串s
 * return: *Claims token的负载结构体
 * return: error 解析token的错误信息
 */
func GetTokenClaims(tokenString string) (*Claims, error) {
	claims := &Claims{}
	parser := jwt.NewParser()
	_, _, err := parser.ParseUnverified(tokenString, claims)
	return claims, err
}

/**
 * 验证token
 * param: tokenString token字符串
 * return: *jwt.Token token结构体
 * return: error 解析token的错误信息
 */
func ParseToken(tokenString string) (*jwt.Token, *Claims, error) {
	// 获取jwt的荷载数据
	claims, err := GetTokenClaims(tokenString)
	if err != nil {
		zap.L().Error("token荷载解析失败: " + err.Error())
	}
	// 判断类型 选择不同的密钥
	var secret []byte
	if claims.TokenType == 0 { // accessToken
		secret = []byte(viper.GetString("security.access_jwt_secret"))
	} else if claims.TokenType == 1 { // refreshToken
		secret = []byte(viper.GetString("security.refresh_jwt_secret"))
	} else {

	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (i interface{}, e error) {
		return secret, nil
	})

	return token, claims, err
}

/**
 * 通过密钥和负载生成token字符串
 * param: key 密钥
 * claims: jwt负载
 * return: token字符串、错误信息
 */
func generateToken(key []byte, claims *Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
