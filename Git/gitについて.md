## Tag
- git clone時、直接Tagを指定することはできない。以下のようにclone後`checkout tags/<tag名>`で切り替える必要がある  
  ```shell
  git clone <GitリポジトリURL>
  cd <Gitリポジトリ名>
  git checkout tags/<tag名>
  ```

#### Tagの付与
1. 現在のブランチを確認  
   ```shell
   git branch
   ```

2. タグを付与したいコミットに移動（最新のコミットにタグを付与する場合はこのステップをスキップ）  
   ```shell
   git checkout <コミットハッシュ>
   ```

3. タグを付与  
   ```shell
   git tag <タグ名>
   ```

  - アノテーション付きタグの場合  
    ```shell
    git tag -a <タグ名> -m "<メッセージ>"
    ```

4. タグをリモートリポジトリにプッシュ（必要に応じて）
   ```shell
   git push origin <タグ名>
   ```

   または、すべてのタグをプッシュする  
   ```shell
   git push origin --tags
   ```