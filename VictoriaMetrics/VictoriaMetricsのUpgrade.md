- VictoriaMetricsは基本(Release Notesで特に言及がなければ)安全にUpgrade/Downgradeすることができる
  > It is safe to upgrade VictoriaMetrics to new versions unless the release notes say otherwise. It is safe to skip multiple versions during the upgrade unless release notes say otherwise. It is recommended to perform regular upgrades to the latest version, since it may contain important bug fixes, performance optimizations or new features.
  >
  > It is also safe to downgrade to the previous version unless release notes say otherwise.
  >
  > The following steps must be performed during the upgrade / downgrade procedure:
  >
  > - Send SIGINT signal to VictoriaMetrics process so that it is stopped gracefully.
  > - Wait until the process stops. This can take a few seconds.
  > - Start the upgraded VictoriaMetrics.
- 参考URL
  - https://docs.victoriametrics.com/BestPractices.html#upgrade-procedure