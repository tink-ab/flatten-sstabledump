#!/usr/bin/env bats

setup() {
    go build ./fss.go
}

@test "parse basic production sstable with no deletion" {
    # Must `sort` here since output can come in random order.
    diff <(cat testdata/testdata1.json| ./fss | sort) testdata/testdata1.stdout
}

@test "parse sstable (from blog example) with cell deletion" {
    # Must `sort` here since output can come in random order.
    diff <(cat testdata/testdata2.json| ./fss | sort) testdata/testdata2.stdout
}

@test "parse basic sstable (from blog example) with no deletion" {
    # Must `sort` here since output can come in random order.
    diff <(cat testdata/testdata3.json| ./fss | sort) testdata/testdata3.stdout
}

@test "parse sstable #1 (from blog example) with no deletion" {
    # Must `sort` here since output can come in random order.
    diff <(cat testdata/testdata4.json| ./fss | sort) testdata/testdata4.stdout
}

@test "parse sstable #2 (from blog example) with cell deletion" {
    # Must `sort` here since output can come in random order.
    diff <(cat testdata/testdata5.json| ./fss | sort) testdata/testdata5.stdout
}

@test "parse sstable setting a map" {
    # Must `sort` here since output can come in random order.
    diff <(cat testdata/testdata6.json| ./fss | sort) testdata/testdata6.stdout
}

@test "parse sstable setting an array" {
    # Must `sort` here since output can come in random order.
    diff <(cat testdata/testdata7.json| ./fss | sort) testdata/testdata7.stdout
}

@test "parse sstable #3 (from blog example) with cell deletion" {
    # Must `sort` here since output can come in random order.
    diff <(cat testdata/testdata8.json| ./fss | sort) testdata/testdata8.stdout
}

@test "parse sstable with clustering key deletion" {
    # Must `sort` here since output can come in random order.
    diff <(cat testdata/testdata9.json| ./fss | sort) testdata/testdata9.stdout
}

@test "parse sstable with range tombstones" {
    # Must `sort` here since output can come in random order.
    diff <(cat testdata/testdata10.json| ./fss | sort) testdata/testdata10.stdout
}

@test "parse sstable with partition key tombstone" {
    # Must `sort` here since output can come in random order.
    diff <(cat testdata/testdata11.json| ./fss | sort) testdata/testdata11.stdout
}
