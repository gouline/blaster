package utils

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
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
	ctx := &gin.Context{
		Request: &http.Request{
			Host: "test.example.com",
		},
	}
	if RelativeURI(ctx, "/context/path") != "http://test.example.com/context/path" {
		t.Errorf("Unexpected relative URI")
	}
}
