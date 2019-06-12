# Technical University of Denmark - Bachelor of engineering thesis project

This project is developed on OpenSUSE Tumbleweed and the guide assume the project is tested and deployed on a Linux machine.

The `$` in front of commands indicates it is a shell command.

## Run on Docker

To run the application on the local docker:

`$ docker-compose up`

This will build microservices and postgres database as well the netork and volumes needed. It will also create and start containers.

The frontend microservice can then be reached on **localhost:8060** and the apigateway can recieve request on **localhost:8079/api**.

If needed to rebuild:

`$ docker-compose up --build`

To Tear down the application:

`$ docker-compose down`

And to remove volumes too:

`$ docker-compose down --volumes`

## Run on Kubernetes

To try the application with Kubernetes one can install the [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) cluster and have the Kubernetes CLI [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed.

Unfortunately images are not pushed up to Docker Hub and therefore the images must be built onto the Kubernetes nodes.

Start the Minikube cluster:

`$ minikube start`

Change docker environment:

`$ eval $(minikube docker-env)`

Build images:

`$ docker-compose build --parllel`

parallel is not strictly needed but speed up build process.

When the images are build then one can use the kubectl to deploy the application.

The following steps will deploy the application:

1. `$ kubectl apply -f conmansyscluster/namespaces`
2. `$ kubectl apply -f conmansyscluster/storage`
3. `$ kubectl apply -f conmansyscluster/configmaps`
4. `$ kubectl apply -f conmansyscluster/secrets`
5. `$ kubectl apply -f conmansyscluster/backend`
6. `$ kubectl apply -f conmansyscluster/microservices`
7. `$ kubectl apply -f conmansyscluster/frontend`

To check pods are running first change namespace for the kubectl:

`$ kubectl config set-context --current --namespace conmansys`

Then:

`$ kubectl get pods`

Which should show that every pod is running.

To access the frontend or apigateway, ask the Minikube for a list of services:

`$ minikube service list`

This will show a list of services which is exposed. These services can then be accessed from the browser.
