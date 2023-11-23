- AthenaにてSQLでtableを作成すると自動的にAWS Glueでtableが作成される
  - CloudFrontログ用のTable作成
    - https://docs.aws.amazon.com/ja_jp/athena/latest/ug/cloudfront-logs.html
    ~~~
    CREATE EXTERNAL TABLE IF NOT EXISTS <Glue内database名>.<table名> (
      `date` DATE,
      time STRING,
      location STRING,
      bytes BIGINT,
      request_ip STRING,
      method STRING,
      host STRING,
      uri STRING,
      status INT,
      referrer STRING,
      user_agent STRING,
      query_string STRING,
      cookie STRING,
      result_type STRING,
      request_id STRING,
      host_header STRING,
      request_protocol STRING,
      request_bytes BIGINT,
      time_taken FLOAT,
      xforwarded_for STRING,
      ssl_protocol STRING,
      ssl_cipher STRING,
      response_result_type STRING,
      http_version STRING,
      fle_status STRING,
      fle_encrypted_fields INT,
      c_port INT,
      time_to_first_byte FLOAT,
      x_edge_detailed_result_type STRING,
      sc_content_type STRING,
      sc_content_len BIGINT,
      sc_range_start BIGINT,
      sc_range_end BIGINT
    )
    ROW FORMAT DELIMITED 
    FIELDS TERMINATED BY '\t'
    LOCATION 's3://<CloudFrontログS3bucket名>/'
    TBLPROPERTIES ( 'skip.header.line.count'='2' )
    ~~~

- GrafanaのAthenaデータソースを使うとAthenaからグラフを作成することもできる。テーブルとしてログを内容を確認することももちろんできる。

- 参考URL
  - https://docs.aws.amazon.com/ja_jp/athena/latest/ug/cloudfront-logs.html
  - https://dev.classmethod.jp/articles/amazon-athena-query-to-retrieve-data-for-specified-period/
  - https://qiita.com/yuji_saito/items/82df1c25813215e0c0ae
  - https://omuron.hateblo.jp/entry/2020/06/09/000000
  - https://qiita.com/aibax/items/6aa8c08e39b824cf85f2

## SQL文サンプル
- 9日前～現在のログからuriでグルーピングし、uriごとの件数をグラフで表示  
  **※ORDER BYがないとエラーになる**
  ~~~
  SELECT uri,date,COUNT(*) as sum
  FROM default.cloudfront_logs
  WHERE date > CURRENT_TIMESTAMP - INTERVAL '9' DAY
  GROUP BY uri,date
  ORDER BY date;
  ~~~
- 7日前～現在のログから`uri`に`.html`を含むものだけ、dateとtimeを降順にして表示 (テーブル形式)
  ~~~
  SELECT *
  FROM default.cloudfront_logs
  WHERE date > CURRENT_TIMESTAMP - INTERVAL '7' DAY AND "uri" LIKE '%.html'
  ORDER BY date desc,time desc;
  ~~~