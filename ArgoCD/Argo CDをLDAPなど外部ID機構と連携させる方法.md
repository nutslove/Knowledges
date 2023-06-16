- https://medium.com/yapi-kredi-teknoloji/argo-cd-ldap-authentication-and-rbac-configuration-7c1b7b0cb7a1

### **LDAPユーザで認証するためには該当LDAPユーザに`mail`属性と`displayName`属性が設定されている必要がある**

## ArgoCDのRBAC
- ArgoCDのRBAC configurationはstored in `argocd-rbac-cm` ConfigMap
- ArgoCDには以下2つのpre-definedのRoleがある
  - `role:readonly` → read-only access to all resources
    - Syncなど更新操作はできない
  - `role:admin` → unrestricted access to all resources
    - すべての操作が可能

## `argocd-cm` ConfigMapに以下を`kubectl patch`コマンドで反映する必要がある
- `patch-dex.yml`(ファイル名は任意)
  ~~~yaml
  apiVersion: v1
  data:
    url: <ArgoCDのURL e.g. https://argocd-nlb-abcdefghijklmnop.elb.ap-northeast-3.amazonaws.com>
    dex.config: |
      connectors:
      - type: ldap
        id: ldap
        name: AD
        config:
          host: test.domain.ad1:389
          insecureNoSSL: true
          insecureSkipVerify: true
          bindDN: CN=joiner,OU=Users,OU=test,DC=test,DC=domain,DC=ad1
          bindPW: <bindDNのPW>
          usernamePrompt: AD Username  ## ---> ログイン画面でユーザ名入力するところの表示名
          userSearch:
            baseDN: OU=Users,OU=test,DC=test,DC=domain,DC=ad1
            username: sAMAccountName
            idAttr: distinguishedName
            emailAttr: mail
            nameAttr: displayName
          groupSearch:
            baseDN: OU=Users,OU=test,DC=test,DC=domain,DC=ad1  ## ---> Groupをこのディレクトリ配下に作成した場合
            userAttr: distinguishedName
            groupAttr: member
            nameAttr: name
  ~~~
- patch-dex.ymlファイルがあるディレクトリにて以下コマンドで適用
  - `kubectl -n argocd patch cm argocd-cm --patch "$(cat patch-dex.yml)"`

## LdapグループとArgoCDのRBACを紐づける方法
- `argocd-rbac-cm`のConfigMapを修正する必要がある
- ポリシーを記述した`policy.csv`を`argocd-rbac-cm`のConfigMapの中に配置する
- `argocd-rbac-cm`の設定例  
  → `root_group`というADグループに所属しているADユーザはArgoCDのadmin権限を使えて、その他のADユーザは参照権限が付与される設定例
  ~~~yaml
  apiVersion: v1
  kind: ConfigMap
  metadata:
    labels:
      app.kubernetes.io/instance: argocd
      app.kubernetes.io/name: argocd-rbac-cm
      app.kubernetes.io/part-of: argocd
    name: argocd-rbac-cm
    namespace: argocd
  data:
    policy.default: role:readonly
    scopes: '[groups, email]'
    policy.csv: |
      p, role:none, *, *, */*, deny
      g, "root_group", role:admin
  ~~~
- 参考URL
  - https://medium.com/yapi-kredi-teknoloji/argo-cd-ldap-authentication-and-rbac-configuration-7c1b7b0cb7a1
  - https://argo-cd.readthedocs.io/en/stable/operator-manual/rbac/
  - https://qiita.com/dtn/items/9bcae313b8cb3583977e

## ArgoCDのRBACの項目について
- ArgoCDのRBACのPolicyの記述方法には2種類があるみたい
- 第1フィールドが`p`はPolicyで、どのroleがどのProjectとApplicationにどのようなActionができるかを定義する。  
  `g`はLDAPグループやGithubのOrgなど外部IDプロバイダーのGroupとArgoCDのRoleを関連付ける。
- `p`の設定
  1. All resources except application-specific permissions (see next bullet):
      - `p, <role/user/group>, <resource>, <action>, <object>`
  2. Applications, applicationsets, logs, and exec (which belong to an AppProject):
      - `p, <role/user/group>, <resource>, <action>, <appproject>/<object>`
- 参考URL
  - https://argo-cd.readthedocs.io/en/stable/operator-manual/rbac/#rbac-permission-structure