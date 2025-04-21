- https://argo-cd.readthedocs.io/en/stable/operator-manual/user-management/microsoft/
- やり方が3つある

## Entra ID App Registration Auth using OIDCの方法
- Azure EntraIDのトップページにあるTenant IDを確認する  
  ![](./image/azure_entraid_tenant_id.jpg)
- Azure EntraIDの該当Enterprise applicationsの中で、「Manage」→「Users and groups」で、Object typeを確認し、権限を付与したいGroupやUserをクリックし、Object IDを確認する  
  ![](./image/azure_entraid_user_and_group_1.jpg)  
  ![](./image/azure_entraid_user_and_group_2.jpg)