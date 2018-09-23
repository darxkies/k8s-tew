
run:
	k8s-tew initialize -f
	k8s-tew node-add -s
	k8s-tew generate
	sudo $$(which k8s-tew) run

dashboard:
	k8s-tew dashboard -o
