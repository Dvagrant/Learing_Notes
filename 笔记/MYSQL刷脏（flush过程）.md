刷脏称为`flush`
应用[[INNODB特性#CHANGE BUFFER|change buffer]]称作`merge`
回收[[日志系统#UNDO LOG|undo log]]成为`purge`

## 引发数据库的 flush 过程
1. InnoDB 的[[日志系统#REDO LOG|redo log]]写满了。这时候系统会停止所有更新操作，把[[INNODB特性#CHECKPOINT|checkpoint]]往前推进，`redo log`留出空间可以继续写。

2. 系统内存不足。当需要新的内存页，而内存不够用的时候，就要淘汰一些数据页，空出内存给别的数据页使用。如果淘汰的是**脏页**，就要先将脏页写到磁盘。

>不能直接把内存淘汰掉，下次需要请求的时候，从磁盘读入数据页，然后拿 redo log 出来应用不就行了？这里其实是从性能考虑的。如果刷脏页一定会写盘，就保证了每个数据页有两种状态：
>- 一种是内存里存在，内存里就肯定是正确的结果，直接返回；
>- 另一种是内存里没有数据，就可以肯定数据文件上是正确的结果，读入内存后返回。这样的效率最高。


3. MySQL 认为系统“空闲”的时候。当然，即使是忙时，也要见缝插针地找时间，只要有机会就刷一点“脏页”。

4. MySQL 正常关闭的情况。这时候，MySQL 会把内存的脏页都 flush 到磁盘上，这样下次 MySQL 启动的时候，就可以直接从磁盘上读数据，启动速度会很快。

---
## 刷新脏页的策略
>InnoDB 的刷盘速度就是要参考两个因素：一个是脏页比例，一个是 redo log 写盘速度。

首先，你要正确地告诉 InnoDB 所在主机的 IO 能力，这样 InnoDB 才能知道需要全力刷脏页的时候，可以刷多快。这就要用到`innodb_io_capacity`这个参数了，它会告诉 InnoDB 你的磁盘能力。这个值我建议你设置成磁盘的`IOPS`。

另一个参数`innodb_max_dirty_pages_pct`是脏页比例上限，默认值是 75%。由`Innodb_buffer_pool_pages_dirty`
/`Innodb_buffer_pool_pages_total`计算而来。

### 小策略
一旦一个查询请求需要在执行过程中先 flush 掉一个脏页时，这个查询就可能要比平时慢了。MySQL 中的一个机制，可能让你的查询会更慢：

在准备刷一个脏页的时候，如果这个数据页旁边的数据页刚好是脏页，就会把这个“邻居”也带着一起刷掉；而且这个把“邻居”拖下水的逻辑还可以继续蔓延，也就是对于每个邻居数据页，如果跟它相邻的数据页也还是脏页的话，也会被放到一起刷。

在 InnoDB 中，`innodb_flush_neighbors`参数就是用来控制这个行为的，值为 1 的时候会有上述的“连坐”机制，值为 0 时表示不找邻居，自己刷自己的。

找“邻居”这个优化在机械硬盘时代是很有意义的，可以减少很多随机 IO。机械硬盘的随机 IOPS 一般只有几百，相同的逻辑操作减少随机 IO 就意味着系统性能的大幅度提升。

而如果使用的是 SSD 这类 IOPS 比较高的设备的话，我就建议你把`innodb_flush_neighbors`的值设置成 0。因为这时候 IOPS 往往不是瓶颈，而“只刷自己”，就能更快地执行完必要的刷脏页操作，减少 SQL 语句响应时间。

在 MySQL 8.0 中，`innodb_flush_neighbors`参数的默认值已经是 0 了。

---
## 刷脏时的细节
内存池`buffer pool`中每个页实例都带有`LSN( log sequence number )`，页每修改一次，LSN增加一次。

redo log，[[三种Page和List.png|flush list]]，[[INNODB特性#CHECKPOINT|checkpoint]]中都带有LSN的信息。

当需要 flush/checkpoint 时，总是会找出 LSN 最小的页向外淘汰。
- flush 时，根据 LRU 算法找出最久未使用的，此页一定是 flush list 中 LSN 最小的页。此时并不影响 redo log 中的 checkpoint 位置，原因如下，即 checkpoint 刷新的时候，会跳过已经刷脏的页。
- checkpoint 时，只需顺序的刷新 checkpoint 位置的页（redo log 顺序写）。若记录在 redo log 中后页无变化，则 LSN 无变化，一定在 flush list 中。记录在 redo log 后变化了，LSN 一定是新的，也在 flush list 中，此时顺序跳过 checkpoint 即可。