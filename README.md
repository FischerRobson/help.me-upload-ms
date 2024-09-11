## Run Local

- Need to do port forward on rabbitmq:
```shell
kubectl port-forward service/helpme-rabbitmq-service 5672:5672 -n helpme

kubectl port-forward service/helpme-rabbitmq-service 15672:15672 -n helpme
```