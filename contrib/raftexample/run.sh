go build .

rm -rf raftexample-1*

./raftexample --id 1 --cluster http://127.0.0.1:12379 --port 12380

