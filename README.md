flatten-sstabledump
===================
`flatten-sstabledump` is a utility that unwraps [Apache
Cassandra](http://cassandra.apache.org/)'s `sstabledump` JSON array output to
multiple small JSON objects, one per partition row. Partition metadata is
inlined into each row (key: `partition`). This utility allows you to easily use
map/reduce jobs (such as [Apache Hive](https://hive.apache.org), [AWS
Athena](https://aws.amazon.com/athena/), [Google
BigQuery](https://cloud.google.com/bigquery/) etc.) to process sstables.

The utility parses the JSON in a streaming fashion using very little memory.

Installation
------------
Install by issuing:

```bash
$ go get github.com/tink-ab/flatten-sstabledump
```

Testing
-------
There are some basic tests using [Bats](https://github.com/sstephenson/bats). Execute them by issuing:

```bash
$ bats test.bats
```

Usage
-----
```bash
$ cat testdata/testdata1.json
[
  {
    "partition" : {
      "key" : [ "d7f50415-3c9e-4a84-bdf2-54cbcbb0df0b", "201806" ],
      "position" : 0
    },
    "rows" : [
      {
        "type" : "row",
        "position" : 408,
        "clustering" : [ "35136e4c-ffa2-4205-82c1-7ce63d2519b9" ],
        "liveness_info" : { "tstamp" : "2018-06-19T11:20:49.363Z" },
        "cells" : [
          { "name" : "accountid", "value" : "457f21b5-69c0-48bc-bffa-037d88c8ecf8" }
        ]
      },
      {
        "type" : "row",
        "position" : 408,
        "clustering" : [ "203b5189-d9e1-4db7-b00c-c1b759790b8f" ],
        "liveness_info" : { "tstamp" : "2018-06-18T11:20:49.363Z" },
        "cells" : [
          { "name" : "accountid", "value" : "ee312163-75bf-4df5-94de-f34146efa502" }
        ]
      }
    ]
  },
  {
    "partition" : {
      "key" : [ "70c2ca4a-84f5-4cc2-b44a-e2f92b4888fb", "201806" ],
      "position" : 0
    },
    "rows" : [
      {
        "type" : "row",
        "position" : 408,
        "clustering" : [ "e1e6fbe4-4ec4-498d-b892-c00f7667bbc8" ],
        "liveness_info" : { "tstamp" : "2018-08-18T11:20:49.363Z" },
        "cells" : [
          { "name" : "accountid", "value" : "501cde77-cf60-468e-954f-e987e7490d4c" }
        ]
      }
    ]
  }
]
$ cat testdata/testdata1.json | fss
{"cells":[{"name":"accountid","value":"ee312163-75bf-4df5-94de-f34146efa502"}],"clustering":["203b5189-d9e1-4db7-b00c-c1b759790b8f"],"liveness_info":{"tstamp":"2018-06-18T11:20:49.363Z"},"partition":{"partition":{"key":["d7f50415-3c9e-4a84-bdf2-54cbcbb0df0b","201806"],"position":0}},"position":408,"type":"row"}
{"cells":[{"name":"accountid","value":"457f21b5-69c0-48bc-bffa-037d88c8ecf8"}],"clustering":["35136e4c-ffa2-4205-82c1-7ce63d2519b9"],"liveness_info":{"tstamp":"2018-06-19T11:20:49.363Z"},"partition":{"partition":{"key":["d7f50415-3c9e-4a84-bdf2-54cbcbb0df0b","201806"],"position":0}},"position":408,"type":"row"}
{"cells":[{"name":"accountid","value":"501cde77-cf60-468e-954f-e987e7490d4c"}],"clustering":["e1e6fbe4-4ec4-498d-b892-c00f7667bbc8"],"liveness_info":{"tstamp":"2018-08-18T11:20:49.363Z"},"partition":{"partition":{"key":["70c2ca4a-84f5-4cc2-b44a-e2f92b4888fb","201806"],"position":0}},"position":408,"type":"row"}
```

FAQ
---

### Can't you use AWS Athena to simply process raw `sstabledump` JSON output?

No. AWS Athena requires every JSON entity to be line-delimited. `sstabledump`
outputs its JSON in multiline JSON. Compacting using something like [`jq -c
.`](https://stedolan.github.io/jq/) doesn't work because AWS Athena requires
the JSON root object to be a JSON object, not array.

Using `jq -c .[]` would work in theory. However, it reads up the entire JSON
content into memory which doesn't scale for larger SSTables. Furthermore, if
you have large partitions, each row will be huge and require a lot of memory in
the mapper.

### Won't the output from `ffs` be huge?

Yes. :-) That's why you want to compress its output. Or, even better, convert
it to compressed [Avro](https://avro.apache.org/),
[ORC](https://orc.apache.org/) or [Parquet](https://parquet.apache.org/).

### What is the format of these JSON objects?

Have a look at these two links (which also describes tombstones):

 * http://thelastpickle.com/blog/2016/07/27/about-deletes-and-tombstones.html
 * http://www.beyondthelines.net/databases/cassandra-tombstones/
