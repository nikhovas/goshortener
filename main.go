package main

import (
	"bytes"
	"github.com/spf13/viper"
	"goshort/src/kernel"
	"goshort/src/modules/dbModules"
	"goshort/src/modules/inputModules"
	"goshort/src/modules/logModules"
	"goshort/src/types"
	"io/ioutil"
	"log"
	"os"
)

// @Title Goshort Swagger API
// @Version 1.0
// @Description Swagger API for Golang Project Goshort.
// @TermsOfService http://swagger.io/terms/

// @Contact.name Nikolay Vasilev
// @Contact.email nikhovas@yandex.ru

// @License.name MIT
// @License.url https://github.com/nikhovas/goshort/blob/master/LICENSE

// @BasePath /api/

var yamlExample = []byte(`
inputs:
 server:
   name: Server
   ip: ''
   port: 80
   mode: tcp
database:
 redis:
  name: Redis
  ip: 127.0.0.1:6379
  port: 6379
  mode: tcp
  pool_size: 10
loggers:
 console:
   name: consoleLogger
   extra_logger: true
   common_logger: true
middlewares:
 - url_normalizer
limits:
 max_connections: 2000
`)

func main() {
	viper.SetConfigType("yaml")
	if len(os.Args) == 1 {
		_ = viper.ReadConfig(bytes.NewBuffer(yamlExample))
	} else {
		fileName := os.Args[1]
		file, err := os.Open(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err = file.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		b, err := ioutil.ReadAll(file)
		_ = viper.ReadConfig(bytes.NewBuffer(b))
	}

	kernelInstance := kernel.Kernel{
		InputCreators: map[string]func(kernel *kernel.Kernel) types.InputInterface{
			"server": inputModules.CreateServer,
		},
		LoggerCreators: map[string]func(kernel types.KernelInterface) types.LoggerInterface{
			"console": logModules.CreateConsole,
			"kafka":   logModules.CreateKafka,
		},
		DatabaseCreators: map[string]func(kernel *kernel.Kernel) types.DatabaseInterface{
			"redis":     dbModules.CreateRedis,
			"in_memory": dbModules.CreateInMemory,
		},
		MiddlewareCreators: map[string]func(kernel *kernel.Kernel) types.MiddlewareInterface{},
	}

	config := map[string]interface{}{
		"inputs":      viper.GetStringMap("inputs"),
		"database":    viper.GetStringMap("database"),
		"loggers":     viper.GetStringMap("loggers"),
		"middlewares": viper.GetStringMap("middlewares"),
	}

	_ = kernelInstance.Init(config)

	err := kernelInstance.Run(nil)
	if err != nil {
		print(err.Error())
	}
}
