# consider references to undefined as an error
MAKEFLAGS += --warn-undefined-variables

# set shell
SHELL := bash

# make bash more error aware.
# -e to stop the script if there is an error
# -u to treat unset parameters as an error
# -o pipefail set the return value of the last non-zero exit status
.SHELLFLAGS := -eu -o pipefail -c

# set the default goal to target 'all'
.DEFAULT_GOAL := all

# delete target if it fails
.DELETE_ON_ERROR:

# use patterns (%.c) insteed of suffixes (.c) for tagets
.SUFFIXES:
	
azure_vm := aks-nodepool1-35188460-0

aks_cluster_name := conmansyscluster
aks_resource_group := conmansys
aks_mc_resource_group := MC_CONMANSYS_CONMANSYSCLUSTER_NORTHEUROPE

acr_name := conmansys

all:
	docker build -t conmansys .

azgroup:
	az group create \
		--name $(aks_resource_group) \
		--location northeurope

aks:
	# create the kubernetes cluster
	az aks create \
		--resource-group $(aks_resource_group) \ 
		--name $(aks_cluster_name) \
		--node-count 1 \
		--ssh-key-value ~/.ssh/azure_rsa.pub
	
	# setup the credentials for kubectl
	az aks get-credentials \
		--resource-group $(aks_resource_group) \
		--name $(aks_cluster_name)

acr:
	# create a new container registry
	az acr create \
		--name $(acr_name) \
		--resource-group $(aks_resource_group) \
		--sku Basic

# let the aks pull images from the acr
role:
	$(eval client_id := \
		$(shell az aks show \
		--resource-group $(aks_resource_group) \
		--name $(aks_cluster_name) \
		--query "servicePrincipalProfile.clientId" \
		--output tsv))
	$(eval acr_id := \
		$(shell az acr show \
		--resource-group $(aks_resource_group) \
		--name $(acr_name) \
		--query "id" \
		--output tsv))
	@echo "client id: $(client_id)"
	@echo "acr id: $(acr_id)"
	az role assignment create --assignee $(client_id) --scope $(acr_id) --role acrpull

nginx:
	kubectl run --image=nginx nginx
	kubectl expose $(shell kubectl get pods --output name | grep 'nginx') --port 80 \
		--target-port 80 \
		--type LoadBalancer

start:
	az vm start \
		--resource-group $(aks_mc_resource_group) \
		--name $(azure_vm)

kubectl-create-volume:
	-kubectl create -f volumes/postgres.yaml

kubectl-delete-volume:
	-kubectl delete -f volumes/postgres.yaml

kubectl-start: kubectl-create-volume
	-kubectl create -f configmaps/conmansys.yaml
	-kubectl create -f services/postgres.yaml
	-kubectl create -f deployments/postgres.yaml

kubectl-stop:
	-kubectl delete -f deployments/postgres.yaml
	-kubectl delete -f services/postgres.yaml
	-kubectl delete -f configmaps/conmansys.yaml

.PHONY: clean
clean:
	az vm deallocate \
		--resource-group $(aks_mc_resource_group) \
		--name $(azure_vm)
