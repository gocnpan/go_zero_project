package conf

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/zeromicro/go-zero/core/mapping"
)

var loaders = map[string]func([]byte, interface{}) error{
	".json": LoadConfigFromJsonBytes,
	".yaml": LoadConfigFromYamlBytes,
	".yml":  LoadConfigFromYamlBytes,
}

// LoadConfig loads config into v from file, .json, .yaml and .yml are acceptable.
func LoadConfig(file string, v interface{}, opts ...Option) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	// path.Ext 获取文件尾缀
	loader, ok := loaders[path.Ext(file)] // 根据不同的文件类型，赋值不同的方法处理
	if !ok {
		return fmt.Errorf("unrecognized file type: %s", file)
	}

	var opt options
	for _, o := range opts {
		o(&opt)
	}

	if opt.env {
		return loader([]byte(os.ExpandEnv(string(content))), v)
	}

	return loader(content, v)
}

// LoadConfigFromJsonBytes loads config into v from content json bytes.
func LoadConfigFromJsonBytes(content []byte, v interface{}) error {
	return mapping.UnmarshalJsonBytes(content, v)
}

// LoadConfigFromYamlBytes loads config into v from content yaml bytes.
func LoadConfigFromYamlBytes(content []byte, v interface{}) error {
	return mapping.UnmarshalYamlBytes(content, v)
}

// MustLoad loads config into v from path, exits on error.
func MustLoad(path string, v interface{}, opts ...Option) {
	if err := LoadConfig(path, v, opts...); err != nil {
		log.Fatalf("error: config file %s, %s", path, err.Error())
	}
}
