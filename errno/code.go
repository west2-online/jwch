package errno

const (
	// Success
	SuccessCode = 10000
	SuccessMsg  = "success"

	// Error
	ServiceErrorCode           = 10001 // 未知微服务错误
	ParamErrorCode             = 10002 // 参数错误
	HTTPQueryErrorCode         = 10003 // HTTP请求出错
	AuthorizationFailedErrCode = 10004 // 鉴权失败
	UnexpectedTypeErrorCode    = 10005 // 未知类型
	NotImplementErrorCode      = 10006 // 未实装

)
