## kubernetesを操作するための権限取得
### pod内で実行される場合
- `rest.InClusterConfig()`を利用
- `rest.InClusterConfig()`は以下の操作を行う
   1. 環境変数 `KUBERNETES_SERVICE_HOST` と `KUBERNETES_SERVICE_PORT` からAPI Serverのホスト名とポートを取得
   2. Kubernetes APIに対して認証を行うために必要なServiceAccountのTokenをPod内のファイルシステム上の`/var/run/secrets/kubernetes.io/serviceaccount/token`を取得
   3. API Serverの証明書を検証するためのCA証明書を、Pod内のファイルシステム上の `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`から取得 
- 例  
  ```go
  package main

  import (
      "context"
      "fmt"

      v1 "k8s.io/api/core/v1"
      metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
      "k8s.io/client-go/kubernetes"
      "k8s.io/client-go/rest"
  )

  func main() {
      // Kubernetesのクラスター設定を取得
      config, err := rest.InClusterConfig()
      if err != nil {
          fmt.Printf("Error creating in-cluster config: %v\n", err)
          return
      }

      // Kubernetesクライアントの作成
      clientset, err := kubernetes.NewForConfig(config)
      if err != nil {
          fmt.Printf("Error creating Kubernetes client: %v\n", err)
          return
      }

      // Namespaceを作成するマニフェストの定義
      namespace := &v1.Namespace{
          ObjectMeta: metav1.ObjectMeta{
              Name: "example-namespace",
          },
      }

      // NamespaceをKubernetesクラスターに適用
      _, err = clientset.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
      if err != nil {
          fmt.Printf("Error creating namespace: %v\n", err)
          return
      }

      fmt.Println("Namespace created successfully")
  }
  ```

### サーバ上で実行される場合
- kubeconfigファイル(`~/.kube/config`)を読み込んで設定
- 例
  ```go
  package main

  import (
      "context"
      "flag"
      "fmt"
      "os"
      "path/filepath"

      v1 "k8s.io/api/core/v1"
      metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
      "k8s.io/client-go/kubernetes"
      "k8s.io/client-go/tools/clientcmd"
      "k8s.io/client-go/util/homedir"
  )

  func main() {
      // ホームディレクトリからkubeconfigのパスを取得
      var kubeconfig *string
      if home := homedir.HomeDir(); home != "" {
          kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(オプション) kubeconfigファイルのパス")
      } else {
          kubeconfig = flag.String("kubeconfig", "", "kubeconfigファイルのパス")
      }
      flag.Parse()

      // kubeconfigファイルを使用して設定をロード
      config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
      if err != nil {
          fmt.Printf("Error creating config from kubeconfig: %v\n", err)
          return
      }

      // Kubernetesクライアントの作成
      clientset, err := kubernetes.NewForConfig(config)
      if err != nil {
          fmt.Printf("Error creating Kubernetes client: %v\n", err)
          return
      }

      // Namespaceを作成するマニフェストの定義
      namespace := &v1.Namespace{
          ObjectMeta: metav1.ObjectMeta{
              Name: "example-namespace",
          },
      }

      // NamespaceをKubernetesクラスターに適用
      _, err = clientset.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
      if err != nil {
          fmt.Printf("Error creating namespace: %v\n", err)
          return
      }

      fmt.Println("Namespace created successfully")
  }
  ```