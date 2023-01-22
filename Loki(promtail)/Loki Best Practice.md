## 各Componentの必要な数(Scaling)の基準について
- これくらいのLog量の時は、これくらいのIngester/Distributorが必要という基準が書かれている(公式)ドキュメントはない。LokiエンジニアによるとまだLimitは見つかってないらしい。  
    <img src="https://github.com/nutslove/Knowledges/blob/main/Loki(promtail)/image/scaling_standard.jpg" width="500" height="500">
- ただ、IngesterはLog量が少なくても最低3つは必要だと考えている。  
  Defaultの`replication_factor`は3で、その場合少なくとも2つのIngesterにpushできなければLogはLokiに連携されない。Ingesterが2つだけある場合は1つでもIngesterがUnhealthyになったら必要最低限のActive状態のIngesterが存在しないためLogはLokiに連携されず、失われてしまうので1つのIngesterがダメになってもLogを受けつけれるようにIngesterは3つ以上を起動した方が良い。

