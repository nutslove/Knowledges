- https://opensearch-project.github.io/opensearch-py/api-ref/clients/opensearch_client.html
- `pip install opensearch-py`でインストール

## 使い方
### 事前準備(共通)
- まず最初に`OpenSearch`クラスからインスタンスを作成
  ~~~python
  opensearch = OpenSearch(
      ["https://opensearch:9200"],
      # index_name="*", ## index_nameはベクトルDBのindex名を指定する。*を指定すると、全indexを検索する。
      index_name="unk",
      http_auth=("admin", "admin"),
      use_ssl = False,
      verify_certs = False,
      ssl_assert_hostname = False,
      ssl_show_warn = False,  
  )
  ~~~
### OpenSearchからデータ検索
- `indices.exists`メソッドでOpenSearchに該当indexがあるか確認し、`search`メソッドでクエリーと一致するデータを検索する
  - https://opensearch-project.github.io/opensearch-py/api-ref/clients/opensearch_client.html#opensearchpy.OpenSearch.search
  ~~~python
  if opensearch.indices.exists(index=index_name):
    response = db.search(
        index='aws',
        body={
            "query": {
                "match": {"metadata.source": "/tmp/ec2-types.pdf"}
            }
        }
    )
    hits = response["hits"]["hits"]
    print("検出数", len(hits))
    for hit in hits:
        print("filename:", hit["_source"]["metadata"]["source"])
        print("text(データ):", hit["_source"]["text"])
  ~~~
### OpenSearchからデータ削除
- `indices.exists`メソッドでOpenSearchに該当indexがあるか確認し、`delete_by_query`メソッドでクエリーに一致するデータを削除する
  - https://opensearch-project.github.io/opensearch-py/api-ref/clients/opensearch_client.html#opensearchpy.OpenSearch.delete_by_query
  ~~~python
  if opensearch.indices.exists(index=index_name):
      try:
          result = opensearch.delete_by_query(
              index=index_name,
              body={
                  "query": {
                      "term": {
                          "metadata.source.keyword": tmp_location ## metadata.sourceをキーとして検索し、一致するものを削除する
                      }
                  }
              }
          )
      except Exception as e:
          print(f"remove_document_from_opensearchでエラーが発生しました: {e}")

      print(f"OpenSearch上Object削除結果: {result}")

      if result['deleted'] == 0:
          print("削除なし")
      else:
          print("削除あり")
  ~~~
#### queryタイプについて（要確認）
- `term`と`match`の２種類があるっぽい
  - `term`は完全一致、`match`は部分一致？
  - `term`はデータのタイプが`text`のものに、`match`は`keyword`データタイプに使えるっぽい  
    ~~~json
    {
      "unk" : {
        "aliases" : { },
        "mappings" : {
          "properties" : {
            "metadata" : {
              "properties" : {
                "page" : {
                  "type" : "long"
                },
                "source" : {
                  "type" : "text",
                  "fields" : {
                    "keyword" : {
                      "type" : "keyword",
                      "ignore_above" : 256
                    }
                  }
                }
              }
            },
            "text" : {
              "type" : "text",
              "fields" : {
                "keyword" : {
                  "type" : "keyword",
                  "ignore_above" : 256
                }
              }
            },
            "vector_field" : {
              "type" : "knn_vector",
              "dimension" : 1536,
              "method" : {
                "engine" : "nmslib",
                "space_type" : "l2",
                "name" : "hnsw",
                "parameters" : {
                  "ef_construction" : 512,
                  "m" : 16
                }
              }
            }
          }
        },
        "settings" : {
          "index" : {
            "replication" : {
              "type" : "DOCUMENT"
            },
            "number_of_shards" : "1",
            "knn.algo_param" : {
              "ef_search" : "512"
            },
            "provided_name" : "unk",
            "knn" : "true",
            "creation_date" : "1708445615760",
            "number_of_replicas" : "1",
            "uuid" : "ZmJdoGuSQQaOjbVWULCNrw",
            "version" : {
              "created" : "136327927"
            }
          }
        }
      }
    }
    ~~~