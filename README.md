# store

store 是stream-stack中的一个模块，用于存储数据。

## 主要特性

- hashicorp/raft 协议,实现leader选举,操作日志复制
- dgraph-io/badger 用于存储数据

## 运行

```shell
main.exe --DataDir=data1 --Address=localhost:2001 --RaftId=node1 --Bootstrap=true
main.exe --DataDir=data2 --Address=localhost:2002 --RaftId=node2
main.exe --DataDir=data3 --Address=localhost:2003 --RaftId=node3

raftadmin localhost:2001 add_voter node2 localhost:2002 0
raftadmin localhost:2001 add_voter node3 localhost:2003 0

raftadmin --leader multi:///localhost:2001,localhost:2002 add_voter node3 localhost:2003 0
```

## 参照
- [Raft 官方文档](https://raft.github.io/raft/)
- [raft实现](https://github.com/hashicorp/raft)
- [badger实现](https://github.com/dgraph-io/badger)
- [raft-badger](https://github.com/BBVA/raft-badger)