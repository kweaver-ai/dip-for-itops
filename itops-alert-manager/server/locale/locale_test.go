package locale

import (
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/kweaver-ai/kweaver-go-lib/i18n"
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
