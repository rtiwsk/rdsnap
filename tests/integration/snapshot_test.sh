#!/bin/bash

set -e

cd `dirname $0`

# check the command path
which terraform 
which psql 
which aws

# go build
mkdir -p bin
go build -o ./bin/rdsnap ../../cmd/rdsnap/

# setup Amazon RDS
cd terraform
terraform init
terraform apply -auto-approve

# setup testdata
export PGPASSWORD='POiu123!!x'
psql -h $(terraform output -raw rds_hostname) \
     -p $(terraform output -raw rds_port) \
	   -U $(terraform output -raw rds_username) \
	   -d example \
	   -f ../testdata.sql

cd ../

# snapshot option error test
./bin/rdsnap \
  -instance exampledb 2>> rdsnap.log || :
grep "\-instance and -engine is required." rdsnap.log

./bin/rdsnap \
  -instance exampledb \
  -engine sqlserver 2>> rdsnap.log || :
grep "Select the correct engine." rdsnap.log

./bin/rdsnap \
  -instance exampledb \
  -engine postgres \
  -user dbadmin 2>> rdsnap.log || :
grep "Insufficient options to truncate DB tables." rdsnap.log

# snapshot test
./bin/rdsnap \
  -instance exampledb \
  -engine postgres \
  -log rdsnap.log
grep "Created DB snapshot: exampledb-" rdsnap.log

# snapshot test with truncated tables
./bin/rdsnap \
  -instance exampledb \
  -engine postgres \
  -user dbadmin \
  -password 'POiu123!!x' \
  -tables example.usertbl,example.grouptbl \
  -log rdsnap.log
grep "Created DB snapshot: exampledb-snapshot-" rdsnap.log

sleep 300

# snapshot error test with truncated tables
./bin/rdsnap \
  -instance exampledb \
  -engine postgres \
  -user dbadmin \
  -password 'POiu123!!x' \
  -tables test.usertbl,test.grouptbl \
  -log rdsnap.log || :
[ $(grep "rdsnap: error:" rdsnap.log | wc -l | awk '{print $1}') -eq 2 ]
[ $(grep "exampledb-snapshot-" rdsnap.log | wc -l | awk '{print $1}') -eq 1 ]

sleep 300

# cleanup
cd terraform
terraform destroy -auto-approve

for instance in exampledb exampledb-snapshot
do 
  snapshotid=$(aws rds describe-db-snapshots \
    --filters Name="db-instance-id",Values=$instance \
    --query "*[].{DBSnapshots:DBSnapshotIdentifier}" \
    --output text)

  aws rds delete-db-snapshot \
      --db-snapshot-identifier $snapshotid \
      --no-cli-pager
done

# print log
cd ../
echo "############################ rdsnap.log ############################"
cat rdsnap.log
echo "############################ rdsnap.log ############################"
rm rdsnap.log
