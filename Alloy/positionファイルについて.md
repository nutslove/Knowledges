- 参考URL
  - https://grafana.com/docs/alloy/latest/set-up/migrate/from-promtail/
  - https://grafana.com/docs/alloy/latest/reference/components/loki/loki.source.file/

# Positionファイルとは
- Alloyがログを収集する際に、どこまでログを読み取ったかを記録するためのファイル。
- これにより、Alloyが再起動した場合や、ログの収集が一時的に停止した場合でも、前回の位置からログの収集を再開することができる。