package format

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestNewAllSymbolsRegexp(t *testing.T) {
	reg := NewAllSymbolsRegexp()

	if reg.ReplaceAllString("test s!t@r#i$n^g [1,2)", "") != "test string 12" {
		t.Errorf("Unexpected sanitization of \"test string 12\"")
	}

	if reg.ReplaceAllString("тестовая с!т@р#о$к^а (9.0]", "") != "тестовая строка 90" {
		t.Errorf("Unexpected sanitization of \"тестовая строка 90\"")
	}

	if reg.ReplaceAllString("测!试@字#符$串%5^6", "") != "测试字符串56" {
		t.Errorf("Unexpected sanitization of \"测试字符串56\"")
	}
}

func TestRelativeURI(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "https://test.example.com/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if actual := RelativeURI(c, "/context/path"); actual != "https://test.example.com/context/path" {
		t.Errorf("Unexpected relative URI: %s", actual)
	}
}
