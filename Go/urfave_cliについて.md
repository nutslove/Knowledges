- CLIの開発を容易にし、より構造化された方法でコマンドラインオプションや引数を扱うことができるようにするライブラリ
- https://github.com/urfave/cli/
- v2とv3があるが、v3は2024/05の段階でまだalpha
- v2の場合、`github.com/urfave/cli/v2`パッケージをimportすることで、そのパッケージ内で定義された型、関数、メソッド等が利用可能になり、`cli.App`という型を利用することができる。

# v2
## 使い方
- アプリ名など、任意の名前の変数で`&cli.App{<CLI設定(e.g. Flags)>}`を初期化し、`<変数名>.Run(os.Args)`でエラーが発生するか確認するだけ

### ■ `Flags`
- CLI実行時に引き渡す引数
#### `Flags`の各項目の説明
- `Name`
  - このフラグの名前で、コマンドラインからアプリケーションを起動するときに使用する。  
    例えば`listen-address`と定義した場合、CLI実行時`--listen-address=5001`のように使う。
- `Required`
  - このフラグの指定が必須かどうかを定義。`true`の場合、フラグが指定されてないとエラーになる。
- `Value`
  - このフラグのデフォルト値。コマンドラインから明示的な値が提供されない場合、このデフォルト値が使用される。
- `Usage`
  - このフラグの説明。通常はヘルプメッセージで表示され、ユーザーがフラグが何を行うかを理解できるようにする。
- `Destination`
  - このフラグから取得した値を格納する変数へのポインタ。下の例では、`listen_port`という変数にフラグの値が格納されます。
- `EnvVars`
  - 環境変数からフラグの値を取得できるようにするための環境変数名のリスト。この例では、`listen-address`という名前の環境変数から値を取得できる。  
    コマンドラインフラグが設定されていない場合、そして環境変数が設定されている場合、この環境変数の値が使用される。  
	  `Value`はデフォルト値で、`EnvVars`から渡される値やコマンドラインから渡される値でデフォルトのValueが上書きされる。

- コマンド実行時`Flags`に指定された値を取得するためには、`cli.Context`の`String`や`Bool`などのメソッドを使用する。  
  例えば、文字列フラグの場合は`c.String("flag-name")`、boolフラグの場合は`c.Bool("flag-name")`を使用する。

- 設定例
  ~~~go
	package main

	import (
		"log"
		"os"

		"github.com/urfave/cli/v2"
	)

	func main() {
		app := &cli.App{
			Name:  "ham3",
			Usage: "CLI for Ham3",
			Commands: []*cli.Command{
				{
					Name:  "caas",
					Usage: "Container as a service",
					Subcommands: []*cli.Command{
						{
							Name:   "create",
							Usage:  "Create a caas cluster",
							Action: CaasCreate,
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:     "cluster-name",
									Aliases:  []string{"c"},
									Usage:    "Name of the cluster",
									Required: true,
								},
							},
						},
						{
							Name:   "delete",
							Usage:  "Delete a caas cluster",
							Action: CaasDelete,
							Flags: []cli.Flag{
								&cli.BoolFlag{
									Name:    "force",
									Aliases: []string{"f"},
									Usage:   "Force delete the cluster",
								},
							},
						},
					},
				},
			},
		}

		err := app.Run(os.Args)
		if err != nil {
			log.Fatal(err)
		}
	}

	func CaasCreate(c *cli.Context) error {
		// フラグの値を取得
		clusterName := c.String("cluster-name")
		if clusterName == "" {
			return cli.Exit("Error: cluster-name is required", 1)
		}
		log.Printf("Creating caas cluster: %s", clusterName)
		return nil
	}

	func CaasDelete(c *cli.Context) error {
		// フラグの値を取得
		force := c.Bool("force")
		log.Printf("Deleting caas cluster with force: %v", force)
		return nil
	}
  ~~~

### ■ `Commands`
- アプリが提供するサブコマンドのリストを保持
- それぞれのサブコマンドは`cli.Command`構造体のインスタンスとして表され、さまざまなプロパティ（`Name`、`Aliases`、`Usage`、`Action`など）を含むことができる
- 設定例 (https://github.com/urfave/cli/blob/main/docs/v2/examples/combining-short-options.md)  
  **以下の例の場合コマンドライン実行時に`-s`,`-o`,`-m`または`-som`で指定できる**
  ~~~go
  package main

  import (
  	"fmt"
  	"log"
  	"os"

  	"github.com/urfave/cli/v2"
  )

  func main() {
  	app := &cli.App{
  		UseShortOptionHandling: true,
  		Commands: []*cli.Command{
  			{
  				Name:  "short",
  				Usage: "complete a task on the list",
  				Flags: []cli.Flag{
  					&cli.BoolFlag{Name: "serve", Aliases: []string{"s"}},
  					&cli.BoolFlag{Name: "option", Aliases: []string{"o"}},
  					&cli.StringFlag{Name: "message", Aliases: []string{"m"}},
  				},
  				Action: func(c *cli.Context) error {
  					fmt.Println("serve:", c.Bool("serve"))
  					fmt.Println("option:", c.Bool("option"))
  					fmt.Println("message:", c.String("message"))
  					return nil
  				},
  			},
  		},
  	}

  	if err := app.Run(os.Args); err != nil {
  		log.Fatal(err)
  	}
  }
	~~~  
	> If you enable `UseShortOptionHandling`, then you must not use any flags that have a single leading `-` or this will result in failures. For example, `-option` can no longer be used. Flags with two leading dashes (such as `--options`) are still valid.

### ■ `Action`
- アプリが実行された時に実行する関数を指定
- この関数は一般に`func(c *cli.Context) error`というシグニチャを持ち、`cli.Context`オブジェクトをパラメータとして受け取る。  
  このオブジェクトを使って、コマンドラインからの入力（例：フラグや引数）を取得できる。
- 設定例
  ~~~go
  func GetMetrics() {
  	for {
  		for ns, v := range namespaces {
  			time.Sleep(1 * time.Second)

  			switch ns {
  			case "oci_computeagent":
  				for i, _ := range v["metricname"] {
  					GetMetric(ns, v["queries"][i], oci_computeagent_gaugevec[i])
  				}
  			case "oci_blockstore":
  				for i, _ := range v["metricname"] {
  					GetMetric(ns, v["queries"][i], oci_blockstore_gaugevec[i])
  				}
  			}
  		}
  		time.Sleep(60 * time.Second)
  	}
  }

  func main() {
  	oci_metrics_exporter := &cli.App{
  		Name:  "OCI Metrics Exporter",
  		Usage: "OCI Metrics Exporter for Prometheus",
  		Flags: []cli.Flag{
  			&cli.StringFlag{
  				Name:        "listen-port",
  				Aliases:     []string{"p"},
  				Usage:       "listen port",
  				Value:       "8080",
  				Destination: &listen_port,
  				EnvVars:     []string{"LISTEN_PORT"},
  			},
  		},
  		Action: func(c *cli.Context) error {
  			fmt.Println("Start OCI Metrics Exporter")
  			go GetMetrics()
  			return nil -------------------> 戻り値がerrorになっているのでこれが必要
  		},
  	}

  	if err := oci_metrics_exporter.Run(os.Args); err != nil {
  		log.Fatal(err)
  	}

  	http.Handle("/metrics", promhttp.Handler())
  	log.Fatal(http.ListenAndServe(":"+listen_port, nil))
  }
	~~~