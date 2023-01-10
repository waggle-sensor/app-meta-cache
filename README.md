# Updating app meta information in wes-app-meta-cache
wes-app-meta-cache service hosts a Redis server to manage meta information of edge applications in Waggle edge stack (WES). We push application's meta information to the service whenever we run applications.

This container is used as Kubernetes's __initContainer__. The initContainer is tied to an edge application and runs before the edge application runs. We add `zone` Kubernetes node label to the meta information to help better discover data measurements that edge applications publish by filtering the zone value. The flow of how we manage app meta information is as follows,
```bash
(): program
[]: data

( node scheduler )
       |
[ app-meta-info ]
       |
( initContainer ) + [ Kubernetes node label ]
       |
[ updated app-meta-info ]
       |
( wes-app-meta-cache )
```

## Testing
To test this tool locally,

```bash
# run a local Redis server
$ docker run -d --name wes-app-meta-cache -p 6379:6379 redis:7.0.4

# --kubeconfig requires a correct kubeconfig to access a Kubernetes cluster that has the node name with a zone label
$ docker run --rm --entrypoint /update-app-cache waggle/app-meta-cache:latest --host localhost set --kubeconfig /tmp/kubeconfig --nodename "000048b02d0766be.ws-nxcore" app-meta.test '{"host": "000048b02d0766be.ws-nxcore"}'
added zone core

# retrieve stored app meta information to ensure the zone attribute is included
$ docker exec wes-app-meta-cache redis-cli get app-meta.test
{"host":"000048b02d0766be.ws-nxcore","zone":"core"}

# clean up testing
$ docker rm -f wes-app-meta-cache
wes-app-meta-cache
```