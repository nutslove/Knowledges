- https://argo-cd.readthedocs.io/en/stable/user-guide/sync-waves/

# Sync Phase
- `PreSync`、`Sync`、`PostSync`の3つのフェーズに分かれている
- `PreSync`の`Application`が全て同期される（Healthyになる）と、`Sync`フェーズが開始される
- `Sync`の`Application`が全て同期される（Healthyになる）と、`PostSync`フェーズが開始される
 
# Sync Waves
- `Sync Waves`は、`Application`の同期を順番に行うための仕組み
- `Application`ではなく、実際にデプロイされる`Job`などのリソースに対して、`sync-wave`というラベルを付与することで、同期の順番を制御できる
- 例  
  ```yaml
  apiVersion: bitnami.com/v1alpha1
  kind: SealedSecret
  metadata:
    creationTimestamp: null
    name: sample-secret
    namespace: default
    annotations:
      argocd.argoproj.io/hook: PreSync
      argocd.argoproj.io/hook-delete-policy: BeforeHookCreation
      argocd.argoproj.io/sync-wave: "1"
  spec:
    encryptedData:
      MYSQL_HOST: AgBPhntyTvUOYPtjJrdZk0fbVyljZFQExAVKAHYjE0rtTNsqaNXygninGs5CPPALfIf/BgmONK35xepHnH7kqc+WNm2nt4MueAMHAtrHQ/8SCkzXbQGLuiyUUds2Q2xVu/rBAVwGfWaKkwi9Db8Y7u6UOLE2wjgk2hpfqnf0KCgR4+u4q936tszBlqQrzk6+ZCqn7oPkOHI3iHJIi4Myo+Xh51Cx0ItEcq18huAK04tTgPoNDJgjLZZnWt83eOJcdqs6mNuI2khsi+/sXizA/naRRDJjiZat5//v/e73ejPRYLJkLCxmbiqBkZUmzBcTgj9PNvOPd+Be/VlKV0I4oO23e4f4CTp+xsq/hxje1qo38zL5UEalbYvtWVCAQCHyA5zcacpD7Sbl0hDSn4JCmXgktrtmgp4iADw2dGDqj7yGyvgQt18VUyWwqIzdhdih9a9LLvqK+gEf9pHs6p6j+rSlvIhKWM9k+vp44h9GLv5ViXB0jFD6Ig5kuyrne6OPceJgj1kuoGQbGdwCvd6thjNytWwBuPujKo5AK4VepLUUh7q7zZwqBv0yXMwSkmFDpJMTmMuy6W32HZvmJw1m/ajaAqlltDKl2pbyDClRLJ/K1n9QDiOEgIdLX3t9C5tJL1G9b1bMTpTfqqYcQbTQnDBEJJ7du4eYQRoT9EP4vI0eyab8beuvBvwIdPvsPWwY+uf7TzuyX0ve6/8=
      MYSQL_PASSWORD: AgAV4O7rOmyqL2JtPZ/XtUIMNW5GPgSMhc8zU0zU7LyaNfu5PPFfZrqmyvV4u0XjiNL0nH0rxCEnbm8ZlYtDQCDTMAZMgQkDSWVq7RGnL0txkE3mtsgzctNvIk54lvoVi7fzsBUHXLWsXu58NgQsGhc/VG25zMIq2ycEKd0nEFoa3EK1EoNsb3gKNS4qGM7hqkS5y7ufOcUR5mVuvXlihrfn9nGcucFsRsMbxsTYWUauM8x0wqxrnw7o74EkA0Gyhu2FI43n4mleubC8pzOEYvurfyfd1UmS7H6j/Cm4VuPtiX/uynvKyb2iVJX3vkeyKnmvhP6AcegfI1UYSD1nLg6taDOUFDdScCRsdI97lF0ehYauinLLp94rRQIXOFWYaSqq8VNbQzApXTDAgSzjazdx17p1IVrspEc7xCzD1Dcn3RiiiSpGnYEbG4rF4pFu8AcKH5j/MSydKN4iJgQ2utOgJ7pUXNeShgRrNf/BI67uus44fgeIXnkG8+LG7/WGhsjnYTRsECZD4g+q2vB3tYeBSFQ7ojUFVN5ChsZgTS+uziIp3jzxu3sWfSt4puGx+BlFp+dxv0Zp1YrRKuje+dBWlCkrf9Sa16SR7DCqKdX+1G0W6rwYYEyXwRW9Y0s6lUZwfpd0kOVcYZABohfH3RkWmxgKjF76xFlyFfHILVtdGJbe1nz1144wAMQeE3jvk/ypeAGy+I8lbg==
      MYSQL_USER: AgDjq7qBJR1DDOWapYF4AhtQi1sQgeglSRhu6tZ+JHPTPPdNFBV+zQk3xzKKRldZeeIlXQ5ZquEdc3lMWYKqg5xvL2e21YhgPv4oWADDvsWxWdklZVLDl0SGPQT64v6K+lPVR5Gia6zIpocvv4v3GbMcCXFkFK/b5HeSU6FcxrAhrQ+vdxSWKfkezO8vrS1wJ3MYLgLSwtiRDsibjpYwVTg0sLDcpTlrDk95IEnmcY8bE6qGJ3BO8mUBPhsBnk5yLJO2MZeOofA9IbpUDIpUJSFcaaXc+PK3C+vKxxJ2Ia94Y07UshUOrl72dTrp9aJv1mdkVfzZ7EHQpHh+2k7b2U1nMFVFwK2QV193vJVSQDM21IzIttE4QOxllv0ei48FVMVw3pJ0ABJMKiYta6KAZvKQ5EJkEIVVHtVqRH6plpBgU72dAPVuJAmdlAQ3I844lj1AKwfydylFRJP/olkHegB5Lu3ScwblicPZt4Q+f+IV79YefKR+Bf8zwLB5bd5uFOVx5Jj3dkiSfTN9P60FE6uA1axxyECafqwcgEmxafODtRpP+94rbAUboWCCXTAqjL+KyclCkXg0lKKLDHBmIzhsV3+eDwlApHWfvs1/x2dyW17Rjs7Nqb6vi/xX+YXT0d67TzNEUFfYFABRN9ABN/YSCWsVxIJXvMulg4RpsGD4g5E4GG5gwj3Jjjfr0yEtsu9fS5Zn
    template:
      metadata:
        creationTimestamp: null
        name: sample-secret
        namespace: default
  ```
  ```yaml
  apiVersion: batch/v1
  kind: Job
  metadata:
    name: sample-job
    annotations:
      argocd.argoproj.io/hook: PreSync
      argocd.argoproj.io/hook-delete-policy: BeforeHookCreation
      argocd.argoproj.io/sync-wave: "2"
  spec:
    backoffLimit: 0
    template:
      spec:
        containers:
          - name: sample-job
            image: istsh/secret-sample-app-job:1.0.0
            args: ["success"]
            envFrom:
              - secretRef:
                  name: sample-secret
        restartPolicy: Never
  ```