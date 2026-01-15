package controller

import (
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/locale"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

const (
	AutoReportMANAGER_InternalError_Error        = "AutoItOpsAlertManager.InternalError.InternalError"
	AutoAlertManager_BadRequest_InvalidParameter = "AutoItOpsAlertManager.BadRequest.InvalidParameter"
)

const (
	// 400 Validate
	AutoItOpsAlertManager_InvalidParameter_Sort      = "AutoItOpsAlertManager.InvalidParameter.SortInvalidParameter"
	AutoItOpsAlertManager_InvalidParameter_Limit     = "AutoItOpsAlertManager.InvalidParameter.LimitInvalidParameter"
	AutoItOpsAlertManager_InvalidParameter_Offset    = "AutoItOpsAlertManager.InvalidParameter.OffsetInvalidParameter"
	AutoItOpsAlertManager_InvalidParameter_Direction = "AutoItOpsAlertManager.InvalidParameter.DirectionInvalidParameter"
	//400 内部校验
	AutoItOpsAlertManager_BadRequest_NameExisted  = "AutoItOpsAlertManager.BadRequest.NameExisted"
	AutoItOpsAlertManager_BadRequest_Unauthorized = "AutoItOpsAlertManager.BadRequest.Unauthorized"
	// 404
	AutoItOpsAlertManager_NotFound_Data = "AutoItOpsAlertManager.NotFound.Data"
	// 500
	AutoItOpsAlertManager_InternalError_GenerateIDFailed   = "AutoItOpsAlertManager.InternalError.GenerateIDFailed"
	AutoItOpsAlertManager_InternalError_DataConvertFailed  = "AutoItOpsAlertManager.InternalError.DataConvertFailed"
	AutoItOpsAlertManager_InternalError_ExecuteSqlError    = "AutoItOpsAlertManager.InternalError.ExecuteSqlError"
	AutoItOpsAlertManager_InternalError_ClientRequestError = "AutoItOpsAlertManager.InternalError.ClientRequestError"
)

var (
	errorCodeList = []string{
		AutoReportMANAGER_InternalError_Error,
		AutoAlertManager_BadRequest_InvalidParameter,
		AutoItOpsAlertManager_InvalidParameter_Limit,
		AutoItOpsAlertManager_InvalidParameter_Offset,
		AutoItOpsAlertManager_InvalidParameter_Sort,
		AutoItOpsAlertManager_InvalidParameter_Direction,
		AutoItOpsAlertManager_InternalError_GenerateIDFailed,
		AutoItOpsAlertManager_InternalError_DataConvertFailed,
		AutoItOpsAlertManager_InternalError_ExecuteSqlError,
		AutoItOpsAlertManager_BadRequest_NameExisted,
		AutoItOpsAlertManager_NotFound_Data,
		AutoItOpsAlertManager_BadRequest_Unauthorized,
		AutoItOpsAlertManager_InternalError_ClientRequestError,
	}
)

func init() {
	locale.Register()
	// 注册
	rest.Register(errorCodeList)
}
