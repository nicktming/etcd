rm -rf default.etcd
ps -ef | grep main.go | awk '{print $2}' | xargs kill -9
lsof -i:2380 | awk 'NR > 1{print $2}' | xargs kill -9
go run main.go
