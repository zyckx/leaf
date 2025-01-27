package service

import (
	"kuukaa.fun/leaf/cache"
	"kuukaa.fun/leaf/util/jwt"
)

func GenerateAccessToken(userId uint) (accessToken string, err error) {
	accessToken = cache.GetAccessToken(userId)
	if accessToken == "" {
		accessToken, err = jwt.GenerateAccessToken(userId)

		// 存入缓存
		cache.SetAccessToken(userId, accessToken)
	}

	return
}
