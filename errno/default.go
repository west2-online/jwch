package errno

var (
	// Success
	Success = NewErrNo(SuccessCode, "Success")

	ServiceError             = NewErrNo(ServiceErrorCode, "service is unable to start successfully")
	ServiceInternalError     = NewErrNo(ServiceErrorCode, "service Internal Error")
	ParamError               = NewErrNo(ParamErrorCode, "parameter error")
	AuthorizationFailedError = NewErrNo(AuthorizationFailedErrCode, "suthorization failed")

	// User
	AccountConflictError  = NewErrNo(AuthorizationFailedErrCode, "account conflict")
	SessionExpiredError   = NewErrNo(AuthorizationFailedErrCode, "session expired")
	LoginCheckFailedError = NewErrNo(AuthorizationFailedErrCode, "login check failed")
	SSOLoginFailedError   = NewErrNo(AuthorizationFailedErrCode, "sso login failed")
	GetSessionFailedError = NewErrNo(AuthorizationFailedErrCode, "get session failed")

	// HTTP
	HTTPQueryError = NewErrNo(HTTPQueryErrorCode, "HTTP query failed")
	HTMLParseError = NewErrNo(HTTPQueryErrorCode, "HTML parse failed")
)
