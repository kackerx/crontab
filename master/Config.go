package master

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
)

type Config struct {
    ApiPort         int      `json:"apiPort"`
    ApiReadTimeout  int      `json:"apiReadTimeout"`
    ApiWriteTimeout int      `json:"apiWriteTimeout"`
    EtcdEndpoints   []string `json:"etcdEndpoints"`
    EtcdDialTimeout int      `json:"etcdDialTimeout"`
    WebRoot         string   `json:"webroot"`
}

var (
    // G_config 单例
    G_config *Config
)

func InitConfig(filename string) (err error) {
    var (
        content []byte
        conf    Config
    )

    if content, err = ioutil.ReadFile(filename); err != nil {
        fmt.Println(err)
        return
    }

    // 2 Json反序列化
    if err = json.Unmarshal(content, &conf); err != nil {
        fmt.Println(err)
        return
    }

    // 3 赋值单例
    G_config = &conf
    return
}
