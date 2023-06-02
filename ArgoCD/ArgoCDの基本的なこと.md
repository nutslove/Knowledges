### 最初に払い出されるadminのPW確認方法
- `kubectl get secret argocd-initial-admin-secret -n argocd -o jsonpath="{.data.password}" | base64 -d`
- ログイン後GUI上でパスワード変更ができる

