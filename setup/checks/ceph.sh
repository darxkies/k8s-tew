#!/usr/bin/env bash 

VERSION=devel

function wait_for_pod()
{
	echo "waiting for pod $1" 

	while [[ $(kubectl get pods $1 -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do 
		sleep 1 
	done
}

function write_to_pod()
{
	kubectl exec -t -i $1 -- sh -c "echo $2 > /var/lib/www/$3/index.html;sync"
}

function read_from_pod()
{
	kubectl exec -t -i $1 -- sh -c "cat /var/lib/www/$2/index.html" | tr -d '\r'
}

function check_content()
{
	local content=$(read_from_pod $1 $3)

	if [ "$content" == "$2" ]
	then
		echo "content of pod $1 is [OK]"
	else
		echo "***content of pod $1 is wrong: got '$content' but expected '$2'***"
	fi
}

function wait_for_snapshot()
{
	echo "waiting for snapshot $1"

	while [[ $(kubectl get volumesnapshots $1 -o 'jsonpath={..status.readyToUse}') != "true" ]]
	do 
		sleep 1
	done
}

function create_pod()
{
	kubectl apply -f https://raw.githubusercontent.com/ceph/ceph-csi/$VERSION/examples/$1/$2.yaml
}

function create_pvc()
{
	curl -sSL https://raw.githubusercontent.com/ceph/ceph-csi/$VERSION/examples/$1/$2.yaml | sed "s/csi-$1-sc/csi-$1/" | kubectl apply -f -
}

function create_snapshot()
{
	kubectl apply -f https://raw.githubusercontent.com/ceph/ceph-csi/$VERSION/examples/$1/$2.yaml
}

function clean_up_rbd() 
{
	kubectl delete pod pod-with-raw-block-volume csi-rbd-demo-pod csi-rbd-clone-demo-app csi-rbd-restore-demo-pod
	kubectl delete pvc raw-block-pvc rbd-pvc rbd-pvc-clone rbd-pvc-restore
	kubectl delete volumesnapshot.snapshot.storage.k8s.io rbd-pvc-snapshot
}

function clean_up_cephfs() 
{
	kubectl delete pod csi-cephfs-demo-pod csi-cephfs-clone-demo-app csi-cephfs-restore-demo-pod
	kubectl delete pvc csi-cephfs-pvc cephfs-pvc-clone cephfs-pvc-restore
	kubectl delete volumesnapshot.snapshot.storage.k8s.io cephfs-pvc-snapshot
}

function check_rbd()
{
	clean_up_rbd

	# Raw - /dev/xvda
	create_pvc rbd raw-block-pvc
	create_pod rbd raw-block-pod
	wait_for_pod pod-with-raw-block-volume 

	# PVC
	create_pvc rbd pvc
	create_pod rbd pod
	wait_for_pod csi-rbd-demo-pod 

	# Write
	write_to_pod csi-rbd-demo-pod "xxx" "html"

	# PVC clone
	create_pvc rbd pvc-clone
	create_pod rbd pod-clone
	wait_for_pod csi-rbd-clone-demo-app 

	# Write
	write_to_pod csi-rbd-demo-pod "yyy" "html" 

	# Snapshot
	create_snapshot rbd snapshot
	wait_for_snapshot rbd-pvc-snapshot 

	# Write
	write_to_pod csi-rbd-demo-pod "zzz" "html"

	# PVC restore
	create_pvc rbd pvc-restore
	create_pod rbd pod-restore
	wait_for_pod csi-rbd-restore-demo-pod 

	# Check content
	check_content csi-rbd-clone-demo-app "xxx" "html"
	check_content csi-rbd-restore-demo-pod  "yyy" "html"
	check_content csi-rbd-demo-pod "zzz" "html"

	clean_up_rbd
}

function check_cephfs()
{
	clean_up_cephfs

	# PVC
	create_pvc cephfs pvc
	create_pod cephfs pod
	wait_for_pod csi-cephfs-demo-pod 

	# Write
	write_to_pod csi-cephfs-demo-pod "xxx" 

	# PVC clone
	create_pvc cephfs pvc-clone
	create_pod cephfs pod-clone
	wait_for_pod csi-cephfs-clone-demo-app 

	# Write
	write_to_pod csi-cephfs-demo-pod "yyy" 

	# Snapshot
	create_snapshot cephfs snapshot
	wait_for_snapshot cephfs-pvc-snapshot 

	# Write
	write_to_pod csi-cephfs-demo-pod "zzz" 

	# PVC restore
	create_pvc cephfs pvc-restore
	create_pod cephfs pod-restore
	wait_for_pod csi-cephfs-restore-demo-pod 

	# Check content
	check_content csi-cephfs-clone-demo-app "xxx" "html"
	check_content csi-cephfs-restore-demo-pod  "yyy" "html" 
	check_content csi-cephfs-demo-pod "zzz" 

	clean_up_cephfs
}

function check_rados_gateway()
{
	local access_key=$(k8s-tew dashboard ceph-rados-gateway -q | cut -d ' ' -f 1)
	local secret_key=$(k8s-tew dashboard ceph-rados-gateway -q | cut -d ' ' -f 2)
	local url=$(k8s-tew dashboard ceph-rados-gateway -q | cut -d ' ' -f 3)
	local config=/tmp/s3cfg
	local put_file=/tmp/s3put
	local get_file=/tmp/s3get
	local content="xxx"
	local cmd="s3cmd -c $config --no-check-certificate" 

	cat <<EOF > $config
# Setup endpoint
host_base = $url
host_bucket = $url
bucket_location = us-east-1
use_https = True

# Setup access keys
access_key =  $access_key
secret_key = $secret_key

EOF

	echo $content > $put_file

	[ -f $get_file ] && rm $get_file

	$cmd mb s3://checks
	$cmd ls s3://
	$cmd put $put_file s3://checks/content
	$cmd get s3://checks/content $get_file
	$cmd del s3://checks/content 
	$cmd rb s3://checks

	if [ "$(cat $put_file)" == "$(cat $get_file)" ]
	then
		echo "Rados Gateway check [OK]"
	else
		echo "***Rados Gateway content is not ok***"
	fi

	rm  $config $put_file $get_file
}

check_rbd
check_cephfs
check_rados_gateway
