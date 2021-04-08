#!/usr/bin/env bash

function wait_for_pod()
{
	echo "waiting for pod $1" 

	while [[ $(kubectl get pods -n showcase $1 -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do 
		sleep 1 
	done
}

function get_wordpress_pod_name()
{
	kubectl get pods --selector tier=frontend -n showcase -o 'jsonpath={..metadata.name}'
}

function get_mysql_pod_name()
{
	kubectl get pods --selector tier=mysql -n showcase -o 'jsonpath={..metadata.name}'
}

function write_to_pod()
{
	kubectl exec -t -i -n showcase $1 -- sh -c "echo $2 > $3; sync"
}

function read_from_pod()
{
	kubectl exec -t -i -n showcase $1 -- sh -c "cat $2" | tr -d '\r'
}

function check_content()
{
	local content=$(read_from_pod $1 $3)

	if [ "$content" == "$2" ]
	then
		echo "content of pod $1 is OK"
	else
		echo "***content of pod $1 is wrong: got '$content' but expected '$2'***"
	fi
}

function check_showcase()
{
	local wordpress_content="wordpress-check"
	local wordpress_file="/var/www/html/check"
	local mysql_content="mysql-check"
	local mysql_file="/var/lib/mysql/check"
	local wordpress_pod_name=$(get_wordpress_pod_name)
	local mysql_pod_name=$(get_mysql_pod_name)

	write_to_pod $wordpress_pod_name $wordpress_content $wordpress_file
	write_to_pod $mysql_pod_name $mysql_content $mysql_file

	velero backup delete showcase --confirm

	velero backup create showcase --include-namespaces showcase --wait

	kubectl delete namespace showcase

	velero restore create --from-backup showcase --wait

	wait_for_pod $wordpress_pod_name 
	wait_for_pod $mysql_pod_name 

	check_content $wordpress_pod_name $wordpress_content $wordpress_file
	check_content $mysql_pod_name $mysql_content $mysql_file
}

check_showcase
