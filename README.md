hxscanner
============

scanner for hx blockchain

# Usage

* install golang
* `./install_deps.sh`
* `docker swarm init`
* `docker stack deploy -c stack.yml postgres` and then `docker service ls` to view postgresql instance
* import `sqls/init.sql` to db
* `go build`
* `./hxscanner`
