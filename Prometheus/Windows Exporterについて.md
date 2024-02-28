- https://github.com/prometheus-community/windows_exporter
- Service(プロセス)の監視もできる
  - https://github.com/prometheus-community/windows_exporter/blob/master/docs/collector.service.md
  - 関連Metrics(`name`ラベルがService名)  
    | Name | Description | Type | Labels |
    | --- | --- | --- | --- |
    | `windows_service_info` | Contains service information in labels, constant 1 | gauge | name, display_name, process_id, run_as |
    | `windows_service_state` | The state of the service, 1 if the current state, 0 otherwise | gauge | name, state |
    | `windows_service_start_mode` | The start mode of the service, 1 if the current start mode, 0 otherwise | gauge | name, start_mode |
    | `windows_service_status` | The status of the service, 1 if the current status, 0 otherwise | gauge | name, status |