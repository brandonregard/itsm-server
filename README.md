# itsm-server
Simple ITSM API

This is a very simple proof of concept.

Developed with Go, Echo, Gorm.
IAC with Terraform
Deployed to AWS Fargate using GitHub Actions CI/CD.
AWS RDS MySQL Data store.

Data set: https://archive.ics.uci.edu/ml/datasets/Incident+management+process+enriched+event+log

NOTES:
* Obviously, something like this would require authentication.
* The data set is not very good. It's a single table. Should be using normalized data.
* Need to modularize the the API.
* Normally, I would have made the API private. However, time was of the essence.
