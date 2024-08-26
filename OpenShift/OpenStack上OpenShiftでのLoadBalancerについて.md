- OpenStack上のOpenShift環境で`LoadBalancer`タイプの`Service`を作成すると、OpenStack上でOctavia LoadBalancerが作成される
  - **`[LoadBalancer]`セクションの`enabled`の値を`false`（defaultは`true`）にしたらOctavia LoadBalancerが作成されないようにできる**
  - https://github.com/kubernetes/cloud-provider-openstack/blob/master/docs/octavia-ingress-controller/using-octavia-ingress-controller.md
  - https://github.com/kubernetes/cloud-provider-openstack/blob/master/docs/openstack-cloud-controller-manager/using-openstack-cloud-controller-manager.md
- `openshift-cloud-controller-manager` namespaceに`openstack-cloud-controller-manager`という`Deployments`(`Pod`)があり、
そいつが`LoadBalancer`タイプの`Service`を検知してOctavia LoadBalancerを作成する
- `openshift-cloud-controller-manager` namespaceに`cloud-conf`という`ConfigMap`があり、Openstack関連Confが設定されている
`openstack-cloud-controller-manager` Pod内に`/etc/openstack/config/cloud.conf`に配置され、さらにそこから`/etc/openstack/secret/clouds.yaml`を参照している
  - clouds.yamlの内容は`/etc/openstack/secret/clouds.yaml`に記載されている
  - `cloud-conf` ConfigMapの`[LoadBalancer]`セッションでLoadBalancer（Octavia）関連設定ができる