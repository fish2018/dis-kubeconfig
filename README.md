# diskubecfg

把合并的kubeconfig文件按集群分割成独立配置文件

编译：`GOOS=linux go build -ldflags="-w -s"`

使用方式：`./dis -f [configFile]`，不指定路径默认`~/.kube/config`
