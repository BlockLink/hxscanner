hxscanner
============

scanner for hx blockchain

# Usage

* install golang
* git clone this repo to $GOPATH or `go get github.com/BlockLink/hxscanner`
* cd to this repo path
* `./install_deps.sh`
* `docker swarm init`
* `docker stack deploy -c stack.yml postgres` and then `docker service ls` to view postgresql instance. You can also create postgresql database manually.
* import `sqls/init.sql` to db
* `go build`
* `./hxscanner` (you can use ./hxscanner -h to see help info)
