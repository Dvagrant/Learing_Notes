## WAL技术（write-ahead logging）

---

## CHECKPOINT
- `sharp checkpoint`：数据库关闭时，所有脏页需要刷回磁盘
- `fuzzy checkpoint`：触发条件后按一定比例刷回磁盘
	- `master thread checkpoint`（定时刷
	- `flush_lru_list checkpoint`（保证 LRU 中有100页可用
	- `Async/Sync flush checkpoint`（redo log 占满
	- `dirty page too mucn point`（内存不足


---
## DOUBLE WRITE技术

---
## 2PL协议，用来实现序列化隔离级别


---
## ONLINE DDL

---
## CHANGE BUFFER
当需要更新一个数据页时，如果数据页在内存中就直接更新，而如果这个数据页还没有在内存中的话，在不影响数据一致性的前提下，InnoDB 会将这些更新操作缓存在`change buffer`中，这样就不需要从磁盘中读入这个数据页了。

需要说明的是，虽然名字叫作`change buffer`，实际上它是可以持久化的数据。也就是说，`change buffer`在内存中有拷贝，也会被写入到磁盘上。

将`change buffer`中的操作应用到原数据页，得到最新结果的过程称为 merge 。除了访问这个数据页会触发 merge 外，系统有后台线程会定期 merge。在数据库正常关闭（shutdown）的过程中，也会执行 merge 操作。

显然，如果能够将更新操作先记录在`change buffer`，减少读磁盘，语句的执行速度会得到明显的提升。而且，数据读入内存是需要占用`buffer pool`的，所以这种方式还能够避免占用内存，提高内存利用率。

由于[[索引|唯一索引]]插入时需要检查是否有冲突，所以必须读入内存，丧失了`change buffer`的优势，所以只能用于普通索引。

因为 merge 的时候是真正进行数据更新的时刻，而`change buffer`的主要目的就是将记录的变更动作缓存下来，所以在一个数据页做 merge 之前，`change buffer`记录的变更越多（也就是这个页面上要更新的次数越多），收益就越大。

- 因此，对于写多读少的业务来说，页面在写完以后马上被访问到的概率比较小，此时`change buffer`的使用效果最好。这种业务模型常见的就是账单类、日志类的系统。
- 反过来，假设一个业务的更新模式是写入之后马上会做查询，那么即使满足了条件，将更新先记录在`change buffer`，但之后由于马上要访问这个数据页，会立即触发 merge 过程。这样随机访问 IO 的次数不会减少，反而增加了`change buffer`的维护代价。所以，对于这种业务模式来说，`change buffer`反而起到了副作用。

### change buffer 与 redo log

###### 带有 change buffer 的更新流程
![[带change buffer的更新流程.png]]

这条更新语句做了如下的操作（按照图中的数字顺序）：
`insert into t(id,k) values(id1,k1),(id2,k2)`
1. Page 1 在内存中，直接更新内存
2. Page 2 没有在内存中，就在内存的 change buffer 区域，记录下`我要往 Page 2 插入一行`这个信息
3. 将上述两个动作记入 redo log 中（图中 3 和 4）

做完上面这些，事务就可以完成了。所以，执行这条更新语句的成本很低，就是写了两处内存，然后写了一处磁盘（两次操作合在一起写了一次磁盘），而且还是顺序写的。同时，图中的两个虚线箭头，是后台操作，不影响更新的响应时间。

###### 更新后的读流程
![[带change buffer的读流程.png]]

`select * from t where k in (k1, k2)`

1. 读 Page 1 的时候，直接从内存返回。

>有几位同学在前面文章的评论中问到，WAL 之后如果读数据，是不是一定要读盘，是不是一定要从 redo log 里面把数据更新以后才可以返回？其实是不用的。你可以看一下图中的这个状态，虽然磁盘上还是之前的数据，但是这里直接从内存返回结果，结果是正确的。

2. 要读 Page 2 的时候，需要把 Page 2 从磁盘读入内存中，然后应用 change buffer 里面的操作日志，生成一个正确的版本并返回结果。

可以看到，直到需要读 Page 2 的时候，这个数据页才会被读入内存。

所以，如果要简单地对比这两个机制在提升更新性能上的收益的话，
- `redo log`主要节省的是随机写磁盘的 IO 消耗（转成顺序写）因为需要将脏数据写入离散的页中
- `change buffer`主要节省的则是随机读磁盘的 IO 消耗。因为要从离散的页中读取数据来UPDATE