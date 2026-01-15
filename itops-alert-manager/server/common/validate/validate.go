package validate

import (
	"strings"
	"unicode/utf8"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/common"
	"github.com/go-playground/validator/v10"
)

// TagsValidation 校验标签
func TagsValidation(fl validator.FieldLevel) bool {
	tags := fl.Field().Interface().([]string)
	if len(tags) > 5 {
		return false
	}
	for _, tag := range tags {
		if tag == "" {
			return false
		}
		if utf8.RuneCountInString(tag) > 40 {
			return false
		}
		if bo := strings.ContainsAny(common.SpecialCharacters, tag); bo {
			return false
		}
	}
	return true
}

// FieldValidation 校验特殊字符
func FieldValidation(fl validator.FieldLevel) bool {
	fieldValue := fl.Field().Interface().(string)
	if utf8.RuneCountInString(fieldValue) > 40 {
		return false
	}
	// 不能包含特殊字符
	if bo := strings.ContainsAny(common.SpecialCharacters, fieldValue); bo {
		return false
	}
	return true
}
