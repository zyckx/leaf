package api

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/wangzmgit/jigsaw"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"kuukaa.fun/leaf/cache"
	"kuukaa.fun/leaf/domain/dto"
	"kuukaa.fun/leaf/domain/model"
	"kuukaa.fun/leaf/domain/resp"
	"kuukaa.fun/leaf/domain/valid"
	"kuukaa.fun/leaf/service"
	"kuukaa.fun/leaf/util/jwt"
	"kuukaa.fun/leaf/util/mail"
	"kuukaa.fun/leaf/util/random"
)

func SendEmailCode(ctx *gin.Context) {
	// 获取邮箱
	email := ctx.PostForm("email")

	zap.L().Debug(email)

	// 生成code
	code := random.GenerateNumberCode(4)

	zap.L().Debug(code)

	// 发送code
	if err := mail.SendCaptcha(email, code); err != nil {
		resp.Response(ctx, resp.SendMailError, "邮箱验证码发送失败", nil)
		zap.L().Error("邮箱验证码发送失败")
		return
	}

	// code放入缓存
	cache.SetEmailCode(email, code)

	resp.OK(ctx, "验证码已发送到您的邮箱", nil)

}

func Register(ctx *gin.Context) {
	// 获取参数
	var registerDTO dto.RegisterDTO
	if err := ctx.Bind(&registerDTO); err != nil {
		resp.Response(ctx, resp.RequestParamError, "", nil)
		zap.L().Error("请求参数有误")
		return
	}

	// 参数校验
	if !valid.Email(registerDTO.Email) { // 邮箱格式验证
		resp.Response(ctx, resp.RequestParamError, valid.EMAIL_ERROR, nil)
		zap.L().Error(valid.EMAIL_ERROR)
		return
	}

	if !valid.Password(registerDTO.Password) { // 密码格式验证
		resp.Response(ctx, resp.RequestParamError, valid.PASSWORD_ERROR, nil)
		zap.L().Error(valid.PASSWORD_ERROR)
		return
	}

	if !valid.EmailCode(registerDTO.EmailCode) { //邮箱验证码格式验证
		resp.Response(ctx, resp.RequestParamError, valid.EMAIL_CODE_ERROR, nil)
		zap.L().Error(valid.EMAIL_CODE_ERROR)
		return
	}

	if cache.GetEmailCode(registerDTO.Email) != registerDTO.EmailCode { // 验证邮箱验证码
		resp.Response(ctx, resp.EmailCodeError, "", nil)
		zap.L().Error("邮箱验证错误")
		return
	}

	// 保存到数据库
	user := dto.RegisterToUser(registerDTO)
	service.InsertUser(user)

	// 删除邮箱验证码
	cache.DelEmailCode(user.Email)

	// 返回
	resp.OK(ctx, "注册成功", nil)
}

func Login(ctx *gin.Context) {
	// 获取参数
	var loginDTO dto.LoginDTO
	if err := ctx.Bind(&loginDTO); err != nil {
		resp.Response(ctx, resp.RequestParamError, "", nil)
		zap.L().Error("请求参数有误")
		return
	}

	// 参数校验
	if !valid.Email(loginDTO.Email) { // 邮箱格式验证
		resp.Response(ctx, resp.RequestParamError, valid.EMAIL_ERROR, nil)
		zap.L().Error(valid.EMAIL_ERROR)
		return
	}

	if !valid.Password(loginDTO.Password) { // 密码格式验证
		resp.Response(ctx, resp.RequestParamError, valid.PASSWORD_ERROR, nil)
		zap.L().Error(valid.PASSWORD_ERROR)
		return
	}

	// 读取登录尝试次数，超过3次进行滑块验证
	loginTryCount := cache.GetLoginTryCount(loginDTO.Email)
	if loginTryCount >= 3 {
		// 进行滑块验证
		slider, bg, x, y, err := jigsaw.Create()
		if err != nil {
			zap.L().Error("滑块验证资源生成失败")
		}
		// 保存x坐标到缓存
		cache.SetSliderX(loginDTO.Email,x)

		resp.

		return 

	}

	// 读取数据库
	user := service.SelectUserByEmail(loginDTO.Email)
	// 验证账号密码
	passwordError := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginDTO.Password))
	if reflect.DeepEqual(user, model.User{}) == false || passwordError != nil {
		resp.Response(ctx, resp.UsernamePasswordNotMatchError, "", nil)
		zap.L().Info("用户名密码不匹配")
		// 记录登录尝试次数
		loginTryCount++
		cache.SetLoginTryCount(loginDTO.Email, loginTryCount)
		return
	}

	// 生成验证token
	var err error
	var accessToken string
	var refreshToken string
	if accessToken, err = jwt.GenerateAccessToken(user.ID); err != nil {
		resp.Response(ctx, resp.Error, "验证token生成失败", nil)
		zap.L().Error("验证token生成失败")
		return
	}
	// 生成刷新token
	if refreshToken, err = jwt.GenerateRefreshToken(user.ID); err != nil {
		resp.Response(ctx, resp.Error, "刷新token生成失败", nil)
		zap.L().Error("刷新token生成失败")
		return
	}

	// 存入缓存
	cache.SetAccessToken(user.ID, accessToken)
	cache.SetRefreshToken(user.ID, refreshToken)

	// 返回给前端
	resp.OK(ctx, "", gin.H{"accessToken": accessToken, "refreshToken": refreshToken})
}
