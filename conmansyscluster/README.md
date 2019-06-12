# Instructions

To deploy this project to kubernetes do the following:

## create a unique namespace

kubectl apply -f namespaces

## make the new namespace the default

kubectl config set-context --current --namespace conmansys

## create new storage for database

kubectl apply -f storage

## create configmaps (environment variables) for pods

kubectl apply -f configmaps

## create secrets  (secret environment variables) for pods

kubectl apply -f secrets

## create the database and a manager

kubectl apply -f backend

## create microservices

kubectl apply -f microservices

## create frontend

kubectl apply -f frontend
