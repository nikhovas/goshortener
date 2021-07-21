package middlewareModules

import (
	"fmt"
	"github.com/PuerkitoBio/purell"
	"goshort/kernel"
	"goshort/types"
	"sync"
)

type CantNormalizeError struct {
	Url     string
	Details error
}

func (e *CantNormalizeError) ToMap() map[string]interface{} {
	data := make(map[string]interface{})
	data["name"] = "Middleware.UrlNormalize.CantNormalize"
	data["type"] = "error"
	data["url"] = e.Url
	data["details"] = e.Details.Error()
	return data
}

func (e *CantNormalizeError) Error() string {
	return fmt.Sprintf("Error Middleware.UrlNormalize.CantNormalize url=%s details=%s", e.Url, e.Details.Error())
}

type UrlNormalizer struct {
	types.ModuleBase
	breakOnError bool
	Kernel       *kernel.Kernel
	Name         string
}

func CreateUrlNormalizer(kernel *kernel.Kernel) types.MiddlewareInterface {
	return &UrlNormalizer{Kernel: kernel}
}

func (middleware *UrlNormalizer) Init(config map[string]interface{}) error {
	middleware.breakOnError = config["break_on_error"].(bool)
	return nil
}

func (middleware *UrlNormalizer) Run(wg *sync.WaitGroup) error {
	wg.Done()
	return nil
}

func (middleware *UrlNormalizer) Stop() error {
	return nil
}

func (middleware *UrlNormalizer) Exec(url *types.Url) error {
	if url.Code == 0 {
		url.Code = 301
	}

	if url.Key == "" {
		url.Autogenerated = true
	}

	newUrl, err := purell.NormalizeURLString(url.Url,
		purell.FlagLowercaseScheme|purell.FlagLowercaseHost|purell.FlagUppercaseEscapes)
	if err != nil {
		return &CantNormalizeError{Url: url.Url, Details: err}
	}

	url.Url = newUrl
	return nil
}

func (middleware *UrlNormalizer) GetName() string {
	return middleware.Name
}

func (middleware *UrlNormalizer) GetType() string {
	return "UrlNormalizer"
}

func (middleware *UrlNormalizer) BreakOnError() bool {
	return middleware.breakOnError
}
