# Admin Passwordの環境変数での変更ができない
- https://github.com/grafana/grafana/issues/49055
- https://community.grafana.com/t/potential-bug-admin-credentials-set-via-env-variables-are-not-picked-up-even-when-docker-container-is-brought-down-and-up-again/65794
- Admin Passwordを初期設定後、`GF_SECURITY_ADMIN_PASSWORD`環境変数でPasswordを変えても反映されない
- 2025/07時点ではまだ修正されていない