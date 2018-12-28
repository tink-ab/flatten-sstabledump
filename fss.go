package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	app   = kingpin.New("fss", "Converts Cassandra's `sstabledump` JSON arrays to a list of smaller objects. Useful to be able to run batch jobs (AWS Athena/Hive etc.) over Cassandra dumps (without requiring mappers with lots of memory).").Version("0.0.1").Author("Jens Rantil")
	files = app.Arg("file", "An `sstabledump` JSON file. Can be `-` to read from stdin. Defaults to stdin if no file is given.").Strings()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if len(*files) == 0 {
		// Default to stdin.
		*files = []string{"-"}
	}

	if err := processFiles(*files); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func processFiles(files []string) error {
	buffers := NewBufferPool()

	out := make(chan *bytes.Buffer, 100)
	var writer sync.WaitGroup
	writer.Add(1)
	go func() {
		write(out, buffers)
		writer.Done()
	}()
	defer writer.Wait()
	defer close(out)

	c := make(chan map[string]interface{})
	var encoders sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		encoders.Add(1)
		go func() {
			encode(c, out, buffers)
			encoders.Done()
		}()
	}
	defer encoders.Wait()
	defer close(c)

	for _, filename := range files {
		var err error
		if filename == "-" {
			err = process(os.Stdin, c)
		} else {
			err = processFile(filename, c)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

type BufferPool struct {
	pool sync.Pool
}

func NewBufferPool() *BufferPool {
	return &BufferPool{sync.Pool{New: func() interface{} { return bytes.NewBuffer(nil) }}}
}

func (bp *BufferPool) Put(b *bytes.Buffer) {
	bp.pool.Put(b)
}

func (bp *BufferPool) Get() *bytes.Buffer {
	return bp.pool.Get().(*bytes.Buffer)
}

func write(in <-chan *bytes.Buffer, pool *BufferPool) {
	out := bufio.NewWriter(os.Stdout)
	for e := range in {
		if _, err := io.Copy(out, e); err != nil {
			log.Fatalln("could not write to stdout:", err)
		}
		out.WriteRune('\n')
	}
	out.Flush()
}

func encode(in chan map[string]interface{}, out chan *bytes.Buffer, pool *BufferPool) {
	for e := range in {
		fluffybuf := pool.Get()
		enc := json.NewEncoder(fluffybuf)
		if err := enc.Encode(e); err != nil {
			log.Fatalln("could not compact JSON:", err)
		}

		compbuf := pool.Get()
		// TODO/PERF: There are likely other JSON encoders out there that can
		// encode immediately to compact format.
		if err := json.Compact(compbuf, fluffybuf.Bytes()); err != nil {
			log.Fatalln("could not compact JSON:", err)
		}
		out <- compbuf

		fluffybuf.Reset()
		pool.Put(fluffybuf)
	}
}

type outChan chan<- map[string]interface{}

func processFile(filename string, out outChan) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return process(f, out)
}

const megabyte = 1024 * 1024

func process(f io.Reader, out outChan) error {
	i := jsoniter.Parse(jsoniter.ConfigFastest, bufio.NewReader(f), 10*megabyte)
	if i.WhatIsNext() != jsoniter.ArrayValue {
		return errors.New("expected first element to be an array")
	}

	for i.ReadArray() {
		if i.WhatIsNext() != jsoniter.ObjectValue {
			return errors.New("expected all array values to be objects")
		}

		// For each field in the partition.
		partition := make(map[string]interface{})
	partFields:
		for {
			fieldname := i.ReadObject()
			if i.Error != nil {
				return errors.Wrap(i.Error, "unable to read field name")
			}
			switch fieldname {
			case "":
				fallthrough
			case "rows":
				break partFields
			default:
				partition[fieldname] = i.Read()
				if i.Error != nil {
					return errors.Wrapf(i.Error, "could not read fieldname `%s`", fieldname)
				}
			}
		}

		if i.WhatIsNext() != jsoniter.ArrayValue {
			return errors.New("expected `rows` to be an array")
		}
		for i.ReadArray() {
			if i.WhatIsNext() != jsoniter.ObjectValue {
				return errors.New("expected every row to be an object")
			}
			row := i.Read().(map[string]interface{})
			if i.Error != nil {
				return errors.Wrap(i.Error, "")
			}
			if _, exist := row["partition"]; exist {
				return errors.New("did not expect any row to contain a `partition` key")
			}
			row["partition"] = partition
			out <- row
		}
		if i.Error != nil {
			return errors.Wrap(i.Error, "unable to read the row array")
		}

		// TODO: Assert object ends.
		if fieldname := i.ReadObject(); fieldname != "" {
			return errors.Errorf("expected `rows` field to be the last. Last field was:", fieldname)
		}
	}

	return errors.Wrap(i.Error, "end of array not found")
}
