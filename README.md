# rdsnap

rdsnap creates a Amazon RDS snapshot.

## Install

```bash
$ git clone https://github.com/rtiwsk/rdsnap
$ cd rdsnap
$ go build ./cmd/rdsnap/
```

## Usage

```bash
$ ./rdsnap -h
Usage:
  rdsnap [options...]

Options:
  -instance    Amazon RDS DB Instance ID. (Not support Aurora)
               Required option.
  -engine      DB Engine. Select from 'mysql' and 'postgres'.
               Required option.
  -log         Logfile path. Default is Stdout.
  -user        DB Username. Specify when truncating tables.
  -password    DB Password. Specify when truncating tables.
  -tables      DB tables. Specify when truncating tables.
               table as <database>.<tablename>.
               e.g. company.sales,company.account
```

### Example

```bash
$ ./rdsnap \
  -instance exampledb \
  -engine postgres
rdsnap: Start creating a DB snapshot.
rdsnap: Created DB snapshot: exampledb-20220406-211651
```

```bash
$ ./rdsnap \
  -instance exampledb \
  -engine postgres \
  -user dbadmin \
  -password 'password' \
  -tables example.usertbl,example.grouptbl
rdsnap: Start creating a DB snapshot.
rdsnap: Created DB snapshot: exampledb-20220406-213921
rdsnap: Restored DB instance: exampledb-snapshot
rdsnap: Truncated the table: example.usertbl
rdsnap: Truncated the table: example.grouptbl
rdsnap: Created DB snapshot: exampledb-snapshot-20220406-214927
rdsnap: Delete DB instance: exampledb-snapshot
rdsnap: Delete DB snapshot: exampledb-20220406-213921
```
