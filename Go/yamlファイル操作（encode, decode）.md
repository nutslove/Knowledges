> [!CAUTION]
> [`gopkg.in/yaml`パッケージ](https://github.com/go-yaml/yaml)はアーカイブされていて、メンテナンスされてないので、代わりに`"github.com/goccy/go-yaml"`（https://github.com/goccy/go-yaml）を使うこと

## `"github.com/goccy/go-yaml"`パッケージの使い方
- jsonと同様に、Goの構造体をYAMLにエンコードしたり、YAMLをGoの構造体にデコードすることができる
```go
package main
import (
    "github.com/goccy/go-yaml"
)

func main() {
    // YAMLファイルのパス
    ConfigFile := "config.yaml"

    // YAMLファイルを読み込む
    yamlFile, err := os.ReadFile(ConfigFile)
    if err != nil {
        panic(err)
    }

    // 構造体にデコードする
    var config map[string]interface{}
    err = yaml.Unmarshal(yamlFile, &config)
    if err != nil {
        panic(err)
    }

    // 構造体をYAMLにエンコードする
    encodedYaml, err := yaml.Marshal(config)
    if err != nil {
        panic(err)
    }

    // エンコードしたYAMLを出力
    fmt.Println(string(encodedYaml))

    // またはファイルに書き込む
    err = os.WriteFile("output.yaml", encodedYaml, 0644)
}
```
