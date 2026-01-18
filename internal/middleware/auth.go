// Package middleware 提供 HTTP 中间件
// 中间件是在 Handler 之前/之后执行的函数，用于通用处理逻辑
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/proyuen/simple-bank-v2/internal/dto/response"
	apperrors "github.com/proyuen/simple-bank-v2/internal/errors"
	"github.com/proyuen/simple-bank-v2/pkg/token"
)

// ==================== 常量定义 ====================

const (
	// AuthorizationHeaderKey 是 HTTP 请求头中的认证字段名
	// 标准 HTTP 认证头格式: Authorization: Bearer <token>
	AuthorizationHeaderKey = "Authorization"

	// AuthorizationTypeBearer 是我们支持的认证类型
	// Bearer Token 是 OAuth 2.0 标准的认证方式
	AuthorizationTypeBearer = "bearer"

	// AuthorizationPayloadKey 是存储在 Gin Context 中的 payload 键名
	// Handler 可以通过这个 key 获取当前登录用户的信息
	AuthorizationPayloadKey = "authorization_payload"
)

// ==================== 中间件实现 ====================

// AuthMiddleware 创建一个 JWT 认证中间件
//
// 工作流程:
//  1. 从 HTTP 头获取 Authorization
//  2. 验证格式是否为 "Bearer <token>"
//  3. 使用 TokenMaker 验证 token 有效性
//  4. 将 payload 存入 Gin Context，供后续 Handler 使用
//
// 参数:
//   - tokenMaker: JWT token 验证器接口
//
// 返回:
//   - gin.HandlerFunc: Gin 中间件函数
//
// 使用示例:
//
//	router.Use(AuthMiddleware(tokenMaker))
//	// 或者只对特定路由组使用
//	authRoutes := router.Group("/").Use(AuthMiddleware(tokenMaker))
func AuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	// 返回一个闭包函数，捕获 tokenMaker 变量
	return func(c *gin.Context) {
		// Step 1: 获取 Authorization 请求头
		// 如果没有提供认证头，返回 401 Unauthorized
		authHeader := c.GetHeader(AuthorizationHeaderKey)
		if authHeader == "" {
			// 创建错误响应
			err := apperrors.New(apperrors.CodeUnauthorized)
			// AbortWithStatusJSON 会停止后续处理并返回 JSON 响应
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.NewErrorResponse(err))
			return
		}

		// Step 2: 解析认证头
		// 预期格式: "Bearer <token>"
		// strings.Fields 按空白字符分割字符串
		fields := strings.Fields(authHeader)
		if len(fields) != 2 {
			// 格式错误: 不是两部分
			err := apperrors.NewWithMessage(apperrors.CodeUnauthorized, "invalid authorization header format")
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.NewErrorResponse(err))
			return
		}

		// Step 3: 验证认证类型
		// 转为小写进行比较，支持大小写不敏感
		authType := strings.ToLower(fields[0])
		if authType != AuthorizationTypeBearer {
			// 不支持的认证类型 (我们只支持 Bearer)
			err := apperrors.NewWithMessage(apperrors.CodeUnauthorized, "unsupported authorization type")
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.NewErrorResponse(err))
			return
		}

		// Step 4: 提取并验证 Token
		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			// Token 验证失败 (过期、无效签名等)
			// 根据错误类型返回不同的错误码
			var appErr *apperrors.AppError
			if err == token.ErrExpiredToken {
				appErr = apperrors.New(apperrors.CodeTokenExpired)
			} else {
				appErr = apperrors.New(apperrors.CodeInvalidToken)
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.NewErrorResponse(appErr))
			return
		}

		// Step 5: 将 payload 存入 Context
		// 后续的 Handler 可以通过 c.MustGet(AuthorizationPayloadKey) 获取
		c.Set(AuthorizationPayloadKey, payload)

		// Step 6: 调用下一个处理器
		// c.Next() 继续处理链中的下一个中间件或 Handler
		c.Next()
	}
}

// GetAuthPayload 从 Gin Context 中获取认证 payload
//
// 这是一个辅助函数，简化 Handler 中获取当前用户信息的代码
//
// 参数:
//   - c: Gin Context
//
// 返回:
//   - *token.Payload: 当前用户的 Token payload
//   - bool: 是否成功获取 (如果 Context 中没有 payload 则为 false)
//
// 使用示例:
//
//	func (h *AccountHandler) CreateAccount(c *gin.Context) {
//	    payload, ok := middleware.GetAuthPayload(c)
//	    if !ok {
//	        // 理论上不会发生，因为已经过了 AuthMiddleware
//	        c.JSON(http.StatusUnauthorized, ...)
//	        return
//	    }
//	    username := payload.Username
//	}
func GetAuthPayload(c *gin.Context) (*token.Payload, bool) {
	// c.Get 返回值和是否存在的布尔值
	value, exists := c.Get(AuthorizationPayloadKey)
	if !exists {
		return nil, false
	}

	// 类型断言: 将 interface{} 转换为 *token.Payload
	payload, ok := value.(*token.Payload)
	if !ok {
		return nil, false
	}

	return payload, true
}

// MustGetAuthPayload 从 Gin Context 中获取认证 payload
//
// 与 GetAuthPayload 类似，但在获取失败时会 panic
// 仅在确定已经过 AuthMiddleware 的路由中使用
//
// 参数:
//   - c: Gin Context
//
// 返回:
//   - *token.Payload: 当前用户的 Token payload
//
// 使用示例:
//
//	func (h *AccountHandler) CreateAccount(c *gin.Context) {
//	    payload := middleware.MustGetAuthPayload(c)
//	    username := payload.Username
//	}
func MustGetAuthPayload(c *gin.Context) *token.Payload {
	// c.MustGet 在 key 不存在时会 panic
	return c.MustGet(AuthorizationPayloadKey).(*token.Payload)
}
