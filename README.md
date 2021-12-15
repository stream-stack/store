# store

```shell
main.exe --DataDir=data1 --Address=localhost:2001 --RaftId=node1 --Bootstrap=true
main.exe --DataDir=data2 --Address=localhost:2002 --RaftId=node2
main.exe --DataDir=data3 --Address=localhost:2003 --RaftId=node3

raftadmin localhost:2001 add_voter node2 localhost:2002 0
raftadmin localhost:2001 add_voter node3 localhost:2003 0

raftadmin --leader multi:///localhost:2001,localhost:2002 add_voter node3 localhost:2003 0
```