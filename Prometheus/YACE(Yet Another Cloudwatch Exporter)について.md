- https://github.com/nerdswords/yet-another-cloudwatch-exporter

## ■ CloudWatchメトリクス取得APIの種類について
- CloudWatchでメトリクスを取得するAPIには`GetMetricStatistics`と`GetMetricData`の2種類がある
- `GetMetricStatistics`は1回のAPIコールで1つのメトリクスしか取得できない
- `GetMetricData`は複数（500個まで）のメトリクスを一度に取得でき、CloudWatchメトリクス取得の料金を抑えられる
  - YACEは`GetMetricData`を使用している

> [!NOTE]  
> ここでいうメトリクスとは1つのデータポイントではなく、メトリクスの種類（例：CPUUtilization、NetworkInなど）を指す  
> https://aws.amazon.com/jp/cloudwatch/pricing/  

### 30個のメトリクス(の種類)を`GetMetricData`APIで5分間隔で取得した場合の料金
- 1回のリクエストで30種類のメトリクス取得
- 1日のリクエスト数：12回/時間 × 24時間 = 288回
- 1ヶ月のリクエスト数：288回 × 30日 = 8,640回
- 月間メトリクス取得数：8,640回 × 30メトリクス = 259,200メトリクス
- **料金：259,200 ÷ 1,000 × $0.01 = 約$2.59/月**

## ■ YACEの設定
### scraping間隔
- defaultでは5分間隔でscrapingする
- `-scraping-interval` 起動flagで変更できる
- https://github.com/nerdswords/yet-another-cloudwatch-exporter#decoupled-scraping