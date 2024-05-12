## Batching queue

In order to instantiate a queue you need to invoke `CreateQueue()` factory function and pass required parameters:
- `maxBatchSize int`: the amount of jobs which will incur immediate sending of the jobs to the `BatchProcessor` instance (mentioned below).
- `batchCollectionTime time.Duration`: period of time to gather jobs after receiving the first job into the queue. Once timed out all gathered jobs will be sent to the `BatchProcessor`.
- `processor BatchProcessor`: external dependency which can handle jobs in batches.

### Job and JobResult

Both types are very generic and having the purpose to demonstate the idea rather than to make production ready solution. The proper payload of each of them depends on business requirements.