- `HostAliases`でPodの中の`/etc/hosts`にレコードを追加することができる
- サンプル
  ~~~yaml
  apiVersion: v1
  kind: Pod
  metadata:
    name: hostaliases-pod
  spec:
    restartPolicy: Never
    hostAliases:
    - ip: "127.0.0.1"
      hostnames:
      - "foo.local"
      - "bar.local"
    - ip: "10.1.2.3"
      hostnames:
      - "foo.remote"
      - "bar.remote"
    containers:
    - name: cat-hosts
      image: busybox:1.28
      command:
      - cat
      args:
      - "/etc/hosts"
  ~~~
  - 上記のPod内の`/etc/hosts`の内容は以下になる
    ~~~
    # Kubernetes-managed hosts file.
    127.0.0.1	localhost
    ::1	localhost ip6-localhost ip6-loopback
    fe00::0	ip6-localnet
    fe00::0	ip6-mcastprefix
    fe00::1	ip6-allnodes
    fe00::2	ip6-allrouters
    10.200.0.5	hostaliases-pod

    # Entries added by HostAliases.
    127.0.0.1	foo.local	bar.local
    10.1.2.3	foo.remote	bar.remote
    ~~~
- 参考URL
  - https://kubernetes.io/docs/tasks/network/customize-hosts-file-for-pods/