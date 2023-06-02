- https://medium.com/yapi-kredi-teknoloji/argo-cd-ldap-authentication-and-rbac-configuration-7c1b7b0cb7a1

### **LDAPユーザで認証するためには該当LDAPユーザに`mail`属性と`displayName`属性が設定されている必要がある**

## ArgoCDのRBAC
- ArgoCDのRBAC configurationはstored in `argocd-rbac-cm` ConfigMap
- ArgoCDには以下2つのpre-definedのRoleがある
  - `role:readonly` → read-only access to all resources
    - Syncなど更新操作はできない
  - `role:admin` → unrestricted access to all resources
    - すべての操作が可能


## LdapグループとArgoCDのRBACを紐づける方法
- `argocd-rbac-cm`のConfigMapを修正する必要がある
- `argocd-rbac-cm`の設定例  
  → `AWS Delegated Administrators`というADグループに所属しているADユーザはArgoCDのadmin権限を使えて、その他のADユーザは参照権限が付与される設定例
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
      g, AWS Delegated Administrators, role:admin
  ~~~
- 参考URL
  - https://medium.com/yapi-kredi-teknoloji/argo-cd-ldap-authentication-and-rbac-configuration-7c1b7b0cb7a1
  - https://argo-cd.readthedocs.io/en/stable/operator-manual/rbac/