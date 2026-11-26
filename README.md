# Key-Value Storage Engine (LSM-Tree)

A persistent key-value storage engine written from scratch in **Go**, based on the **LSM-tree** architecture used by databases such as RocksDB and Cassandra.

## Architecture

```
write ──► WAL ──► Memtable ──(flush)──► SSTables ──(compaction)──► merged SSTables
read  ──► Memtable ──► Cache (LRU) ──► Bloom filter ──► SSTable index ──► data
```

- **Write-ahead log (WAL):** segmented log for durability and crash recovery
- **Memtable:** configurable in-memory structure (skip list, B-tree or hash map)
- **SSTables:** immutable sorted files with an index, summary and compression
- **Bloom filters:** skip SSTables that cannot contain a key
- **Merkle trees:** data integrity checks
- **Compaction:** merges SSTables across LSM levels
- **Block manager + LRU cache:** block-based disk I/O with caching
- **Rate limiting:** token bucket

## Features

- `PUT`, `GET`, `DELETE`
- **Prefix and range scans** with paginated iterators
- Built-in **probabilistic structures**: Bloom filter, Count-Min Sketch, HyperLogLog and SimHash (text similarity)
- Interactive **CLI** with user login
- External configuration for memtable type, sizes and LSM parameters

## Running

```bash
go run .
```

Type `help` in the CLI to see the available commands.

## Team

- Mihajlo Milojević
- Anđela Bulatović
- Marko Milutin
- Luka Prlinčević
- Dejan Ostojić
