package inputModules

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"goshort/kernel"
	"goshort/modules/dbModules"
	"goshort/types"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var NotImplementedError = errors.New("not implemented")

func getDefaultFunc(_ string) (types.Url, error) {
	return types.Url{}, NotImplementedError
}

func postDefaultFunc(_ types.Url) (types.Url, error) {
	return types.Url{}, NotImplementedError
}

func patchDefaultFunc(_ types.Url) error {
	return NotImplementedError
}

func deleteDefaultFunc(_ string) error {
	return NotImplementedError
}

func genericKeySupportDefaultFunc() bool {
	return false
}

func genericKeySupportTrueFunc() bool {
	return true
}

func TestSimpleGet(t *testing.T) {
	url := types.Url{Key: "testKey", Url: "http://example.com", Code: 301, Autogenerated: false}

	getFunc := func(key string) (types.Url, error) {
		assert.Equal(t, "testKey", key)
		return url, nil
	}

	var kernelInstance kernel.Kernel
	server := Server{Kernel: &kernelInstance}

	db := dbModules.Generic{
		GetFunc:               getFunc,
		PostFunc:              postDefaultFunc,
		PatchFunc:             patchDefaultFunc,
		DeleteFunc:            deleteDefaultFunc,
		GenericKeySupportFunc: genericKeySupportDefaultFunc,
		Name:                  "Generic",
	}

	kernelInstance = kernel.Kernel{}

	kernelInstance.Logger = &kernel.LoggingKernel{Kernel: &kernelInstance}
	kernelInstance.Database = &kernel.DatabaseKernel{Kernel: &kernelInstance, Database: &db}
	kernelInstance.Input = &kernel.InputKernel{Kernel: &kernelInstance, Inputs: []types.InputInterface{&server}}
	kernelInstance.Middleware = &kernel.MiddlewareKernel{Kernel: &kernelInstance}
	kernelInstance.Reconnection = kernel.ReconnectionKernel{Kernel: &kernelInstance}
	kernelInstance.Signal = kernel.SignalKernel{Kernel: &kernelInstance}

	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
	rec := httptest.NewRecorder()

	server.echo = echo.New()
	context := server.echo.NewContext(req, rec)
	context.SetPath("/api/urls/:id/")
	context.SetParamNames("id")
	context.SetParamValues("testKey")

	if assert.NoError(t, server.urlsGetHandler(context)) {
		var resUrl types.Url
		_ = json.NewDecoder(rec.Body).Decode(&resUrl)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, url, resUrl)
	}
}

func postTestHelper(t *testing.T, postFunc func(newUrl types.Url) (types.Url, error), genericKeySupport func() bool, inputString string, comparingUrl types.Url) {
	var kernelInstance kernel.Kernel
	server := Server{Kernel: &kernelInstance}

	db := dbModules.Generic{
		GetFunc:               getDefaultFunc,
		PostFunc:              postFunc,
		PatchFunc:             patchDefaultFunc,
		DeleteFunc:            deleteDefaultFunc,
		GenericKeySupportFunc: genericKeySupport,
		Name:                  "Generic",
	}

	kernelInstance = kernel.Kernel{}

	kernelInstance.Logger = &kernel.LoggingKernel{Kernel: &kernelInstance}
	kernelInstance.Database = &kernel.DatabaseKernel{Kernel: &kernelInstance, Database: &db}
	kernelInstance.Input = &kernel.InputKernel{Kernel: &kernelInstance, Inputs: []types.InputInterface{&server}}
	kernelInstance.Middleware = &kernel.MiddlewareKernel{Kernel: &kernelInstance}
	kernelInstance.Reconnection = kernel.ReconnectionKernel{Kernel: &kernelInstance}
	kernelInstance.Signal = kernel.SignalKernel{Kernel: &kernelInstance}

	req := httptest.NewRequest(http.MethodPost, "/api/urls/", strings.NewReader(inputString))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	server.echo = echo.New()
	context := server.echo.NewContext(req, rec)

	if assert.NoError(t, server.urlsPostHandler(context)) {
		var resUrl types.Url
		_ = json.NewDecoder(rec.Body).Decode(&resUrl)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, comparingUrl, resUrl)
	}
}

func TestSimplePost(t *testing.T) {
	postFunc := func(newUrl types.Url) (types.Url, error) {
		assert.Equal(t, types.Url{Key: "aaa", Url: "https://yandex.ru", Code: 301, Autogenerated: false}, newUrl)
		return newUrl, nil
	}
	postTestHelper(t, postFunc, genericKeySupportDefaultFunc, `{"url":"https://yandex.ru","key":"aaa"}`,
		types.Url{Key: "aaa", Url: "https://yandex.ru", Code: 301, Autogenerated: false})
}

func TestGenericPost(t *testing.T) {
	postFunc := func(newUrl types.Url) (types.Url, error) {
		assert.Equal(t, types.Url{Key: "", Url: "https://yandex.ru", Code: 301, Autogenerated: true}, newUrl)
		newUrl.Key = "a"
		return newUrl, nil
	}
	postTestHelper(t, postFunc, genericKeySupportTrueFunc, `{"url":"https://yandex.ru"}`,
		types.Url{Key: "a", Url: "https://yandex.ru", Code: 301, Autogenerated: true})
}
