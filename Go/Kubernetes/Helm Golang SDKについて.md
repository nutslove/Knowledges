- https://helm.sh/ko/docs/topics/advanced/#go-sdk  
  > This is a list of the most commonly used packages with a simple explanation about each one:
  >
  > - **`pkg/action`**: Contains the main “client” for performing Helm actions. This is the same package that the CLI is using underneath the hood. If you just need to perform basic Helm commands from another Go program, this package is for you
  > - **`pkg/{chart,chartutil}`**: Methods and helpers used for loading and manipulating charts
  > - **`pkg/cli`** and its subpackages: Contains all the handlers for the standard Helm environment variables and its subpackages contain output and values file handling
  > **`pkg/release`**: Defines the Release object and statuses