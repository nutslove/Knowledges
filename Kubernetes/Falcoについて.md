## Falcoとは
- https://falco.org/docs/  
  > Falco is a cloud native runtime security tool for Linux operating systems. It is designed to detect and alert on abnormal behavior and potential security threats in real-time.
  >
  > At its core, Falco is a kernel monitoring and detection agent that observes events, such as syscalls, based on custom rules. Falco can enhance these events by integrating metadata from the container runtime and Kubernetes. The collected events can be analyzed off-host in SIEM or data lake systems.

- Falcoのディレクトリは`/etc/falco/`
- defaultのruleは`/etc/falco/falco_rules.yaml`
  - 追加/上書きするruleは`/etc/falco/falco_rules.local.yaml`に記述
- rule修正/追加後の反映は`systemctl restart falco`
- FalcoのOutputなど、一般的な設定は`/etc/falco/falco.yaml`に定義
  - https://falco.org/docs/outputs/
  - fileにoutputを書き出すためには`file_output`
    - https://falco.org/docs/reference/daemon/config-options/