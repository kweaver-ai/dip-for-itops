package controller

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"net/http"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/DE_go-lib/rest"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/domain"
)

var (
	ModuleName = "AutoItOpsAlertManager"
	HTTPError  = map[string]ErrorInfo{
		// 500
		"InternalError": {
			httpCode:  http.StatusInternalServerError,
			errorCode: AutoReportMANAGER_InternalError_Error,
		},
		"GenerateIDFailed": {
			httpCode:  http.StatusInternalServerError,
			errorCode: ModuleName + ".InternalError.GenerateIDFailed",
		},
		"DataConvertFailed": {
			httpCode:  http.StatusInternalServerError,
			errorCode: ModuleName + ".InternalError.DataConvertFailed",
		},
		"ExecuteSqlError": {
			httpCode:  http.StatusInternalServerError,
			errorCode: ModuleName + ".InternalError.ExecuteSqlError",
		},
		"ClientRequestError": {
			httpCode:  http.StatusInternalServerError,
			errorCode: ModuleName + ".InternalError.ClientRequestError",
		},
		// 400
		"ValidateParamError": {
			httpCode:  http.StatusBadRequest,
			errorCode: ModuleName + ".InvalidParameter.%sInvalidParameter",
		},
		"NameExisted": {
			httpCode:  http.StatusBadRequest,
			errorCode: ModuleName + ".BadRequest.NameExisted",
		},
		"Unauthorized": {
			httpCode:  http.StatusUnauthorized,
			errorCode: ModuleName + ".BadRequest.Unauthorized",
		},
		// http错误 404
		"NotFound": {
			httpCode:  http.StatusNotFound,
			errorCode: ModuleName + ".NotFound.Data",
		},
	}
	InvalidParameter = ErrorInfo{
		httpCode:  http.StatusBadRequest,
		errorCode: AutoAlertManager_BadRequest_InvalidParameter,
	}
)

type ErrorInfo struct {
	httpCode  int
	errorCode string
	format    map[string]interface{}
}

func (errInfo *ErrorInfo) WithFormat(format map[string]interface{}) *ErrorInfo {
	errInfo.format = format
	return errInfo
}

func HandleValidateError(ctx context.Context, err error) *rest.HTTPError {
	for _, e := range err.(validator.ValidationErrors) {
		errInfo := HTTPError["ValidateParamError"]
		errInfo.errorCode = fmt.Sprintf(errInfo.errorCode, e.StructField())
		// errInfo.WithFormat(map[string]interface{}{
		// 	"param": e.Param(),
		// })
		// return NewRestHTTPError(ctx, errInfo).WithErrorDetails(err.Error())
		return NewRestHTTPError(ctx, errInfo)
	}
	// return NewRestHTTPError(ctx, InvalidParameter).WithErrorDetails(err.Error())
	return NewRestHTTPError(ctx, InvalidParameter)
}

func NewRestHTTPError(ctx context.Context, info ErrorInfo) *rest.HTTPError {
	return rest.NewHTTPError(ctx, info.httpCode, info.errorCode)
}

func HandDomainError(ctx context.Context, err domain.DomainError) *rest.HTTPError {
	return NewRestHTTPError(ctx, HTTPError[err.Type()])
}
