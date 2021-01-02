
run:
	k8s-tew initialize -f
	# Remove this line if you do not use Ubuntu
	k8s-tew configure --resolv-conf=/run/systemd/resolve/resolv.conf 
	k8s-tew node-add -s
	k8s-tew generate
	sudo $$(which k8s-tew) run

dashboard:
	k8s-tew dashboard -o
