package locale

import (
	"os"
	"testing"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/DE_go-lib/i18n"
	"github.com/agiledragon/gomonkey/v2"
)

func TestRegister(t *testing.T) {
	p1 := gomonkey.ApplyFunc(i18n.RegisterI18n, func(localeDir string) {
		// Do nothing
	})

	defer p1.Reset()

	os.Setenv("I18N_MODE_UT", "false")
	Register()
	os.Setenv("I18N_MODE_UT", "true")
	Register()
}
