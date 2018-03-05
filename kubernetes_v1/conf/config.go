package conf

import (
	"bufio"
	"io"
	"os"
	"strings"
)

type Config struct {
	configMap map[string]string
}

func (c *Config) InitConfig(path *string) {
	c.configMap = make(map[string]string)
	file, err := os.Open(*path)
	if err != nil {
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		d, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}
		s := strings.TrimSpace(string(d))
		if strings.Index(s, "#") == 0 {
			continue
		}
		index := strings.Index(s, "=")
		if index < 0 {
			continue
		}
		key := strings.TrimSpace(s[:index])
		if len(key) == 0 {
			continue
		}
		value := strings.TrimSpace(s[index+1:])
		index = strings.Index(value, "\t#")
		if index > -1 {
			value = value[:index]
		}
		index = strings.Index(value, " #")
		if index > -1 {
			value = value[:index]
		}
		index = strings.Index(value, "\t//")
		if index > -1 {
			value = value[:index]
		}
		index = strings.Index(value, " //")
		if index > -1 {
			value = value[:index]
		}
		if len(value) == 0 {
			continue
		}
		c.configMap[key] = value
	}
}

func (c *Config) Read(key string) string {
	value, found := c.configMap[key]
	if !found {
		return ""
	}
	return value
}
